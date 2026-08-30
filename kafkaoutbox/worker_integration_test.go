//go:build integration

package kafkaoutbox_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/kafka"
	"github.com/devctllabs/go-libs/kafkaoutbox"
	"github.com/devctllabs/go-libs/retry"
	"github.com/stretchr/testify/require"
	"github.com/zeebo/xxh3"
	"go.uber.org/mock/gomock"
)

func TestWorkerPublishesCommittedEventAndDeletesIt(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
			Topic:         "orders",
			AggregateType: "Order",
			AggregateID:   "42",
			Type:          "OrderPaid",
			Value:         []byte("wire"),
		})
	}))

	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	runCtx, cancel := context.WithCancel(ctx)
	publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, messages []kafka.OutgoingMessage[[]byte]) error {
			require.Len(t, messages, 1)
			require.Equal(t, "orders", messages[0].Topic)
			require.Equal(t, []byte("Order::42"), messages[0].Key)
			require.Equal(t, []byte("wire"), messages[0].Value)
			require.Equal(t, []string{"id", "type", "aggregatetype"}, headerKeys(messages[0].Headers))
			cancel()
			return nil
		},
	)
	worker, err := kafkaoutbox.NewWorker(kafkaoutbox.WorkerConfig{
		MaxBatchSize:    100,
		PollInterval:    10 * time.Millisecond,
		DatabaseTimeout: time.Second,
		PublishTimeout:  time.Second,
		LeaseDuration:   4 * time.Second,
		MaxAttempts:     1,
	}, store, publisher)
	require.NoError(t, err)

	err = worker.Run(runCtx)
	require.NoError(t, err)
	require.Equal(t, 0, eventCount(t, ctx, db.Writer()))
}

func TestWorkerReportsLifecycleWithoutTreatingDeliveryAsFencing(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
			Topic: "orders", AggregateType: "Order", AggregateID: "42", Type: "OrderPaid",
		})
	}))
	observer := &recordingOutboxObserver{}
	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	runCtx, cancel := context.WithCancel(ctx)
	publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ []kafka.OutgoingMessage[[]byte]) error {
			cancel()
			return nil
		},
	)
	config := workerConfig(kafkaoutbox.TopologyConfig{})
	config.Observer = observer
	worker, err := kafkaoutbox.NewWorker(config, store, publisher)
	require.NoError(t, err)

	require.NoError(t, worker.Run(runCtx))
	require.Contains(t, observer.operationPhases(), kafkaoutbox.OperationTopology)
	require.Contains(t, observer.operationPhases(), kafkaoutbox.OperationClaim)
	require.Contains(t, observer.operationPhases(), kafkaoutbox.OperationRead)
	require.Contains(t, observer.operationPhases(), kafkaoutbox.OperationPublish)
	require.Contains(t, observer.operationPhases(), kafkaoutbox.OperationFinalize)
	require.Equal(t, []kafkaoutbox.BatchOutcome{kafkaoutbox.BatchDelivered}, observer.batchOutcomes())
	require.Len(t, observer.topologies, 1)
	require.True(t, observer.topologies[0].Changed)
	require.Equal(t, uint(4), observer.topologies[0].ShardCount)
	require.Empty(t, observer.fencing)
}

func TestWorkerRetriesTheSameBatchAfterPublishFailure(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	for _, aggregateID := range sameShardAggregateIDs() {
		require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
			return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
				Topic: "orders", AggregateType: "Order", AggregateID: aggregateID,
				Type: "OrderPaid", Value: []byte(aggregateID),
			})
		}))
	}
	policy, err := retry.NewExponential(retry.ExponentialConfig{
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Millisecond,
		Multiplier:   2,
	})
	require.NoError(t, err)
	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	runCtx, cancel := context.WithCancel(ctx)
	var firstIDs []string
	gomock.InOrder(
		publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, messages []kafka.OutgoingMessage[[]byte]) error {
				firstIDs = messageIDs(messages)
				return errors.New("broker unavailable")
			},
		),
		publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, messages []kafka.OutgoingMessage[[]byte]) error {
				require.Equal(t, firstIDs, messageIDs(messages))
				cancel()
				return nil
			},
		),
	)
	worker, err := kafkaoutbox.NewWorker(kafkaoutbox.WorkerConfig{
		MaxBatchSize: 2, PollInterval: time.Millisecond,
		DatabaseTimeout: time.Second, PublishTimeout: time.Second, LeaseDuration: 4 * time.Second,
		RetryPolicy: policy,
	}, store, publisher)
	require.NoError(t, err)

	require.NoError(t, worker.Run(runCtx))
	require.Equal(t, 0, eventCount(t, ctx, db.Writer()))
}

