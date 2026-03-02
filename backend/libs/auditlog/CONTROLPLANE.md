# Guia de Implementação: Control Plane

Este documento detalha como a biblioteca de auditoria deve ser integrada no **Control Plane** do Lockarivault.

---

## 🛠️ Inicialização via Providers

No diretório `backend/services/controlplane/cmd/api/providers`, seguimos o padrão de criar uma função de provedor que se integra ao ciclo de vida (Lifecycle) do Uber fx.

### Passo 1: Criar o arquivo `audit.go`
Crie o arquivo `backend/services/controlplane/cmd/api/providers/audit.go` com o seguinte conteúdo:

```go
package providers

import (
    "context"
    "github.com/lockarivault/lockari-app/backend/libs/auditlog"
    "github.com/lockarivault/lockari-app/backend/libs/database/nosql"
    "go.uber.org/fx"
)

// ProvideAudit inicializa a lib de auditoria com workers e canal interno.
func ProvideAudit(lc fx.Lifecycle, db nosql.DatabaseService) auditlog.Service {
    // 1. Criamos a instância da lib
    // Passamos o repositório de MongoDB (que deve implementar a interface auditlog.Store)
    service := auditlog.New(auditlog.Config{
        Store:   db,   // O serviço de NoSQL injetado
        Workers: 5,    // 5 goroutines processando em paralelo
        Buffer:  1000, // Fila de espera de até 1000 logs
    })

    // 2. Integramos ao Lifecycle do Go
    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            // Aqui iniciamos o Worker Pool
            return service.Start(ctx)
        },
        OnStop: func(ctx context.Context) error {
            // Shutdown Gracioso: espera processar o que sobrou no canal
            return service.Stop(ctx)
        },
    })

    return service
}
```

### Passo 2: Registrar no `module.go`
Após criar a função, você deve registrá-la no arquivo `backend/services/controlplane/cmd/api/providers/module.go` para que ela fique disponível para todo o projeto:

```go
// No arquivo providers/module.go
var Module = fx.Module("api-providers",
    fx.Provide(
        // ... outros providers
        ProvideNoSQL,
        ProvideAudit, // <-- Adicione aqui
    ),
)
```

---

## 🎯 Uso nos Usecases

Os desenvolvedores do Control Plane não precisam se preocupar com a infraestrutura de workers. Eles apenas injetam a interface e chamam o método.

### Exemplo: Tenant Creation
Ao criar um novo Tenant, o usecase deve registrar essa ação.

```go
type CreateTenantUsecase struct {
    repo  tenant.Repository
    audit audit.Service
}

func (u *CreateTenantUsecase) Execute(ctx context.Context, input CreateTenantInput) error {
    // 1. Regra de Negócio
    tenant, err := u.repo.Create(ctx, input)
    if err != nil {
        return err
    }

    // 2. Auditoria (Sem impacto na performance da regra acima)
    _ = u.audit.CreateAuditLog(ctx, audit.AuditEntry{
        TenantID:     tenant.ID,
        UserID:       ctx.Value("user_id").(string),
        ResourceType: "tenant",
        ResourceID:   tenant.ID,
        Action:       "create",
        Metadata: map[string]any{
            "name": tenant.Name,
            "plan": tenant.Plan,
        },
    })

    return nil
}
```

---

## 🌐 Contexto e Metadados

No Control Plane, muitas informações (IP, User Agent) já estão disponíveis nos middlewares de API.

- **Dica para Juniores:** Certifiquem-se de que o `context.Context` passado para o usecase contenha essas informações, ou passe-as explicitamente no campo `Metadata` da `AuditEntry`.

---

## 💾 Armazenamento (Storage)

No Control Plane, os logs serão salvos em uma coleção dedicada no **MongoDB** chamada `audit_logs`.
- **Importante:** A lib deve garantir que essa coleção tenha índices por `tenant_id` e `timestamp` para que o método `ListAuditLogs` seja rápido quando exibirmos isso no Dashboard do administrador.
