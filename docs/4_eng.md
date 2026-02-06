Este documento detalha a especificação técnica para a implementação da Plataforma de Gestão de Secrets SaaS Multi-Tenant, conforme a arquitetura proposta. Focaremos na aplicação estrita da pilha tecnológica definida, padrões de desenvolvimento e estratégias de segurança.

Especificação Técnica de Implementação: Plataforma de Gestão de Secrets SaaS Multi-Tenant
1. Visão Geral e Mapeamento da Stack Técnica
A arquitetura de microserviços será implementada utilizando Go para o backend, React/TypeScript para o frontend, e uma série de tecnologias para garantir escalabilidade, segurança e observabilidade. A tabela abaixo mapeia os componentes arquiteturais da visão geral para as tecnologias de implementação específicas.

Componente Arquitetural (Documento Original)    Mapeamento da Stack Técnica
Control Plane  
Proxy de Borda/Load Balancer    Kubernetes Ingress Controller (Nginx, Traefik) + Cloud Load Balancer
API Gateway Golang (Gin Framework): Centraliza autenticação, autorização inicial, roteamento e GraphQL API (opcionalmente)
Serviço de Frontend React com TypeScript + Tailwind CSS: Aplicação web SPA (Single Page Application)
Serviço de Usuários Golang (Gin/gRPC): Gerenciamento de usuários, perfis, roles. Interage com MongoDB.
Serviço de Tenant   Golang (Gin/gRPC): Gerenciamento de ciclo de vida de tenants, planos, cotas. Interage com MongoDB.
Serviço de Faturamento  Golang (Gin/gRPC): Integração com provedores de pagamento, cálculo de uso, aplicação de limites. Interage com MongoDB e RabbitMQ.
Serviço de Notificação  Golang (gRPC): Consome eventos RabbitMQ para enviar notificações (e-mail, push, etc.).
Serviço de Auditoria    Golang (gRPC): Consome eventos RabbitMQ para registrar logs de auditoria imutáveis no MongoDB.
Serviço de Autorização  Golang (gRPC): Motor de autorização fine-grained (RBAC/ABAC). Consulta políticas no MongoDB/Redis.
Serviço de Agentes  Golang (gRPC): Gerenciamento de registro e status de Agentes. Interage com MongoDB.
Serviço de Chaves Mestras da Nuvem  Golang (gRPC): Abstração para KMS (Key Management Service) da nuvem (ex: AWS KMS, GCP KMS, Azure Key Vault) ou solução interna baseada em HSM. Gerencia Root Keys e KEKs.
Serviço de Integração de IDP/MFA    Golang (Gin/gRPC): Integração OAuth2, SAML/OIDC, TOTP/U2F. Emite JWTs internos.
Serviço de Gerenciamento de Vaults  Golang (Gin/gRPC): Gerencia metadados de vaults, políticas de acesso, KEKs associadas. Interage com MongoDB e Serviço de Chaves Mestras.
Serviço de Gerenciamento de Secrets Golang (Gin/gRPC): CRUD de secrets/certificados/chaves. Interage com MongoDB, Serviço de Gerenciamento de Vaults (para DEKs).
Serviço de Rotação/Certificados Golang (gRPC) + Temporal Workflows: Orquestra a rotação automática de DEKs, KEKs, secrets e certificados. Interage com MongoDB, RabbitMQ, Serviço de Chaves Mestras.
Data Plane 
Banco de Dados de Secrets   MongoDB (NoSQL): Armazenamento de secrets criptografados.
Banco de Dados de Certificados  MongoDB (NoSQL): Armazenamento de certificados/chaves criptografados.
Banco de Dados de Metadados MongoDB (NoSQL): Armazenamento de metadados de vaults, tenants, usuários, etc.
Agent Plane
Agente (On-premise) Golang (Executável/Contêiner Docker): Comunica via mTLS com Control Plane. Atua como um Worker Temporal local e orquestrador de microsserviços do Data Plane on-premise.
Serviços Compartilhados/Terceiros  
Provedor de Identidade, MFA, Pagamentos APIs REST externas.
Sistema de Filas de Mensagens   RabbitMQ: Comunicação assíncrona entre microsserviços.
Serviço de Cache Distribuído    Redis: Cache de dados frequentemente acessados.
Serviço de Logging Centralizado OpenTelemetry (coleta) + Prometheus/Grafana (visualização) + Loki/Elasticsearch (storage/search)
Serviço de Monitoramento/Alerta OpenTelemetry (coleta) + Prometheus (métricas) + Grafana (dashboards/alertas)
Componentes de Infraestrutura  
Plataforma de Orquestração  Kubernetes (K8s): Orquestração de contêineres.
Registros de Contêineres    Docker Registry (ex: GitHub Container Registry, ECR, GCR, ACR)
Armazenamento de Objetos    Cloud Object Storage (ex: S3, GCS, Blob Storage) para backups, artefatos.
2. Estrutura do Projeto (Golang)
A estrutura de pastas Golang será organizada para promover a modularidade, separação de responsabilidades e clareza, alinhada com as melhores práticas para aplicações Go de microserviços.

