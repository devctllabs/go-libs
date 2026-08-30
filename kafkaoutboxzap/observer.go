// Package kafkaoutboxzap adapts kafkaoutbox observer events to structured zap logs.
package kafkaoutboxzap

import (
	"context"
	"errors"

	"github.com/devctllabs/go-libs/kafkaoutbox"
	"go.uber.org/zap"
)

// Observer logs actionable low-volume outbox lifecycle events.
type Observer struct {
	logger *zap.Logger
}

// New constructs a kafkaoutbox Observer backed by logger.
func New(logger *zap.Logger) (*Observer, error) {
	if logger == nil {
		return nil, errors.New("kafkaoutboxzap: logger must not be nil")
	}
	return &Observer{logger: logger}, nil
}

// StartOperation leaves terminal runtime error logging to the Run caller.
func (*Observer) StartOperation(
	ctx context.Context,
	_ kafkaoutbox.Operation,
) (context.Context, kafkaoutbox.OperationDone) {
	return ctx, func(error) {}
}

// Enqueued avoids per-event success logs.
func (*Observer) Enqueued(context.Context) {}

// BatchCompleted avoids high-volume success and failure logs; Retry and the
// Run return value carry the actionable failure paths.
func (*Observer) BatchCompleted(context.Context, kafkaoutbox.BatchResult) {}

// Retry logs every durably scheduled publish retry.
func (observer *Observer) Retry(_ context.Context, event kafkaoutbox.RetryEvent) {
	observer.logger.Warn(
		"kafka outbox publish retry scheduled",
		zap.Uint64("generation", event.Generation),
		zap.Uint("shard_id", event.ShardID),
		zap.Uint("attempt", event.Attempt),
		zap.Duration("next_delay", event.NextDelay),
		zap.Error(event.Err),
	)
}

// FencingConflict logs a lost shard claim separately from publish failures.
func (observer *Observer) FencingConflict(_ context.Context, event kafkaoutbox.FencingEvent) {
	observer.logger.Warn(
		"kafka outbox shard claim lost",
		zap.String("phase", string(event.Phase)),
		zap.Uint64("generation", event.Generation),
		zap.Uint("shard_id", event.ShardID),
	)
}

// TopologyReconciled logs only actual persisted topology changes.
func (observer *Observer) TopologyReconciled(_ context.Context, result kafkaoutbox.TopologyResult) {
	if !result.Changed {
		return
	}
	observer.logger.Info(
		"kafka outbox topology changed",
		zap.Uint64("revision", result.Revision),
		zap.Uint64("generation", result.Generation),
		zap.Uint("shard_count", result.ShardCount),
	)
}

var _ kafkaoutbox.Observer = (*Observer)(nil)
