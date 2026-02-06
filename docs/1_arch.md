Arquitetura SaaS Multi-Tenant para Plataforma de Gestão de Secrets
Este documento descreve a arquitetura proposta para uma plataforma de gestão de secrets, certificados e chaves OpenSSH, desenhada como um Software as a Service (SaaS) multi-tenant. O objetivo é fornecer uma solução robusta, segura e escalável que atenda às necessidades de empresas de diversos portes, com foco em conformidade e eficiência.

1. Visão Geral da Arquitetura
A arquitetura será dividida em três planos distintos, conforme solicitado:

Control Plane: Hospedado na nuvem, responsável pela interface do usuário (UI), APIs de gerenciamento, autenticação, autorização, gerenciamento de tenants e usuários, faturamento e coordenação geral.
Data Plane: Pode ser hospedado na nuvem ou on-premise, é onde os dados sensíveis (secrets, certificados, chaves) são armazenados, criptografados e processados.
Agent Plane: Composto por agentes que podem ser instalados na infraestrutura do cliente (on-premise ou em outras nuvens) para facilitar a comunicação segura, integração local e, em casos específicos, hospedar o Data Plane.
A arquitetura adota uma abordagem de microserviços para garantir flexibilidade, escalabilidade e resiliência.

'''mermaid graph TD subgraph Cliente A[Navegador Web/SDK] -->|Acesso UI/API| B(Proxy de Borda/Load Balancer); C[Agent Plane (On-premise)] -->|Comunicação Segura mTLS| B; end

subgraph Control Plane (Nuven)
   B --> D(API Gateway);
   D --> E(Serviço de Frontend);
   D --> F(Serviço de Usuários);
   D --> G(Serviço de Tenant);
   D --> H(Serviço de Faturamento);
   D --> I(Serviço de Notificação);
   D --> J(Serviço de Auditoria);
   D --> K(Serviço de Autorização - OpenFGA);
   D --> L(Serviço de Agentes);
   D --> M(Serviço de Gestão de Chaves Raiz - GCP Secret Manager);
   D --> N(Serviço de Integração de IDP/MFA);
   D --> O(Serviço de Gerenciamento de Vaults);
   D --> P(Serviço de Gerenciamento de Secrets);
   D --> Q(Serviço de Rotação/Certificados);
end

subgraph Data Plane (Nuven ou On-premise)
   P -.-> |Armazenamento de Secrets Criptografados| R(Banco de Dados de Secrets);
   Q -.-> |Armazenamento de Certificados/Chaves Criptografados| S(Banco de Dados de Certificados);
   O -.-> |Armazenamento de Metadados de Vaults| T(Banco de Dados de Metadados);
end

subgraph Serviços Compartilhados/Terceiros
   U[Provedor de Identidade (IDP)] <-- OAuth2/SAML/OIDC --> N;
   V[Provedor de MFA] <-- TOTP/U2F --> N;
   W[Stripe] <-- API --> H;
   X[Serviço de Mensageria] <-- Publicar/Consumir --> I;
   Y[Serviço de Caching] <-- Cache de Dados --> D, E, F, G, O, P, Q;
   Z[Serviço de Logging Centralizado] <-- Logs --> J, E, F, G, H, I, K, L, N, O, P, Q;
   AA[Serviço de Monitoramento/Alerta] <-- Métricas/Eventos --> Z, E, F, G, H, I, K, L, N, O, P, Q;
end

subgraph Componentes de Infraestrutura (Comum a Control/Data Plane na Nuvem)
   BB[Kubernetes Cluster] --> CC[Registros de Contêineres];
   DD[Storage de Objetos];
   EE[Load Balancers/APIGateways];
end

subgraph Fluxo de Comunicação On-premise
   C -->|Proxy HTTP/gRPC| DD_onprem(Data Plane On-premise Microservices);
   DD_onprem -->|Acesso Local| R_onprem(Banco de Dados de Secrets On-premise);
   DD_onprem -->|Acesso Local| S_onprem(Banco de Dados de Certificados On-premise);
   DD_onprem -->|Acesso Local| T_onprem(Banco de Dados de Metadados On-premise);
   DD_onprem -->|Mensageria Local| X_onprem(Serviço de Mensageria On-premise);
end

style B fill:#f9f,stroke:#333,stroke-width:2px;
style C fill:#ccf,stroke:#333,stroke-width:2px;
style DD_onprem fill:#cff,stroke:#333,stroke-width:2px;
style R_onprem fill:#cff,stroke:#333,stroke-width:2px;
style S_onprem fill:#cff,stroke:#333,stroke-width:2px;
style T_onprem fill:#cff,stroke:#333,stroke-width:2px;
style X_onprem fill:#cff,stroke:#333,stroke-width:2px;
style M fill:#ffc,stroke:#333,stroke-width:2px;
style U fill:#ffc,stroke:#333,stroke-width:2px;
style V fill:#ffc,stroke:#333,stroke-width:2px;
style W fill:#ffc,stroke:#333,stroke-width:2px;
style K fill:#ddf,stroke:#333,stroke-width:2px;
style J fill:#eed,stroke:#333,stroke-width:2px;
style Z fill:#eed,stroke:#333,stroke-width:2px;
style AA fill:#eed,stroke:#333,stroke-width:2px;
'''