project_folder:
├── cmd/
│   └── api/                  # Ponto de entrada da aplicação principal (serviço web/API Gateway)
│       ├── hooks/            # Funções de ciclo de vida do Uber Fx (e.g., conectar/desconectar DB)
│       ├── providers/        # Funções para provisionar dependências globais (DB, Cache, Mensageria, Telemetria)
│       ├── types/            # Tipos usados pelos providers e hooks
│       └── main.go           # Configura e inicia o container Uber Fx
├── configs/                  # Arquivos de configuração (YAML, TOML, env vars)
├── pkgs/                     # Pacotes utilitários gerais que podem ser compartilhados por múltiplos projetos (raro em microserviços)
├── scripts/                  # Scripts úteis (build, deploy, setup local)
├── docs/
│   ├── api/                  # Documentação da API (OpenAPI/Swagger, GraphQL Schema)
│   ├── deployments/          # Documentos de deploy (Kubernetes manifests, Helm charts)
│   └── architecture/         # Documentação arquitetural
└── internal/                 # Código privado da aplicação, não destinado a ser importado por outros projetos
   ├── core/                 # Lógica de negócio principal, dividida por módulos/microsserviços
   │   └── [module_name]/    # Ex: users, tenants, vaults, secrets
   │       ├── models/       # Estruturas de dados (domínio, ORM/ODM)
   │       ├── services/     # Implementação da lógica de negócio (funções puras ou com dependências mínimas)
   │       ├── repositories/ # Abstração para acesso a dados (interface)
   │       ├── usecases/     # Orquestra a lógica de negócio, interage com services e repositories
   │       ├── tools/        # Funções utilitárias específicas do módulo
   │       ├── handlers/     # Pontos de entrada da API para o módulo
   │       │   ├── grpc/     # Implementações de servidores gRPC
   │       │   ├── http/     # Implementações de handlers HTTP (Gin)
   │       │   └── middleware/ # Middlewares específicos do módulo
   │       └── module.go     # Configuração de injeção de dependências do módulo (Uber Fx)
   └── infra/                # Implementações de infraestrutura e clientes externos
       ├── database/         # Implementação de acesso a banco de dados (MongoDB client wrapper)
       ├── telemetry/        # Configuração e instrumentação OpenTelemetry
       ├── messaging/        # Implementação de cliente RabbitMQ
       ├── cache/            # Implementação de cliente Redis
       ├── security/         # Utilitários de segurança (JWT, criptografia, KMS client)
       ├── handlers/         # Middlewares globais (autenticação, autorização, CORS, tracing)
       │   ├── grpc/         # Interceptors gRPC (global)
       │   └── http/         # Middlewares HTTP (global Gin)
       └── clients/          # Clientes para serviços de terceiros (IDP, pagamentos)
3. Implementação Multi-Tenancy
A multi-tenancy será implementada de forma robusta, garantindo o isolamento lógico dos dados e contexto do tenant em todas as camadas da aplicação.

