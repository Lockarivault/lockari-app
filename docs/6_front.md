# Estratégia de Implementação Frontend: Gestão de Secrets SaaS

Este documento descreve a estratégia de implementação para o frontend da Plataforma de Gestão de Secrets SaaS Multi-Tenant, utilizando **React** com **TypeScript** e **Tailwind CSS**. A arquitetura visa fornecer uma experiência de usuário robusta, segura, performática e acessível.

O frontend será uma Single Page Application (SPA), comunicando-se com o Control Plane (API Gateway) primariamente através de APIs REST e GraphQL.

---

## 1. Arquitetura de Componentes

Nossa abordagem de componentes se inspira nos princípios do Google Material Design/Material You para usabilidade e estética, mantendo a flexibilidade do Tailwind CSS.

### 1.1. Roteamento
Utilizaremos **React Router DOM** para gerenciar as rotas da aplicação.

*   **Rotas Protegidas**: Componentes de `ProtectedRoute` verificarão o status de autenticação e permissões. Usuários não autenticados serão redirecionados para o login.
*   **Rotas Públicas**: Páginas como login, registro e recuperação de senha.
*   **Organização**: Rotas definidas em um arquivo central (`src/app/router.tsx`) e agrupadas por área.

### 1.2. Layout Shells
Serão definidos layouts base para garantir consistência visual:

*   **AuthLayout**: Para páginas de autenticação, com foco na marca e no formulário.
*   **DashboardLayout**: Layout principal para usuários autenticados, incluindo Sidebar, Header e Content Area.
*   **TenantSetupLayout**: Para o fluxo de onboarding inicial do tenant.

### 1.3. Design Tokens e Theming
A estratégia de theming permitirá a personalização da marca por tenant.

*   **tailwind.config.js**: Ponto central para tokens de design base (cores, tipografia, espaçamento).
*   **Cores Dinâmicas**: Definiremos variáveis CSS (e.g., `--primary-color`) para personalização por tenant.
*   **Injeção de Tema**: O backend fornecerá os valores das variáveis CSS que serão injetadas dinamicamente no Shell.
*   **Classes Tailwind**: Uso de classes que referem a essas variáveis (e.g., `bg-primary`, `text-secondary`).

### 1.4. Módulos Atômicos e Funcionais
*   **Módulos Atômicos**: Componentes básicos reutilizáveis (Button, Input, Checkbox).
*   **Módulos de Funcionalidade (Features)**: Agrupam componentes, hooks e serviços relacionados a um domínio (e.g., `features/secrets`).
*   **Primitivos UI (Headless)**: Uso de bibliotecas como **Headless UI** ou **Radix UI** para componentes complexos (Modal, Dropdown).

---

## 2. Fluxo de Estado e Dados

### 2.1. Estado do Cliente
*   **Global**: Uso da **React Context API** para estados que afetam múltiplos níveis (tema, auth).
*   **Local**: `useState` e `useReducer` para estados internos.
*   **Formulários**: **React Hook Form** para gestão de validação e submissão.

### 2.2. Estado do Servidor
*   **React Query (TanStack Query)**: Biblioteca principal para fetching, caching e sincronização.
    *   `useQuery` para leituras (GET).
    *   `useMutation` para escritas com suporte a atualizações otimistas.
*   **Service Layer**: Cada feature terá um serviço que encapsula chamadas via **Axios**.
*   **Propagação Multi-Tenant**: O `Tenant ID` é obtido do JWT e incluído em todos os headers de autorização para o API Gateway.

---

## 3. Gerenciamento de Sessão

*   **Access Token (JWT)**: Armazenado em memória e incluído no header `Authorization: Bearer {token}`. Vida útil curta.
*   **Refresh Token**: Enviado pelo backend via **Cookie HttpOnly, Secure, SameSite=Strict**.
*   **Ciclo de Vida**: Interceptor Axios para detectar `401 Unauthorized` e realizar o refresh automático.
*   **Logout**: Endpoint `/logout` invalida a sessão no backend e limpa o estado local.

---

## 4. Segurança e Privacidade

*   **CSP (Content Security Policy)**: Configurado no servidor de estáticos para mitigar XSS.
*   **Sanitização**: Uso de **DOMPurify** para qualquer conteúdo gerado pelo usuário.
*   **RBAC Guards**: Hook `usePermissions` ou componente `<Can />` para controle de visibilidade baseado em roles.
*   **Telemetria**: Mascaramento automático de PII (Informações Pessoais) antes do envio para o coletor.

---

## 5. Performance e Acessibilidade

*   **Core Web Vitals**: Foco em Lighthouse scores e mitigação de CLS (Cumulative Layout Shift).
*   **Code-Splitting**: Uso de `React.lazy` e `Suspense` para divisão do bundle por rota.
*   **Acessibilidade**: Linting com `eslint-plugin-jsx-a11y` e testes com **Axe-core**.
*   **Internacionalização (i18n)**: Uso de `react-i18next` com carregamento dinâmico de JSONs.

---

## 6. Observabilidade e Qualidade

*   **OpenTelemetry JS SDK**: Captura de traces, métricas e logs estruturados.
*   **Feature Flags**: Controle de funcionalidades via API do backend.
*   **Testes**:
    *   **Unitários**: Vitest + React Testing Library.
    *   **Visuais**: Storybook + Chromatic.
    *   **E2E**: Playwright ou Cypress.
*   **CI/CD**: Pipeline no GitHub Actions para validação de cada PR.

---

## 7. Estrutura de Pastas

```text
src/
├── app/                      # Nível superior (App, Router, Layouts, Contexts)
├── features/                 # Módulos por domínio (authentication, secrets, vaults)
│   └── [feature_name]/
│       ├── components/
│       ├── hooks/
│       ├── pages/
│       ├── services/
│       └── types/
├── shared/                   # Código reutilizável (Atoms, Utils, Types)
│   ├── components/           # UI Components (Button, Modal, etc)
│   ├── hooks/                # Hooks utilitários
│   ├── styles/               # Tailwind config, themes
│   └── api/                  # Axios instance, interceptors
├── assets/                   # Imagens, ícones, fontes
└── tests/                    # E2E e utilitários de teste
```

---
*Estratégia Frontend - v1.0 - Fevereiro de 2026*
