CREATE TABLE outbox_topology (
    singleton boolean NOT NULL PRIMARY KEY DEFAULT true CHECK (singleton),
    revision bigint NOT NULL CHECK (revision > 0),
    generation bigint NOT NULL CHECK (generation > 0),
    shard_count integer NOT NULL CHECK (shard_count > 0)
);

CREATE TABLE outbox_shards (
    generation bigint NOT NULL,
    shard_id integer NOT NULL,
    hash_from bigint,
    hash_to bigint,
    scan_after_hash bigint,
    owner_token uuid,
    lease_until timestamptz,
    next_attempt_at timestamptz NOT NULL DEFAULT clock_timestamp(),
    failure_count bigint NOT NULL DEFAULT 0 CHECK (failure_count >= 0),
    PRIMARY KEY (generation, shard_id),
    CHECK (hash_from IS NULL OR hash_to IS NULL OR hash_from < hash_to),
    CHECK ((owner_token IS NULL) = (lease_until IS NULL))
);

CREATE INDEX outbox_shards_due_idx ON outbox_shards (next_attempt_at, shard_id);

