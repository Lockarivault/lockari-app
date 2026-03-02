# AuditLog Library - Technical Specification

Esta biblioteca foi projetada para fornecer uma solução de auditoria robusta, resiliente e de alta performance para o ecossistema Lockarivault. O objetivo é registrar todas as ações críticas do sistema sem impactar a experiência do usuário final.

---

## 🎯 Objetivo
Criar uma trilha de auditoria imutável (quem fez o quê, quando e onde). A biblioteca deve ser capaz de lidar com volumes variados de dados, garantindo que o serviço principal (API) não seja bloqueado por operações de escrita em banco de dados ou mensageria.

---

## 🏗️ Conceitos Fundamentais (Para o Desenvolvedor)

Para garantir a performance, esta lib utiliza três pilares do Go:

### 1. Channels (Canais)
Imagine o canal como uma esteira de fábrica. Quando a aplicação gera um log, ela coloca o "pacote" na esteira e volta a trabalhar imediatamente. Ela não espera o pacote chegar ao fim da linha.
- **Vantagem:** Operação não-bloqueante. A API responde ao usuário em microssegundos.

### 2. Worker Pool (Piscina de Trabalhadores)
Teremos um número fixo de "trabalhadores" (goroutines) parados ao lado da esteira. Assim que um log aparece, um trabalhador disponível o pega e o leva para o destino final (Banco de Dados ou Mensageria).
- **Vantagem:** Controle de recursos. Mesmo que tragam milhares de logs em um segundo, teremos apenas X trabalhadores processando-os, evitando que o servidor fique sem memória.

### 3. Providers (Abstração)
A lib não deve saber se está salvando no MongoDB, Postgres ou enviando para o RabbitMQ. Ela deve usar uma **Interface**.
- **Vantagem:** Flexibilidade. Podemos começar salvando direto no banco de dados e, no futuro, mudar para mensageria apenas trocando o "provedor", sem mexer em nenhuma linha do código de auditoria.

---

## 📋 Estrutura de Dados (AuditEntry)

Para que a biblioteca seja consistente, o desenvolvedor deve utilizar a seguinte estrutura (ou similar) em Go:

```go
type AuditEntry struct {
    TenantID     string         `json:"tenant_id"`     // ID do cliente (SaaS)
    UserID       string         `json:"user_id"`       // ID de quem fez a ação
    ResourceType string         `json:"resource_type"` // ex: "secret"
    ResourceID   string         `json:"resource_id"`   // ID do recurso
    Action       string         `json:"action"`        // ex: "read", "update"
    Timestamp    time.Time      `json:"timestamp"`     // Gerado automaticamente se vazio
    IPAddress    string         `json:"ip_address"`    // IP do solicitante
    UserAgent    string         `json:"user_agent"`    // Browser/Client info
    Metadata     map[string]any `json:"metadata"`      // Dados extras
}
```

---

## 🛠️ Detalhamento: CreateAuditLog(ctx, entry)

Este é o método principal que a aplicação irá chamar. É fundamental entender que **ele não salva no banco de dados imediatamente**.

### Assinatura
```go
func (a *Audit) CreateAuditLog(ctx context.Context, entry AuditEntry) error
```

### O que acontece dentro deste método? (Fluxo Interno)

1.  **Validação:** A lib verifica se os campos obrigatórios (`TenantID`, `Action`, `ResourceType`) foram preenchidos. Se faltar algo crítico, ele retorna um **erro imediato** para a aplicação.
2.  **Enriquecimento:** Se o campo `Timestamp` estiver vazio, a própria lib preenche com `time.Now().UTC()`.
3.  **Enfileiramento (The Magic Part):** A lib tenta colocar o `entry` dentro do **Channel (canal)** interno.
    *   **Se o canal tiver espaço:** O log entra na fila e o método retorna `nil` (sucesso) instantaneamente.
    *   **Se o canal estiver cheio:** Dependendo da configuração, ele pode retornar um erro de "Buffer Full" ou aguardar um curto período.
4.  **Retorno:** O controle volta para a regra de negócio da aplicação.

### ❓ Ele conecta com o Banco de Dados?
**NÃO.** O método `CreateAuditLog` nunca abre conexões com o banco ou mensageria.
- Quem faz a conexão e o `Insert`/`Publish` são os **Workers** que rodam em background.
- Isso garante que, mesmo que o banco de dados esteja com lentidão de 10 segundos, o método `CreateAuditLog` continuará respondendo em menos de 1 milisegundo.

---

## 🚀 Fluxo de Implementação Passo a Passo

Para desenvolver esta lib, siga esta ordem:

### Passo 1: Definir a Interface do Provider
Crie uma interface `Store` que tenha os métodos de salvamento e busca. Isso permite que a lib seja "plugável".

### Passo 2: O Buffer (Canal Interno)
Ao inicializar a lib, crie um canal `chan AuditEntry` com um tamanho configurável (ex: buffer de 1000).

### Passo 3: Inicializar o Worker Pool
No `Init` da biblioteca, use um laço `for` para disparar X goroutines (workers).
- Cada worker deve rodar um `for range` no canal de logs.
- Dentro do worker, chame o método `Store.Save()` do provider.

### Passo 4: Implementar o Graceful Shutdown
Este é o ponto mais crítico para um pleno/sênior. Quando a aplicação for desligada:
1.  A lib deve parar de aceitar novos logs.
2.  Ela deve **fechar o canal**.
3.  Ela deve esperar que os workers processem todos os logs que ainda estavam na "esteira" antes de encerrar o processo.
*Use `sync.WaitGroup` para controlar isso.*

---

## 💡 Por que fazer assim? (Resumo para Juniores)

Se tentarmos salvar no banco de dados sincronamente (direto na thread da API), e o banco estiver lento, o usuário ficará esperando a tela carregar por vários segundos.

Com **Goroutines + Channels**:
- O usuário recebe "OK" instantaneamente.
- O trabalho pesado de escrita é feito em "segundo plano".
- Se o banco de dados cair temporariamente, os logs ficam guardados no canal (buffer) esperando a volta do serviço, aumentando a resiliência do sistema.

---

## 🛠️ Tecnologias Envolvidas
- **Go 1.21+**
- **Concurrency patterns** (Worker Pools, Channels, WaitGroups)
- **JSON Serialization** para o campo `metadata`.
