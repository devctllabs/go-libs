package sqlitedb_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/devctllabs/go-libs/sqlitedb"
	"github.com/devctllabs/go-libs/txmanager"
	"github.com/stretchr/testify/require"
)

func TestOpenRejectsBlankDSN(t *testing.T) {
	t.Parallel()
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{})

	require.Nil(t, db)
	require.Error(t, err)
}

func TestOpenFileConfiguresWriterAndStrictReader(t *testing.T) {
	t.Parallel()
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{
		DSN: filepath.Join(t.TempDir(), "app.sqlite"),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Writer().ExecContext(context.Background(), `CREATE TABLE messages (id INTEGER PRIMARY KEY)`)
	require.NoError(t, err)
	_, err = db.Writer().ExecContext(context.Background(), `INSERT INTO messages DEFAULT VALUES`)
	require.NoError(t, err)

	var count int
	err = db.Reader().QueryRowContext(context.Background(), `SELECT COUNT(*) FROM messages`).Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	_, err = db.Reader().ExecContext(context.Background(), `INSERT INTO messages DEFAULT VALUES`)
	require.Error(t, err)
	require.NoError(t, db.Writer().Check(context.Background()))
	require.NoError(t, db.Reader().Check(context.Background()))

	var journalMode string
	err = db.Writer().QueryRowContext(context.Background(), `PRAGMA journal_mode`).Scan(&journalMode)
	require.NoError(t, err)
	require.Equal(t, "wal", journalMode)
	var queryOnly int
	err = db.Reader().QueryRowContext(context.Background(), `PRAGMA query_only`).Scan(&queryOnly)
	require.NoError(t, err)
	require.Equal(t, 1, queryOnly)
}

func TestOpenNamedMemorySharesStateWithStrictReader(t *testing.T) {
	t.Parallel()
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{
		DSN: "file:sqlitedb-memory-test?mode=memory&cache=shared",
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	var journalMode string
	err = db.Writer().QueryRowContext(context.Background(), `PRAGMA journal_mode`).Scan(&journalMode)
	require.NoError(t, err)
	require.Equal(t, "memory", journalMode)

	_, err = db.Writer().ExecContext(context.Background(), `CREATE TABLE values_table (value TEXT NOT NULL)`)
	require.NoError(t, err)
	_, err = db.Writer().ExecContext(context.Background(), `INSERT INTO values_table VALUES ('shared')`)
	require.NoError(t, err)

	var value string
	err = db.Reader().QueryRowContext(context.Background(), `SELECT value FROM values_table`).Scan(&value)
	require.NoError(t, err)
	require.Equal(t, "shared", value)
	_, err = db.Reader().ExecContext(context.Background(), `DELETE FROM values_table`)
	require.Error(t, err)
}

func TestOpenRejectsBareMemoryDSN(t *testing.T) {
	t.Parallel()
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{DSN: ":memory:"})

	require.Nil(t, db)
	require.Error(t, err)
}

func TestWriterTransactionRoutesReaderAndRollsBack(t *testing.T) {
	t.Parallel()
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{
		DSN: filepath.Join(t.TempDir(), "transactions.sqlite"),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	_, err = db.Writer().ExecContext(context.Background(), `CREATE TABLE entries (value TEXT NOT NULL)`)
	require.NoError(t, err)

	callbackErr := errors.New("rollback requested")
	err = db.Writer().WithinTx(context.Background(), func(ctx context.Context) error {
		_, insertErr := db.Writer().ExecContext(ctx, `INSERT INTO entries VALUES ('pending')`)
		require.NoError(t, insertErr)
		var count int
		queryErr := db.Reader().QueryRowContext(ctx, `SELECT COUNT(*) FROM entries`).Scan(&count)
		require.NoError(t, queryErr)
		require.Equal(t, 1, count)
		return callbackErr
	})
	require.ErrorIs(t, err, callbackErr)

	var count int
	err = db.Reader().QueryRowContext(context.Background(), `SELECT COUNT(*) FROM entries`).Scan(&count)
	require.NoError(t, err)
	require.Zero(t, count)
	require.Same(t, db.Reader(), db.TxManagers().Reader())
	require.Same(t, db.Writer(), db.TxManagers().Writer())
}

func TestWriterTransactionCannotEscalateReaderTransaction(t *testing.T) {
	t.Parallel()
	db, err := sqlitedb.Open(context.Background(), sqlitedb.Config{
		DSN: filepath.Join(t.TempDir(), "read-only.sqlite"),
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	err = db.Reader().WithinTx(context.Background(), func(ctx context.Context) error {
		return db.Writer().WithinTx(ctx, func(context.Context) error { return nil })
	})

	require.ErrorIs(t, err, txmanager.ErrReadOnlyEscalation)
}
