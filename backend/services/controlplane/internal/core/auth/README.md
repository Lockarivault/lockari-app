# Lockarivault - Authentication Core (OAuth2 + JWT)


Este documento descreve detalhadamente o funcionamento do núcleo de autenticação do **Control Plane**. Aqui, explicamos como transformamos uma identidade externa (**Google Identity Platform**) em uma identidade interna soberana e segura para o nosso ecossistema SaaS.

---

## 1. O Conceito: Token Exchange (BFF Pattern)

Em vez de permitir que o token do provedor externo circule livremente por todos os nossos microsserviços do **Data Plane**, adotamos o padrão de **Troca de Token**.

### Por que fazemos isso?
1.  **Independência**: Se trocarmos o provedor de identidade (ex: de Google para Auth0 ou Keycloak), os serviços internos não precisam mudar nada.
2.  **Enriquecimento**: O provedor externo não sabe qual é o `InternalUserID` ou o `TenantID` do nosso banco. O backend injeta esses dados no novo token.
3.  **Performance**: Validar um token externo exige chamadas de rede (JWKS). Validar um token interno é ultra-rápido (apenas matemática de chave pública).

---

## 2. O Fluxo de Autenticação (Passo a Passo)

1.  **Frontend (Login)**: O usuário faz login no **Google Identity Platform** via OIDC/OAuth2.
2.  **Google Token**: O Google devolve um `ID_TOKEN` ou `ACCESS_TOKEN` assinado por eles para o Frontend.
3.  **Handoff**: O Frontend envia esse token para o nosso backend via Header `Authorization: Bearer <google_token>`.
4.  **Validação Externa**: O Backend (este módulo) verifica se o token é válido, se não expirou e se a assinatura é legítima do Google Identity Platform.
5.  **Provisionamento JIT (Just-In-Time)**:
    *   O backend lê o `sub` (Subject) ou `email` do Google.
    *   Busca no nosso banco de dados por esse usuário.
    *   Se for um novo usuário, criamos o registro no banco e associamos ao Tenant.
6.  **Geração do JWT Interno**: O backend gera um novo JWT usando a **nossa Chave Privada (RS256)** contendo as nossas Claims internas.
7.  **Sessão**: O Backend devolve esse novo JWT para o Frontend (geralmente salvo em Cookie HttpOnly).

---

## 3. Estrutura de Arquivos e Propósitos

O pacote está localizado em `backend/services/controlplane/internal/core/auth/`.

| Arquivo | Propósito |
| :--- | :--- |
| `models.go` | Contém as structs de dados (Claims, Request/Response payloads). |
| `service.go` | Define a interface `AuthService` (O contrato que os outros usam). |
| `jwt_manager.go` | Lógica de baixo nível para assinar e validar tokens usando RS256 (RSA). |
| `google_auth_validator.go` | Implementação específica para validar o token externo do Google Identity Platform. |
| `module.go` | Configuração para injeção de dependência via **Uber fx**. |

---

## 4. Contratos e Estruturas de Dados

### Claims Internas (`InternalClaims`)
Este é o coração da nossa segurança. É o que o `AuditLog` lê para saber quem fez o quê.

```go
type InternalClaims struct {
    jwt.RegisteredClaims
    UserID    string   `json:"uid"` // UUID interno do banco
    TenantID  string   `json:"tid"` // ID do Tenant no qual o usuário está operando
    ActorType string   `json:"act"` // "human", "app" ou "system"
    Roles     []string `json:"rol"` // Lista de permissões (ex: "admin", "viewer")
}
```

### Interface do Serviço (`AuthService`)
Os desenvolvedores devem injetar esta interface, nunca a implementação concreta.

```go
type AuthService interface {
    // Exchange: Troca o token do Google pelo nosso JWT interno
    Exchange(ctx context.Context, externalToken string) (string, error)
    
    // Parse: Decodifica e valida o JWT interno que vem nas requests
    Parse(tokenString string) (*InternalClaims, error)
}
```

---

## 5. Detalhes Técnicos de Implementação

