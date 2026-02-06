Estou desenvolvendo uma plataforma para gestão de secrets, certificates, opensshkey, etc. a proposta é que seja uma plataforma Software As a Service multi tenant.


Fluxo de excução:
- Control Plane: Plataforma na nuvem, onde o usuário irá acessar a interface web, API, etc.
- Data Plane: Plataforma na nuvem ou onpremise, onde os dados serão armazenados e processados.
- Agent Plane: Agentes que poderão ser instalados na infraestrutura do cliente, para facilitar a comunicação e integração com a plataforma.

- Sendo o objetivo principal, ter o frontend como um microserviço, a exposição de APIs e recursos como outro microserviço, e o backend (data plane) como outro microserviço, podendo ser na nuvem ou onpremise.

Funcionalidades:
1. O usuário poderá pertencer a um ou any tenants;
2. Cada tenant poderá ter any vaults.  Porém cada vault pode estar vinculado a um único tenant
3. Cada vault pode ter any secrets, certificates, opensshkey, etc.
4. Cada tenant terá um ou any admins, que poderão gerenciar os users, roles, vaults, etc.
5. Cada tenant poderá ter any users, que poderão ter permissões diferentes, dependendo das roles atribuídas a ele.
6. Apenas os **VAULTS** poderão ser compartilhados entre outros tenants/usuários, nunca os tenants.
7. Uma secret, certificate, opensshkey, etc. poderá ser compartilhada com usuários externos no modelo onetouch, ou seja, apenas uma visualização, sem permissão de download ou edição. e depois automaticamente é expirada.
8. Haverá integração com provedores de identidade via SSO ( SAML, OIDC) para autenticação dos usuários.
9. Haverá integração com provedores de MFA (TOTP, U2F, etc.) para aumentar a segurança na autenticação dos usuários.
10. Haverá logs de auditoria para todas as ações realizadas na plataforma, para garantir a rastreabilidade e segurança.  Os logs serão organizados por tenant, user, data/hora, ação realizada, etc.
11. Haverá API para integração com outras ferramentas e automações.
12. Haverá SDKs para facilitar a integração com outras linguagens de programação, usndo o gRPC como base.
13. Haverá proxy agent , caso o cliente queira que toda a comunicação e dados sejam trafegados e armazenados na infraestrutura do cliente, ou seja, onpremise.  Apenas a interface web e API estarão na nuvem, mas toda a comunicação com o banco de dados, serviço de mensageria, etc. estarão na infraestrutura do cliente.
14. Gestão de certificados do Let's Encrypt, com renovação automática.
15.Geração de certificados self-signed.
16. Rotação automática de secrets, certificates, opensshkey, etc.
17. Usar o novo padrão de gestão de certificados e secrets da NIST (NIST SP 800-57 Part 1 Revision 5).

Pagamentos:
1. Integração com Stripe para gestão de pagamentos, assinaturas, etc.
2. Importante ter uma gestão de trials, cupons, descontos, etc.
3. Importante ter uma estrutura referente a tenat e pagamento, de forma descentralizada, pois como podemos ter proxy agent, as infromações, precisam ficar no control plane

CLente target:
Empresas de pequeno, médio e grande porte, que precisam gerenciar secrets, certificates, opensshkey, etc. de forma segura e eficiente.
- Clientes com modelo PCIDSS, HIPAA, GDPR, LGPD, etc. que precisam garantir a segurança e conformidade na gestão de secrets.
- Equipes de desenvolvimento, operações, segurança da informação, etc. que precisam de uma solução centralizada e fácil de usar para gerenciar secrets.
- Clientes com dificuldade de gerenciamento de certificados expiration, rotação de secrets, etc.

Ferramentas concorrentes:
- Hashicorp Vault
- AWS Secrets Manager
- Azure Key Vault
- Google Secret Manager
- 1Password Business
- LastPass Enterprise
- Akeyless

Teremos 4 tipos de planos de serviços, que são:
- 1. Free: Plano individual, onde pode compartilhar os vaults com até 4 users, 100 secrets, 1 vault, sem SLA, sem suporte.
- 2. Basic: com limite de 20 users, 1.000 secrets, 5 vaults, SLA de 72h, suporte via email.
- 3. Professional: com limite de 100 users, 10.000 secrets, 50 vaults, SLA de 48h, suporte via email e chat.
- 4. Enterprise: sem limites, SLA de 4h, suporte via email, chat e telefone, onboarding dedicado. Authenticação via SSO ( SAML, OIDC).

Permissionamentos:
A proposta é que toda a gestão de permissionamento, seja realizada através de roles com o OpenFGA (OpenFGA: Fine-Grained Authorization as a Service), onde o tenant admin poderá criar roles customizadas, e atribuir essas roles para os users do tenant.

Segurança dos tenants:
1. Cada tenant terá sua própria KEK (key encryption key) para criptografar as DEK (data encryption key) de cada vault.
2. Cada vault terá sua própria DEK (data encryption key) para criptografar os secrets, certificates, opensshkey, etc.
3. A root key, que será usada para criptografar as KEK, ficará armazenada no serviço de gestão de secrets da GCP (Google Cloud Secret Manager).
4. Toda a comunicação entre o cliente e a plataforma será feita via TLS 1.3.
5. Haverá integração com provedores de MFA (TOTP, U2F, etc.) para aumentar a segurança na autenticação dos usuários.
6. Haverá logs de auditoria para todas as ações realizadas na plataforma, para garantir a rastreabilidade e segurança.


Plataforma:
1. Autheticação é via OUTH2 com Google, Microsoft, Okta, etc.
2. Poderá registrar de forma individual, sem ser autenticação através do provider;
3. Para planos enterprise, poderá ser integrado com SSO via SAML ou OIDC;
4. Planos básicos, os convites devem ser enviados via email, e o usuário criará sua senha na plataforma;
5. Para planos Basic e Professional, os usuáreios precisaram fazer parte do mesmo domínio de email;
6. O usuário ao criar a conta, irá criar o tenant, e já será o admin do tenant;
7. O admin do tenant, poderá convidar outros usuários para o tenant, e atribuir roles para esses usuários;



Sempre siga as melhores práticas de mercado, podendo usar as soluções do Google como base ou inspiração.


