package auth

import (
	"context"
)

// AuthService define o contrato público para gerenciamento de autenticação.
// Segue o princípio de interfaces definidas pelo consumidor ou próximas ao core.
type AuthService interface {
	// Exchange realiza a troca de um token do Google Identity Platform pelo nosso JWT interno.
	// 1. Valida a assinatura e expiração do Google Token.
	// 2. Provisiona ou atualiza o usuário no banco (JIT - Just In Time).
	// 3. Emite e assina o nosso JWT interno (RS256).
	Exchange(ctx context.Context, externalToken string) (*TokenResponse, error)

	// Parse decodifica e valida o nosso JWT interno soberano.
	// Usado principalmente em middlewares e chamadas inter-serviços.
	Parse(ctx context.Context, tokenString string) (*InternalClaims, error)
}
