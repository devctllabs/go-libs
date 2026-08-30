//go:build kafka_integration

package kafka_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/kafka"
	"github.com/stretchr/testify/require"
)

func TestProducerAgainstRealKafka(t *testing.T) {
	brokersValue := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if brokersValue == "" {
		t.Skip("KAFKA_BROKERS is not set")
	}
	topic := strings.TrimSpace(os.Getenv("KAFKA_TEST_TOPIC"))
	if topic == "" {
		topic = "go-libs-kafka-integration"
	}
	producer, err := kafka.NewProducer(
		kafka.ProducerConfig{Brokers: strings.Split(brokersValue, ",")},
		kafka.EncoderFunc[string](func(_ context.Context, value string) ([]byte, error) {
			return []byte(value), nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, producer.Close(context.Background())) })

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	require.NoError(t, producer.Send(ctx, kafka.OutgoingMessage[string]{
		Topic: topic,
		Key:   []byte("go-libs-smoke"),
		Value: "ok",
	}))
}
