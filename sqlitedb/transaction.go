package sqlitedb

import (
	"context"
	"database/sql"

	"github.com/devctllabs/go-libs/txmanager"
)

type transactionBackend struct {
	reader *sql.DB
	writer *sql.DB
}

var _ txmanager.Backend[*sql.Tx] = (*transactionBackend)(nil)

func (b *transactionBackend) Begin(ctx context.Context, spec txmanager.BeginSpec) (*sql.Tx, error) {
	pool := b.writer
	readOnly := false
	if spec.Role == txmanager.RoleReader {
		pool = b.reader
		readOnly = true
	}
	return pool.BeginTx(ctx, &sql.TxOptions{ReadOnly: readOnly})
}

func (*transactionBackend) Commit(_ context.Context, tx *sql.Tx) error {
	return tx.Commit()
}

func (*transactionBackend) Rollback(_ context.Context, tx *sql.Tx) error {
	return tx.Rollback()
}
