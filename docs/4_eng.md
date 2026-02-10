# Especificação Técnica de Implementação: Gestão de Secrets SaaS

Este documento detalha a especificação técnica para a implementação da Plataforma de Gestão de Secrets SaaS Multi-Tenant, conforme a arquitetura proposta. Focaremos na aplicação estrita da pilha tecnológica definida, padrões de desenvolvimento e estratégias de segurança.

---

## 1. Visão Geral e Mapeamento da Stack Técnica

A arquitetura de microserviços será implementada utilizando Go para o backend, React/TypeScript para o frontend, e uma série de tecnologias para garantir escalabilidade, segurança e observabilidade. A tabela abaixo mapeia os componentes arquiteturais da visão geral para as tecnologias de implementação específicas.

| Componente Arquitetural | Mapeamento da Stack Técnica |
| :--- | :--- |
| **Control Plane** | |
| Proxy de Borda/Load Balancer | Kubernetes Ingress Controller (Nginx, Traefik) + Cloud Load Balancer |
| API Gateway | Golang (Gin Framework): Centraliza autenticação, autorização inicial, roteamento e GraphQL API |
| Serviço de Frontend | React com TypeScript + Tailwind CSS: Aplicação web SPA (Single Page Application) |
| Serviço de Usuários | Golang (Gin/gRPC): Gerenciamento de usuários, perfis, roles. Interage com MongoDB. |
| Serviço de Tenant | Golang (Gin/gRPC): Gerenciamento de ciclo de vida de tenants, planos, cotas. Interage com MongoDB. |
| Serviço de Faturamento | Golang (Gin/gRPC): Integração com provedores de pagamento, cálculo de uso. Interage com MongoDB e RabbitMQ. |
| Serviço de Notificação | Golang (gRPC): Consome eventos RabbitMQ para enviar notificações. |
| Serviço de Auditoria | Golang (gRPC): Consome eventos RabbitMQ para registrar logs imutáveis no MongoDB. |
| Serviço de Autorização | Golang (gRPC): Motor de autorização fine-grained (RBAC/ABAC). |
| Serviço de Agentes | Golang (gRPC): Gerenciamento de registro e status de Agentes. |
| Serviço de Chaves Mestras | Golang (gRPC): Abstração para KMS da nuvem (AWS KMS, GCP KMS, Azure Key Vault). |
| Integração IDP/MFA | Golang (Gin/gRPC): Integração OAuth2, SAML/OIDC, TOTP/U2F. Emite JWTs internos. |
| Gerenciamento de Vaults | Golang (Gin/gRPC): Gerencia metadados de vaults, políticas de acesso. |
| Gerenciamento de Secrets | Golang (Gin/gRPC): CRUD de secrets/certificados/chaves. |
| Rotação/Certificados | Golang (gRPC) + Temporal Workflows: Orquestra a rotação automática. |
| **Data Plane** | |
| Banco de Dados de Secrets | MongoDB (NoSQL): Armazenamento de secrets criptografados. |
| Banco de Dados de Metadados | MongoDB (NoSQL): Armazenamento de metadados de vaults, tenants, etc. |
| **Agent Plane** | |
| Agente (On-premise) | Golang: Comunica via mTLS. Atua como Worker Temporal local. |
| **Serviços Compartilhados** | |
| Filas de Mensagens | RabbitMQ: Comunicação assíncrona entre microsserviços. |
| Cache Distribuído | Redis: Cache de dados frequentemente acessados. |
| Monitoramento | OpenTelemetry + Prometheus + Grafana + Loki/Elasticsearch |
| **Infraestrutura** | |
| Orquestração | Kubernetes (K8s) |
| Armazenamento de Objetos | Cloud Object Storage (S3, GCS, Blob Storage) |

---

## 2. Estrutura do Projeto (Golang)

A estrutura de pastas Golang será organizada para promover a modularidade e clareza.

