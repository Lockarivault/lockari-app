# Estratégia de Atualização de Tenant

## Objetivo
Implementar um fluxo robusto para a edição de dados de um tenant, garantindo integridade, segurança (imutabilidade de chaves) e flexibilidade (propriedades dinâmicas).

## Cenário Atual
O microsserviço já possui um esqueleto para atualização (`Update` no Handler, Usecase e Repository), mas carece de validações de negócio refinadas, tratamento adequado de propriedades e bloqueio de campos sensíveis.

## Regras de Negócio e Validações

### 1. Campos Editáveis
Apenas os seguintes campos devem ser permitidos na operação de Update:
- **Name**: Nome de exibição ou referência interna.
- **Description**: Texto descritivo.
- **OwnerID**: Identificador do proprietário (pode haver transferência de propriedade).
- **Properties**: Mapa de propriedades dinâmicas (configurações customizadas).

### 2. Campos Imutáveis (Segurança e Integridade)
Os seguintes campos **NÃO** podem ser alterados via rota de edição padrão:
- **ID**: Identificador único (imutável por definição).
- **Slug**: Identidade pública na URL. Alterar slug quebra links e configurações de DNS.
- **Status**: Gerenciado apenas pela esteira de provisionamento (Active, Suspended, etc).
- **SecurityMetadata**: **CRÍTICO**. Contém chaves de criptografia (KEK). Jamais deve ser exposto ou alterado via API de usuário.
- **CreatedAt / DeletedAt**: Auditoria do sistema.

### 3. Validações Necessárias
Antes de persistir, o sistema deve validar:
- **Existência**: O tenant deve existir (já implementado).
- **Status**: Tenants marcados como `DELETED` não devem ser editáveis.
- **Propriedades**: Se o campo `Properties` for enviado, ele deve respeitar o contrato de tipos (ex: se houver validação de schema JSON no futuro). Por enquanto, garantir que é um mapa válido.

## Plano de Implementação

### 1. Camada de Modelo (`internal/core/tenant/model`)
- [ ] Criar método `UpdateValidate()` na struct `TenantModel` ou em um DTO específico de domínio.
    - Este método deve mesclar os dados novos sobre os antigos, garantindo que campos nulos na request não apaguem dados existentes (patch vs put) ou definir explicitamente que a operação é um "Replace" parcial.
    - **Recomendação**: Utilizar a abordagem de "Dirty Checking" ou DTO de Update explícito na camada de domínio para blindar os campos imutáveis.

### 2. Camada de Usecase (`internal/core/tenant/usecase/manage.go`)
- [ ] Atualizar o método `Update`.
    - Recuperar o tenant atual do repositório (já feito).
    - Aplicar as alterações permitidas (Name, Description, OwnerID).
    - **Properties**: Implementar lógica de merge ou substituição das propriedades.
    - **Segurança**: Garantir explicitamente que `SecurityMetadata` do objeto a ser salvo seja idêntico ao do objeto recuperado do banco, ignorando qualquer dado vindo do handler.
    - Atualizar `UpdatedAt`.

### 3. Camada de Handler (`internal/core/tenant/handler/tenant_handler.go`)
- [ ] Revisar `UpdateTenantRequest` dto.
    - Garantir que não existam campos como `SecurityMetadata` no JSON de entrada.
- [ ] Ajustar o mapeamento.
    - Mapear `Properties` do DTO para o Model.

## Lista de Tarefas (Task List)

### Backend

- [ ] **Model Enhancement**
    - [ ] Adicionar método `Update(newData TenantModel)` em `TenantModel` para centralizar a lógica de quais campos podem ser copiados.

- [ ] **Usecase Update Logic**
    - [ ] Alterar `manageTenant.Update` para usar o método de merge do model.
    - [ ] Garantir preservação de `SecurityMetadata` e `Slug`.
    - [ ] Tratar atualização de `Properties` (Maps no Go precisam de cuidado para não serem nil).

- [ ] **DTO Adjustment**
    - [ ] Verificar se `UpdateTenantRequest` suporta `Properties` (map[string]interface{}).
