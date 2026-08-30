package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/devctllabs/go-libs/kafka"

// OTelObserverConfig configures standard franz-go hooks and wrapper-level
// tracing and metrics.
type OTelObserverConfig struct {
	MeterProvider     metric.MeterProvider
	TracerProvider    trace.TracerProvider
	Propagator        propagation.TextMapPropagator
	MaxBatchSpanLinks int
	// DisablePartitionAttribute removes partition from custom metrics when its
	// cardinality is too high for the deployment.
	DisablePartitionAttribute bool
}

// OTelObserver implements Observer and supplies standard franz-go OTel hooks.
type OTelObserver struct {
	meterProvider  metric.MeterProvider
	tracerProvider trace.TracerProvider
	propagator     propagation.TextMapPropagator
	tracer         trace.Tracer
	maxLinks       int
	partition      bool

	attempts     metric.Int64Counter
	retries      metric.Int64Counter
	messages     metric.Int64Counter
	dispositions metric.Int64Counter
	inflight     metric.Int64UpDownCounter
	duration     metric.Float64Histogram
	batchSize    metric.Int64Histogram
}

// NewOTelObserver constructs wrapper instruments. A zero link cap uses 128;
// a negative cap is invalid.
func NewOTelObserver(config OTelObserverConfig) (*OTelObserver, error) {
	if config.MaxBatchSpanLinks < 0 {
		return nil, errors.New("kafka: maximum batch span links must not be negative")
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
	if config.Propagator == nil {
		config.Propagator = otel.GetTextMapPropagator()
	}
	meter := config.MeterProvider.Meter(instrumentationName)
	observer := &OTelObserver{
		meterProvider: config.MeterProvider, tracerProvider: config.TracerProvider,
		propagator: config.Propagator, tracer: config.TracerProvider.Tracer(instrumentationName),
		maxLinks: config.MaxBatchSpanLinks, partition: !config.DisablePartitionAttribute,
	}
	var err error
	observer.attempts, err = meter.Int64Counter("kafka.consumer.operation.attempts")
	if err != nil {
		return nil, fmt.Errorf("kafka: create attempts counter: %w", err)
	}
	observer.retries, err = meter.Int64Counter("kafka.consumer.operation.retries")
	if err != nil {
		return nil, fmt.Errorf("kafka: create retries counter: %w", err)
	}
	observer.messages, err = meter.Int64Counter("kafka.consumer.messages")
	if err != nil {
		return nil, fmt.Errorf("kafka: create messages counter: %w", err)
	}
	observer.dispositions, err = meter.Int64Counter("kafka.consumer.record.dispositions")
	if err != nil {
		return nil, fmt.Errorf("kafka: create dispositions counter: %w", err)
	}
	observer.inflight, err = meter.Int64UpDownCounter("kafka.consumer.operation.inflight")
	if err != nil {
		return nil, fmt.Errorf("kafka: create in-flight counter: %w", err)
	}
	observer.duration, err = meter.Float64Histogram(
		"kafka.consumer.operation.duration",
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, fmt.Errorf("kafka: create duration histogram: %w", err)
	}
	observer.batchSize, err = meter.Int64Histogram("kafka.consumer.batch.size", metric.WithUnit("{message}"))
	if err != nil {
		return nil, fmt.Errorf("kafka: create batch size histogram: %w", err)
	}
	return observer, nil
}

// Hooks returns standard broker, fetch, produce, and propagation hooks for a
// franz-go client. group may be blank for producers.
func (observer *OTelObserver) Hooks(group string) []kgo.Hook {
	tracerOptions := []kotel.TracerOpt{
		kotel.TracerProvider(observer.tracerProvider),
		kotel.TracerPropagator(observer.propagator),
	}
	if group != "" {
		tracerOptions = append(tracerOptions, kotel.ConsumerGroup(group))
	}
	return kotel.NewKotel(
		kotel.WithMeter(kotel.NewMeter(kotel.MeterProvider(observer.meterProvider))),
		kotel.WithTracer(kotel.NewTracer(tracerOptions...)),
	).Hooks()
}

// StartAttempt starts one root consumer span linked to unique upstream record
// contexts and updates attempt metrics.
func (observer *OTelObserver) StartAttempt(ctx context.Context, attempt Attempt) (context.Context, AttemptDone) {
	attributes := observer.attemptAttributes(attempt)
	observer.attempts.Add(ctx, 1, metric.WithAttributes(attributes...))
	observer.inflight.Add(ctx, 1, metric.WithAttributes(attributes...))
	started := time.Now()
	spanCtx, span := observer.tracer.Start(
		ctx,
		"kafka "+string(attempt.Phase)+" attempt",
		trace.WithNewRoot(),
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(attributes...),
		trace.WithLinks(observer.links(attempt.Records)...),
	)
	var once sync.Once
	return spanCtx, func(err error) {
		once.Do(func() {
			observer.inflight.Add(spanCtx, -1, metric.WithAttributes(attributes...))
			observer.duration.Record(
				spanCtx,
				time.Since(started).Seconds(),
				metric.WithAttributes(attributes...),
			)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "operation failed")
			}
			span.End()
		})
	}
}

// Retry records one retryable failure. Error text is not used as a metric
// attribute.
func (observer *OTelObserver) Retry(ctx context.Context, event RetryEvent) {
	observer.retries.Add(ctx, 1, metric.WithAttributes(attribute.String("kafka.phase", string(event.Phase))))
}

// BatchCompleted records batch size and outcome counters after commit.
func (observer *OTelObserver) BatchCompleted(ctx context.Context, result BatchResult) {
	observer.batchSize.Record(ctx, int64(result.Size))
	observer.addMessages(ctx, "processed", result.Processed)
	observer.addMessages(ctx, "skipped", result.Skipped)
	observer.addMessages(ctx, "rejected", result.Rejected)
}

// RecordDisposition records low-cardinality disposition counts.
func (observer *OTelObserver) RecordDisposition(ctx context.Context, disposition RecordDisposition) {
	observer.dispositions.Add(ctx, 1, metric.WithAttributes(
		attribute.String("kafka.disposition", string(disposition.Kind)),
		attribute.String("messaging.destination.name", disposition.Record.Topic),
	))
}

func (observer *OTelObserver) addMessages(ctx context.Context, outcome string, count int) {
	if count > 0 {
		observer.messages.Add(ctx, int64(count), metric.WithAttributes(attribute.String("kafka.outcome", outcome)))
	}
}

func (observer *OTelObserver) attemptAttributes(attempt Attempt) []attribute.KeyValue {
	attributes := []attribute.KeyValue{
		attribute.String("kafka.phase", string(attempt.Phase)),
		attribute.Int("kafka.attempt", int(attempt.Attempt)),
		attribute.Int("messaging.batch.message_count", len(attempt.Records)),
	}
	if len(attempt.Records) == 0 {
		return attributes
	}
	first := attempt.Records[0]
	attributes = append(attributes, attribute.String("messaging.destination.name", first.Topic))
	if observer.partition {
		attributes = append(attributes, attribute.Int("messaging.kafka.partition", int(first.Partition)))
	}
	return attributes
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
