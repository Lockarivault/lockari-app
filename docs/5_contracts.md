Como Arquiteto de Dados Especialista, apresento as definições de baixo nível para o esquema MongoDB, definições gRPC, tipos GraphQL e estratégia Redis, garantindo a aderência aos requisitos de multi-tenancy e isolamento de dados conforme especificado.

Definições de Dados e Interfaces para Plataforma de Gestão de Secrets Multi-Tenant
Este documento detalha as estruturas de dados e contratos de interface para a plataforma de gestão de secrets, com foco rigoroso em isolamento multi-tenant e desempenho.

1. MONGODB SCHEMA
O MongoDB será utilizado para persistência de dados. Todas as coleções específicas de tenant incluirão o campo tenant_id obrigatório e terão índices compostos começando com este campo para otimizar consultas e garantir isolamento lógico.

Convenções
_id: Campo padrão do MongoDB, usaremos UUIDs (string) para identificadores de recursos quando possível, para consistência e facilidade de referência entre serviços.
tenant_id: Campo string (UUID) obrigatório em todas as coleções que contêm dados específicos de um tenant.
created_at, updated_at: Campos ISODate para registro de tempo.
encryption_metadata: Objeto para armazenar metadados de criptografia (e.g., iv, algoritmo, dek_version_id) para dados sensíveis.
Relacionamentos: Implementados por referência (IDs) entre documentos em diferentes coleções.
Estruturas de Documentos
Coleção: tenants
{
 "_id": "UUID_DO_TENANT", // Ex: "a1b2c3d4-e5f6-7890-1234-567890abcdef"
 "name": "Nome do Tenant",
 "status": "active", // "active", "suspended", "inactive"
 "plan_id": "UUID_DO_PLANO_DE_ASSINATURA",
 "kek_id": "UUID_DA_KEK_ASSOCIADA", // ID da Key Encryption Key deste tenant
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "name": 1 } (único)
{ "status": 1 }
Coleção: users
{
 "_id": "UUID_DO_USUARIO", // Ex: "b2c3d4e5-f6a7-8901-2345-67890abcdef0"
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "email": "usuario@exemplo.com",
 "username": "nome.usuario",
 "password_hash": "$2a$10$hashedpassword...", // Se houver autenticação local
 "first_name": "Nome",
 "last_name": "Sobrenome",
 "roles": ["admin", "viewer"], // Papéis dentro do tenant
 "status": "active", // "active", "pending", "disabled"
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "email": 1 } (único, otimiza busca por email dentro do tenant)
{ "tenant_id": 1, "username": 1 } (único, otimiza busca por username dentro do tenant)
{ "tenant_id": 1, "status": 1 }
Coleção: vaults
{
 "_id": "UUID_DO_VAULT", // Ex: "c3d4e5f6-a7b8-9012-3456-7890abcdef01"
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "name": "Vault de Produção",
 "description": "Secrets para ambiente de produção",
 "dek_id": "UUID_DA_DEK_ASSOCIADA", // ID da Data Encryption Key deste vault
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "name": 1 } (único, otimiza busca por nome dentro do tenant)
Coleção: secrets
{
 "_id": "UUID_DO_SECRET", // Ex: "d4e5f6a7-b8c9-0123-4567-890abcdef012"
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "vault_id": "UUID_DO_VAULT", // Referência ao Vault
 "name": "ChaveAPI_ServicoXYZ",
 "type": "password", // "password", "api_key", "certificate_ref", "ssh_key"
 "description": "Chave para acesso ao serviço XYZ",
 "encrypted_value": BinData(0, "...base64_do_valor_criptografado..."), // Valor sensível criptografado
 "encryption_metadata": {
   "iv": "base64_do_vetor_de_inicializacao",
   "algorithm": "AES256-GCM",
   "dek_version_id": "UUID_DA_VERSAO_DA_DEK" // Para rastrear qual versão da DEK foi usada
 },
 "tags": ["servico-xyz", "producao"],
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z"),
 "expires_at": ISODate("2024-10-27T10:00:00Z") // Opcional
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "vault_id": 1, "name": 1 } (único, otimiza busca por nome dentro do vault/tenant)
{ "tenant_id": 1, "vault_id": 1, "type": 1 }
{ "tenant_id": 1, "tags": 1 }
{ "tenant_id": 1, "expires_at": 1 }
Coleção: certificates
{
 "_id": "UUID_DO_CERTIFICADO", // Ex: "e5f6a7b8-c9d0-1234-5678-90abcdef0123"
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "vault_id": "UUID_DO_VAULT", // Referência ao Vault
 "name": "SSL_WebApp",
 "description": "Certificado SSL para Web App de Produção",
 "encrypted_certificate": BinData(0, "...base64_do_certificado_criptografado..."),
 "encrypted_private_key": BinData(0, "...base64_da_chave_privada_criptografada..."), // Opcional, se a chave é gerenciada
 "encryption_metadata": {
   "iv": "base64_do_vetor_de_inicializacao",
   "algorithm": "AES256-GCM",
   "dek_version_id": "UUID_DA_VERSAO_DA_DEK"
 },
 "issuer": "Let's Encrypt",
 "subject": "CN=webapp.example.com",
 "serial_number": "1234567890ABCDEF",
 "not_before": ISODate("2023-01-01T00:00:00Z"),
 "not_after": ISODate("2024-01-01T00:00:00Z"),
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "vault_id": 1, "name": 1 } (único)
{ "tenant_id": 1, "not_after": 1 } (para expiração)
Coleção: agents
{
 "_id": "UUID_DO_AGENTE", // Ex: "f6a7b8c9-d0e1-2345-6789-0abcdef01234"
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "name": "Agent_Prod_Server1",
 "status": "online", // "online", "offline", "decommissioned"
 "version": "1.0.0",
 "last_heartbeat_at": ISODate("2023-10-27T10:30:00Z"),
 "metadata": {
   "ip_address": "192.168.1.100",
   "os": "Linux",
   "hostname": "server1.example.local"
 },
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "name": 1 } (único)
{ "tenant_id": 1, "status": 1 }
{ "tenant_id": 1, "last_heartbeat_at": 1 }
Coleção: audit_logs
{
 "_id": ObjectId("653b6f0e4a7b8c9d0e1f2a3b"), // ObjectId é preferível para logs sequenciais
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "user_id": "UUID_DO_USUARIO_OU_NULO", // Quem realizou a ação
 "agent_id": "UUID_DO_AGENTE_OU_NULO", // Se a ação veio de um agente
 "event_type": "SecretAccessed", // Ex: "SecretCreated", "VaultPolicyUpdated", "UserLoggedIn"
 "resource_type": "secret", // "secret", "vault", "user", "policy", "tenant"
 "resource_id": "UUID_DO_RECURSO_AFETADO",
 "timestamp": ISODate("2023-10-27T10:35:00Z"),
 "details": {
   "action": "read_value",
   "field": "value",
   "client_ip": "192.0.2.1"
 },
 "outcome": "success" // "success", "failure"
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "timestamp": -1 } (otimiza busca por logs mais recentes)
{ "tenant_id": 1, "resource_type": 1, "resource_id": 1, "timestamp": -1 }
{ "tenant_id": 1, "user_id": 1, "timestamp": -1 }
{ "tenant_id": 1, "event_type": 1 }
Coleção: authorization_policies
{
 "_id": "UUID_DA_POLITICA",
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "name": "Politica_Admins_VaultProducao",
 "description": "Permissões de administradores para Vault de Produção",
 "rules": [
   {
     "action": "read",
     "resource_type": "secret",
     "resource_id": "*", // "*" significa todos, pode ser um UUID específico
     "vault_id": "UUID_DO_VAULT_PRODUCAO", // Opcional, para granularidade por vault
     "roles": ["admin"]
   },
   {
     "action": "write",
     "resource_type": "secret",
     "resource_id": "*",
     "vault_id": "UUID_DO_VAULT_PRODUCAO",
     "roles": ["admin"]
   }
 ],
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "name": 1 } (único)
{ "tenant_id": 1, "rules.roles": 1 }
Coleção: keks (Key Encryption Keys)
{
 "_id": "UUID_DA_KEK", // Ex: "a1b2c3d4-e5f6-7890-1234-567890abcdef" (referenciado pelo tenant)
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "kms_key_identifier": "arn:aws:kms:region:account:key/key_id", // Identificador no KMS da nuvem
 "status": "active", // "active", "pending_rotation", "rotated", "deactivated"
 "version": 1, // Versão da KEK para rotação
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "status": 1 } (para encontrar a KEK ativa de um tenant)
{ "tenant_id": 1, "version": 1 }
Coleção: deks (Data Encryption Keys)
{
 "_id": "UUID_DA_DEK", // Ex: "c3d4e5f6-a7b8-9012-3456-7890abcdef01" (referenciado pelo vault)
 "tenant_id": "UUID_DO_TENANT", // Obrigatório
 "vault_id": "UUID_DO_VAULT", // Referência ao vault que esta DEK protege
 "kek_id": "UUID_DA_KEK", // Referência à KEK que criptografou esta DEK
 "encrypted_dek_value": BinData(0, "...base64_da_dek_criptografada_pela_kek..."),
 "status": "active", // "active", "pending_rotation", "rotated", "deactivated"
 "version": 1, // Versão da DEK para rotação
 "created_at": ISODate("2023-10-27T10:00:00Z"),
 "updated_at": ISODate("2023-10-27T10:00:00Z")
}
Índices:

