package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errClaimLost = errors.New("kafkaoutbox: shard claim lost")

type shardClaim struct {
	generation uint64
	id         uint
	from       *int64
	to         *int64
	cursor     *int64
	token      uuid.UUID
}

type storedEvent struct {
	id            uuid.UUID
	topic         string
	aggregateType string
	aggregateKey  string
	eventType     string
	payload       []byte
	routingHash   int64
	traceparent   *string
	tracestate    *string
}

func (store *PollingStore) reconcileTopology(ctx context.Context, config WorkerConfig) (TopologyResult, error) {
	operationCtx, cancel := context.WithTimeout(ctx, config.DatabaseTimeout)
	defer cancel()
	var result TopologyResult
	err := store.endpoint.WithinTx(operationCtx, func(txCtx context.Context) error {
		_, err := store.endpoint.Exec(txCtx, `
			SELECT pg_advisory_xact_lock(
				($1::bigint << 32) | 'outbox_topology'::regclass::oid::bigint
			)
		`, config.Topology.AdvisoryLockNamespace)
		if err != nil {
			return fmt.Errorf("endpoint.Exec advisory lock: %w", err)
		}
		var revision, generation uint64
		var shardCount uint
		err = store.endpoint.QueryRow(txCtx, `
			SELECT revision, generation, shard_count
			FROM outbox_topology
			WHERE singleton = true
			FOR UPDATE
		`).Scan(&revision, &generation, &shardCount)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			result = TopologyResult{
				Revision: config.Topology.Revision, Generation: 1,
				ShardCount: config.Topology.ShardCount, Changed: true,
			}
			return store.replaceTopology(txCtx, result.Revision, result.Generation, result.ShardCount)
		case err != nil:
			return fmt.Errorf("endpoint.QueryRow topology: %w", err)
		case config.Topology.Revision < revision:
			result = TopologyResult{Revision: revision, Generation: generation, ShardCount: shardCount}
			return nil
		case config.Topology.Revision == revision && config.Topology.ShardCount != shardCount:
			return errors.New("configured shard count conflicts with persisted topology revision")
		case config.Topology.Revision == revision:
			result = TopologyResult{Revision: revision, Generation: generation, ShardCount: shardCount}
			return nil
		default:
			result = TopologyResult{
				Revision: config.Topology.Revision, Generation: generation + 1,
				ShardCount: config.Topology.ShardCount, Changed: true,
			}
			return store.replaceTopology(txCtx, result.Revision, result.Generation, result.ShardCount)
		}
	})
	return result, err
}

func (store *PollingStore) replaceTopology(ctx context.Context, revision, generation uint64, count uint) error {
	if _, err := store.endpoint.Exec(ctx, `DELETE FROM outbox_shards`); err != nil {
		return fmt.Errorf("endpoint.Exec delete shards: %w", err)
	}
	if _, err := store.endpoint.Exec(ctx, `
		INSERT INTO outbox_topology (singleton, revision, generation, shard_count)
		VALUES (true, $1, $2, $3)
		ON CONFLICT (singleton) DO UPDATE
		SET revision = EXCLUDED.revision,
		    generation = EXCLUDED.generation,
		    shard_count = EXCLUDED.shard_count
	`, revision, generation, count); err != nil {
		return fmt.Errorf("endpoint.Exec upsert topology: %w", err)
	}
	for shardID := uint(0); shardID < count; shardID++ {
		from, to := shardBounds(shardID, count)
		if _, err := store.endpoint.Exec(ctx, `
			INSERT INTO outbox_shards (generation, shard_id, hash_from, hash_to)
			VALUES ($1, $2, $3, $4)
		`, generation, shardID, from, to); err != nil {
			return fmt.Errorf("endpoint.Exec insert shard %d: %w", shardID, err)
		}
	}
	return nil
}

func shardBounds(shardID, count uint) (*int64, *int64) {
	var from, to *int64
	step := ^uint64(0)/uint64(count) + 1
	if shardID > 0 {
		value := int64(uint64(shardID)*step ^ (uint64(1) << 63))
		from = &value
	}
	if shardID+1 < count {
		value := int64(uint64(shardID+1)*step ^ (uint64(1) << 63))
		to = &value
	}
	return from, to
}

