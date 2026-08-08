package postgresdb

import (
	"context"

	"github.com/devctllabs/go-libs/txmanager"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type transactionBackend struct {
	reader *pgxpool.Pool
	writer *pgxpool.Pool
}

var _ txmanager.Backend[pgx.Tx] = (*transactionBackend)(nil)

func (b *transactionBackend) Begin(ctx context.Context, spec txmanager.BeginSpec) (pgx.Tx, error) {
	pool := b.writer
	accessMode := pgx.ReadWrite
	if spec.Role == txmanager.RoleReader {
		pool = b.reader
		accessMode = pgx.ReadOnly
	}
	return pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgxIsolation(spec.Isolation),
		AccessMode: accessMode,
	})
}

func (*transactionBackend) Commit(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

func (*transactionBackend) Rollback(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}

func pgxIsolation(isolation *txmanager.Isolation) pgx.TxIsoLevel {
	if isolation == nil {
		return ""
	}
	switch *isolation {
	case txmanager.ReadCommitted:
		return pgx.ReadCommitted
	case txmanager.RepeatableRead:
		return pgx.RepeatableRead
	case txmanager.Serializable:
		return pgx.Serializable
	default:
		return ""
	}
}
