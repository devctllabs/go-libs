package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/devctllabs/go-libs/kafkaoutbox"

// OTelObserverConfig configures outbox tracing and metrics.
type OTelObserverConfig struct {
	MeterProvider     metric.MeterProvider
	TracerProvider    trace.TracerProvider
	MaxBatchSpanLinks int
}

// OTelObserver implements Observer using OpenTelemetry.
type OTelObserver struct {
	tracer   trace.Tracer
	maxLinks int

	attempts         metric.Int64Counter
	duration         metric.Float64Histogram
	inflight         metric.Int64UpDownCounter
	enqueued         metric.Int64Counter
	batches          metric.Int64Counter
	events           metric.Int64Counter
	batchSize        metric.Int64Histogram
	retries          metric.Int64Counter
	fencingConflicts metric.Int64Counter
	topologyChanges  metric.Int64Counter
	topologyShards   metric.Int64Gauge
}

// NewOTelObserver constructs standard outbox instruments. A zero link cap
// uses 128; a negative cap is invalid.
func NewOTelObserver(config OTelObserverConfig) (*OTelObserver, error) {
	if config.MaxBatchSpanLinks < 0 {
		return nil, errors.New("kafkaoutbox: maximum batch span links must not be negative")
	}
	if config.MaxBatchSpanLinks == 0 {
		config.MaxBatchSpanLinks = 128
	}
	if config.MeterProvider == nil {
		config.MeterProvider = otel.GetMeterProvider()
	}
	if config.TracerProvider == nil {
		config.TracerProvider = otel.GetTracerProvider()
	}
	meter := config.MeterProvider.Meter(instrumentationName)
	observer := &OTelObserver{
		tracer: config.TracerProvider.Tracer(instrumentationName), maxLinks: config.MaxBatchSpanLinks,
	}
	if err := observer.createOperationInstruments(meter); err != nil {
		return nil, err
	}
	if err := observer.createWorkerInstruments(meter); err != nil {
		return nil, err
	}
	if err := observer.createTopologyInstruments(meter); err != nil {
		return nil, err
	}
	return observer, nil
}

func (observer *OTelObserver) createOperationInstruments(meter metric.Meter) error {
	var err error
	observer.attempts, err = meter.Int64Counter("kafkaoutbox.operation.attempts")
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create operation attempts counter: %w", err)
	}
	observer.duration, err = meter.Float64Histogram("kafkaoutbox.operation.duration", metric.WithUnit("s"))
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create operation duration histogram: %w", err)
	}
	observer.inflight, err = meter.Int64UpDownCounter("kafkaoutbox.operation.inflight")
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create operation inflight counter: %w", err)
	}
	observer.enqueued, err = meter.Int64Counter("kafkaoutbox.enqueue.events", metric.WithUnit("{event}"))
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create enqueue events counter: %w", err)
	}
	return nil
}

func (observer *OTelObserver) createWorkerInstruments(meter metric.Meter) error {
	var err error
	observer.batches, err = meter.Int64Counter("kafkaoutbox.worker.batches", metric.WithUnit("{batch}"))
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create worker batches counter: %w", err)
	}
	observer.events, err = meter.Int64Counter("kafkaoutbox.worker.events", metric.WithUnit("{event}"))
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create worker events counter: %w", err)
	}
	observer.batchSize, err = meter.Int64Histogram("kafkaoutbox.worker.batch.size", metric.WithUnit("{event}"))
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create worker batch size histogram: %w", err)
	}
	observer.retries, err = meter.Int64Counter("kafkaoutbox.worker.retries")
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create worker retries counter: %w", err)
	}
	observer.fencingConflicts, err = meter.Int64Counter("kafkaoutbox.worker.fencing_conflicts")
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create worker fencing conflicts counter: %w", err)
	}
	return nil
}

func (observer *OTelObserver) createTopologyInstruments(meter metric.Meter) error {
	var err error
	observer.topologyChanges, err = meter.Int64Counter("kafkaoutbox.topology.changes")
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create topology changes counter: %w", err)
	}
	observer.topologyShards, err = meter.Int64Gauge("kafkaoutbox.topology.shards", metric.WithUnit("{shard}"))
	if err != nil {
		return fmt.Errorf("kafkaoutbox: create topology shards gauge: %w", err)
	}
	return nil
}

