package kafka

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/mock/gomock"
)

func TestPartitionedConsumerNeverMixesPartitionsAndClosesHandlers(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	first := &kgo.Record{Topic: "invoices", Partition: 1, Offset: 10, Value: []byte("first")}
	second := &kgo.Record{Topic: "invoices", Partition: 2, Offset: 20, Value: []byte("second")}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gomock.InOrder(
		client.EXPECT().PollRecords(gomock.Any(), 2).Return(kgo.Fetches{{Topics: []kgo.FetchTopic{{
			Topic: "invoices",
			Partitions: []kgo.FetchPartition{
				{Partition: 1, Records: []*kgo.Record{first}},
				{Partition: 2, Records: []*kgo.Record{second}},
			},
		}}}}),
		client.EXPECT().PollRecords(gomock.Any(), 2).DoAndReturn(
			func(ctx context.Context, _ int) kgo.Fetches {
				<-ctx.Done()
				return nil
			},
		),
	)
	decoder.EXPECT().Decode(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, value []byte) (string, error) { return string(value), nil },
	).Times(2)

	var mu sync.Mutex
	handled := make(map[TopicPartition][]int64)
	closed := make(map[TopicPartition]int)
	factory := PartitionHandlerFactoryFunc[string](func(
		_ context.Context,
		partition TopicPartition,
	) (PartitionHandler[string], error) {
		return partitionHandlerFuncs[string]{
			handle: func(_ context.Context, batch *Batch[string]) error {
				mu.Lock()
				defer mu.Unlock()
				for _, message := range batch.Messages() {
					require.Equal(t, TopicPartition{Topic: message.Topic, Partition: message.Partition}, partition)
					handled[partition] = append(handled[partition], message.Offset)
				}
				return nil
			},
			close: func(context.Context) error {
				mu.Lock()
				defer mu.Unlock()
				closed[partition]++
				return nil
			},
		}, nil
	})

	commits := 0
	client.EXPECT().CommitRecords(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, _ ...*kgo.Record) error {
			mu.Lock()
			defer mu.Unlock()
			commits++
			if commits == 2 {
				cancel()
			}
			return nil
		},
	).Times(2)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 2, FlushInterval: time.Millisecond}
	consumer := newPartitionedConsumerWithClient(
		PartitionedConsumerConfig{Consumer: cfg, MaxConcurrentPartitions: 2},
		client,
		decoder,
		factory,
	)

	require.NoError(t, consumer.Run(runCtx))
	mu.Lock()
	defer mu.Unlock()
	require.Equal(t, []int64{10}, handled[TopicPartition{Topic: "invoices", Partition: 1}])
	require.Equal(t, []int64{20}, handled[TopicPartition{Topic: "invoices", Partition: 2}])
	require.Equal(t, 1, closed[TopicPartition{Topic: "invoices", Partition: 1}])
	require.Equal(t, 1, closed[TopicPartition{Topic: "invoices", Partition: 2}])
}

type partitionHandlerFuncs[T any] struct {
	handle func(ctx context.Context, batch *Batch[T]) error
	close  func(ctx context.Context) error
}

func (handler partitionHandlerFuncs[T]) Handle(ctx context.Context, batch *Batch[T]) error {
	return handler.handle(ctx, batch)
}

func (handler partitionHandlerFuncs[T]) Close(ctx context.Context) error {
	return handler.close(ctx)
}
