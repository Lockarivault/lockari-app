# Arquitetura SaaS Multi-Tenant para Gestão de Secrets

Este documento descreve a arquitetura proposta para uma plataforma de gestão de **secrets, certificados e chaves OpenSSH**, desenhada como um Software as a Service (SaaS) multi-tenant. O objetivo é fornecer uma solução robusta, segura e escalável, com foco em conformidade (PCIDSS, HIPAA, GDPR, LGPD).

---

## 1. Visão Geral da Arquitetura

A arquitetura é dividida em três planos fundamentais para garantir a separação de responsabilidades e segurança:

*   **Control Plane**: Hospedado na nuvem. Gerencia UI, APIs, AuthN/AuthZ, Tenancy e Faturamento.
*   **Data Plane**: Nuvem ou On-premise. Armazena e processa dados sensíveis criptografados.
*   **Agent Plane**: Agentes instalados na infraestrutura do cliente para comunicação segura (mTLS) e integração local.

### Diagrama de Arquitetura

```mermaid
graph TD
   subgraph Cliente
       A[Navegador Web/SDK] -->|Acesso UI/API| B(Proxy de Borda/Load Balancer)
       C[Agent Plane On-premise] -->|mTLS| B
   end

   subgraph Control Plane (Nuvem)
       B --> D(API Gateway)
       D --> E(Serviço de Frontend)
       D --> F(Serviço de Usuários)
       D --> G(Serviço de Tenant)
       D --> H(Serviço de Faturamento)
       D --> I(Serviço de Notificação)
       D --> J(Serviço de Auditoria)
       D --> K(Autorização - OpenFGA)
       D --> L(Serviço de Agentes)
       D --> M(Gestão de Chaves - GCP Secret Manager)
       D --> N(Integração IDP/MFA)
       D --> O(Gestão de Vaults)
       D --> P(Gestão de Secrets)
       D --> Q(Rotação/Certificados)
   end

   subgraph Data Plane (Nuvem ou On-premise)
       P -.-> |Criptografia| R(DB Secrets)
       Q -.-> |Criptografia| S(DB Certificados)
       O -.-> |Metadados| T(DB Metadados)
   end

   subgraph Serviços Compartilhados
       U[IDP Google/Okta] <-- OIDC --> N
       V[Provedor MFA] <-- TOTP/U2F --> N
       W[Stripe] <-- API --> H
       X[Pub/Sub / RabbitMQ] <-- Eventos --> I
       Y[Redis Cache] <-- Cache --> D,E,F,G,O,P,Q
       Z[Log Suite] <-- Logs --> J,E,F,G,H,I,K,L,N,O,P,Q
       AA[Prometheus/Grafana] <-- Métricas --> Z
   end

   subgraph Infraestrutura
       BB[K8s Cluster] --> CC[Container Registry]
       DD[Object Storage]
       EE[Cloud Load Balancer]
   end

   subgraph On-premise Flow
       C -->|gRPC| DD_onprem(Data Plane Local)
       DD_onprem -->|Local Access| R_onprem(Local DB)
   end

   style B fill:#f9f,stroke:#333
   style C fill:#ccf,stroke:#333
   style M fill:#ffc,stroke:#333
   style K fill:#ddf,stroke:#333
```

---

## 2. Modelo de Multi-Tenancy

Adotamos o modelo **Shared Database with Tenant Identifier** para metadados e **Isolamento Criptográfico** para dados sensíveis.

> [!IMPORTANT]
> **Isolamento de Dados**: Embora o banco possa ser compartilhado, um Secret de um Tenant **nunca** pode ser aberto por outro, pois as chaves de criptografia são exclusivas por Tenant.

*   **Metadados**: (Usuários, Planos, Roles) possuem um `tenant_id` em cada registro.
*   **Dados Sensíveis**: Cada Vault possui sua **DEK** (Data Encryption Key) e cada Tenant possui sua **KEK** (Key Encryption Key).
*   **Chave Raiz (Root Key)**: Gerenciada pelo **GCP Secret Manager** ou HSM.

---

## 3. Segurança e Criptografia Hierárquica

A segurança baseia-se na hierarquia de chaves da NIST SP 800-57:

| Nível | Chave | Descrição | Armazenamento |
| :--- | :--- | :--- | :--- |
| **L1** | **Root Key** | Chave mestra que cifra as KEKs. | GCP Secret Manager / HSM |
| **L2** | **KEK** | Exclusiva por Tenant. Cifra as DEKs. | Control Plane (Protegida pela Root) |
| **L3** | **DEK** | Exclusiva por Vault. Cifra os Secrets. | Data Plane |

### Medidas Adicionais
*   **Trânsito**: TLS 1.3 obrigatório para clientes; mTLS para comunicação interna.
*   **Acesso**: Controle granular via **OpenFGA** (Zanzibar style).
*   **Hardening**: Imagens de containers escaneadas continuamente; privilégio mínimo.

---

## 4. Estratégias de Escalabilidade

*   **Microserviços**: Escalabilidade horizontal via Kubernetes (GKE).
*   **Bancos de Dados**: Cloud SQL (PostgreSQL) com réplicas para metadados; Firestore/DynamoDB para volume de secrets.
*   **Mensageria**: Cloud Pub/Sub ou RabbitMQ para tarefas assíncronas (ex: rotação).

---

## 5. Autenticação e Autorização (AuthN/AuthZ)

### Autenticação
*   **Standard**: OAuth2/OIDC (Google, Microsoft, Okta).
*   **Enterprise**: SSO via SAML/OIDC customizado.
*   **MFA**: Obrigatório via TOTP ou U2F (YubiKey).

### Autorização
*   **Fine-Grained**: OpenFGA gerencia permissões como "Usuário A pode ler Secret B no Vault C".
*   **Context Aware**: Validação baseada em IP, Horário e Tenant ID.

---

## 6. Observabilidade e Auditoria

O **Serviço de Auditoria** é central e registra:
1.  `Tenant ID` e `User ID`.
2.  `Timestamp` e `Ação` (ex: "Secret Read").
3.  `Recurso` afetado e `IP de origem`.
4.  `Resultado` (Sucesso/Falha).

> [!CAUTION]
> Os logs de auditoria são imutáveis. Qualquer tentativa de alteração deve disparar alertas críticos de segurança.

---

## 7. Planos e Funcionalidades

| Recurso | Free | Basic | Professional | Enterprise |
| :--- | :--- | :--- | :--- | :--- |
| **Usuários** | Até 4 | Até 20 | Até 100 | Ilimitado |
| **Secrets** | 100 | 1.000 | 10.000 | Ilimitado |
| **Vaults** | 1 | 5 | 50 | Ilimitado |
| **SLA** | N/A | 72h | 48h | 4h |
| **Suporte** | Comunidade | E-mail | Chat/E-mail | Dedicado |
| **Auth** | Social | Social/Email | Mesmos Domínios | SSO (SAML) |

---

## 8. Conformidade

A plataforma foi desenhada para facilitar auditorias:
*   **GDPR/LGPD**: Isolamento de PII e logs de acesso.
*   **HIPAA**: Criptografia de dados de saúde em repouso e trânsito.
*   **PCIDSS**: Auditoria imutável e rotação de chaves.

---
*Documento atualizado em: Fevereiro de 2026*
