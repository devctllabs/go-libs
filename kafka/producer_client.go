package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

//go:generate go tool mockgen -source=producer_client.go -destination=producer_client.gen_test.go -package=kafka -typed

type recordProducer interface {
	// ProduceSync sends records and waits for all delivery results.
	ProduceSync(ctx context.Context, records ...*kgo.Record) kgo.ProduceResults
	// Close closes the owned Kafka client.
	Close()
}
