# kafka

Typed, instance-owned Kafka producer and consumer runtimes built on
[`franz-go`](https://github.com/twmb/franz-go).

`franz-go` is the base client because it is pure Go, actively implements the
Kafka protocol, exposes precise polling and commit control, supports Kafka
transactions when an application needs them directly, and has first-party
OpenTelemetry hooks. `segmentio/kafka-go` has a smaller API and is a reasonable
choice for simple consumers, while Confluent's Go client is appropriate when
`librdkafka` and its native deployment dependency are already standard. This
wrapper deliberately targets `franz-go` only.

## Delivery model

- Delivery is **at least once**. The library does not claim exactly-once
  processing and does not enable transactions.
- A batch is committed only after decode policy, a successful handler attempt,
  DLQ delivery when required, and commit retries have completed.
- Handler retries repeat the whole decoded batch. Values are decoded once;
  `Skip` and `Reject` decisions from a failed attempt are cleared before the
  next attempt.
- Commit retries never rerun the handler or DLQ phase.
- `retry.Permanent(err)` stops handler retries immediately.
- `Run` is single-use and owns client shutdown. Plain context cancellation is a
  clean stop; an active partial batch is drained under `ShutdownTimeout`.

This is intentionally an atomic scheduling/commit model, not atomic business
side effects. A handler may run more than once after process failure, so
business operations still need idempotency.

## Batching and single-message mode

All consumers use one `BatchHandler[T]` contract. Set `Batch.MaxSize` and
`Batch.FlushInterval` to seal a batch on the first reached boundary. For
single-message processing use `MaxSize: 1`; `FlushInterval` may then be zero.
There is no second single-record handler abstraction because it would duplicate
retry, reject, and commit semantics.

```go
consumer, err := kafka.NewConsumer(
    kafka.ConsumerConfig{
        Brokers: []string{"localhost:9092"},
        Group:   "invoice-worker",
        Topics:  []string{"invoices"},
        Batch: kafka.BatchConfig{
            MaxSize:       100,
            FlushInterval: 250 * time.Millisecond,
        },
        Retry: kafka.RetryConfig{
            Policy:      backoff,
            MaxAttempts: 5,
        },
        OnReject:              kafka.RejectStop,
        RebalanceTimeout:      time.Minute,
        RebalanceDrainTimeout: 30 * time.Second,
        ShutdownTimeout:       time.Minute,
    },
    decoder,
    kafka.BatchHandlerFunc[Invoice](func(ctx context.Context, batch *kafka.Batch[Invoice]) error {
        for index, message := range batch.Messages() {
            if alreadyProcessed(message.Value.ID) {
                if err := batch.Skip(index); err != nil {
                    return err
                }
            }
        }
        return process(ctx, batch.Messages())
    }),
)
if err != nil {
    return err
}
return consumer.Run(ctx)
```

`Messages()` is a borrowed immutable view. Handlers must not mutate or retain
the slice, headers, keys, or values after returning.

## Codecs

Use `NewJSONEncoder[T]` and `NewJSONDecoder[T]` for standard JSON payloads.
The decoder creates a fresh value for each record and is safe to share across
partition workers. Use `NewBytesEncoder` only when the payload is already in
its final wire format; it avoids a copy, so the caller must not mutate the
slice until the send or enqueue operation returns.

## Reject and DLQ policies

Decode failures are permanent rejects and never enter the generic handler.
The handler can reject a decoded record with `batch.Reject(index, cause)`.
Every consumer must explicitly choose one policy:

| Policy | Result |
|---|---|
| `RejectStop` | Return `RejectedMessageError`; do not commit the input batch. |
| `RejectDrop` | Commit the input batch and emit a dropped disposition. |
| `RejectDLQ` | Publish the original record to the configured same-cluster topic, then commit. |

DLQ records preserve the original key, value, timestamp, and non-reserved
headers. Reserved provenance headers contain original topic, partition, offset,
and failure kind (`decode` or `handler`). Error text is never written to Kafka.
`DLQFailureStop` favors durability; `DLQFailureDrop` favors continued processing
and emits a dropped disposition. DLQ and commit retries inherit the baseline
retry config unless an explicit override is supplied.

## Partition isolation

`NewPartitionedConsumer` creates one `PartitionHandler` per observed topic
partition. Different partitions may run concurrently; each handler is called
sequentially and is closed under a bounded context. `MaxConcurrentPartitions`
sets the concurrency limit and zero means unlimited. Per-partition partial
batches accumulate until their own size or interval boundary. Offset commit
calls are serialized.

## Producer

`Producer.Send` and `SendBatch` encode the complete input before sending any
record, wait synchronously for broker acknowledgements, and use all-ISR acks.
The producer is safe for concurrent sends. Partial broker delivery is returned
as `BatchDeliveryError`, with per-input-index failure metadata and a success
count. `Close(ctx)` prevents new sends, waits for active sends, and is
idempotent.

## Observability

Core code does not log implicitly. Configure an `Observer`, or compose several
with `NewMultiObserver`.

`NewOTelObserver` supplies franz-go's standard broker/fetch/produce metrics and
propagation hooks plus wrapper instruments:

- `kafka.consumer.operation.attempts`
- `kafka.consumer.operation.retries`
- `kafka.consumer.operation.inflight`
- `kafka.consumer.operation.duration`
- `kafka.consumer.batch.size`
- `kafka.consumer.messages` with outcome
- `kafka.consumer.record.dispositions`

Message throughput is the rate of the monotonic `kafka.consumer.messages`
counter; a separate rate gauge would be aggregation-window dependent and is
therefore not emitted. Exact consumer lag should come from committed offsets
versus broker log-end offsets in the monitoring backend; the runtime does not
guess it from locally fetched records.

Each handler, DLQ, and commit attempt creates a consumer span. A span is linked
to unique upstream record contexts, capped at 128 links by default. Error text
is recorded on spans but never used as a metric attribute.

The sibling `kafkazap` module logs every retry at Warn, confirmed DLQ delivery
at Info, and drops at Warn. Terminal `Run` errors are returned, not logged, so
the application has one owner for terminal error reporting. Treat rejection
causes as potentially sensitive when configuring log sinks.

The sibling `kafkaproto` module supplies concurrency-safe protobuf encoders and
fresh-message decoders without adding protobuf to the core module.

## Tests

```sh
go test -race ./kafka/... ./kafkaproto/... ./kafkazap/...
golangci-lint run ./kafka/... ./kafkaproto/... ./kafkazap/...
```

The default suite includes an in-memory `kfake` protocol round-trip. A real
broker smoke test is opt-in:

```sh
KAFKA_BROKERS=localhost:9092 \
KAFKA_TEST_TOPIC=go-libs-kafka-integration \
go test -tags=kafka_integration ./kafka/...
```
