package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/devctllabs/go-libs/kafka"
	"github.com/devctllabs/go-libs/retry"
	"go.opentelemetry.io/otel/propagation"
)

// ErrAlreadyRun reports a second Run call on one Worker instance.
var ErrAlreadyRun = errors.New("kafkaoutbox: worker has already run")

var errWorkerStopped = errors.New("kafkaoutbox: worker stopped")

// AttemptsExhaustedError reports a shard that reached its consecutive publish-failure limit.
type AttemptsExhaustedError struct {
	Generation uint64
	ShardID    uint
	Attempts   uint
	Err        error
}

// Error implements error.
func (err *AttemptsExhaustedError) Error() string {
	return fmt.Sprintf(
		"kafkaoutbox: shard %d/%d exhausted %d publish attempts: %v",
		err.Generation,
		err.ShardID,
		err.Attempts,
		err.Err,
	)
}

// Unwrap returns the last publish failure.
func (err *AttemptsExhaustedError) Unwrap() error { return err.Err }

// TopologyConfig selects a monotonic persisted virtual-shard topology.
type TopologyConfig struct {
	Revision              uint64
	ShardCount            uint
	AdvisoryLockNamespace int32
}

// WorkerConfig controls polling, delivery, retries, and claim lifetime.
type WorkerConfig struct {
	MaxBatchSize    int
	PollInterval    time.Duration
	DatabaseTimeout time.Duration
	PublishTimeout  time.Duration
	LeaseDuration   time.Duration
	RetryPolicy     retry.Policy
	MaxAttempts     uint
	Topology        TopologyConfig
	Observer        Observer
}

// Worker claims one virtual shard at a time and publishes its committed events.
type Worker struct {
	config    WorkerConfig
	store     *PollingStore
	publisher BatchPublisher
	observer  Observer

	mu      sync.Mutex
	started bool
}

// NewWorker constructs a single-use polling worker without starting goroutines.
func NewWorker(config WorkerConfig, store *PollingStore, publisher BatchPublisher) (*Worker, error) {
	config = defaultWorkerConfig(config)
	if err := validateWorkerConfig(config); err != nil {
		return nil, err
	}
	if store == nil {
		return nil, errors.New("kafkaoutbox: polling store must not be nil")
	}
	if publisher == nil {
		return nil, errors.New("kafkaoutbox: batch publisher must not be nil")
	}
	return &Worker{
		config: config, store: store, publisher: publisher,
		observer: effectiveObserver(config.Observer),
	}, nil
}

// Run publishes committed events until ctx is canceled or a terminal error occurs.
// Context cancellation is a clean stop.
func (worker *Worker) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("kafkaoutbox: context must not be nil")
	}
	if err := worker.beginRun(); err != nil {
		return err
	}
	if err := worker.reconcileTopology(ctx); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	var nextClaim *shardClaim
	for ctx.Err() == nil {
		var err error
		nextClaim, err = worker.runOnce(ctx, nextClaim)
		if err != nil {
			if errors.Is(err, errWorkerStopped) {
				return nil
			}
			return err
		}
	}
	return nil
}

func (worker *Worker) reconcileTopology(ctx context.Context) error {
	topologyCtx, done := worker.startOperation(ctx, Operation{Phase: OperationTopology})
	topology, err := worker.store.reconcileTopology(topologyCtx, worker.config)
	done(err)
	if err != nil {
		return fmt.Errorf("kafkaoutbox: reconcile topology: %w", err)
	}
	worker.observer.TopologyReconciled(topologyCtx, topology)
	return nil
}

func (worker *Worker) runOnce(ctx context.Context, claim *shardClaim) (*shardClaim, error) {
	if claim == nil {
		claimCtx, done := worker.startOperation(ctx, Operation{Phase: OperationClaim})
		var err error
		claim, err = worker.store.claimShard(claimCtx, worker.config)
		done(err)
		if err != nil {
			return nil, worker.operationError(ctx, "claim shard", err)
		}
	}
	if claim == nil {
		if waitForPoll(ctx, worker.config.PollInterval) != nil {
			return nil, errWorkerStopped
		}
		return nil, nil
	}
	return worker.processClaim(ctx, *claim)
}

