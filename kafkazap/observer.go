// Package kafkazap adapts kafka observer events to structured zap logs.
package kafkazap

import (
	"context"
	"errors"

	"github.com/devctllabs/go-libs/kafka"
	"go.uber.org/zap"
)

// Observer logs retries and explicit record dispositions. It intentionally
// does not log terminal runtime errors, which remain the Run caller's concern.
type Observer struct {
	logger *zap.Logger
}

// New constructs a kafka Observer backed by logger.
func New(logger *zap.Logger) (*Observer, error) {
	if logger == nil {
		return nil, errors.New("kafkazap: logger must not be nil")
	}
	return &Observer{logger: logger}, nil
}

// StartAttempt leaves attempt logging to Retry and the Run caller.
func (*Observer) StartAttempt(
	ctx context.Context,
	_ kafka.Attempt,
) (context.Context, kafka.AttemptDone) {
	return ctx, func(error) {}
}

// Retry logs every retryable failure at Warn level.
func (observer *Observer) Retry(_ context.Context, event kafka.RetryEvent) {
	observer.logger.Warn(
		"kafka operation retry",
		zap.String("phase", string(event.Phase)),
		zap.Uint("attempt", event.Attempt),
		zap.Duration("next_delay", event.NextDelay),
		zap.Error(event.Err),
	)
}

// BatchCompleted does not log successful high-volume batches.
func (*Observer) BatchCompleted(context.Context, kafka.BatchResult) {}

// RecordDisposition logs DLQ delivery at Info and drops at Warn.
func (observer *Observer) RecordDisposition(_ context.Context, disposition kafka.RecordDisposition) {
	fields := []zap.Field{
		zap.String("topic", disposition.Record.Topic),
		zap.Int32("partition", disposition.Record.Partition),
		zap.Int64("offset", disposition.Record.Offset),
		zap.Error(disposition.Cause),
	}
	switch disposition.Kind {
	case kafka.DispositionDLQ:
		observer.logger.Info("kafka record sent to DLQ", fields...)
	case kafka.DispositionDropped:
		observer.logger.Warn("kafka record dropped", fields...)
	}
}

var _ kafka.Observer = (*Observer)(nil)
