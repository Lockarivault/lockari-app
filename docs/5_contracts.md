# Definições de Dados e Interfaces para Plataforma de Gestão de Secrets

Este documento detalha as estruturas de dados e contratos de interface para a plataforma de gestão de secrets, com foco rigoroso em isolamento multi-tenant e desempenho.

---

## 1. MongoDB Schema

O MongoDB será utilizado para persistência de dados. Todas as coleções específicas de tenant incluirão o campo `tenant_id` obrigatório e terão índices compostos começando com este campo para otimizar consultas e garantir isolamento lógico.

### Convenções
| Campo | Descrição |
| :--- | :--- |
| **`_id`** | UUID (string) para identificadores de recursos. |
| **`tenant_id`** | UUID (string) obrigatório para isolamento lógico. |
| **`created_at`, `updated_at`** | ISODate para registro de auditoria temporal. |
| **`encryption_metadata`** | Objeto contendo IV, algoritmo e versão da DEK. |

### Estruturas de Documentos

#### Coleção: `tenants`
```json
{
  "_id": "UUID_DO_TENANT",
  "name": "Nome do Tenant",
  "status": "active",
  "plan_id": "UUID_DO_PLANO",
  "kek_id": "UUID_DA_KEK",
  "created_at": "ISODate",
  "updated_at": "ISODate"
}
```
**Índices:**
- `{ "_id": 1 }` (Primário)
- `{ "name": 1 }` (Único)
- `{ "status": 1 }`

#### Coleção: `users`
```json
{
  "_id": "UUID_DO_USUARIO",
  "tenant_id": "UUID_DO_TENANT",
  "email": "usuario@exemplo.com",
  "username": "nome.usuario",
  "password_hash": "$2a$10$hashedpassword...",
  "roles": ["admin", "viewer"],
  "status": "active",
  "created_at": "ISODate"
}
```
**Índices:**
- `{ "tenant_id": 1, "email": 1 }` (Único)
- `{ "tenant_id": 1, "username": 1 }` (Único)

#### Coleção: `vaults`
```json
{
  "_id": "UUID_DO_VAULT",
  "tenant_id": "UUID_DO_TENANT",
  "name": "Vault de Produção",
  "dek_id": "UUID_DA_DEK",
  "created_at": "ISODate"
}
```
**Índices:**
- `{ "tenant_id": 1, "name": 1 }` (Único)

#### Coleção: `secrets`
```json
{
  "_id": "UUID_DO_SECRET",
  "tenant_id": "UUID_DO_TENANT",
  "vault_id": "UUID_DO_VAULT",
  "name": "ChaveAPI_ServicoXYZ",
  "type": "password",
  "encrypted_value": "BinData",
  "encryption_metadata": {
    "iv": "base64",
    "algorithm": "AES256-GCM",
    "dek_version_id": "UUID"
  },
  "tags": ["prod"],
  "expires_at": "ISODate"
}
```
**Índices:**
- `{ "tenant_id": 1, "vault_id": 1, "name": 1 }` (Único)
- `{ "tenant_id": 1, "tags": 1 }`

---

## 2. gRPC Definitions (.proto)

As definições gRPC serão usadas para a comunicação inter-serviços. O `tenant_id` e `user_id` são propagados via metadados por interceptors.

### common.proto
```proto
syntax = "proto3";
package common;

message OperationResult {
   bool success = 1;
   string message = 2;
   string error_code = 3;
}

message Pagination {
   int32 limit = 1;
   int32 offset = 2;
}

message PaginationResult {
   int32 total_count = 1;
   bool has_next_page = 4;
}
```

### secrets.proto
```proto
syntax = "proto3";
package secrets;

import "google/protobuf/timestamp.proto";
import "common.proto";

service SecretService {
   rpc CreateSecret(CreateSecretRequest) returns (SecretResponse);
   rpc GetSecret(GetSecretRequest) returns (SecretResponse);
   rpc AccessSecret(AccessSecretRequest) returns (AccessSecretResponse);
}

message Secret {
   string id = 1;
   string tenant_id = 2;
   string vault_id = 3;
   string name = 4;
   string type = 5;
   google.protobuf.Timestamp created_at = 8;
}

message AccessSecretResponse {
   string value = 1;
   google.protobuf.Timestamp accessed_at = 2;
}
```

---

## 3. GraphQL Types

O GraphQL será exposto pelo API Gateway/BFF. O `tenant_id` é gerenciado internamente via JWT.

### schema.graphql
```graphql
type Query {
   me: User!
   tenant: Tenant!
   vault(id: ID!): Vault
   secrets(vaultId: ID, pagination: PaginationInput): SecretConnection!
}

type Mutation {
   createVault(input: CreateVaultInput!): VaultPayload!
   createSecret(input: CreateSecretInput!): SecretPayload!
   accessSecret(id: ID!, reason: String!): AccessSecretPayload!
}

type Secret {
  id: ID!
  name: String!
  type: String!
  tags: [String]
  createdAt: String!
}
```

---
*Relatório de Contratos de Dados - v1.0 - Fevereiro de 2026*
