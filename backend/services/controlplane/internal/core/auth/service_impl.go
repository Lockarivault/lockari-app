package auth

import (
	"context"
	"fmt"
)

// authService implementa a interface AuthService articulando a validação externa e emissão interna.
type authService struct {
	jwtManager      *JWTManager
	googleValidator *GoogleValidator
}

// NewAuthService constrói a implementação concreta do serviço de autenticação.
func NewAuthService(jwtManager *JWTManager, googleValidator *GoogleValidator) AuthService {
	return &authService{
		jwtManager:      jwtManager,
		googleValidator: googleValidator,
	}
}

// Exchange realiza a troca do token do Google pelo nosso JWT soberano.
func (s *authService) Exchange(ctx context.Context, externalToken string) (*TokenResponse, error) {
	// 1. Valida o token do Google Identity Platform
	googleClaims, err := s.googleValidator.Validate(ctx, externalToken)
	if err != nil {
		return nil, fmt.Errorf("auth: external validation failed: %w", err)
	}

	// 2. TODO: Lookup/Provisioning (Just-In-Time)
	// Nesta etapa (Passo 1), focamos na infraestrutura de tokens.
	// No futuro, usaremos o 'sub' ou 'email' de googleClaims para buscar o UserID real no banco.
	sub, _ := (*googleClaims)["sub"].(string)

	internalClaims := &InternalClaims{
		UserID:    sub,                     // Placeholder: deve vir do banco
		TenantID:  "system-default-tenant", // Placeholder: deve vir da relação user-tenant
		ActorType: "human",
		Roles:     []string{"user"},
	}

	// 3. Gera o nosso JWT interno assinado com RS256
	token, err := s.jwtManager.Generate(internalClaims)
	if err != nil {
		return nil, fmt.Errorf("auth: failed to generate internal token: %w", err)
	}

	return &TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(24 * 3600), // 24 horas
	}, nil
}

// Parse valida e decodifica um token interno.
func (s *authService) Parse(ctx context.Context, tokenString string) (*InternalClaims, error) {
	return s.jwtManager.Validate(ctx, tokenString)
}
