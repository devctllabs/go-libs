// Package sqlitedb owns an instrumented SQLite database with separate logical
// reader and writer endpoints.
//
// File databases use WAL mode, a single writer connection, and a read-only
// reader pool. Named shared in-memory databases are supported for tests and
// ephemeral applications. Endpoints route calls through transactions stored in
// context and are directly compatible with database/sql scanners such as scany.
package sqlitedb
