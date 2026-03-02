# RabbitMQ Messaging Configuration Guide

The `mensageria` library uses a declarative approach to manage RabbitMQ infrastructure. You define your exchanges, queues, and bindings in your application's `config.yaml`, and the library ensures they exist upon connection.

## Configuration Structure

The configuration is expected under the `messagequeues` key:

```yaml
messagequeues:
  addr: "localhost"           # RabbitMQ server address
  port: "5672"                # RabbitMQ server port (default: 5672, TLS: 5671)
  user: "guest"               # RabbitMQ username
  password: "guest"           # RabbitMQ password
  tls: true                   # Set to true to use amqps:// (preferred)
  config:
    exchanges:
      - name: "my.exchange"
        kind: "topic"         # direct, fanout, topic, headers
        durable: true         # survives server restarts
        auto_deleted: false   # deleted when last queue unbinds
        internal: false       # if true, clients can't publish to it
        no_wait: false
        args: {}              # extra arguments (e.g., x-delayed-type)

    queues:
      - name: "my.queue"
        durable: true
        auto_delete: false
        exclusive: false
        no_wait: false
        args: {}
        dead_letter:          # optional DLQ configuration
          exchange: "my.dlx"
          key: "my.queue.dlq"
        prefetch_count: 10    # QoS prefetch for consumers
        max_delivery_retries: 3

    bindings:
      - queue: "my.queue"
        exchange: "my.exchange"
        routing_key: "event.created"
        no_wait: false
        args: {}
```

## Best Practices

1. **Use TLS**: In production environments, always set `tls: true` and use port `5671`.
2. **Durable Entities**: Set `durable: true` for both exchanges and queues to prevent data loss on RabbitMQ restarts.
3. **Dead Letter Queues**: Always configure `dead_letter` for queues processing critical data to handle message failures gracefully.
4. **Topic Exchanges**: Use `topic` exchanges for flexibility in routing events based on dot-separated keys (e.g., `tenant.created`, `tenant.updated`).
