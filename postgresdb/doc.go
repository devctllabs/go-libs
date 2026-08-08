// Package postgresdb owns instrumented pgx pools for PostgreSQL reader and
// writer endpoints.
//
// A reader DSN may point at a read-only replica; when omitted, the reader and
// writer endpoints share one physical pool. Endpoints route calls through native
// pgx transactions stored in context and are directly compatible with pgx
// scanners such as scany. The driver uses the extended query protocol without
// a prepared-statement cache so it remains safe behind PgBouncer transaction
// pooling.
package postgresdb