func (worker *Worker) processClaim(ctx context.Context, claim shardClaim) (*shardClaim, error) {
	readCtx, done := worker.startOperation(ctx, claim.operation(OperationRead, nil))
	batch, err := worker.store.readBatch(readCtx, claim, worker.config)
	done(err)
	if err != nil {
		return nil, worker.operationError(ctx, "read batch", err)
	}
	if len(batch) == 0 {
		return worker.finalizeEmpty(ctx, claim)
	}
	return worker.publishBatch(ctx, claim, batch)
}

func (worker *Worker) publishBatch(
	ctx context.Context,
	claim shardClaim,
	batch []storedEvent,
) (*shardClaim, error) {
	publishCtx, done := worker.startOperation(ctx, claim.operation(OperationPublish, batch))
	publishCtx, cancel := context.WithTimeout(publishCtx, worker.config.PublishTimeout)
	err := worker.publisher.SendBatch(publishCtx, outgoingMessages(batch))
	cancel()
	done(err)
	if err != nil {
		worker.observeBatch(publishCtx, claim, batch, BatchFailed)
		if ctx.Err() != nil {
			return nil, errWorkerStopped
		}
		return worker.finalizeFailure(ctx, claim, err)
	}
	return worker.finalizeSuccess(ctx, claim, batch)
}

func (worker *Worker) finalizeEmpty(ctx context.Context, claim shardClaim) (*shardClaim, error) {
	finalizeCtx, done := worker.startFinalize(ctx, claim)
	nextClaim, err := worker.store.finalizeEmpty(finalizeCtx, claim, ctx.Err() == nil, worker.config)
	done(err)
	if errors.Is(err, errClaimLost) {
		worker.observeFencing(finalizeCtx, claim, OperationFinalize)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("kafkaoutbox: finalize empty shard: %w", err)
	}
	return nextClaim, nil
}

func (worker *Worker) finalizeFailure(
	ctx context.Context,
	claim shardClaim,
	publishErr error,
) (*shardClaim, error) {
	finalizeCtx, done := worker.startFinalize(ctx, claim)
	failures, delay, nextClaim, err := worker.store.finalizeFailure(finalizeCtx, claim, worker.config)
	done(err)
	if errors.Is(err, errClaimLost) {
		worker.observeFencing(finalizeCtx, claim, OperationFinalize)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("kafkaoutbox: finalize failed batch: %w", err)
	}
	if worker.config.MaxAttempts != 0 && failures >= worker.config.MaxAttempts {
		return nil, &AttemptsExhaustedError{
			Generation: claim.generation, ShardID: claim.id, Attempts: failures, Err: publishErr,
		}
	}
	worker.observer.Retry(finalizeCtx, RetryEvent{
		Generation: claim.generation, ShardID: claim.id,
		Attempt: failures, NextDelay: delay, Err: publishErr,
	})
	return nextClaim, nil
}

func (worker *Worker) finalizeSuccess(
	ctx context.Context,
	claim shardClaim,
	batch []storedEvent,
) (*shardClaim, error) {
	finalizeCtx, done := worker.startFinalize(ctx, claim)
	nextClaim, err := worker.store.finalizeSuccess(
		finalizeCtx, claim, batch, ctx.Err() == nil, worker.config,
	)
	done(err)
	if errors.Is(err, errClaimLost) {
		worker.observeBatch(finalizeCtx, claim, batch, BatchFenced)
		worker.observeFencing(finalizeCtx, claim, OperationFinalize)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("kafkaoutbox: finalize successful batch: %w", err)
	}
	worker.observeBatch(finalizeCtx, claim, batch, BatchDelivered)
	return nextClaim, nil
}

func (worker *Worker) startFinalize(ctx context.Context, claim shardClaim) (context.Context, OperationDone) {
	return worker.startOperation(context.WithoutCancel(ctx), claim.operation(OperationFinalize, nil))
}

func (*Worker) operationError(ctx context.Context, operation string, err error) error {
	if ctx.Err() != nil {
		return errWorkerStopped
	}
	return fmt.Errorf("kafkaoutbox: %s: %w", operation, err)
}

