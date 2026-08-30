package kafkaoutbox

import (
	"context"
	"errors"
	"time"
)

// OperationPhase identifies one bounded outbox operation.
type OperationPhase string

const (
	OperationEnqueue  OperationPhase = "enqueue"
	OperationTopology OperationPhase = "topology"
	OperationClaim    OperationPhase = "claim"
	OperationRead     OperationPhase = "read"
	OperationPublish  OperationPhase = "publish"
	OperationFinalize OperationPhase = "finalize"
)

// RecordMetadata carries only the data needed to link a publish span to an
// originating trace. Context must not be retained after the observer call.
type RecordMetadata struct {
	Topic   string
	Context context.Context
}

// Operation describes one operation attempt.
type Operation struct {
	Phase      OperationPhase
	Generation uint64
	ShardID    uint
	Records    []RecordMetadata
}

// OperationDone completes operation instrumentation.
type OperationDone func(err error)

// BatchOutcome is the low-cardinality terminal result of a publish attempt.
type BatchOutcome string

const (
	BatchDelivered BatchOutcome = "delivered"
	BatchFailed    BatchOutcome = "failed"
	BatchFenced    BatchOutcome = "fenced"
)

// BatchResult reports one completed publish attempt.
type BatchResult struct {
	Outcome    BatchOutcome
	Generation uint64
	ShardID    uint
	Size       int
}

// RetryEvent reports a durable retry schedule after a publish failure.
type RetryEvent struct {
	Generation uint64
	ShardID    uint
	Attempt    uint
	NextDelay  time.Duration
	Err        error
}

// FencingEvent reports a claim that was lost before finalization.
type FencingEvent struct {
	Phase      OperationPhase
	Generation uint64
	ShardID    uint
}

// TopologyResult reports the persisted topology observed by a worker.
type TopologyResult struct {
	Revision   uint64
	Generation uint64
	ShardCount uint
	Changed    bool
}

// Observer receives synchronous lifecycle signals. Implementations must be
// concurrency-safe, return quickly, and must not retain supplied contexts.
type Observer interface {
	StartOperation(ctx context.Context, operation Operation) (context.Context, OperationDone)
	Enqueued(ctx context.Context)
	BatchCompleted(ctx context.Context, result BatchResult)
	Retry(ctx context.Context, event RetryEvent)
	FencingConflict(ctx context.Context, event FencingEvent)
	TopologyReconciled(ctx context.Context, result TopologyResult)
}

type noopObserver struct{}

func (noopObserver) StartOperation(ctx context.Context, _ Operation) (context.Context, OperationDone) {
	return ctx, func(error) {}
}
func (noopObserver) Enqueued(context.Context)                           {}
func (noopObserver) BatchCompleted(context.Context, BatchResult)        {}
func (noopObserver) Retry(context.Context, RetryEvent)                  {}
func (noopObserver) FencingConflict(context.Context, FencingEvent)      {}
func (noopObserver) TopologyReconciled(context.Context, TopologyResult) {}

func effectiveObserver(observer Observer) Observer {
	if observer == nil {
		return noopObserver{}
	}
	return observer
}

type multiObserver struct {
	observers []Observer
}

// NewMultiObserver composes observers in registration order.
func NewMultiObserver(observers ...Observer) (Observer, error) {
	for _, observer := range observers {
		if observer == nil {
			return nil, errors.New("kafkaoutbox: observer must not be nil")
		}
	}
	return multiObserver{observers: append([]Observer(nil), observers...)}, nil
}

func (observer multiObserver) StartOperation(ctx context.Context, operation Operation) (context.Context, OperationDone) {
	doneCallbacks := make([]OperationDone, 0, len(observer.observers))
	for _, child := range observer.observers {
		var done OperationDone
		ctx, done = child.StartOperation(ctx, operation)
		if done == nil {
			done = func(error) {}
		}
		doneCallbacks = append(doneCallbacks, done)
	}
	return ctx, func(err error) {
		for index := len(doneCallbacks) - 1; index >= 0; index-- {
			doneCallbacks[index](err)
		}
	}
}

func (observer multiObserver) Enqueued(ctx context.Context) {
	for _, child := range observer.observers {
		child.Enqueued(ctx)
	}
}

func (observer multiObserver) BatchCompleted(ctx context.Context, result BatchResult) {
	for _, child := range observer.observers {
		child.BatchCompleted(ctx, result)
	}
}

func (observer multiObserver) Retry(ctx context.Context, event RetryEvent) {
	for _, child := range observer.observers {
		child.Retry(ctx, event)
	}
}

func (observer multiObserver) FencingConflict(ctx context.Context, event FencingEvent) {
	for _, child := range observer.observers {
		child.FencingConflict(ctx, event)
	}
}

func (observer multiObserver) TopologyReconciled(ctx context.Context, result TopologyResult) {
	for _, child := range observer.observers {
		child.TopologyReconciled(ctx, result)
	}
}