2. Modelo de Multi-Tenancy
Será adotado um modelo de "Shared Database with Tenant Identifier" para a maioria dos metadados e dados não sensíveis, com uma forte camada de isolamento por criptografia e autorização para dados sensíveis.

Metadados (Tenants, Usuários, Roles, Planos): Armazenados em um banco de dados compartilhado com um identificador de tenant (Tenant ID) em cada registro. Isso simplifica o gerenciamento de tenants, usuários que pertencem a múltiplos tenants e a lógica de faturamento.
Dados Sensíveis (Secrets, Certificados, Chaves): Armazenados em bancos de dados (ou storages de objetos) que, embora possam ser logicamente compartilhados, a criptografia em repouso garante o isolamento físico/lógico. Cada vault terá sua própria Data Encryption Key (DEK), e cada tenant terá sua própria Key Encryption Key (KEK) para criptografar as DEKs dos vaults pertencentes a ele. A chave raiz (Root Key) para criptografar as KEKs será gerenciada por um serviço de gestão de secrets da nuvem (GCP Secret Manager), conforme as melhores práticas da NIST SP 800-57 Part 1 Revision 5.
Compartilhamento de Vaults: Quando um vault é compartilhado entre tenants, a DEK do vault continua sendo criptografada pela KEK do tenant proprietário. O acesso por usuários de outros tenants será mediado pelo Serviço de Gerenciamento de Vaults, que, após verificação de autorização (via OpenFGA), utiliza a KEK do tenant proprietário para descriptografar a DEK do vault e, em seguida, descriptografa o secret para o usuário autorizado. Isso mantém a propriedade da chave KEK com o tenant original, mas permite acesso controlado.
3. Isolamento de Dados e Medidas de Segurança
Criptografia Hierárquica:
Root Key: Armazenada no GCP Secret Manager, usada para criptografar as KEKs.
KEK (Key Encryption Key): Gerada e exclusiva para cada tenant, usada para criptografar as DEKs de todos os vaults pertencentes àquele tenant.
DEK (Data Encryption Key): Gerada e exclusiva para cada vault, usada para criptografar os secrets, certificados e chaves contidos no vault.
Criptografia de Dados: Todos os secrets, certificados e chaves são criptografados em repouso usando suas respectivas DEKs.
Segurança em Trânsito: Toda a comunicação entre o cliente (navegador, SDK, Agent) e a plataforma (Control Plane/Data Plane) será feita via TLS 1.3. A comunicação entre microserviços internos utilizará mTLS (mutual TLS).
Segregação Lógica: Através do Tenant ID em todos os registros de dados e políticas de autorização rigorosas (OpenFGA).
Firewalls e Redes Privadas: Utilização de Virtual Private Clouds (VPCs) com sub-redes privadas e firewalls configurados para permitir apenas o tráfego essencial entre os microserviços e com os serviços externos.
Hardening de Imagens/Contêineres: Uso de imagens base seguras e varredura de vulnerabilidades contínua.
Controle de Acesso Baseado em Função (RBAC): Implementado rigorosamente em todos os microserviços, com OpenFGA fornecendo o controle de acesso fino.
Segurança de Credenciais: Nenhuma credencial de acesso direto aos bancos de dados ou serviços será armazenada em código-fonte ou variáveis de ambiente de forma insegura. Uso de serviços de gestão de secrets da nuvem (por exemplo, GCP Secret Manager, Workload Identity) para o gerenciamento de credenciais internas.
Log de Auditoria: Detalhamento abaixo.
4. Estratégias de Escalabilidade
Arquitetura de Microserviços: Permite escalar serviços individualmente com base na demanda.
Orquestração de Contêineres: Kubernetes (GKE no GCP) será usado para orquestrar e gerenciar a vida útil dos microserviços, incluindo auto-scaling horizontal e vertical.
Bancos de Dados Escaláveis:
Banco de Dados de Metadados: Escolha de um banco de dados relacional gerenciado (por exemplo, Cloud SQL para PostgreSQL) com réplicas de leitura e sharding, se necessário.
Bancos de Dados de Secrets/Certificados: Soluções de armazenamento de dados distribuídas e escaláveis, como bases de dados NoSQL (ex: Firestore, DynamoDB) ou soluções de objetos (Cloud Storage, S3) com metadados em um banco relacional, para lidar com o grande volume de itens criptografados.
Load Balancers: Distribuem o tráfego de entrada para múltiplas instâncias de microserviços.
Filas de Mensagens: Para desacoplar componentes e processar tarefas assíncronas (ex: rotação de secrets, envio de e-mails) de forma escalável.
Redes de Entrega de Conteúdo (CDNs): Para cache de ativos estáticos do frontend e melhor desempenho global.
Caching Distribuído: Camadas de cache para reduzir a latência e a carga sobre os bancos de dados.
5. Mecanismos de Autenticação e Autorização
Autenticação (Control Plane):
OAuth2: Integração com provedores de identidade como Google, Microsoft, Okta para registro e login de usuários.
Registro Individual: Possibilidade de registro direto na plataforma com credenciais próprias (e-mail/senha).
SSO (SAML/OIDC): Para planos Enterprise, integração com o IDP corporativo do cliente.
MFA (Multi-Factor Authentication): Suporte a TOTP (ex: Google Authenticator) e U2F (ex: YubiKey) para todos os usuários, gerenciado pelo Serviço de Integração de IDP/MFA.
Fluxo de Convite: Para planos Basic e Professional, convites via e-mail e criação de senha na plataforma. Usuários devem pertencer ao mesmo domínio de e-mail.
Criação de Tenant: O primeiro usuário a se registrar cria o tenant e se torna o administrador inicial.
Autorização (Control Plane & Data Plane):
OpenFGA: Utilizado como o serviço central de autorização fine-grained. O Serviço de Autorização (OpenFGA) será consultado por outros microserviços para decidir se um usuário tem permissão para realizar uma determinada ação em um recurso (ex: user X can read secret Y in vault Z).
Roles Customizáveis: Administradores de tenant podem criar e gerenciar roles personalizadas, atribuindo permissões específicas a essas roles para seus usuários.
Contexto de Acesso: As verificações de autorização levarão em conta o Tenant ID, User ID, Vault ID e o tipo de recurso/ação.
6. Integração com Serviços de Terceiros
Provedores de Identidade (IDP): Google, Microsoft, Okta (OAuth2), qualquer IDP compatível com SAML/OIDC (para Enterprise).
Provedores de MFA: Serviços compatíveis com TOTP e U2F.
Stripe: Para processamento de pagamentos, gestão de assinaturas, trials, cupons e descontos. O Serviço de Faturamento no Control Plane se comunicará diretamente com a API do Stripe.
Serviços de E-mail: Para envio de convites, notificações e alertas (ex: SendGrid, Mailgun).
Serviços de Log Centralizado: Para agregação e análise de logs (ex: Google Cloud Logging/Operations Suite, ELK Stack).
Serviços de Monitoramento/Alerta: Para rastreamento de métricas e detecção de anomalias (ex: Google Cloud Monitoring/Operations Suite, Prometheus/Grafana).
Let's Encrypt: Para gestão e renovação automática de certificados SSL/TLS, via o Serviço de Rotação/Certificados.
7. Monitoramento e Logging Solutions
Logging Centralizado: Todos os microserviços enviarão seus logs estruturados para um sistema de logging centralizado (ex: Google Cloud Logging).
Logs de Auditoria (Serviço de Auditoria): Um microserviço dedicado (Serviço de Auditoria) registrará todas as ações críticas realizadas na plataforma, incluindo:
Tenant: ID do tenant.
User: ID do usuário que realizou a ação.
Data/Hora: Timestamp da ação.
Ação Realizada: Descrição da operação (ex: "Criou secret 'api-key-prod'", "Acessou vault 'financeiro'").
Recurso: ID do recurso afetado (secret, vault, user, role).
Endereço IP: Endereço IP de origem da requisição.
Resultado: Sucesso ou falha da operação.
Os logs de auditoria serão imutáveis e protegidos contra adulterações, críticos para conformidade (PCIDSS, HIPAA, GDPR, LGPD).
Dashboards e Alertas: Dashboards personalizados serão criados para visualizar logs, métricas e eventos, com alertas configurados para detectar comportamentos anômalos ou problemas de segurança/desempenho.
8. Métricas e Performance Tracking
Coleta de Métricas: Uso de agentes ou bibliotecas de instrumentação em cada microserviço para coletar métricas de desempenho (CPU, memória, latência de requisição, taxa de erro, taxa de sucesso, etc.).
Plataforma de Monitoramento: Integração com uma plataforma de monitoramento (ex: Google Cloud Monitoring, Prometheus com Grafana) para agregação, visualização e análise dessas métricas.
SLAs e SLOs: Definição de Service Level Agreements (SLAs) e Service Level Objectives (SLOs) para cada plano de serviço, com monitoramento proativo para garantir o cumprimento.
Tracing Distribuído: Implementação de tracing distribuído (ex: OpenTelemetry, Jaeger) para rastrear requisições através de múltiplos microserviços, auxiliando na depuração e otimização de desempenho.
9. Estratégias de Caching
Caching Distribuído: Uso de um serviço de cache em memória distribuído (ex: Redis ou Memcached gerenciado) para armazenar:
Dados de Metadados: Informações de tenant, usuário, roles frequentemente acessadas.
Tokens de Autenticação/Autorização: JWTs ou tokens de sessão.
Resultados de Consultas: Resultados de consultas complexas ou frequentes.
Políticas de Autorização: Políticas OpenFGA para reduzir a latência de verificação.
Cache de Borda (CDN): Para ativos estáticos do frontend (JavaScript, CSS, imagens).
Invalidacão de Cache: Estratégias de invalidação baseadas em tempo (TTL) ou eventos (cache-aside, write-through, write-back) para garantir a consistência dos dados.
10. Mensageria e Processamento Assíncrono
Fila de Mensagens: Um sistema de fila de mensagens robusto (ex: Google Cloud Pub/Sub, Kafka, RabbitMQ) será utilizado para:
Rotação Automática: Acionar a rotação de secrets, certificados e chaves.
Geração de Certificados: Processamento assíncrono para geração de certificados Let's Encrypt e self-signed.
Notificações: Envio de e-mails, alertas e outras notificações (via Serviço de Notificação).
Eventos de Auditoria: Publicação de eventos para o Serviço de Auditoria.
Processamento de Pagamentos: Eventos do Stripe para o Serviço de Faturamento.
Tarefas de Longa Duração: Desacoplamento de requisições que podem levar tempo para serem processadas.
Pub/Sub: Os microserviços publicarão eventos em tópicos, e outros microserviços (ou workers) se inscreverão nesses tópicos para processar as mensagens assincronamente. Isso aumenta a resiliência e a escalabilidade da plataforma.
11. Detalhes de Funcionalidades e Planos
Funcionalidades Específicas:
Vaults Compartilháveis: O Serviço de Gerenciamento de Vaults permitirá que administradores de tenant compartilhem vaults específicos com outros tenants ou usuários, com controle granular de permissões via OpenFGA.
Compartilhamento One-Touch: Para secrets, certificados ou chaves, um link temporário, de uso único, será gerado. Este link permitirá uma única visualização (sem download ou edição) e expirará automaticamente após o acesso ou um período pré-definido.
Gestão de Certificados:
Let's Encrypt: O Serviço de Rotação/Certificados se integrará com a API do Let's Encrypt para solicitar, validar e renovar automaticamente certificados.
Self-Signed: Geração de certificados autoassinados para ambientes internos.
Rotação Automática: O Serviço de Rotação/Certificados gerenciará a política de rotação automática de secrets, certificados e chaves OpenSSH, utilizando filas de mensagens para agendar e executar essas operações assincronamente.
SDKs e API: A API (gRPC e REST) será completamente documentada. SDKs serão gerados a partir de definições de serviço (ex: Protobuf para gRPC) para facilitar a integração em várias linguagens.
Proxy Agent (Data Plane On-premise):
O Proxy Agent será uma aplicação leve e segura instalada na infraestrutura do cliente.
Ele atuará como um gateway seguro, estabelecendo comunicação mTLS com o Control Plane na nuvem.
Para clientes que desejam o Data Plane on-premise, o Proxy Agent orquestrará instâncias dos microserviços de Data Plane (Serviço de Gerenciamento de Vaults e Secrets) localmente, que se conectarão a bancos de dados e serviços de mensageria na rede do cliente.
Isso garante que dados sensíveis nunca saiam do ambiente do cliente, enquanto a interface de gerenciamento (Control Plane) permanece na nuvem.
Planos de Serviços:
As regras de faturamento e limites serão impostas pelo Serviço de Faturamento em conjunto com o Serviço de Tenant e as políticas do OpenFGA.

