package kafka

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/mock/gomock"
)

func TestProducerEncodesEntireBatchBeforeSending(t *testing.T) {
	t.Parallel()

	encodeErr := errors.New("marshal invoice")
	producer, err := NewProducer(
		ProducerConfig{Brokers: []string{"localhost:9092"}},
		EncoderFunc[string](func(_ context.Context, value string) ([]byte, error) {
			if value == "invalid" {
				return nil, encodeErr
			}
			return []byte(value), nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, producer.Close(context.Background())) })

	err = producer.SendBatch(context.Background(), []OutgoingMessage[string]{
		{Topic: "invoices", Value: "valid"},
		{Topic: "invoices", Value: "invalid"},
	})

	var got *EncodeError
	require.ErrorAs(t, err, &got)
	require.Equal(t, 1, got.Index)
	require.ErrorIs(t, got, encodeErr)
}

func TestProducerReportsPartialDelivery(t *testing.T) {
	t.Parallel()

	deliveryErr := errors.New("message too large")
	client := NewMockrecordProducer(gomock.NewController(t))
	client.EXPECT().ProduceSync(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, records ...*kgo.Record) kgo.ProduceResults {
			records[0].Partition = 1
			records[0].Offset = 10
			records[1].Partition = 2
			return kgo.ProduceResults{
				{Record: records[0]},
				{Record: records[1], Err: deliveryErr},
			}
		},
	)
	producer := newProducerWithClient(client, EncoderFunc[string](
		func(_ context.Context, value string) ([]byte, error) { return []byte(value), nil },
	))

	err := producer.SendBatch(context.Background(), []OutgoingMessage[string]{
		{Topic: "invoices", Value: "first"},
		{Topic: "invoices", Value: "second"},
	})

	var got *BatchDeliveryError
	require.ErrorAs(t, err, &got)
	require.Equal(t, 1, got.Succeeded())
	require.Equal(t, []DeliveryFailure{{
		Index:     1,
		Topic:     "invoices",
		Partition: 2,
		Err:       deliveryErr,
	}}, got.Failures())
	require.ErrorIs(t, got, deliveryErr)
}

func TestProducerRejectsBlankTopicsBeforeSending(t *testing.T) {
	t.Parallel()

	producer, err := NewProducer(
		ProducerConfig{Brokers: []string{"localhost:9092"}},
		EncoderFunc[string](func(_ context.Context, value string) ([]byte, error) {
			return []byte(value), nil
		}),
	)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, producer.Close(context.Background())) })

	err = producer.Send(context.Background(), OutgoingMessage[string]{Value: "invoice"})

	require.ErrorContains(t, err, "topic")
}
