CREATE TABLE outbox_events (
    id uuid NOT NULL PRIMARY KEY,
    topic text NOT NULL,
    aggregatetype text NOT NULL,
    aggregateid text NOT NULL,
    aggregatekey text NOT NULL,
    type text NOT NULL,
    payload bytea NOT NULL,
    routing_hash bigint NOT NULL,
    traceparent text,
    tracestate text,
    created_at timestamptz NOT NULL DEFAULT statement_timestamp()
);

CREATE INDEX outbox_events_routing_idx ON outbox_events (routing_hash, id);

