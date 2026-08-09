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

### Operations and observability

| Module | Description |
| --- | --- |
| [`debugserver`](debugserver) | A standalone HTTP server for Go pprof endpoints. |
| [`log`](log) | Production JSON zap logger construction without global logger state. |
| [`telemetry`](telemetry) | Instance-owned OpenTelemetry trace and metric providers for Go services. |

### Health

| Module | Description |
| --- | --- |
| [`health`](health) | Transport-neutral liveness and readiness probes. |
| [`healthotel`](healthotel) | OpenTelemetry metrics for health check observations. |
| [`healthserver`](healthserver) | OpenAPI-generated Echo endpoints for liveness and readiness probes. |
| [`healthzap`](healthzap) | Structured zap logging for health check failures and recoveries. |

### Data and infrastructure

| Module | Description |
| --- | --- |
| [`filesystem`](filesystem) | Rooted filesystem operations that compose with the standard `io/fs` package. |
| [`oapivalidator`](oapivalidator) | OpenAPI request validation middleware for Echo. |
| [`postgresdb`](postgresdb) | Instrumented pgx reader and writer pools for PostgreSQL. |
| [`sqlitedb`](sqlitedb) | Instrumented SQLite reader and writer endpoints. |
| [`txmanager`](txmanager) | Shared transaction boundaries for services and database adapters. |

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
```

## Releases

Modules are released independently. Release notes are generated on demand and tags follow the
`<module>/vX.Y.Z` convention. See [RELEASING.md](RELEASING.md) for the release process and module
dependency order.

## License

Licensed under the [MIT License](LICENSE).
