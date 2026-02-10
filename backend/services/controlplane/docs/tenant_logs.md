# Arquitetura de Audit Logs (Trilhas de Auditoria)

Este documento descreve a estrutura e a lógica de persistência dos Logs de Auditoria do Lockari, diferenciando-os dos logs técnicos de aplicação.

## O Padrão AATO (Actor, Action, Target, Outcome)

Para conformidade (Compliance) e segurança, cada evento de auditoria deve responder a quatro perguntas fundamentais:

| Componente | Descrição | Exemplo |
| :--- | :--- | :--- |
| **Actor** (Quem) | Identificador do sujeito que executou a ação. | `user_uuid`, `system_service`, `api_key_id` |
| **Action** (O quê) | A operação realizada no sistema. | `TENANT_CREATE`, `SECRET_DECRYPT`, `POLICY_UPDATE` |
| **Target** (Alvo) | O recurso que sofreu a ação. | `tenant_id`, `secret_uuid`, `key_id` |
| **Outcome** (Resultado)| Se a operação foi bem sucedida ou falhou. | `SUCCESS`, `FAILED (Unauthorized)` |

---

## Estrutura do Objeto (Schema)

```json
{
  "timestamp": "2024-02-09T12:00:00Z",
  "actor": {
    "id": "uuid-v4",
    "type": "USER|SYSTEM|AGENT",
    "ip": "192.168.1.1",
    "user_agent": "Mozilla/5.0..."
  },
  "action": "TENANT_PROVISION_SUCCESS",
  "target": {
    "id": "tenant-uuid-v4",
    "type": "TENANT",
    "name": "Acme Corp"
  },
  "context": {
    "trace_id": "otel-trace-id",
    "parent_action_id": "uuid-referencia"
  },
  "metadata": {
    "reason": "Initial provisioning",
    "region": "us-east-1"
  }
}
```

---

## Modelos de Implementação

### 1. SaaS (Cloud Native)
No modelo 100% SaaS, o Control Plane é responsável por gerar e persistir o log imediatamente na collection central de auditoria (`audit_logs`) após a execução de qualquer usecase crítico.

### 2. Agente On-Premise (Híbrido)
Para operações que ocorrem dentro da infraestrutura do cliente (Data Plane):

1.  **Geração Local**: O Agente gera a trilha de auditoria localmente (mesmo offline).
2.  **Relay (Telemetria)**: Assim que houver conexão, o Agente envia o log via gRPC/mTLS para o Control Plane.
3.  **Visualização Única**: O Control Plane centraliza esses logs para que o cliente veja no Dashboard SaaS todas as ações, independentemente de onde ocorreram.

---

## Segurança e Imutabilidade

*   **Append-Only**: A collection de auditoria não deve permitir `Update` ou `Delete`.
*   **Integridade**: (Futuro) Logs podem ser assinados digitalmente para garantir que não foram alterados no banco de dados.
*   **Retenção**: Definida por política (ex: 365 dias para conformidade SOC2).
