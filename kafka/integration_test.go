package kafka_test

import (
	"context"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/kafka"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
)

func TestProducerRoundTripWithKfake(t *testing.T) {
	t.Parallel()

	cluster, err := kfake.NewCluster(kfake.SeedTopics(1, "invoices"))
	require.NoError(t, err)
	t.Cleanup(cluster.Close)
	producer, err := kafka.NewProducer(
		kafka.ProducerConfig{Brokers: cluster.ListenAddrs()},
		kafka.EncoderFunc[string](func(_ context.Context, value string) ([]byte, error) {
			return []byte(value), nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, producer.Close(context.Background())) })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, producer.Send(ctx, kafka.OutgoingMessage[string]{
		Topic: "invoices", Key: []byte("42"), Value: "invoice-42",
	}))

	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(cluster.ListenAddrs()...),
		kgo.ConsumePartitions(map[string]map[int32]kgo.Offset{
			"invoices": {0: kgo.NewOffset().AtStart()},
		}),
	)
	require.NoError(t, err)
	t.Cleanup(consumer.Close)
	fetches := consumer.PollRecords(ctx, 1)
	require.NoError(t, fetches.Err())
	require.Len(t, fetches.Records(), 1)
	require.Equal(t, []byte("42"), fetches.Records()[0].Key)
	require.Equal(t, []byte("invoice-42"), fetches.Records()[0].Value)
}
