package providers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/lockarivault/lockari-app/backend/services/controlplane/config"
	"github.com/lockarivault/lockari-app/backend/services/controlplane/internal/core/auth"
)

// ProvideAuthService extrai as chaves RSA da configuração e inicializa o núcleo de autenticação.
func ProvideAuthService(cfg *config.Connections) (auth.AuthService, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	// 1. Carregar Chave Privada (Necessária para assinar o nosso JWT)
	var privateKey *rsa.PrivateKey
	if cfg.Auth.RSAPrivateKey != "" {
		pk, err := parseRSAPrivateKey(cfg.Auth.RSAPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("auth-provider: failed to parse private key: %w", err)
		}
		privateKey = pk
	}

	// 2. Carregar Chave Pública (Necessária para validar o nosso JWT)
	var publicKey *rsa.PublicKey
	if cfg.Auth.RSAPublicKey != "" {
		pk, err := parseRSAPublicKey(cfg.Auth.RSAPublicKey)
		if err != nil {
			return nil, fmt.Errorf("auth-provider: failed to parse public key: %w", err)
		}
		publicKey = pk
	}

	// 3. Configurações de Duração e Issuer
	duration, err := time.ParseDuration(cfg.Auth.JWTDuration)
	if err != nil {
		duration = 24 * time.Hour // Default 1 dia
	}

	issuer := cfg.Auth.JWTIssuer
	if issuer == "" {
		issuer = "lockarivault-controlplane"
	}

	// 4. Inicializar Gerenciadores
	jwtManager := auth.NewJWTManager(privateKey, publicKey, issuer, duration)
	googleValidator := auth.NewGoogleValidator(cfg.Auth.GoogleAudience)

	return auth.NewAuthService(jwtManager, googleValidator), nil
}

// Helpers para Parsing de PEM
func parseRSAPrivateKey(pemData string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Tenta PKCS8 se PKCS1 falhar
		key8, err8 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err8 != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err8)
		}
		return key8.(*rsa.PrivateKey), nil
	}
	return key, nil
}

func parseRSAPublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pub.(*rsa.PublicKey), nil
}
