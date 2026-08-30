package kafkaoutbox

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/devctllabs/go-libs/postgresdb"
	"github.com/jackc/pgx/v5"
)

// PartitionGranularity selects UTC calendar boundaries for CDC partitions.
type PartitionGranularity uint8

const (
	// PartitionDaily creates one partition per UTC day.
	PartitionDaily PartitionGranularity = iota + 1
	// PartitionWeekly creates one partition per UTC week starting Monday.
	PartitionWeekly
	// PartitionMonthly creates one partition per UTC calendar month.
	PartitionMonthly
)

// PartitionManagerConfig controls CDC partition creation and retention.
type PartitionManagerConfig struct {
	Granularity           PartitionGranularity
	AheadPartitions       uint
	Retention             time.Duration
	AdvisoryLockNamespace int32
}

// PartitionManager maintains the caller's CDC outbox partitions.
type PartitionManager struct {
	endpoint *postgresdb.Endpoint
	config   PartitionManagerConfig
}

// NewPartitionManager constructs a caller-driven CDC partition manager.
func NewPartitionManager(endpoint *postgresdb.Endpoint, config PartitionManagerConfig) (*PartitionManager, error) {
	if endpoint == nil {
		return nil, errors.New("kafkaoutbox: PostgreSQL endpoint must not be nil")
	}
	if config.Granularity < PartitionDaily || config.Granularity > PartitionMonthly {
		return nil, errors.New("kafkaoutbox: partition granularity is required")
	}
	if config.Retention <= 0 {
		return nil, errors.New("kafkaoutbox: partition retention must be positive")
	}
	return &PartitionManager{endpoint: endpoint, config: config}, nil
}

// Maintain creates the current and configured future UTC partitions.
func (manager *PartitionManager) Maintain(ctx context.Context, now time.Time) error {
	if ctx == nil {
		return errors.New("kafkaoutbox: context must not be nil")
	}
	if now.IsZero() {
		return errors.New("kafkaoutbox: maintenance time must not be zero")
	}
	return manager.endpoint.WithinTx(ctx, func(txCtx context.Context) error {
		if _, err := manager.endpoint.Exec(txCtx, `
			SELECT pg_advisory_xact_lock(
				($1::bigint << 32) | 'outbox_events'::regclass::oid::bigint
			)
		`, manager.config.AdvisoryLockNamespace); err != nil {
			return fmt.Errorf("endpoint.Exec advisory lock: %w", err)
		}
		start := partitionStart(now, manager.config.Granularity)
		for offset := uint(0); offset <= manager.config.AheadPartitions; offset++ {
			end := nextPartition(start, manager.config.Granularity)
			if err := manager.createPartition(txCtx, start, end); err != nil {
				return err
			}
			start = end
		}
		return manager.dropExpiredPartitions(txCtx, now.UTC().Add(-manager.config.Retention))
	})
}

func (manager *PartitionManager) createPartition(ctx context.Context, start, end time.Time) error {
	name := "outbox_events_p" + start.Format("20060102")
	query := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS %s PARTITION OF outbox_events FOR VALUES FROM ('%s') TO ('%s')",
		pgx.Identifier{name}.Sanitize(),
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
	)
	if _, err := manager.endpoint.Exec(ctx, query); err != nil {
		return fmt.Errorf("endpoint.Exec create partition %s: %w", name, err)
	}
	return nil
}

func (manager *PartitionManager) dropExpiredPartitions(ctx context.Context, cutoff time.Time) error {
	rows, err := manager.endpoint.Query(ctx, `
		SELECT child.relname
		FROM pg_inherits
		JOIN pg_class child ON child.oid = inhrelid
		WHERE inhparent = 'outbox_events'::regclass
	`)
	if err != nil {
		return fmt.Errorf("endpoint.Query partitions: %w", err)
	}
	var expired []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return fmt.Errorf("rows.Scan partition: %w", err)
		}
		start, ok := canonicalPartitionStart(name)
		if ok && !nextPartition(start, manager.config.Granularity).After(cutoff) {
			expired = append(expired, name)
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("rows.Err partitions: %w", err)
	}
	rows.Close()
	for _, name := range expired {
		if _, err := manager.endpoint.Exec(ctx, "DROP TABLE "+pgx.Identifier{name}.Sanitize()); err != nil {
			return fmt.Errorf("endpoint.Exec drop partition %s: %w", name, err)
		}
	}
	return nil
}

func canonicalPartitionStart(name string) (time.Time, bool) {
	const prefix = "outbox_events_p"
	date, ok := strings.CutPrefix(name, prefix)
	if !ok || len(date) != len("20060102") {
		return time.Time{}, false
	}
	start, err := time.Parse("20060102", date)
	return start, err == nil
}

func partitionStart(value time.Time, granularity PartitionGranularity) time.Time {
	value = value.UTC()
	switch granularity {
	case PartitionWeekly:
		start := time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
		daysSinceMonday := (int(start.Weekday()) + 6) % 7
		return start.AddDate(0, 0, -daysSinceMonday)
	case PartitionMonthly:
		return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
	}
}

func nextPartition(start time.Time, granularity PartitionGranularity) time.Time {
	switch granularity {
	case PartitionWeekly:
		return start.AddDate(0, 0, 7)
	case PartitionMonthly:
		return start.AddDate(0, 1, 0)
	default:
		return start.AddDate(0, 0, 1)
	}
}