func (worker *Worker) startOperation(ctx context.Context, operation Operation) (context.Context, OperationDone) {
	observedCtx, done := worker.observer.StartOperation(ctx, operation)
	if observedCtx == nil {
		observedCtx = ctx
	}
	if done == nil {
		done = func(error) {}
	}
	return observedCtx, done
}

func (worker *Worker) observeBatch(ctx context.Context, claim shardClaim, batch []storedEvent, outcome BatchOutcome) {
	worker.observer.BatchCompleted(ctx, BatchResult{
		Outcome: outcome, Generation: claim.generation, ShardID: claim.id, Size: len(batch),
	})
}

func (worker *Worker) observeFencing(ctx context.Context, claim shardClaim, phase OperationPhase) {
	worker.observer.FencingConflict(ctx, FencingEvent{
		Phase: phase, Generation: claim.generation, ShardID: claim.id,
	})
}

func (claim shardClaim) operation(phase OperationPhase, batch []storedEvent) Operation {
	return Operation{
		Phase: phase, Generation: claim.generation, ShardID: claim.id,
		Records: recordMetadata(batch),
	}
}

func recordMetadata(batch []storedEvent) []RecordMetadata {
	records := make([]RecordMetadata, len(batch))
	for index, event := range batch {
		carrier := propagation.MapCarrier{}
		if event.traceparent != nil {
			carrier.Set("traceparent", *event.traceparent)
		}
		if event.tracestate != nil {
			carrier.Set("tracestate", *event.tracestate)
		}
		records[index] = RecordMetadata{
			Topic:   event.topic,
			Context: propagation.TraceContext{}.Extract(context.Background(), carrier),
		}
	}
	return records
}

func (worker *Worker) beginRun() error {
	worker.mu.Lock()
	defer worker.mu.Unlock()
	if worker.started {
		return ErrAlreadyRun
	}
	worker.started = true
	return nil
}

func defaultWorkerConfig(config WorkerConfig) WorkerConfig {
	if config.Topology.Revision == 0 {
		config.Topology.Revision = 1
	}
	if config.Topology.ShardCount == 0 {
		config.Topology.ShardCount = 4
	}
	return config
}

func validateWorkerConfig(config WorkerConfig) error {
	if config.MaxBatchSize <= 0 {
		return errors.New("kafkaoutbox: maximum batch size must be positive")
	}
	if config.PollInterval <= 0 {
		return errors.New("kafkaoutbox: poll interval must be positive")
	}
	if config.DatabaseTimeout <= 0 {
		return errors.New("kafkaoutbox: database timeout must be positive")
	}
	if config.PublishTimeout <= 0 {
		return errors.New("kafkaoutbox: publish timeout must be positive")
	}
	minimumLease := config.PublishTimeout + 2*config.DatabaseTimeout
	if config.LeaseDuration <= minimumLease {
		return fmt.Errorf("kafkaoutbox: lease duration must exceed %s", minimumLease)
	}
	if config.Topology.Revision > math.MaxInt64 {
		return errors.New("kafkaoutbox: topology revision exceeds PostgreSQL bigint")
	}
	if !validShardCount(config.Topology.ShardCount) {
		return errors.New("kafkaoutbox: shard count must be a power of two between 1 and 1024")
	}
	if config.MaxAttempts != 1 && config.RetryPolicy == nil {
		return errors.New("kafkaoutbox: retry policy is required when retries are possible")
	}
	return nil
}

func validShardCount(count uint) bool {
	return count >= 1 && count <= 1024 && count&(count-1) == 0
}

func outgoingMessages(batch []storedEvent) []kafka.OutgoingMessage[[]byte] {
	messages := make([]kafka.OutgoingMessage[[]byte], len(batch))
	for index, event := range batch {
		messages[index] = kafka.OutgoingMessage[[]byte]{
			Topic: event.topic,
			Key:   []byte(event.aggregateKey),
			Headers: []kafka.Header{
				{Key: "id", Value: []byte(event.id.String())},
				{Key: "type", Value: []byte(event.eventType)},
				{Key: "aggregatetype", Value: []byte(event.aggregateType)},
			},
			Value: event.payload,
		}
	}
	return messages
}

func waitForPoll(ctx context.Context, interval time.Duration) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