func TestWorkerStopsAfterBoundedPublishFailuresAndPreservesEvent(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
			Topic: "orders", AggregateType: "Order", AggregateID: "42",
			Type: "OrderPaid", Value: []byte("wire"),
		})
	}))
	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	publishErr := errors.New("authorization denied")
	publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).Return(publishErr)
	worker, err := kafkaoutbox.NewWorker(kafkaoutbox.WorkerConfig{
		MaxBatchSize: 1, PollInterval: time.Millisecond,
		DatabaseTimeout: time.Second, PublishTimeout: time.Second, LeaseDuration: 4 * time.Second,
		MaxAttempts: 1,
	}, store, publisher)
	require.NoError(t, err)

	err = worker.Run(ctx)
	var exhausted *kafkaoutbox.AttemptsExhaustedError
	require.ErrorAs(t, err, &exhausted)
	require.ErrorIs(t, err, publishErr)
	require.Equal(t, uint(1), exhausted.Attempts)
	require.Equal(t, 1, eventCount(t, ctx, db.Writer()))
	var failures uint
	require.NoError(t, db.Writer().QueryRow(ctx, `SELECT MAX(failure_count) FROM outbox_shards`).Scan(&failures))
	require.Equal(t, uint(1), failures)
}

func TestWorkerDoesNotLoseEnqueueDuringPartialBatchPublish(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	ids := sameShardAggregateIDs()
	appendEvent := func(aggregateID string) error {
		return db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
			return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
				Topic: "orders", AggregateType: "Order", AggregateID: aggregateID,
				Type: "OrderPaid", Value: []byte(aggregateID),
			})
		})
	}
	require.NoError(t, appendEvent(ids[0]))
	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	runCtx, cancel := context.WithCancel(ctx)
	gomock.InOrder(
		publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, messages []kafka.OutgoingMessage[[]byte]) error {
				require.Len(t, messages, 1)
				require.NoError(t, appendEvent(ids[1]))
				return nil
			},
		),
		publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).DoAndReturn(
			func(_ context.Context, messages []kafka.OutgoingMessage[[]byte]) error {
				require.Len(t, messages, 1)
				require.Equal(t, []byte(ids[1]), messages[0].Value)
				cancel()
				return nil
			},
		),
	)
	worker, err := kafkaoutbox.NewWorker(kafkaoutbox.WorkerConfig{
		MaxBatchSize: 2, PollInterval: 5 * time.Millisecond,
		DatabaseTimeout: time.Second, PublishTimeout: time.Second, LeaseDuration: 4 * time.Second,
		MaxAttempts: 1,
	}, store, publisher)
	require.NoError(t, err)

	require.NoError(t, worker.Run(runCtx))
	require.Equal(t, 0, eventCount(t, ctx, db.Writer()))
}

func TestWorkerRotatesToOlderDueShardAfterFullBatch(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	idsByShard := aggregateIDsForShards(2, 1)
	for _, aggregateID := range append(idsByShard[0], idsByShard[1]...) {
		require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
			return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
				Topic: "orders", AggregateType: "Order", AggregateID: aggregateID, Type: "OrderPaid",
			})
		}))
	}
	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	runCtx, cancel := context.WithCancel(ctx)
	seenShards := make([]int, 0, 2)
	observer := &recordingOutboxObserver{}
	publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).Times(2).DoAndReturn(
		func(_ context.Context, messages []kafka.OutgoingMessage[[]byte]) error {
			require.Len(t, messages, 1)
			seenShards = append(seenShards, defaultShardFor(int64(xxh3.Hash(messages[0].Key))))
			if len(seenShards) == 2 {
				cancel()
			}
			return nil
		},
	)
	worker, err := kafkaoutbox.NewWorker(kafkaoutbox.WorkerConfig{
		MaxBatchSize: 1, PollInterval: time.Second,
		DatabaseTimeout: time.Second, PublishTimeout: time.Second, LeaseDuration: 4 * time.Second,
		MaxAttempts: 1, Observer: observer,
	}, store, publisher)
	require.NoError(t, err)

	require.NoError(t, worker.Run(runCtx))
	require.Equal(t, []int{0, 1}, seenShards)
	require.Equal(t, 1, countPhase(observer.operationPhases(), kafkaoutbox.OperationClaim))
}

