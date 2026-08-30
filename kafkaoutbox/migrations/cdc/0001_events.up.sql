CREATE TABLE outbox_events (
    id uuid NOT NULL,
    topic text NOT NULL,
    aggregatetype text NOT NULL,
    aggregateid text NOT NULL,
    aggregatekey text NOT NULL,
    type text NOT NULL,
    payload bytea NOT NULL,
    tracingspancontext text,
    created_at timestamptz NOT NULL DEFAULT statement_timestamp(),
    PRIMARY KEY (created_at, id)
) PARTITION BY RANGE (created_at);