{ "_id": 1 } (primário)
{ "tenant_id": 1, "vault_id": 1, "status": 1 } (para encontrar a DEK ativa de um vault)
{ "tenant_id": 1, "kek_id": 1 }
{ "tenant_id": 1, "vault_id": 1, "version": 1 }
2. gRPC DEFINITIONS (.proto)
As definições gRPC (.proto) serão usadas para a comunicação inter-serviços. O tenant_id e user_id serão propagados através de metadados da requisição gRPC por interceptors, conforme especificado. No entanto, para operações de criação/atualização de recursos, o tenant_id será incluído explicitamente nas mensagens para garantir consistência e permitir validação na camada de serviço antes da persistência.

Convenções
google/protobuf/timestamp.proto: Para campos de data/hora.
common.proto: Um arquivo proto comum para tipos e estruturas reutilizáveis (e.g., OperationResult, Pagination, Sorting).
UUIDs: Representados como string.
common.proto
syntax = "proto3";

package common;

option go_package = "github.com/your-org/project/internal/pb/common";

message OperationResult {
   bool success = 1;
   string message = 2;
   string error_code = 3; // Opcional, para erros específicos
}

message Pagination {
   int32 limit = 1;
   int32 offset = 2;
}

message Sorting {
   string field = 1;
   string order = 2; // "asc" ou "desc"
}