func TestWorkerDoesNotDeleteAfterLosingClaimToRepartition(t *testing.T) {
	ctx, stop := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(stop)
	db := openTestDatabase(t, ctx)
	applyMigrations(t, ctx, db.Writer(), kafkaoutbox.PollingMigrations())
	store, err := kafkaoutbox.NewPollingStore(db.Writer())
	require.NoError(t, err)
	require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
		return store.Append(txCtx, kafkaoutbox.Event[[]byte]{
			Topic: "orders", AggregateType: "Order", AggregateID: "42",
			Type: "OrderPaid", Value: []byte("wire"),
		})
	}))
	ctrl := gomock.NewController(t)
	publisher := kafkaoutbox.NewMockBatchPublisher(ctrl)
	runCtx, cancel := context.WithCancel(ctx)
	publisher.EXPECT().SendBatch(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ []kafka.OutgoingMessage[[]byte]) error {
			require.NoError(t, db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
				_, execErr := db.Writer().Exec(txCtx, `DELETE FROM outbox_shards`)
				return execErr
			}))
			cancel()
			return nil
		},
	)
	worker, err := kafkaoutbox.NewWorker(workerConfig(kafkaoutbox.TopologyConfig{}), store, publisher)
	require.NoError(t, err)

	require.NoError(t, worker.Run(runCtx))
	require.Equal(t, 1, eventCount(t, ctx, db.Writer()))
}

func sameShardAggregateIDs() []string {
	byShard := make(map[int][]string)
	for candidate := 0; ; candidate++ {
		id := fmt.Sprintf("candidate-%d", candidate)
		shard := defaultShardFor(int64(xxh3.HashString("Order::" + id)))
		byShard[shard] = append(byShard[shard], id)
		if len(byShard[shard]) == 2 {
			return byShard[shard]
		}
	}
}

func aggregateIDsForShards(firstCount, secondCount int) map[int][]string {
	result := map[int][]string{0: {}, 1: {}}
	for candidate := 0; len(result[0]) < firstCount || len(result[1]) < secondCount; candidate++ {
		id := fmt.Sprintf("rotation-%d", candidate)
		shard := defaultShardFor(int64(xxh3.HashString("Order::" + id)))
		limit := firstCount
		if shard == 1 {
			limit = secondCount
		}
		if (shard == 0 || shard == 1) && len(result[shard]) < limit {
			result[shard] = append(result[shard], id)
		}
	}
	return result
}

func defaultShardFor(hash int64) int {
	switch {
	case hash < -(int64(1) << 62):
		return 0
	case hash < 0:
		return 1
	case hash < int64(1)<<62:
		return 2
	default:
		return 3
	}
}

func headerKeys(headers []kafka.Header) []string {
	keys := make([]string, len(headers))
	for index, header := range headers {
		keys[index] = header.Key
	}
	return keys
}

func messageIDs(messages []kafka.OutgoingMessage[[]byte]) []string {
	ids := make([]string, len(messages))
	for index, message := range messages {
		ids[index] = string(message.Headers[0].Value)
	}
	return ids
}

type recordingOutboxObserver struct {
	mu         sync.Mutex
	operations []kafkaoutbox.Operation
	batches    []kafkaoutbox.BatchResult
	topologies []kafkaoutbox.TopologyResult
	fencing    []kafkaoutbox.FencingEvent
}

func (observer *recordingOutboxObserver) StartOperation(
	ctx context.Context,
	operation kafkaoutbox.Operation,
) (context.Context, kafkaoutbox.OperationDone) {
	observer.mu.Lock()
	observer.operations = append(observer.operations, operation)
	observer.mu.Unlock()
	return ctx, func(error) {}
}

func (*recordingOutboxObserver) Enqueued(context.Context) {}

func (observer *recordingOutboxObserver) BatchCompleted(_ context.Context, result kafkaoutbox.BatchResult) {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	observer.batches = append(observer.batches, result)
}

func (*recordingOutboxObserver) Retry(context.Context, kafkaoutbox.RetryEvent) {}

func (observer *recordingOutboxObserver) FencingConflict(_ context.Context, event kafkaoutbox.FencingEvent) {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	observer.fencing = append(observer.fencing, event)
}

func (observer *recordingOutboxObserver) TopologyReconciled(_ context.Context, result kafkaoutbox.TopologyResult) {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	observer.topologies = append(observer.topologies, result)
}

func (observer *recordingOutboxObserver) operationPhases() []kafkaoutbox.OperationPhase {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	phases := make([]kafkaoutbox.OperationPhase, len(observer.operations))
	for index, operation := range observer.operations {
		phases[index] = operation.Phase
	}
	return phases
}

func (observer *recordingOutboxObserver) batchOutcomes() []kafkaoutbox.BatchOutcome {
	observer.mu.Lock()
	defer observer.mu.Unlock()
	outcomes := make([]kafkaoutbox.BatchOutcome, len(observer.batches))
	for index, batch := range observer.batches {
		outcomes[index] = batch.Outcome
	}
	return outcomes
}

func countPhase(phases []kafkaoutbox.OperationPhase, want kafkaoutbox.OperationPhase) int {
	count := 0
	for _, phase := range phases {
		if phase == want {
			count++
		}
	}
	return count
}
