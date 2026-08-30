package postgresdb

import (
	"context"

	"github.com/devctllabs/go-libs/txmanager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Endpoint executes PostgreSQL queries against a role-bound pool or the active transaction in ctx.
type Endpoint struct {
	pool        *pgxpool.Pool
	coordinator *txmanager.Coordinator[pgx.Tx]
	manager     txmanager.Manager
}

// InTransaction reports whether ctx carries a transaction for this database.
// It does not change the autocommit behavior of Endpoint operations.
func (e *Endpoint) InTransaction(ctx context.Context) bool {
	_, ok := e.coordinator.Current(ctx)
	return ok
}

// Exec executes query using the active transaction when ctx carries one.
func (e *Endpoint) Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if tx, ok := e.coordinator.Current(ctx); ok {
		return tx.Exec(ctx, query, args...)
	}
	return e.pool.Exec(ctx, query, args...)
}

// Query queries rows using the active transaction when ctx carries one.
func (e *Endpoint) Query(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
	if tx, ok := e.coordinator.Current(ctx); ok {
		return tx.Query(ctx, query, args...)
	}
	return e.pool.Query(ctx, query, args...)
}

// QueryRow queries one row using the active transaction when ctx carries one.
func (e *Endpoint) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	if tx, ok := e.coordinator.Current(ctx); ok {
		return tx.QueryRow(ctx, query, args...)
	}
	return e.pool.QueryRow(ctx, query, args...)
}

// WithinTx executes fn inside a transaction bound to this endpoint's role.
func (e *Endpoint) WithinTx(
	ctx context.Context,
	fn func(ctx context.Context) error,
	options ...txmanager.Option,
) error {
	return e.manager.WithinTx(ctx, fn, options...)
}

// Check verifies that this endpoint can execute a bounded query.
func (e *Endpoint) Check(ctx context.Context) error {
	var one int
	return e.pool.QueryRow(ctx, `SELECT 1`).Scan(&one)
}
