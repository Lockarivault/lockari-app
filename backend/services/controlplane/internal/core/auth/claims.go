package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// InternalClaims representa o payload do nosso JWT interno soberano.
// É usado para comunicação entre microserviços e auditoria.
type InternalClaims struct {
	jwt.RegisteredClaims
	UserID    string   `json:"uid"` // UUID interno do usuário no nosso banco
	TenantID  string   `json:"tid"` // ID do Tenant ativo na sessão
	ActorType string   `json:"act"` // Tipo de ator: "human", "app" ou "system"
	Roles     []string `json:"rol"` // Lista de permissões ou papéis associados
}
