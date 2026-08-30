package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

//go:generate go tool mockgen -source=consumer_client.go -destination=consumer_client.gen_test.go -package=kafka -typed

type consumerClient interface {
	// PollRecords fetches at most maxRecords records.
	PollRecords(ctx context.Context, maxRecords int) kgo.Fetches
	// ProduceSync waits for delivery results for all records.
	ProduceSync(ctx context.Context, records ...*kgo.Record) kgo.ProduceResults
	// CommitRecords synchronously commits records in partition order.
	CommitRecords(ctx context.Context, records ...*kgo.Record) error
	// AllowRebalance permits a rebalance blocked by the latest poll.
	AllowRebalance()
	// Close closes the owned Kafka client.
	Close()
}