```text
project_folder:
├── cmd/
│   └── api/                  # Ponto de entrada da aplicação principal (serviço web/API Gateway)
│       ├── hooks/            # Funções de ciclo de vida do Uber Fx (e.g., conectar/desconectar DB)
│       ├── providers/        # Funções para provisionar dependências globais
│       ├── types/            # Tipos usados pelos providers e hooks
│       └── main.go           # Configura e inicia o container Uber Fx
├── configs/                  # Arquivos de configuração (YAML, TOML, env vars)
├── pkgs/                     # Pacotes utilitários gerais
├── docs/
│   ├── api/                  # Documentação da API (OpenAPI/Swagger)
│   ├── deployments/          # Documentos de deploy (K8s manifests, Helm)
│   └── architecture/         # Documentação arquitetural
└── internal/                 # Código privado da aplicação
    ├── core/                 # Lógica de negócio principal por módulo
    │   └── [module_name]/    # Ex: users, tenants, vaults, secrets
    │       ├── models/       # Estruturas de dados
    │       ├── services/     # Implementação da lógica de negócio
    │       ├── repositories/ # Abstração para acesso a dados (interface)
    │       ├── usecases/     # Orquestra a lógica de negócio
    │       ├── tools/        # Funções utilitárias específicas
    │       ├── handlers/     # Pontos de entrada (HTTP/gRPC)
    │       └── module.go     # Configuração Uber Fx do módulo
    └── infra/                # Implementações de infraestrutura
        ├── database/         # MongoDB client wrapper
        ├── telemetry/        # Instrumentação OpenTelemetry
        ├── messaging/        # Cliente RabbitMQ
        ├── cache/            # Cliente Redis
        ├── security/         # JWT, criptografia, KMS client
        └── clients/          # Clientes para serviços de terceiros
```

---

## 3. Implementação Multi-Tenancy

A multi-tenancy será implementada garantindo o isolamento lógico em todas as camadas.

### 3.1. Contexto do Tenant com JWT
- **sub**: ID do usuário.
- **tid**: ID do tenant.
- **roles**: Roles no contexto do tenant.

Um middleware Gin validará o JWT, injetará o `Tenant ID` no `context.Context` e consultará o Serviço de Autorização.

### 3.2. Propagação gRPC
A comunicação interna usará Interceptors gRPC para propagar o context via metadados (Client-Side e Server-Side).

### 3.3. Isolamento de Dados (MongoDB)
- **Modelo**: Shared Database com `tenant_id` em cada documento.
- **Indexação**: Índices compostos iniciando com `tenant_id`.
- **Repositório**: Adição automática do filtro de tenant em todas as operações.

### 3.4. Propagação RabbitMQ
O `TenantID` será incluído em todos os DTOs de eventos e, opcionalmente, nas Routing Keys (`secrets.rotated.tenant.<id>`).

---

## 4. Arquitetura de Serviços e Camadas

### 4.1. API Gateway (Gin)
Centraliza autenticação, rate limiting e roteamento. Pode atuar como BFF (Backend for Frontend).

### 4.2. Serviços de Negócio
Organizados em `internal/core`, seguindo o padrão de separação entre `usecases`, `services` e `repositories`.

### 4.3. Comunicação Interna (gRPC)
Uso rigoroso de Protocol Buffers para contratos de serviço e interceptores para segurança (mTLS).

### 4.4. GraphQL
Recomenda-se o uso de `gqlgen` no BFF para agregar dados de múltiplos microserviços.

### 4.5. Data Plane On-Premise (Proxy Agent)
- **Executável Go**: Binário leve e independente.
- **Orquestrador Local**: Capaz de gerenciar contêineres locais.
- **Worker Temporal**: Processa workflows locais (rotação/updates OTA).

---

## 5. Gerenciamento de Dados

*   **MongoDB**: Wrapper customizado para injeção de `tenant_id`, suporte a transações e backups point-in-time.
*   **Redis**: Chaves de cache prefixadas com `{tenant_id}` e estratégia Cache-Aside.

---

## 6. Mensageria e Workflows

*   **RabbitMQ**: Uso de Exchanges (Direct/Topic) e Filas duráveis com DLQ para resiliência.
*   **Temporal**: Essencial para workflows complexos (Rotação de Chaves, Renovação de Certificados, Onboarding/Offboarding).
    - **Multi-tenancy**: O Workflow ID incluirá o `tenant_id`.

---
*Especificação Técnica - v1.0 - Fevereiro de 2026*