3.1. Contexto do Tenant com JWT
Geração do JWT: Após a autenticação bem-sucedida (via OAuth2/SAML/OIDC no Serviço de Integração de IDP/MFA), um JWT interno será gerado. Este JWT conterá no mínimo as seguintes claims:
sub (Subject): ID do usuário.
tid (Tenant ID): ID do tenant ao qual o usuário pertence (ou tem contexto de acesso).
roles: Roles do usuário no contexto do tenant.
exp: Tempo de expiração. O JWT será assinado digitalmente pelo Serviço de Integração de IDP/MFA com uma chave privada.
Middleware de Autenticação/Autorização (Gin):
Um middleware Gin (internal/infra/handlers/http/auth_middleware.go) será configurado no API Gateway para interceptar todas as requisições autenticadas.
Ele validará o JWT (assinatura, expiração).
Extrairá o tid do JWT.
Injetará o Tenant ID e User ID num context.Context (ex: ctx = context.WithValue(ctx, tenantIDKey, tenantID)).
Consultará o Serviço de Autorização para verificações iniciais de autorização baseadas em políticas globais/permissões de API.
3.2. Propagação para Serviços Internos (gRPC)
A comunicação entre microserviços (backend para backend) usará gRPC. O Tenant ID e User ID serão propagados via metadados de requisição gRPC.

Interceptors gRPC (Client-Side):
No API Gateway ou em qualquer microserviço que atue como cliente gRPC, um interceptor internal/infra/handlers/grpc/client_interceptor.go será implementado.
Este interceptor extrairá o Tenant ID e User ID do context.Context da requisição HTTP de entrada (ou do contexto atual) e os adicionará aos metadados gRPC (ex: metadata.NewOutgoingContext).
Interceptors gRPC (Server-Side):
Em cada microserviço que atua como servidor gRPC, um interceptor internal/infra/handlers/grpc/server_interceptor.go será configurado.
Ele extrairá o Tenant ID e User ID dos metadados da requisição gRPC.
Injetará esses IDs em um novo context.Context que será passado para os handlers do serviço.
Este interceptor também poderá realizar verificações de autorização intermediárias consultando o Serviço de Autorização.
3.3. Isolamento de Dados (MongoDB)
O modelo "Shared Database with Tenant Identifier" será aplicado ao MongoDB, com criptografia adicional para dados sensíveis.

Estrutura de Coleções: Todas as coleções que armazenam dados específicos do tenant terão um campo tenant_id obrigatório (tipo string ou UUID).
Exemplo: db.collection("secrets").insertOne({ tenant_id: "...", vault_id: "...", data: "..." }).
Indexação: Índices compostos serão criados em todas as coleções sensíveis, com tenant_id como o primeiro campo. Isso otimiza as consultas e garante que as operações de busca sejam eficientes para um tenant específico.
Exemplo: db.collection("secrets").createIndex({ tenant_id: 1, vault_id: 1, name: 1 }).
Abstração de Repositório (internal/core/[module_name]/repositories):
As implementações dos repositórios (por exemplo, secretsRepository, vaultsRepository) serão responsáveis por adicionar automaticamente o tenant_id a todas as queries de leitura, escrita, atualização e exclusão.
Exemplo de pseudocódigo no repositório:
func (r *secretsRepository) FindSecret(ctx context.Context, secretID string) (*Secret, error) {
   tenantID := GetTenantIDFromContext(ctx) // Função utilitária para extrair do contexto
   filter := bson.M{"_id": secretID, "tenant_id": tenantID}
   // ... db.Collection("secrets").FindOne(ctx, filter) ...
}

func (r *secretsRepository) CreateSecret(ctx context.Context, secret *Secret) error {
   tenantID := GetTenantIDFromContext(ctx)
   secret.TenantID = tenantID // Garante que o tenant_id é setado antes de persistir
   // ... db.Collection("secrets").InsertOne(ctx, secret) ...
}
Criptografia Hierárquica: Para dados sensíveis (secrets, certificados), a criptografia em repouso com KEKs (por tenant) e DEKs (por vault) será a camada primária de isolamento de segurança. O tenant_id atua como uma camada lógica de segregação e otimização, enquanto a criptografia garante a confidencialidade mesmo em caso de comprometimento do banco de dados.
3.4. Propagação para Mensageria (RabbitMQ)
O Tenant ID será incluído nas mensagens do RabbitMQ para que os consumidores possam processar eventos no contexto correto.

