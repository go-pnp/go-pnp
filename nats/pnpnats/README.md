# NATS Module

This module provides a NATS clients to fx di container.

## Configuration

### Custom configuration

To use custom configuration, you can use `pnpnats.WithConfigFromContainer()` option when using `pnpnats.Module`, and
then provide your custom configuration to the di container by `fx.Supply(*pnpnats.Config{...})`
or `fx.Provide(func() *pnpnats.Config {...})`.

### Config from env variables

| Env variable                       | Default Value | Required       | Comment                                              |
|------------------------------------|---------------|----------------|------------------------------------------------------|
| NATS_ADDR                          | 127.0.0.1:443 | No             | Nats server address                                  |
| NATS_TLS_ENABLED                   | false         | No             | Use TLS to communicate with server                   |
| NATS_TLS_CERT_PATH                 |               | if TLS_ENABLED | Path to client certificate                           |
| NATS_TLS_KEY_PATH                  |               | if TLS_ENABLED | Path to client key                                   |
| NATS_TLS_INSECURE_SKIP_VERIFY      | false         | if TLS_ENABLED | Do not verify server certificate                     |
| NATS_TLS_ROOT_CA_PATH              |               | if TLS_ENABLED | Path to root CA Certificate                          |
| NATS_TLS_APPEND_SYSTEM_CAS_TO_ROOT | false         | if TLS_ENABLED | Append system CA Certificates to root CA Certificate |
| NATS_RECONNECTS_MAX_COUNT          | -1            | No             | Reconnects count (-1 for infinity)                   |
| NATS_RECONNECTS_ALLOW              | true          | No             | Allow reconnects                                     |
| NATS_RECONNECTS_WAIT               | 500ms         | No             | Time until next reconnect                            |

### Customizing nats connection options

If you need to pass some custom client options, you can provide it to fx DI container like this:

```go
fx.Provide(pnpnats.ClientOptionProvider(func () nats.Option {
return nats.MaxReconnects(10)
}))
```

### Customizing jetstream options

If you need to pass some custom jetstream options, you can provide it to fx DI container like this:

```go
fx.Provide(pnpnats.JetstreamOptionProvider(func () nats.JSOpt {
return nats.MaxWait(time.Second)
}))
```