Free: Plano individual.
Até 4 usuários para compartilhamento de vaults.
100 secrets, 1 vault.
Sem SLA, sem suporte.
Basic:
Limite de 20 usuários, 1.000 secrets, 5 vaults.
SLA de 72h, suporte via e-mail.
Usuários do mesmo domínio de e-mail.
Professional:
Limite de 100 usuários, 10.000 secrets, 50 vaults.
SLA de 48h, suporte via e-mail e chat.
Usuários do mesmo domínio de e-mail.
Enterprise:
Sem limites de usuários, secrets, vaults.
SLA de 4h, suporte via e-mail, chat e telefone, onboarding dedicado.
Autenticação via SSO (SAML, OIDC).
12. Conformidade e Segurança para Clientes Target
A arquitetura é projetada para atender às rigorosas exigências de conformidade de clientes PCIDSS, HIPAA, GDPR, LGPD, etc.:

Criptografia Robusta: KEK por tenant, DEK por vault e Root Key em HSM/serviço gerenciado.
Logs de Auditoria Imutáveis: Rastreabilidade completa de todas as ações.
Controle de Acesso Fino: OpenFGA garante que apenas usuários autorizados acessem recursos específicos.
Isolamento de Dados: Criptografia garante a separação lógica e confidencialidade.
MFA Obrigatório: Aumenta a segurança de autenticação.
Comunicação Segura: TLS 1.3 e mTLS em toda a plataforma.
Data Plane On-premise: Opção para clientes que precisam manter dados sensíveis dentro de seu perímetro de rede para conformidade.
Esta arquitetura fornece uma base sólida para a construção de uma plataforma SaaS de gestão de secrets segura, escalável e flexível.
