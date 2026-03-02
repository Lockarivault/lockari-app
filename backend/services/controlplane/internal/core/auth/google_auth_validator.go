package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	googleJWKSURL = "https://www.googleapis.com/oauth2/v3/certs"
)

// GoogleValidator lida com a validação de tokens OIDC vindos do Google Identity Platform.
type GoogleValidator struct {
	audience  string
	issuer    string
	keys      map[string]any
	mu        sync.RWMutex
	lastFetch time.Time
}

// NewGoogleValidator cria uma nova instância de validador Google.
func NewGoogleValidator(audience string) *GoogleValidator {
	return &GoogleValidator{
		audience: audience,
		issuer:   "https://accounts.google.com", // Emissor padrão do Google
		keys:     make(map[string]any),
	}
}

// Validate verifica a assinatura, emissor e audiência de um Google ID Token.
func (v *GoogleValidator) Validate(ctx context.Context, tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, v.getKey)
	if err != nil {
		return nil, fmt.Errorf("google-auth: failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("google-auth: invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("google-auth: failed to extract claims")
	}

	// Validação de Audience (aud)
	if aud, ok := claims["aud"].(string); !ok || aud != v.audience {
		return nil, errors.New("google-auth: audience mismatch")
	}

	// Validação de Issuer (iss)
	if iss, ok := claims["iss"].(string); !ok || (iss != v.issuer && iss != "accounts.google.com") {
		return nil, errors.New("google-auth: issuer mismatch")
	}

	return &claims, nil
}

// getKey busca a chave pública correspondente ao 'kid' do token, com cache.
func (v *GoogleValidator) getKey(token *jwt.Token) (interface{}, error) {
	// 1. Verifica se o algoritmo é RSA
	if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
		return nil, fmt.Errorf("google-auth: unexpected signing method: %v", token.Header["alg"])
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("google-auth: missing 'kid' header")
	}

	// 2. Tenta buscar do cache
	v.mu.RLock()
	key, exists := v.keys[kid]
	v.mu.RUnlock()

	if exists {
		return key, nil
	}

	// 3. Se não existe ou expirou (cache TTL simples de 1 hora), busca do Google
	if err := v.refreshKeys(); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	if key, exists := v.keys[kid]; exists {
		return key, nil
	}

	return nil, fmt.Errorf("google-auth: key not found for kid: %s", kid)
}

func (v *GoogleValidator) refreshKeys() error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Double-check para evitar múltiplas chamadas simultâneas
	if time.Since(v.lastFetch) < 5*time.Minute {
		return nil
	}

	resp, err := http.Get(googleJWKSURL)
	if err != nil {
		return fmt.Errorf("google-auth: failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("google-auth: failed to decode JWKS: %w", err)
	}

	for _, key := range jwks.Keys {
		// 1. Decode Modulus (N)
		nb, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}

		// 2. Decode Exponent (E)
		eb, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}

		// 3. Convert E to int
		var e int
		for _, b := range eb {
			e = e<<8 | int(b)
		}

		// 4. Create Public Key
		pub := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nb),
			E: e,
		}

		v.keys[key.Kid] = pub
	}

	v.lastFetch = time.Now()
	return nil
}
