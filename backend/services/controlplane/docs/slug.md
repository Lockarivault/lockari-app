# Estudo de Validação de Slug do Tenant

## Objetivo
Desenvolver um mecanismo eficiente para validar a disponibilidade de um "slug" para novos tenants, garantindo unicidade e performance dentro do microsserviço `controlplane`.

## Análise do Cenário Atual

O sistema utiliza **MongoDB** como persistência principal para os tenants via `TenantMongoRepository`. O modelo de dados (`TenantModel`) já contempla o campo `Slug`.

### Questão: Iterar ou Consultar?
**Não devemos iterar sobre todos os tenants.**
O MongoDB (e qualquer banco de dados moderno) é otimizado para buscas diretas através de índices. A validação deve ser feita através de uma **query direta** ao banco de dados buscando pelo slug específico.

A iteração ("scan") de todos os registros seria o pior cenário possível (O(n)), causando lentidão exponencial conforme a base cresce. A consulta indexada é praticamente instantânea (O(1) ou O(log n)).

## Solução Proposta

### 1. Camada de Dados (Repository)
O repositório `tenantMongoRepository` já possui o método:
```go
GetBySlug(ctx context.Context, slug string) (tenantmodel.TenantModel, error)
```
Isso é suficiente para a validação.
**Recomendação Crítica:** Deve-se garantir que existe um **Índice Único (Unique Index)** no campo `slug` da collection `tenants` no MongoDB. Isso garante que, mesmo em condições de corrida (duas requisições simultâneas), o banco de dados impedirá a duplicidade.

### 2. Camada de Negócio (Usecase)
Devemos criar um método específico no Usecase para verificar a disponibilidade.
A struct de retorno já existe em `model/slug_available.go`: `SlugAvailable`.

**Lógica sugerida:**
1. Recebe o slug desejado.
2. Normaliza o slug (lowercase, trim).
3. Chama `repo.GetBySlug`.
4. Se retornar erro `NotFound`, o slug está **Disponível**.
5. Se retornar um tenant (sucesso), o slug está **Indisponível**.
6. (Opcional/Futuro) Se indisponível, gerar sugestões (ex: `minha-empresa-1`, `minha-empresa-br`).

### 3. Camada de Interface (Handler)
Criar um novo endpoint dedicado para essa verificação. Isso permite que o frontend valide o slug em tempo real enquanto o usuário digita, sem submeter o formulário completo de criação.

**Path Sugerido:**
`GET /api/v1/tenants/check-slug?slug={valor}`

## Lista de Tarefas (Task List)

### Backend

- [x] **Database Index (Infra/Migration)**
    - [x] Verificar e criar índice único no MongoDB para o campo `slug`.
    - `db.tenants.createIndex({ "slug": 1 }, { unique: true })`

- [ ] **Usecase Implementation (`internal/core/tenant/usecase`)**
    - [ ] Adicionar método `CheckSlugAvailability(ctx, slug)` na interface `UsecaseTenant`. ```backend/services/controlplane/internal/core/tenant/usecase/check_slug.go````
    - [ ] Implementar o método (provavelmente em um novo arquivo `check.go` ou em `get.go`).
    - [ ] A lógica deve mapear o resultado do repository para a struct `SlugAvailable`.

- [ ] **Handler Implementation (`internal/core/tenant/handler`)**
    - [ ] Adicionar método `CheckSlug(c *gin.Context)` no handler.
    - [ ] Ler o query parameter `slug`.
    - [ ] Validar formato básico do slug (regex).
    - [ ] Chamar o usecase.
    - [ ] Retornar JSON com `SlugAvailable`.

- [ ] **Router Registration**
    - [ ] Registrar a rota `GET /api/v1/tenants/check-slug` no grupo de rotas de tenant.

### Frontend (Referência)
- [ ] Integrar no formulário de criação de tenant para consultar essa rota no evento `onBlur` ou com `debounce` no input de slug.

## Recomendações Adicionais

1. **Normalização**: O slug deve ser sempre salvo e buscado em minúsculas (lowercase).
2. **Concorrência**: Confie no índice único do banco (“Unique Constraint”) como a "source of truth" final para evitar duplicatas no momento da criação (`CreateTenant`). A rota de validação é apenas para UX (User Experience).
