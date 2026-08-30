//go:build integration

package kafkaoutbox_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/kafkaoutbox"
	"github.com/devctllabs/go-libs/postgresdb"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWorkerReconcilesVersionedEqualRangeTopology(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)

	runIdleWorker(t, ctx, store, kafkaoutbox.TopologyConfig{})
	assertTopology(t, ctx, db.Writer(), 1, 1, 4)
	assertFourShardBounds(t, ctx, db.Writer())

	runIdleWorker(t, ctx, store, kafkaoutbox.TopologyConfig{Revision: 2, ShardCount: 8})
	assertTopology(t, ctx, db.Writer(), 2, 2, 8)

	runIdleWorker(t, ctx, store, kafkaoutbox.TopologyConfig{Revision: 1, ShardCount: 4})
	assertTopology(t, ctx, db.Writer(), 2, 2, 8)

	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	worker, err := kafkaoutbox.NewWorker(workerConfig(kafkaoutbox.TopologyConfig{Revision: 2, ShardCount: 4}), store, publisher)
	require.NoError(t, err)
	err = worker.Run(ctx)
	require.ErrorContains(t, err, "conflicts with persisted topology revision")
}

func runIdleWorker(t *testing.T, parent context.Context, store *kafkaoutbox.PollingStore, topology kafkaoutbox.TopologyConfig) {
	t.Helper()
	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	worker, err := kafkaoutbox.NewWorker(workerConfig(topology), store, publisher)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(parent, 30*time.Millisecond)
	defer cancel()
	require.NoError(t, worker.Run(ctx))
}

func workerConfig(topology kafkaoutbox.TopologyConfig) kafkaoutbox.WorkerConfig {
	return kafkaoutbox.WorkerConfig{
		MaxBatchSize: 10, PollInterval: 5 * time.Millisecond,
		DatabaseTimeout: time.Second, PublishTimeout: time.Second, LeaseDuration: 4 * time.Second,
		MaxAttempts: 1, Topology: topology,
	}
}

func assertTopology(t *testing.T, ctx context.Context, endpoint *postgresdb.Endpoint, revision, generation uint64, count int) {
	t.Helper()
	var gotRevision, gotGeneration uint64
	var gotCount int
	err := endpoint.QueryRow(ctx, `SELECT revision, generation, shard_count FROM outbox_topology`).Scan(
		&gotRevision,
		&gotGeneration,
		&gotCount,
	)
	require.NoError(t, err)
	require.Equal(t, revision, gotRevision)
	require.Equal(t, generation, gotGeneration)
	require.Equal(t, count, gotCount)
}

func assertFourShardBounds(t *testing.T, ctx context.Context, endpoint *postgresdb.Endpoint) {
	t.Helper()
	rows, err := endpoint.Query(ctx, `SELECT hash_from, hash_to FROM outbox_shards ORDER BY shard_id`)
	require.NoError(t, err)
	defer rows.Close()
	want := [][2]sql.NullInt64{
		{{}, {Int64: -(int64(1) << 62), Valid: true}},
		{{Int64: -(int64(1) << 62), Valid: true}, {Int64: 0, Valid: true}},
		{{Int64: 0, Valid: true}, {Int64: int64(1) << 62, Valid: true}},
		{{Int64: int64(1) << 62, Valid: true}, {}},
	}
	var got [][2]sql.NullInt64
	for rows.Next() {
		var bounds [2]sql.NullInt64
		require.NoError(t, rows.Scan(&bounds[0], &bounds[1]))
		got = append(got, bounds)
	}
	require.NoError(t, rows.Err())
	require.Equal(t, want, got)
}