message PaginationResult {
   int32 total_count = 1;
   int32 limit = 2;
   int32 offset = 3;
   bool has_next_page = 4;
   bool has_previous_page = 5;
}
secrets.proto
Define o serviço e as mensagens para o gerenciamento de secrets.

syntax = "proto3";

package secrets;

option go_package = "github.com/your-org/project/internal/core/secrets/pb";

import "google/protobuf/timestamp.proto";
import "common.proto";

// SecretService define o contrato gRPC para o gerenciamento de secrets.
service SecretService {
   // Cria um novo secret.
   rpc CreateSecret(CreateSecretRequest) returns (SecretResponse);
   // Obtém os metadados de um secret específico.
   rpc GetSecret(GetSecretRequest) returns (SecretResponse);
   // Atualiza um secret existente.
   rpc UpdateSecret(UpdateSecretRequest) returns (SecretResponse);
   // Deleta um secret.
   rpc DeleteSecret(DeleteSecretRequest) returns (common.OperationResult);
   // Lista os secrets com base em critérios de filtro e paginação.
   rpc ListSecrets(ListSecretsRequest) returns (ListSecretsResponse);
   // Acessa o valor descriptografado de um secret. Requer autorização e log de auditoria.
   rpc AccessSecret(AccessSecretRequest) returns (AccessSecretResponse);
}

// Secret representa os metadados de um secret.
message Secret {
   string id = 1;
   string tenant_id = 2;
   string vault_id = 3;
   string name = 4;
   string type = 5; // e.g., "password", "api_key", "certificate_ref"
   string description = 6;
   repeated string tags = 7;
   google.protobuf.Timestamp created_at = 8;
   google.protobuf.Timestamp updated_at = 9;
   google.protobuf.Timestamp expires_at = 10; // Opcional
}

