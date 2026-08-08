package sqlitedb

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOpenAppliesBusyTimeoutAndReaderPoolConfig(t *testing.T) {
	t.Parallel()
	db, err := Open(context.Background(), Config{
		DSN:         filepath.Join(t.TempDir(), "configured.sqlite"),
		BusyTimeout: 1250 * time.Millisecond,
		ReaderPool: PoolConfig{
			MaxOpenConnections:    3,
			MaxIdleConnections:    2,
			ConnectionMaxLifetime: time.Minute,
			ConnectionMaxIdleTime: 30 * time.Second,
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	require.Equal(t, 1, db.writerPool.Stats().MaxOpenConnections)
	require.Equal(t, 3, db.readerPool.Stats().MaxOpenConnections)
	for _, endpoint := range []*Endpoint{db.Writer(), db.Reader()} {
		var timeoutMillis int
		err = endpoint.QueryRowContext(context.Background(), `PRAGMA busy_timeout`).Scan(&timeoutMillis)
		require.NoError(t, err)
		require.Equal(t, 1250, timeoutMillis)
	}
}
