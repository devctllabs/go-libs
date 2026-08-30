# kafkaoutbox

Transactional outbox primitives for PostgreSQL and Kafka. The module supports
two explicit delivery profiles:

- `PollingStore` plus `Worker`: an application-owned publisher with leased
  virtual shards, whole-batch retry, and fenced deletion.
- `CDCStore` plus `PartitionManager`: an append-only, time-partitioned table
  for Debezium's Outbox Event Router.

Both stores implement the same `Appender` contract and require an active
`postgresdb.Endpoint.WithinTx` transaction. Migrations are embedded but remain
caller-owned through `PollingMigrations()` and `CDCMigrations()`.

## Enqueue

The typed enqueuer encodes before appending. The application transaction is the
only transaction boundary:

```go
store, err := kafkaoutbox.NewPollingStore(db.Writer())
if err != nil {
    return err
}
enqueuer, err := kafkaoutbox.NewEnqueuer(
    kafkaoutbox.EnqueuerConfig{Observer: observer},
    store,
    kafka.NewJSONEncoder[OrderPaid](),
)
if err != nil {
    return err
}

return db.Writer().WithinTx(ctx, func(txCtx context.Context) error {
    if err := updateOrder(txCtx); err != nil {
        return err
    }
    return enqueuer.Enqueue(txCtx, kafkaoutbox.Event[OrderPaid]{
        Topic: "orders.events", AggregateType: "Order", AggregateID: "42",
        Type: "OrderPaid", Value: event,
    })
})
```

`AggregateType + "::" + AggregateID` is the Kafka key and routing identity.
One aggregate therefore always maps to one virtual shard and, with normal
Kafka key partitioning, one Kafka partition. The delimiter is reserved.

The sibling codec choices are `kafka.NewBytesEncoder()`,
`kafka.NewJSONEncoder[T]()`, and
`kafkaproto.NewMessageEncoder[*mypb.Event]()`. A nil encoded byte slice is
stored as an empty non-null payload.

## Polling worker

```go
worker, err := kafkaoutbox.NewWorker(kafkaoutbox.WorkerConfig{
    MaxBatchSize:    100,
    PollInterval:    100 * time.Millisecond,
    DatabaseTimeout: 2 * time.Second,
    PublishTimeout:  10 * time.Second,
    LeaseDuration:   15 * time.Second,
    RetryPolicy:     backoff,
    MaxAttempts:     0, // retry forever; non-zero stops Run after this many consecutive failures
    Topology: kafkaoutbox.TopologyConfig{
        Revision:   1,
        ShardCount: 4,
    },
    Observer: observer,
}, store, publisher)
if err != nil {
    return err
}
return worker.Run(ctx)
```

Four virtual shards are the default and a good starting point for moderate
load. Shard count bounds publisher concurrency: use at least as many shards as
the maximum useful worker count, and increase to 8 or 16 only when metrics show
one worker or shard is saturated. Counts are powers of two from 1 through
1024. Changing the count requires a strictly higher topology revision.

Workers use `FOR UPDATE SKIP LOCKED` only to lease tiny rows in
`outbox_shards`; event scans themselves need no row locks because one live
lease owns a virtual shard. Kafka I/O never holds a database transaction. A
successful finalize locks and fences the exact generation/shard/token before
deleting the acknowledged IDs.

There is deliberately no `ready` flag. Shards are selected by durable
`next_attempt_at`, and an empty or partial scan schedules another poll. An
enqueue does not need to race with a worker that clears a readiness bit, so
there is no lost-wakeup window. The finalize transaction releases the current
lease and, while the worker is continuing, claims the oldest next due shard.
Updating `next_attempt_at` after each scan makes this atomic context switch
rotate away from a hot shard instead of pinning the worker to it.

Delivery is **at least once**. A crash after Kafka acknowledges a batch but
before PostgreSQL deletes it sends the same event again. The UUIDv7 `id` header
is the consumer deduplication key. Whole batches retry unchanged. With
`MaxAttempts == 0`, failures retry forever; a positive limit returns
`AttemptsExhaustedError` and leaves events intact. There is no publisher DLQ:
moving an unpublished source event to another topic would turn an infrastructure
failure into data loss. The consumer-side `kafka` module owns reject/drop/DLQ
policy after an event reaches Kafka.

## CDC profile

Apply `CDCMigrations()`, run `PartitionManager.Maintain` on startup and on a
schedule, then pass `CDCStore` to the same enqueuer. The manager creates UTC
daily, weekly, or monthly partitions ahead of time under an advisory lock and
drops only partitions whose complete upper bound is older than retention.
Retention must exceed the worst credible connector outage and replication lag.

[`testdata/debezium-postgres.json`](testdata/debezium-postgres.json) is a
starting connector fragment. Important choices are:

- include only the outbox table and use `publish.via.partition.root=true`, so
  child partitions appear as the root table;
- route by `topic`, key by `aggregatekey`, and copy `type` and
  `aggregatetype` into headers;
- use `BinaryDataConverter` so PostgreSQL `bytea` reaches Kafka unchanged;
- read `tracingspancontext` and set `tracing.with.context.field.only=true`.

Use one polling or CDC profile per PostgreSQL schema/search path. The embedded
SQL intentionally uses stable unqualified table names rather than adding a
runtime schema dimension to every query.

## Observability

Core code does not log. Compose `NewOTelObserver` and the sibling
`kafkaoutboxzap` observer with `NewMultiObserver`.

The OTel observer emits:

- `kafkaoutbox.operation.attempts`, `.duration`, and `.inflight`, by bounded
  phase and outcome;
- `kafkaoutbox.enqueue.events`;
- `kafkaoutbox.worker.batches`, `.events`, `.batch.size`, and `.retries`;
- `kafkaoutbox.worker.fencing_conflicts`, kept separate from delivery outcome;
- `kafkaoutbox.topology.changes` and `.shards`.

Rates such as event or batch throughput should be derived from monotonic
counters in the metrics backend. Raw errors are recorded on spans and logs,
never metric attributes. Polling publish spans are new producer roots linked
to the persisted originating W3C contexts, capped at 128 unique links by
default. CDC stores the same context in Debezium's Java Properties format.

The zap adapter logs scheduled retries and lost claims at Warn, and actual
topology changes at Info. The application remains responsible for logging a
terminal `Run` error once.

## Inbox decision

No transactional inbox is included in v1. The outbox cannot provide
exactly-once business effects: consumers still need idempotency because
at-least-once delivery can repeat an event. Start with a domain uniqueness
constraint or a small application-owned processed-event table keyed by the
outbox `id`. A generic inbox becomes worthwhile only when multiple services
need the same atomic "deduplicate + mutate business state" transaction and can
share database semantics; adding it pre-emptively would couple this publisher
library to consumer storage and retention policy.

## Tests

```sh
go test -race ./kafkaoutbox/... ./kafkaoutboxzap/...
go test -race -tags=integration ./kafkaoutbox/...
```
