package sqlitedb

import (
	"context"
	"database/sql"

	"github.com/devctllabs/go-libs/txmanager"
)

// Endpoint executes SQLite queries against a role-bound pool or the active transaction in ctx.
type Endpoint struct {
	pool        *sql.DB
	coordinator *txmanager.Coordinator[*sql.Tx]
	manager     txmanager.Manager
}

// ExecContext executes query using the active transaction when ctx carries one.
func (e *Endpoint) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tx, ok := e.coordinator.Current(ctx); ok {
		return tx.ExecContext(ctx, query, args...)
	}
	return e.pool.ExecContext(ctx, query, args...)
}

// QueryContext queries rows using the active transaction when ctx carries one.
func (e *Endpoint) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if tx, ok := e.coordinator.Current(ctx); ok {
		return tx.QueryContext(ctx, query, args...)
	}
	return e.pool.QueryContext(ctx, query, args...)
}

// QueryRowContext queries one row using the active transaction when ctx carries one.
func (e *Endpoint) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	if tx, ok := e.coordinator.Current(ctx); ok {
		return tx.QueryRowContext(ctx, query, args...)
	}
	return e.pool.QueryRowContext(ctx, query, args...)
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
	return e.pool.QueryRowContext(ctx, `SELECT 1`).Scan(&one)
}