// StartOperation starts a bounded operation span and updates attempt metrics.
func (observer *OTelObserver) StartOperation(
	ctx context.Context,
	operation Operation,
) (context.Context, OperationDone) {
	metricAttributes := metric.WithAttributes(attribute.String("kafkaoutbox.phase", string(operation.Phase)))
	observer.attempts.Add(ctx, 1, metricAttributes)
	observer.inflight.Add(ctx, 1, metricAttributes)
	started := time.Now()
	spanOptions := []trace.SpanStartOption{
		trace.WithSpanKind(operationSpanKind(operation.Phase)),
		trace.WithAttributes(
			attribute.String("kafkaoutbox.phase", string(operation.Phase)),
			attribute.Int64("kafkaoutbox.generation", int64(operation.Generation)),
			attribute.Int("kafkaoutbox.shard.id", int(operation.ShardID)),
			attribute.Int("messaging.batch.message_count", len(operation.Records)),
		),
	}
	if operation.Phase == OperationPublish {
		spanOptions = append(spanOptions, trace.WithNewRoot(), trace.WithLinks(observer.links(operation.Records)...))
	}
	spanCtx, span := observer.tracer.Start(ctx, "kafkaoutbox "+string(operation.Phase), spanOptions...)
	var once sync.Once
	return spanCtx, func(err error) {
		once.Do(func() {
			outcome := "success"
			if err != nil {
				outcome = "error"
				span.RecordError(err)
				span.SetStatus(codes.Error, "operation failed")
			}
			observer.inflight.Add(spanCtx, -1, metricAttributes)
			observer.duration.Record(spanCtx, time.Since(started).Seconds(), metric.WithAttributes(
				attribute.String("kafkaoutbox.phase", string(operation.Phase)),
				attribute.String("kafkaoutbox.outcome", outcome),
			))
			span.End()
		})
	}
}

// Enqueued records one event appended successfully.
func (observer *OTelObserver) Enqueued(ctx context.Context) {
	observer.enqueued.Add(ctx, 1)
}

// BatchCompleted records publish-attempt throughput and size by outcome.
func (observer *OTelObserver) BatchCompleted(ctx context.Context, result BatchResult) {
	attributes := metric.WithAttributes(attribute.String("kafkaoutbox.outcome", string(result.Outcome)))
	observer.batches.Add(ctx, 1, attributes)
	observer.events.Add(ctx, int64(result.Size), attributes)
	observer.batchSize.Record(ctx, int64(result.Size), attributes)
}

// Retry records one durably scheduled publish retry.
func (observer *OTelObserver) Retry(ctx context.Context, _ RetryEvent) {
	observer.retries.Add(ctx, 1)
}

// FencingConflict records coordination conflicts separately from delivery outcomes.
func (observer *OTelObserver) FencingConflict(ctx context.Context, event FencingEvent) {
	observer.fencingConflicts.Add(ctx, 1, metric.WithAttributes(
		attribute.String("kafkaoutbox.phase", string(event.Phase)),
	))
}

// TopologyReconciled records current shard count and actual topology changes.
func (observer *OTelObserver) TopologyReconciled(ctx context.Context, result TopologyResult) {
	observer.topologyShards.Record(ctx, int64(result.ShardCount))
	if result.Changed {
		observer.topologyChanges.Add(ctx, 1)
	}
}

func operationSpanKind(phase OperationPhase) trace.SpanKind {
	if phase == OperationEnqueue || phase == OperationPublish {
		return trace.SpanKindProducer
	}
	return trace.SpanKindInternal
}

func (observer *OTelObserver) links(records []RecordMetadata) []trace.Link {
	links := make([]trace.Link, 0, min(len(records), observer.maxLinks))
	type linkKey struct {
		traceID trace.TraceID
		spanID  trace.SpanID
	}
	seen := make(map[linkKey]struct{}, cap(links))
	for _, record := range records {
		spanContext := trace.SpanContextFromContext(record.Context)
		if !spanContext.IsValid() {
			continue
		}
		key := linkKey{traceID: spanContext.TraceID(), spanID: spanContext.SpanID()}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		links = append(links, trace.Link{SpanContext: spanContext})
		if len(links) == observer.maxLinks {
			break
		}
	}
	return links
}

var _ Observer = (*OTelObserver)(nil)
