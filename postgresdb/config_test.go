package postgresdb

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestOpenAppliesTypedPoolOverridesWithoutConnecting(t *testing.T) {
	t.Parallel()
	db, err := Open(context.Background(), Config{
		Writer: EndpointConfig{
			DSN: "postgres://user:password@127.0.0.1:1/app?pool_max_conns=9&default_query_exec_mode=cache_statement",
			Pool: PoolConfig{
				MaxConnections:              17,
				MinIdleConnections:          2,
				MaxConnectionLifetime:       45 * time.Minute,
				MaxConnectionLifetimeJitter: 3 * time.Minute,
				MaxConnectionIdleTime:       5 * time.Minute,
				HealthCheckPeriod:           20 * time.Second,
			},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	poolConfig := db.writerPool.Config()
	require.Equal(t, int32(17), poolConfig.MaxConns)
	require.Equal(t, int32(2), poolConfig.MinIdleConns)
	require.Equal(t, 45*time.Minute, poolConfig.MaxConnLifetime)
	require.Equal(t, 3*time.Minute, poolConfig.MaxConnLifetimeJitter)
	require.Equal(t, 5*time.Minute, poolConfig.MaxConnIdleTime)
	require.Equal(t, 20*time.Second, poolConfig.HealthCheckPeriod)
	require.Equal(t, pgx.QueryExecModeExec, poolConfig.ConnConfig.DefaultQueryExecMode)
	require.Same(t, db.writerPool, db.readerPool)
	require.NotSame(t, db.Writer(), db.Reader())
}
