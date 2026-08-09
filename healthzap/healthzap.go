package healthzap

import (
	"context"
	"errors"
	"sync"

	"github.com/devctllabs/go-libs/health"
	"go.uber.org/zap"
)

// Observer writes every failed health observation and one recovery after failures.
type Observer struct {
	logger *zap.Logger

	mu     sync.Mutex
	failed map[string]bool
}

var _ health.Observer = (*Observer)(nil)

// New creates an instance-owned logging observer.
func New(logger *zap.Logger) (*Observer, error) {
	if logger == nil {
		return nil, errors.New("healthzap: logger must not be nil")
	}
	return &Observer{logger: logger, failed: make(map[string]bool)}, nil
}

// Observe logs failures at a level selected by criticality and logs recovery once.
func (o *Observer) Observe(_ context.Context, observation health.Observation) {
	if observation.Outcome != health.OutcomeOK {
		o.mu.Lock()
		o.failed[observation.Name] = true
		o.mu.Unlock()

		fields := observationFields(observation)
		if observation.Err != nil {
			fields = append(fields, zap.Error(observation.Err))
		}
		if observation.Critical {
			o.logger.Error("health check failed", fields...)
			return
		}
		o.logger.Warn("health check failed", fields...)
		return
	}

	o.mu.Lock()
	wasFailed := o.failed[observation.Name]
	if wasFailed {
		o.failed[observation.Name] = false
	}
	o.mu.Unlock()
	if wasFailed {
		o.logger.Info("health check recovered", observationFields(observation)...)
	}
}

func observationFields(observation health.Observation) []zap.Field {
	return []zap.Field{
		zap.String("check.name", observation.Name),
		zap.Bool("check.critical", observation.Critical),
		zap.String("check.outcome", string(observation.Outcome)),
		zap.Duration("check.duration", observation.Duration),
	}
}