func (store *PollingStore) claimShard(ctx context.Context, config WorkerConfig) (*shardClaim, error) {
	operationCtx, cancel := context.WithTimeout(ctx, config.DatabaseTimeout)
	defer cancel()
	var claim *shardClaim
	err := store.endpoint.WithinTx(operationCtx, func(txCtx context.Context) error {
		var err error
		claim, err = store.claimShardTx(txCtx, config)
		return err
	})
	return claim, err
}

func (store *PollingStore) claimShardTx(ctx context.Context, config WorkerConfig) (*shardClaim, error) {
	row := store.endpoint.QueryRow(ctx, `
		SELECT generation, shard_id, hash_from, hash_to, scan_after_hash
		FROM outbox_shards
		WHERE next_attempt_at <= clock_timestamp()
		  AND (owner_token IS NULL OR lease_until <= clock_timestamp())
		ORDER BY next_attempt_at, shard_id
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`)
	candidate := shardClaim{}
	if err := row.Scan(&candidate.generation, &candidate.id, &candidate.from, &candidate.to, &candidate.cursor); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("endpoint.QueryRow due shard: %w", err)
	}
	token, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("uuid.NewV7: %w", err)
	}
	candidate.token = token
	tag, err := store.endpoint.Exec(ctx, `
		UPDATE outbox_shards
		SET owner_token = $3,
		    lease_until = clock_timestamp() + $4 * interval '1 microsecond'
		WHERE generation = $1 AND shard_id = $2
	`, candidate.generation, candidate.id, candidate.token, config.LeaseDuration.Microseconds())
	if err != nil {
		return nil, fmt.Errorf("endpoint.Exec claim shard: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return nil, errClaimLost
	}
	return &candidate, nil
}

func (store *PollingStore) readBatch(ctx context.Context, claim shardClaim, config WorkerConfig) ([]storedEvent, error) {
	operationCtx, cancel := context.WithTimeout(ctx, config.DatabaseTimeout)
	defer cancel()
	batch, err := store.queryEvents(operationCtx, claim, false, config.MaxBatchSize)
	if err != nil || claim.cursor == nil || len(batch) == config.MaxBatchSize {
		return batch, err
	}
	wrapped, err := store.queryEvents(operationCtx, claim, true, config.MaxBatchSize-len(batch))
	return append(batch, wrapped...), err
}

