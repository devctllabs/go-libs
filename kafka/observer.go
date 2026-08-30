package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// AttemptPhase identifies an independently retried operation.
type AttemptPhase string

const (
	AttemptHandler AttemptPhase = "handler"
	AttemptDLQ     AttemptPhase = "dlq"
	AttemptCommit  AttemptPhase = "commit"
)

// RecordMetadata identifies a record without retaining its payload.
type RecordMetadata struct {
	Topic     string
	Partition int32
	Offset    int64
	// Context carries extracted trace state and must not be retained after the
	// observer call.
	Context context.Context
}

// Attempt describes one operation call. Attempt numbering starts at one.
type Attempt struct {
	Phase   AttemptPhase
	Attempt uint
	Records []RecordMetadata
}

// AttemptDone completes instrumentation for an attempt.
type AttemptDone func(err error)

// RetryEvent reports a retryable failure and its next delay.
type RetryEvent struct {
	Phase     AttemptPhase
	Attempt   uint
	NextDelay time.Duration
	Err       error
}

// BatchResult reports outcomes after a successful offset commit.
type BatchResult struct {
	Size      int
	Processed int
	Skipped   int
	Rejected  int
	Dropped   int
	DLQ       int
}

// DispositionKind identifies terminal handling of a rejected record.
type DispositionKind string

const (
	DispositionDropped DispositionKind = "dropped"
	DispositionDLQ     DispositionKind = "dlq"
)

// RecordDisposition reports a rejected record that was dropped or delivered
// to a DLQ.
type RecordDisposition struct {
	Record RecordMetadata
	Kind   DispositionKind
	Cause  error
}

// Observer receives synchronous lifecycle signals. Implementations must be
// concurrency-safe and should return quickly.
type Observer interface {
	StartAttempt(ctx context.Context, attempt Attempt) (context.Context, AttemptDone)
	Retry(ctx context.Context, event RetryEvent)
	BatchCompleted(ctx context.Context, result BatchResult)
	RecordDisposition(ctx context.Context, disposition RecordDisposition)
}

type noopObserver struct{}

func (noopObserver) StartAttempt(ctx context.Context, _ Attempt) (context.Context, AttemptDone) {
	return ctx, func(error) {}
}
func (noopObserver) Retry(context.Context, RetryEvent)                    {}
func (noopObserver) BatchCompleted(context.Context, BatchResult)          {}
func (noopObserver) RecordDisposition(context.Context, RecordDisposition) {}

func effectiveObserver(observer Observer) Observer {
	if observer == nil {
		return noopObserver{}
	}
	return observer
}

type multiObserver struct {
	observers []Observer
}

// NewMultiObserver composes observers in registration order. Nil observers
// are rejected as configuration errors.
func NewMultiObserver(observers ...Observer) (Observer, error) {
	for _, observer := range observers {
		if observer == nil {
			return nil, errors.New("kafka: observer must not be nil")
		}
	}
	return multiObserver{observers: append([]Observer(nil), observers...)}, nil
}

func (observer multiObserver) StartAttempt(ctx context.Context, attempt Attempt) (context.Context, AttemptDone) {
	doneCallbacks := make([]AttemptDone, 0, len(observer.observers))
	for _, child := range observer.observers {
		var done AttemptDone
		ctx, done = child.StartAttempt(ctx, attempt)
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

func (observer multiObserver) Retry(ctx context.Context, event RetryEvent) {
	for _, child := range observer.observers {
		child.Retry(ctx, event)
	}
}

func (observer multiObserver) BatchCompleted(ctx context.Context, result BatchResult) {
	for _, child := range observer.observers {
		child.BatchCompleted(ctx, result)
	}
}

func (observer multiObserver) RecordDisposition(ctx context.Context, disposition RecordDisposition) {
	for _, child := range observer.observers {
		child.RecordDisposition(ctx, disposition)
	}
}

func (observer multiObserver) Hooks(group string) []kgo.Hook {
	var hooks []kgo.Hook
	for _, child := range observer.observers {
		if provider, ok := child.(interface{ Hooks(group string) []kgo.Hook }); ok {
			hooks = append(hooks, provider.Hooks(group)...)
		}
	}
	return hooks
}

func metadata(records []*kgo.Record) []RecordMetadata {
	result := make([]RecordMetadata, len(records))
	for index, record := range records {
		result[index] = recordMetadata(record)
	}
	return result
}

func recordMetadata(record *kgo.Record) RecordMetadata {
	return RecordMetadata{
		Topic: record.Topic, Partition: record.Partition, Offset: record.Offset,
		Context: record.Context,
	}
}

func observerHooks(observer Observer, group string) []kgo.Hook {
	provider, ok := observer.(interface{ Hooks(group string) []kgo.Hook })
	if !ok {
		return nil
	}
	return provider.Hooks(group)
}
