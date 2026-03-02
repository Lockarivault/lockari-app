package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager lida com a lógica de baixo nível para emissão e validação de JWTs RS256.
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	expiration time.Duration
}

// NewJWTManager cria uma nova instância de gerenciador de JWT.
func NewJWTManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, issuer string, expiration time.Duration) *JWTManager {
	if expiration <= 0 {
		expiration = 24 * time.Hour // Default: 1 dia
	}
	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		expiration: expiration,
	}
}

// Generate assina um novo InternalClaims e retorna o token formatado.
func (m *JWTManager) Generate(claims *InternalClaims) (string, error) {
	if m.privateKey == nil {
		return "", errors.New("jwt: private key is missing for signing")
	}

	// Enriquecimento das claims registradas
	now := time.Now().UTC()
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Subject:   claims.UserID,
		ExpiresAt: jwt.NewNumericDate(now.Add(m.expiration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// Validate valida a assinatura e as claims de um token interno.
func (m *JWTManager) Validate(ctx context.Context, tokenString string) (*InternalClaims, error) {
	if m.publicKey == nil {
		return nil, errors.New("jwt: public key is missing for validation")
	}

	token, err := jwt.ParseWithClaims(tokenString, &InternalClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Valida se o método de assinatura é RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("jwt: unexpected signing method: %v", token.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("jwt: validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*InternalClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("jwt: invalid claims")
}