Corpo da Mensagem (DTO): Todos os DTOs de eventos enviados para o RabbitMQ incluirão um campo TenantID.
Exemplo: { "TenantID": "...", "EventType": "SecretRotated", "SecretID": "..." }.
Routing Keys: Para eventos que necessitam de roteamento específico do tenant, o Tenant ID pode ser incorporado na chave de roteamento.
Exemplo: Publicar em secrets.rotated.tenant.<tenant_id>.
Consumidores podem se ligar a secrets.rotated.tenant.# e filtrar internamente pelo TenantID do corpo da mensagem, ou ter filas específicas que se ligam a secrets.rotated.tenant.<specific_tenant_id> para cenários de workers dedicados por tenant (raro em SaaS).
Consumidores (Workers): Os consumidores do RabbitMQ (incluindo Temporal Workers) extrairão o Tenant ID do corpo da mensagem e o injetarão no context.Context das funções de processamento, garantindo que todas as operações subsequentes (acesso ao DB, chamadas gRPC) respeitem o contexto do tenant.
4. Arquitetura de Serviços e Camadas
4.1. API Gateway (Gin Framework)
Ponto de Entrada: Todas as requisições HTTP externas para o Control Plane.
Middleware Global: Autenticação (JWT validation), autorização básica, rate limiting, CORS, tracing OpenTelemetry, logging.
Roteamento: Mapeamento de rotas HTTP para os handlers específicos dos serviços de negócio.
GraphQL Endpoint: Pode ser implementado aqui como um serviço de "Backend For Frontend" (BFF), que agrega dados de múltiplos microsserviços via gRPC. Isso oferece flexibilidade para o Frontend.
Comunicação: Usa clientes gRPC para interagir com os microsserviços internos.
4.2. Serviços de Negócio (internal/core/[module_name])
Cada module_name representará um microsserviço ou um domínio de negócio coeso.

models/: Definições de estruturas de dados Go que representam entidades de domínio (ex: User, Tenant, Vault, Secret). Incluem tags bson para MongoDB e protobuf para gRPC.
repositories/: Interfaces que definem as operações de persistência de dados para uma entidade. As implementações estarão em internal/infra/database e internal/infra/cache para MongoDB e Redis, respectivamente.
usecases/: Contém a lógica de negócio orquestrada. Um usecase orquestra chamadas a vários repositórios, serviços internos e valida a lógica de negócio. Recebem context.Context e dependências injetadas.
services/: Lógica de negócio mais fina, que pode ser reutilizada por múltiplos usecases.
handlers/http/: Funções Gin que lidam com requisições HTTP, convertem DTOs de entrada, chamam usecases e retornam respostas HTTP. Injetam o Tenant ID e User ID do contexto.
handlers/grpc/: Implementações de servidores gRPC gerados a partir de .proto files. Recebem requisições gRPC, convertem para DTOs internos, chamam usecases e retornam respostas gRPC. Injetam o Tenant ID e User ID do contexto.
module.go: Arquivo para configuração de injeção de dependências do Uber Fx para este módulo, provendo usecases, handlers, etc.
4.3. Comunicação Interna (gRPC)
Protocol Buffers: Usados para definir os contratos de serviço (.proto files) entre os microserviços. Isso garante tipagem forte e serialização eficiente.
Serviços gRPC: Cada microserviço expõe uma API gRPC para comunicação interna.
Interceptors: Conforme detalhado na seção de Multi-Tenancy, usados para mTLS, tracing e propagação de contexto (Tenant ID).
4.4. GraphQL
A implementação de GraphQL pode residir como um endpoint no API Gateway (Gin) ou como um serviço GraphQL dedicado.
Recomenda-se um serviço GraphQL (Serviço de Frontend) que atua como um Backend For Frontend (BFF), utilizando bibliotecas Go como gqlgen para gerar o schema e os resolvers.
Os resolvers farão chamadas gRPC aos microsserviços internos para buscar e agregar dados, minimizando a complexidade do frontend.
4.5. Data Plane On-Premise (Proxy Agent)
O Agent em Go é um componente crítico para o Data Plane on-premise.