// SecretValue representa o valor descriptografado de um secret.
message SecretValue {
   string value = 1;
   google.protobuf.Timestamp accessed_at = 2;
}

// CreateSecretRequest é a requisição para criar um secret.
message CreateSecretRequest {
   string tenant_id = 1; // Explicitamente incluído para criação e validação
   string vault_id = 2;
   string name = 3;
   string value = 4; // Valor em texto claro, será criptografado pelo serviço
   string type = 5;
   string description = 6;
   repeated string tags = 7;
   google.protobuf.Timestamp expires_at = 8; // Opcional
}

// SecretResponse contém o secret criado ou recuperado.
message SecretResponse {
   Secret secret = 1;
}

// GetSecretRequest é a requisição para obter um secret por ID.
message GetSecretRequest {
   string secret_id = 1;
}

// UpdateSecretRequest é a requisição para atualizar um secret.
message UpdateSecretRequest {
   string tenant_id = 1; // Explicitamente incluído para atualização e validação
   string secret_id = 2;
   // Campos opcionais para atualização. Use go.protobuf.wrappers para tipagem nullable se necessário.
   // Para simplificar, assumimos que campos não preenchidos não serão atualizados.
   string name = 3;
   string value = 4; // Novo valor em texto claro, será criptografado
   string type = 5;
   string description = 6;
   repeated string tags = 7;
   google.protobuf.Timestamp expires_at = 8; // Opcional
}

// DeleteSecretRequest é a requisição para deletar um secret.
message DeleteSecretRequest {
   string secret_id = 1;
}

// ListSecretsRequest é a requisição para listar secrets.
message ListSecretsRequest {
   string vault_id = 1; // Filtro opcional por vault
   common.Pagination pagination = 2;
   common.Sorting sorting = 3;
   map<string, string> filters = 4; // Filtros genéricos (e.g., name_contains, type_eq)
}

// ListSecretsResponse contém uma lista de secrets e informações de paginação.
message ListSecretsResponse {
   repeated Secret secrets = 1;
   common.PaginationResult pagination_result = 2;
}

// AccessSecretRequest é a requisição para acessar o valor de um secret.
message AccessSecretRequest {
   string secret_id = 1;
   string agent_id = 2; // Opcional: ID do agente, se o acesso for via agente
   string reason = 3;   // Razão do acesso para fins de auditoria
}

// AccessSecretResponse contém o valor descriptografado do secret.
message AccessSecretResponse {
   SecretValue secret_value = 1;
}
vaults.proto
Define o serviço e as mensagens para o gerenciamento de vaults.

syntax = "proto3";

package vaults;

option go_package = "github.com/your-org/project/internal/core/vaults/pb";

import "google/protobuf/timestamp.proto";
import "common.proto";
import "secrets.proto"; // Para listar secrets aninhados

// VaultService define o contrato gRPC para o gerenciamento de vaults.
service VaultService {
   rpc CreateVault(CreateVaultRequest) returns (VaultResponse);
   rpc GetVault(GetVaultRequest) returns (VaultResponse);
   rpc UpdateVault(UpdateVaultRequest) returns (VaultResponse);
   rpc DeleteVault(DeleteVaultRequest) returns (common.OperationResult);
   rpc ListVaults(ListVaultsRequest) returns (ListVaultsResponse);
}

// Vault representa um contêiner lógico para secrets e certificados.
message Vault {
   string id = 1;
   string tenant_id = 2;
   string name = 3;
   string description = 4;
   string dek_id = 5; // ID da DEK associada a este vault
   google.protobuf.Timestamp created_at = 6;
   google.protobuf.Timestamp updated_at = 7;
}

// CreateVaultRequest é a requisição para criar um vault.
message CreateVaultRequest {
   string tenant_id = 1; // Explicitamente incluído para criação e validação
   string name = 2;
   string description = 3;
}

// VaultResponse contém o vault criado ou recuperado.
message VaultResponse {
   Vault vault = 1;
}

// GetVaultRequest é a requisição para obter um vault por ID.
message GetVaultRequest {
   string vault_id = 1;
}

// UpdateVaultRequest é a requisição para atualizar um vault.
message UpdateVaultRequest {
   string tenant_id = 1; // Explicitamente incluído para atualização e validação
   string vault_id = 2;
   string name = 3; // Opcional
   string description = 4; // Opcional
}