func (store *PollingStore) queryEvents(ctx context.Context, claim shardClaim, wrapped bool, limit int) ([]storedEvent, error) {
	comparison := `($3::bigint IS NULL OR routing_hash > $3)`
	if wrapped {
		comparison = `routing_hash <= $3`
	}
	rows, err := store.endpoint.Query(ctx, `
		SELECT id, topic, aggregatetype, aggregatekey, type, payload,
		       routing_hash, traceparent, tracestate
		FROM outbox_events
		WHERE ($1::bigint IS NULL OR routing_hash >= $1)
		  AND ($2::bigint IS NULL OR routing_hash < $2)
		  AND `+comparison+`
		ORDER BY routing_hash, id
		LIMIT $4
	`, claim.from, claim.to, claim.cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("endpoint.Query events: %w", err)
	}
	defer rows.Close()
	batch := make([]storedEvent, 0, limit)
	for rows.Next() {
		event := storedEvent{}
		if err := rows.Scan(
			&event.id, &event.topic, &event.aggregateType, &event.aggregateKey,
			&event.eventType, &event.payload, &event.routingHash,
			&event.traceparent, &event.tracestate,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan event: %w", err)
		}
		batch = append(batch, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err events: %w", err)
	}
	return batch, nil
}

func (store *PollingStore) finalizeEmpty(
	ctx context.Context,
	claim shardClaim,
	claimNext bool,
	config WorkerConfig,
) (*shardClaim, error) {
	return store.finalize(ctx, claim, nil, claim.cursor, config.PollInterval, claimNext, config)
}

func (store *PollingStore) finalizeSuccess(
	ctx context.Context,
	claim shardClaim,
	batch []storedEvent,
	claimNext bool,
	config WorkerConfig,
) (*shardClaim, error) {
	delay := time.Duration(0)
	if len(batch) < config.MaxBatchSize {
		delay = config.PollInterval
	}
	return store.finalize(
		ctx, claim, eventIDs(batch), &batch[len(batch)-1].routingHash, delay, claimNext, config,
	)
}

func (store *PollingStore) finalizeFailure(
	ctx context.Context,
	claim shardClaim,
	config WorkerConfig,
) (uint, time.Duration, *shardClaim, error) {
	operationCtx, cancel := context.WithTimeout(ctx, config.DatabaseTimeout)
	defer cancel()
	var failures uint
	var delay time.Duration
	var nextClaim *shardClaim
	err := store.endpoint.WithinTx(operationCtx, func(txCtx context.Context) error {
		current, err := store.lockClaim(txCtx, claim)
		if err != nil {
			return err
		}
		failures = current + 1
		delay = config.PollInterval
		if config.RetryPolicy != nil {
			delay = config.RetryPolicy.Delay(failures)
			if delay < 0 {
				return errors.New("retry policy returned a negative delay")
			}
		}
		tag, err := store.endpoint.Exec(txCtx, `
			UPDATE outbox_shards
			SET owner_token = NULL,
			    lease_until = NULL,
			    next_attempt_at = clock_timestamp() + $4 * interval '1 microsecond',
			    failure_count = $5
			WHERE generation = $1 AND shard_id = $2 AND owner_token = $3
		`, claim.generation, claim.id, claim.token, delay.Microseconds(), failures)
		if err != nil {
			return fmt.Errorf("endpoint.Exec fail shard: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return errClaimLost
		}
		if config.MaxAttempts != 0 && failures >= config.MaxAttempts {
			return nil
		}
		nextClaim, err = store.claimShardTx(txCtx, config)
		return err
	})
	return failures, delay, nextClaim, err
}

func (store *PollingStore) finalize(
	ctx context.Context,
	claim shardClaim,
	ids []uuid.UUID,
	cursor *int64,
	delay time.Duration,
	claimNext bool,
	config WorkerConfig,
) (*shardClaim, error) {
	operationCtx, cancel := context.WithTimeout(ctx, config.DatabaseTimeout)
	defer cancel()
	var nextClaim *shardClaim
	err := store.endpoint.WithinTx(operationCtx, func(txCtx context.Context) error {
		if _, err := store.lockClaim(txCtx, claim); err != nil {
			return err
		}
		if len(ids) > 0 {
			if err := store.deleteEvents(txCtx, ids); err != nil {
				return fmt.Errorf("endpoint.Exec delete events: %w", err)
			}
		}
		tag, err := store.endpoint.Exec(txCtx, `
			UPDATE outbox_shards
			SET scan_after_hash = $4,
			    owner_token = NULL,
			    lease_until = NULL,
			    next_attempt_at = clock_timestamp() + $5 * interval '1 microsecond',
			    failure_count = 0
			WHERE generation = $1 AND shard_id = $2 AND owner_token = $3
		`, claim.generation, claim.id, claim.token, cursor, delay.Microseconds())
		if err != nil {
			return fmt.Errorf("endpoint.Exec finalize shard: %w", err)
		}
		if tag.RowsAffected() != 1 {
			return errClaimLost
		}
		if !claimNext {
			return nil
		}
		nextClaim, err = store.claimShardTx(txCtx, config)
		return err
	})
	return nextClaim, err
}

func (store *PollingStore) lockClaim(ctx context.Context, claim shardClaim) (uint, error) {
	var failures uint
	err := store.endpoint.QueryRow(ctx, `
		SELECT failure_count
		FROM outbox_shards
		WHERE generation = $1 AND shard_id = $2 AND owner_token = $3
		FOR UPDATE
	`, claim.generation, claim.id, claim.token).Scan(&failures)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, errClaimLost
	}
	if err != nil {
		return 0, fmt.Errorf("endpoint.QueryRow fence: %w", err)
	}
	return failures, nil
}

func (store *PollingStore) deleteEvents(ctx context.Context, ids []uuid.UUID) error {
	placeholders := make([]string, len(ids))
	arguments := make([]any, len(ids))
	for index, id := range ids {
		placeholders[index] = fmt.Sprintf("$%d::uuid", index+1)
		arguments[index] = id.String()
	}
	_, err := store.endpoint.Exec(
		ctx,
		`DELETE FROM outbox_events WHERE id IN (`+strings.Join(placeholders, ",")+`)`,
		arguments...,
	)
	return err
}

func eventIDs(batch []storedEvent) []uuid.UUID {
	ids := make([]uuid.UUID, len(batch))
	for index, event := range batch {
		ids[index] = event.id
	}
	return ids
}