### Algoritmo RS256 (RSA)
Utilizamos chaves assimétricas. 
*   **Chave Privada (Private Key)**: Fica apenas no Control Plane. É usada para **Assinar** o token.
*   **Chave Pública (Public Key)**: Pode ser distribuída. É usada para **Verificar** a assinatura.
*   *Benefício:* O Data Plane pode validar o token sem nunca precisar da Chave Privada.

### Cache de JWKS
O Google publica chaves públicas via URL (JSON Web Key Set). Para evitar latência, o `google_auth_validator.go` deve implementar um cache em memória das chaves do Google, atualizando-as apenas periodicamente.

---

## 6. Guia para o Desenvolvedor Junior/Pleno

1.  **Nunca ignore o Contexto**: Sempre passe o `ctx` para o `Exchange`, pois ele contém logs de rastreabilidade (TraceID).
2.  **Fail Fast**: Se o token externo falhar, retorne `401 Unauthorized` imediatamente. Não processe lógica de banco antes de validar o token.
3.  **Logs Estruturados**: Use `slog.Error` ou `slog.Warn` se a assinatura falhar, mas **nunca logue o Token completo** no console por questões de segurança.
4.  **Uber fx**: Certifique-se de que o `AuthService` está registrado no `module.go` para que o seu Usecase ou Middleware possa recebê-lo via construtor.

---

## 7. Passo 2: O Middleware de Autenticação

O Middleware é o "segurança" da nossa API. Ele intercepta todas as requisições protegidas e garante que apenas usuários com um JWT interno válido possam prosseguir.

### Responsabilidades do Middleware
1.  **Extração**: Ler o header `Authorization: Bearer <token>`.
2.  **Validação**: Chamar o método `authService.Parse(ctx, tokenString)` que criamos.
3.  **Metadados**: Extrair o IP do cliente e o User-Agent da requisição.
4.  **Injeção de Contexto**: Pegar os dados das `InternalClaims` (UserID, TenantID, ActorType) + Metadados e injetar no `context.Context`.
5.  **Bloqueio**: Se o token for inválido ou ausente, retornar `401 Unauthorized`.

### Fluxo de Implementação (Pseudo-código para o Desenvolvedor)

O desenvolvedor deve criar o middleware seguindo este padrão:

```go
func AuthMiddleware(svc auth.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Pega o token do Header
        token := extractToken(c.Request)
        if token == "" {
            c.AbortWithStatusJSON(401, errorResponse("Token is required"))
            return
        }

        // 2. Valida o token usando o nosso Core Auth
        claims, err := svc.Parse(c.Request.Context(), token)
        if err != nil {
            c.AbortWithStatusJSON(401, errorResponse("Invalid or expired token"))
            return
        }

        // 3. Injeta no Contexto usando nossos Helpers (tenanttools)
        ctx := c.Request.Context()
        ctx = tenanttools.SetUserIDFromContext(ctx, claims.UserID)
        ctx = tenanttools.SetTenantIDFromContext(ctx, claims.TenantID)
        ctx = tenanttools.SetActorTypeFromContext(ctx, claims.ActorType)
        
        // Novos campos solicitados:
        ctx = tenanttools.SetIPAddressFromContext(ctx, c.ClientIP())
        ctx = tenanttools.SetUserAgentFromContext(ctx, c.Request.UserAgent())

        // 4. Atualiza a request com o novo contexto populado
        c.Request = c.Request.WithContext(ctx)

        c.Next()
    }
}
```

### Por que usar os Helpers (`tenanttools`) no Middleware?
O nosso Usecase (`CreateTenant`) já está configurado para ler esses valores do contexto. Ao injetá-los no Middleware, garantimos que toda a rastreabilidade e auditoria funcionem automaticamente sem que o desenvolvedor do Handler precise fazer nada.

---
> [!IMPORTANT]
> A segurança de todo o sistema Lockarivault depende da integridade da nossa **Chave Privada**. Em produção, ela deve ser carregada via Variável de Ambiente ou Vault, nunca via Hardcoded no código.