// DeleteVaultRequest é a requisição para deletar um vault.
message DeleteVaultRequest {
   string vault_id = 1;
}

// ListVaultsRequest é a requisição para listar vaults.
message ListVaultsRequest {
   common.Pagination pagination = 1;
   common.Sorting sorting = 2;
   map<string, string> filters = 3; // Filtros genéricos (e.g., name_contains)
}

// ListVaultsResponse contém uma lista de vaults e informações de paginação.
message ListVaultsResponse {
   repeated Vault vaults = 1;
   common.PaginationResult pagination_result = 2;
}
3. GRAPHQL TYPES
O GraphQL será exposto pelo API Gateway ou por um BFF (Backend For Frontend) para o frontend. O tenant_id é gerenciado internamente pelo gateway/BFF através do JWT do usuário autenticado e propagado para os microsserviços via gRPC, portanto, não é necessário expô-lo diretamente como argumento em todas as queries/mutations, exceto quando a operação é diretamente sobre o tenant (e.g., updateTenant).

Convenções
IDs: ID! para identificadores únicos.
Datas: String! formatadas (e.g., ISO 8601).
Conexões: Uso do padrão Relay para paginação (e.g., SecretConnection, SecretEdge).
Entradas: Tipos Input para mutações.
Cargas (Payloads): Tipos Payload para retornos de mutações, incluindo success e message.
schema.graphql
# Esquema GraphQL para a Plataforma de Gestão de Secrets Multi-Tenant

# --- Query (Operações de Leitura) ---
type Query {
   "Retorna os detalhes do usuário autenticado."
   me: User!

   "Retorna os detalhes do tenant atual do usuário autenticado."
   tenant: Tenant!

   "Obtém um vault por ID."
   vault(id: ID!): Vault

   "Lista vaults, com filtros e paginação opcionais."
   vaults(filter: VaultFilterInput, pagination: PaginationInput): VaultConnection!

   "Obtém um secret por ID."
   secret(id: ID!): Secret

   "Lista secrets, com filtros, agrupados por vault ou em todo o tenant, com paginação."
   secrets(vaultId: ID, filter: SecretFilterInput, pagination: PaginationInput): SecretConnection!

   "Obtém um certificado por ID."
   certificate(id: ID!): Certificate

   "Lista certificados, com filtros e paginação opcionais."
   certificates(vaultId: ID, filter: CertificateFilterInput, pagination: PaginationInput): CertificateConnection!

   "Obtém um agente por ID."
   agent(id: ID!): Agent

   "Lista agentes, com filtros e paginação opcionais."
   agents(filter: AgentFilterInput, pagination: PaginationInput): AgentConnection!
}

# --- Mutation (Operações de Escrita) ---
type Mutation {
   "Cria um novo usuário no tenant atual."
   createUser(input: CreateUserInput!): UserPayload!
   "Atualiza um usuário existente no tenant atual."
   updateUser(id: ID!, input: UpdateUserInput!): UserPayload!
   "Deleta um usuário do tenant atual."
   deleteUser(id: ID!): DeletePayload!

   "Cria um novo vault no tenant atual."
   createVault(input: CreateVaultInput!): VaultPayload!
   "Atualiza um vault existente no tenant atual."
   updateVault(id: ID!, input: UpdateVaultInput!): VaultPayload!
   "Deleta um vault do tenant atual."
   deleteVault(id: ID!): DeletePayload!

   "Cria um novo secret dentro de um vault no tenant atual."
   createSecret(input: CreateSecretInput!): SecretPayload!
   "Atualiza um secret existente dentro de um vault no tenant atual."
   updateSecret(id: ID!, input: UpdateSecretInput!): SecretPayload!
   "Deleta um secret do tenant atual."
   deleteSecret(id: ID!): DeletePayload!
   "Acessa e retorna o valor descriptografado de um secret. Requer justificativa para auditoria."
   accessSecret(id: ID!, reason: String!): AccessSecretPayload!

   "Cria um novo certificado dentro de um vault no tenant atual."
   createCertificate(input: CreateCertificateInput!): CertificatePayload!
   "Atualiza um certificado existente dentro de um vault no tenant atual."
   updateCertificate(id: ID!, input: UpdateCertificateInput!): CertificatePayloa
