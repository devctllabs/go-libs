package healthotel

import (
	"context"
	"errors"
	"fmt"

	"github.com/devctllabs/go-libs/health"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const instrumentationScope = "github.com/devctllabs/go-libs/healthotel"

// Observer records health observations with a caller-owned OpenTelemetry provider.
type Observer struct {
	status     metric.Int64Gauge
	executions metric.Int64Counter
	duration   metric.Float64Histogram
}

var _ health.Observer = (*Observer)(nil)

// New constructs all health metric instruments once from provider.
func New(provider metric.MeterProvider) (*Observer, error) {
	if provider == nil {
		return nil, errors.New("healthotel: MeterProvider must not be nil")
	}

	meter := provider.Meter(instrumentationScope)
	status, err := meter.Int64Gauge(
		"health.check.status",
		metric.WithDescription("Latest observed health check status, where 1 is healthy and 0 is unhealthy."),
		metric.WithUnit("1"),
	)
	if err != nil {
		return nil, fmt.Errorf("healthotel: create status gauge: %w", err)
	}
	executions, err := meter.Int64Counter(
		"health.check.executions",
		metric.WithDescription("Number of completed health check executions."),
		metric.WithUnit("{check}"),
	)
	if err != nil {
		return nil, fmt.Errorf("healthotel: create executions counter: %w", err)
	}
	duration, err := meter.Float64Histogram(
		"health.check.duration",
		metric.WithDescription("Duration of health check executions."),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("healthotel: create duration histogram: %w", err)
	}

	return &Observer{status: status, executions: executions, duration: duration}, nil
}

// Observe records one bounded-cardinality health observation.
func (o *Observer) Observe(ctx context.Context, observation health.Observation) {
	outcome := boundedOutcome(observation.Outcome)
	common := attribute.NewSet(
		attribute.String("check.name", observation.Name),
		attribute.Bool("check.critical", observation.Critical),
	)
	status := int64(0)
	if outcome == health.OutcomeOK {
		status = 1
	}
	o.status.Record(ctx, status, metric.WithAttributeSet(common))

	result := attribute.NewSet(
		attribute.String("check.name", observation.Name),
		attribute.Bool("check.critical", observation.Critical),
		attribute.String("result", string(outcome)),
	)
	o.executions.Add(ctx, 1, metric.WithAttributeSet(result))
	o.duration.Record(ctx, observation.Duration.Seconds(), metric.WithAttributeSet(result))
}

func boundedOutcome(outcome health.Outcome) health.Outcome {
	switch outcome {
	case health.OutcomeOK, health.OutcomeError, health.OutcomeTimeout:
		return outcome
	default:
		return health.OutcomeError
	}
}
