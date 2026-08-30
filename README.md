# go-libs

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![CI](https://github.com/devctllabs/go-libs/actions/workflows/ci.yml/badge.svg)](https://github.com/devctllabs/go-libs/actions/workflows/ci.yml)

`go-libs` is a monorepo of focused Go modules for building and operating services. Each
top-level module is versioned independently with tags named `<module>/vX.Y.Z`.

## Installation

Install only the module you need:

```sh
go get github.com/devctllabs/go-libs/<module>@latest
```

## Modules

### Application foundations

| Module | Description |
| --- | --- |
| [`config`](config) | Ordered, composable configuration loaders for defaults, files, dotenv data, and environment variables. |
| [`di`](di) | A small, type-safe dependency container with explicit resource ownership and shutdown. |
| [`lifecycle`](lifecycle) | Coordination for long-running tasks and graceful shutdown. |
| [`retry`](retry) | Explicit context-aware retry loops and capped exponential backoff. |

### Operations and observability

| Module | Description |
| --- | --- |
| [`buildinfo`](buildinfo) | Build metadata embedded in the current Go executable. |
| [`debugserver`](debugserver) | A standalone HTTP server for Go pprof endpoints. |
| [`log`](log) | Production JSON zap logger construction without global logger state. |
| [`telemetry`](telemetry) | Instance-owned OpenTelemetry trace and metric providers for Go services. |

### Health

| Module | Description |
| --- | --- |
| [`health`](health) | Transport-neutral liveness and readiness probes. |
| [`healthgrpc`](healthgrpc) | Standard gRPC Health service backed by transport-neutral probes. |
| [`healthotel`](healthotel) | OpenTelemetry metrics for health check observations. |
| [`healthserver`](healthserver) | OpenAPI-generated Echo endpoints for liveness and readiness probes. |
| [`healthzap`](healthzap) | Structured zap logging for health check failures and recoveries. |

### gRPC

| Module | Description |
| --- | --- |
| [`grpcclient`](grpcclient) | Explicit gRPC client connections with file-backed TLS, interceptor chains, and opt-in OpenTelemetry. |
| [`grpcserver`](grpcserver) | Application-owned gRPC runtime with validation, panic recovery, reflection, TLS, and graceful shutdown. |
| [`grpczap`](grpczap) | Completion and recovered-panic logging adapters for zap. |

### Messaging

| Module | Description |
| --- | --- |
| [`kafka`](kafka) | Typed franz-go producer and batching consumer runtimes with JSON codecs, retry, reject/DLQ policy, and OpenTelemetry. |
| [`kafkaproto`](kafkaproto) | Protobuf encoders and fresh-message decoders for Kafka values. |
| [`kafkazap`](kafkazap) | Structured zap logging for Kafka consumer retries and dispositions. |
| [`kafkaoutbox`](kafkaoutbox) | PostgreSQL transactional outbox with polling virtual shards or Debezium CDC delivery. |
| [`kafkaoutboxzap`](kafkaoutboxzap) | Structured zap logging for outbox retries, fencing, and topology changes. |

### Data and infrastructure

| Module | Description |
| --- | --- |
| [`filesystem`](filesystem) | Rooted filesystem operations that compose with the standard `io/fs` package. |
| [`postgresdb`](postgresdb) | Instrumented pgx reader and writer pools for PostgreSQL. |
| [`sqlitedb`](sqlitedb) | Instrumented SQLite reader and writer endpoints. |
| [`txmanager`](txmanager) | Shared transaction boundaries for services and database adapters. |

### Identity and API security

| Module | Description |
| --- | --- |
| [`oapivalidator`](oapivalidator) | OpenAPI request validation and authentication middleware for Echo. |
| [`oapivalidatorjwt`](oapivalidatorjwt) | JWT bearer and cookie authenticator for `oapivalidator`. |
| [`oidcsession`](oidcsession) | OIDC provider lifecycle, encrypted login state, and browser session HTTP flows. |
| [`oidcsessionredis`](oidcsessionredis) | Redis-backed opaque refresh sessions with encrypted provider tokens. |

### Codex integration

| Module | Description |
| --- | --- |
| [`codexapp`](codexapp) | A process-owning Go client for the Codex App Server JSONL JSON-RPC protocol. See its [package guide](codexapp/README.md). |

## Development

The repository pins its development tools with `mise`:

```sh
mise install
mise run lint
mise run test:race
mise run check-generated
```

Run the PostgreSQL integration suite separately when changing `postgresdb` or `txmanager`:

```sh
mise run postgresdb:test-integration
mise run oidcsessionredis:test-integration
mise run kafkaoutbox:test-integration
```

## Releases

Modules are released independently. Release notes are generated on demand and tags follow the
`<module>/vX.Y.Z` convention. See [RELEASING.md](RELEASING.md) for the release process and module
dependency order.

## License

Licensed under the [MIT License](LICENSE).