Executável Go: Compilado como um binário leve e independente.
Comunicação Segura: Usa gRPC com mTLS para se comunicar exclusivamente com o Serviço de Agentes no Control Plane.
Orquestrador Local: O Agent não é apenas um proxy, mas um micro-orquestrador. Ele pode baixar e executar contêineres Docker (ou executáveis Go) dos Serviço de Gerenciamento de Vaults e Serviço de Gerenciamento de Secrets localmente, no ambiente do cliente.
Worker Temporal: O Agent atuará como um Temporal Worker, consumindo Workflows e Atividades específicas para o Data Plane on-premise (e.g., rotação de secrets locais, atualizações OTA).
Conexão Local: Os microsserviços do Data Plane localmente orquestrados pelo Agent se conectarão a instâncias locais de MongoDB e RabbitMQ, utilizando as mesmas convenções de tenant_id para isolamento de dados e roteamento de mensagens.
Configuração/Credenciais: O Agent receberá configurações e credenciais de forma segura (e.g., via ambiente ou KMS local) para acessar bancos de dados e serviços de mensageria na rede do cliente.
Atualização OTA: O Agent terá um mecanismo de atualização over-the-air (OTA) orquestrado por um Temporal Workflow, garantindo a integridade da atualização via verificação de assinatura de código.
5. Gerenciamento de Dados (MongoDB e Redis)
5.1. MongoDB (NoSQL)
Modelagem de Dados: Os modelos (internal/core/[module_name]/models) incluirão o campo TenantID em todas as entidades relacionadas a tenant.
internal/infra/database/mongodb.go: Fornecerá um wrapper para o driver go.mongodb.org/mongo-driver. Este wrapper encapsulará a lógica de adicionar automaticamente o tenant_id às queries, como demonstrado na Seção 3.3.
Transações: Para operações que exigem atomicidade em múltiplos documentos ou coleções, serão utilizadas transações MongoDB (para replica sets e sharded clusters).
Backup/Restauração: Utilizar as capacidades de backup point-in-time e restauração gerenciadas por provedores de nuvem (MongoDB Atlas ou similar) ou ferramentas como mongodump/mongorestore e LVM snapshots para instalações self-managed.
5.2. Redis (Caching)
internal/infra/cache/redis.go: Abstração para o cliente Redis (github.com/go-redis/redis/v8).
Chaves de Cache: As chaves no Redis incluirão o Tenant ID para garantir o isolamento lógico.
Exemplo: cache:tenant:{tenant_id}:user:{user_id}:session ou cache:tenant:{tenant_id}:vault:{vault_id}:policy.
Estratégias de Invalidação: Serão usadas estratégias como "Cache-Aside" com TTL (Time-To-Live) apropriados. Para dados críticos, eventos RabbitMQ podem ser usados para disparar a invalidação de cache em todos os serviços interessados.
6. Mensageria (RabbitMQ) e Workflows (Temporal)
6.1. RabbitMQ
internal/infra/messaging/rabbitmq.go: Wrapper para o cliente RabbitMQ (github.com/rabbitmq/amqp091-go).
Exchanges e Queues:
Exchanges: direct, topic e fanout serão usados para diferentes propósitos (e.g., fanout para eventos de auditoria, topic para eventos específicos do tenant, direct para tarefas point-to-point).
Queues: Filas duráveis e com DLE (Dead Letter Exchange) para resiliência.
Publicação/Consumo:
Produtores: Publicarão mensagens com Tenant ID no corpo e, se necessário, na routing key.
Consumidores: Assinarão as filas e usarão o Tenant ID da mensagem para processar no contexto correto.
Mecanismo de Retentativa: Implementação de retentativas e DLQ (Dead Letter Queue) para mensagens que falham no processamento.
6.2. Temporal (Workflows)
O Temporal será essencial para orquestrar operações de longa duração, assíncronas e confiáveis.

Casos de Uso:
Rotação de Chaves: Workflows para RotateDEKWorkflow, RotateKEKWorkflow, RotateRootKeyWorkflow.
Geração/Renovação de Certificados: GenerateCertificateWorkflow, RenewCertificateWorkflow.
Provisionamento/Atualização de Agentes: ProvisionAgentWorkflow, UpdateAgentWorkflow.
Offboarding de Tenant: TenantOffboardingWorkflow (exclusão segura de dados e chaves).
Workflows e Atividades:
Workflows: Definem a lógica de orquestração e a sequência de etapas. Go routines serão usadas para implementar os Workflows.
Atividades: Funções Go que contêm a lógica de negócio real, executadas por Temporal Workers. Atividades interagem com MongoDB, RabbitMQ, Serviço de Chaves Mestras, etc.
Temporal Workers:
Implementados em Go, serão executados em clusters Kubernetes.
Os Agents (Data Plane On-premise) também podem ser configurados como Temporal Workers para executar atividades localmente.
Multi-tenancy com Temporal:
Workflow ID: O Workflow ID pode incluir o Tenant ID (ex: rotate-dek-{tenant_id}-{vault_id}).
Input do Workflow: Os inputs dos Workflows sempre incluirão o Tenant ID.
Atividades: As 
