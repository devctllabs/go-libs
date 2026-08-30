package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/mock/gomock"
)

func TestConsumerProcessesAndCommitsFullBatch(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	first := &kgo.Record{Topic: "invoices", Partition: 1, Offset: 10, Value: []byte("first")}
	second := &kgo.Record{Topic: "invoices", Partition: 1, Offset: 11, Value: []byte("second")}
	fetches := fetchedInvoiceRecords(1, first, second)
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.EXPECT().PollRecords(gomock.Any(), 2).Return(fetches)
	decoder.EXPECT().Decode(gomock.Any(), []byte("first")).Return("decoded-first", nil)
	decoder.EXPECT().Decode(gomock.Any(), []byte("second")).Return("decoded-second", nil)
	handled := make(chan []Message[string], 1)
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, batch *Batch[string]) error {
			handled <- append([]Message[string](nil), batch.Messages()...)
			return batch.Skip(1)
		},
	)
	client.EXPECT().CommitRecords(gomock.Any(), first, second).DoAndReturn(
		func(_ context.Context, _ ...*kgo.Record) error {
			cancel()
			return nil
		},
	)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 2, FlushInterval: 10}
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
	require.Equal(t, []Message[string]{
		{Topic: "invoices", Partition: 1, Offset: 10, Value: "decoded-first"},
		{Topic: "invoices", Partition: 1, Offset: 11, Value: "decoded-second"},
	}, <-handled)
}

func TestNewConsumerValidatesDependencies(t *testing.T) {
	t.Parallel()

	cfg := validConsumerConfig(t)
	decoder := DecoderFunc[string](func(_ context.Context, value []byte) (string, error) {
		return string(value), nil
	})
	handler := BatchHandlerFunc[string](func(context.Context, *Batch[string]) error { return nil })

	_, err := NewConsumer[string](cfg, nil, handler)
	require.ErrorContains(t, err, "decoder")
	_, err = NewConsumer[string](cfg, decoder, nil)
	require.ErrorContains(t, err, "handler")

	cfg.Group = ""
	_, err = NewConsumer(cfg, decoder, handler)
	require.ErrorContains(t, err, "group")
}

func TestConsumerFlushesPartialBatchAfterInterval(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{Topic: "invoices", Partition: 1, Offset: 10, Value: []byte("first")}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gomock.InOrder(
		client.EXPECT().PollRecords(gomock.Any(), 2).
			Return(fetchedInvoiceRecords(1, record)),
		client.EXPECT().PollRecords(gomock.Any(), 1).DoAndReturn(
			func(ctx context.Context, _ int) kgo.Fetches {
				<-ctx.Done()
				return nil
			},
		),
	)
	decoder.EXPECT().Decode(gomock.Any(), []byte("first")).Return("decoded-first", nil)
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, batch *Batch[string]) error {
			require.Len(t, batch.Messages(), 1)
			return nil
		},
	)
	client.EXPECT().CommitRecords(gomock.Any(), record).DoAndReturn(
		func(_ context.Context, _ ...*kgo.Record) error {
			cancel()
			return nil
		},
	)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 2, FlushInterval: time.Millisecond}
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
}

func TestConsumerDrainsActiveBatchOnShutdown(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{Topic: "invoices", Partition: 1, Offset: 10, Value: []byte("first")}
	runCtx, cancel := context.WithCancel(context.Background())

	gomock.InOrder(
		client.EXPECT().PollRecords(gomock.Any(), 2).
			Return(fetchedInvoiceRecords(1, record)),
		client.EXPECT().PollRecords(gomock.Any(), 1).DoAndReturn(
			func(_ context.Context, _ int) kgo.Fetches {
				cancel()
				return nil
			},
		),
	)
	decoder.EXPECT().Decode(gomock.Any(), []byte("first")).Return("decoded-first", nil)
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, _ *Batch[string]) error {
			require.NoError(t, ctx.Err())
			_, hasDeadline := ctx.Deadline()
			require.True(t, hasDeadline)
			return nil
		},
	)
	client.EXPECT().CommitRecords(gomock.Any(), record).Return(nil)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 2, FlushInterval: time.Hour}
	cfg.ShutdownTimeout = time.Second
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
}

func TestConsumerRetriesHandlerWithResetDecisionsAndOneDecode(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{Topic: "invoices", Partition: 1, Offset: 10, Value: []byte("first")}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.EXPECT().PollRecords(gomock.Any(), 1).Return(fetchedInvoiceRecords(1, record))
	decoder.EXPECT().Decode(gomock.Any(), []byte("first")).Return("decoded-first", nil).Times(1)
	attempt := 0
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, batch *Batch[string]) error {
			attempt++
			if attempt == 1 {
				require.NoError(t, batch.Skip(0))
				return errors.New("temporary")
			}
			require.Equal(t, []messageDecision{{outcome: messageProcessed}}, batch.decisions())
			return nil
		},
	).Times(2)
	client.EXPECT().CommitRecords(gomock.Any(), record).DoAndReturn(
		func(_ context.Context, _ ...*kgo.Record) error {
			cancel()
			return nil
		},
	)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 1}
	cfg.Retry = RetryConfig{Policy: immediatePolicy{}, MaxAttempts: 2}
	observer := &recordingObserver{}
	cfg.Observer = observer
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
	require.Equal(t, 2, attempt)
	require.Equal(t, []AttemptPhase{AttemptHandler, AttemptHandler, AttemptCommit}, observer.phases)
	require.Equal(t, []AttemptPhase{AttemptHandler}, observer.retries)
	require.Equal(t, []BatchResult{{Processed: 1, Size: 1}}, observer.batches)
}

func TestConsumerStopsWithoutCommitOnRejectedMessage(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{Topic: "invoices", Partition: 3, Offset: 42, Value: []byte("invalid")}
	rejected := errors.New("invalid invoice")

	client.EXPECT().PollRecords(gomock.Any(), 1).Return(fetchedInvoiceRecords(3, record))
	decoder.EXPECT().Decode(gomock.Any(), []byte("invalid")).Return("decoded", nil)
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, batch *Batch[string]) error {
			return batch.Reject(0, rejected)
		},
	)
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 1}
	cfg.OnReject = RejectStop
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	err := consumer.Run(context.Background())
	var rejectedErr *RejectedMessageError
	require.ErrorAs(t, err, &rejectedErr)
	require.ErrorIs(t, err, rejected)
	require.Equal(t, "invoices", rejectedErr.Topic)
	require.Equal(t, int32(3), rejectedErr.Partition)
	require.Equal(t, int64(42), rejectedErr.Offset)
}

func TestConsumerDropsDecodeFailureWithoutCallingHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{Topic: "invoices", Partition: 3, Offset: 42, Value: []byte("invalid")}
	decodeErr := errors.New("bad wire format")
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.EXPECT().PollRecords(gomock.Any(), 1).Return(fetchedInvoiceRecords(3, record))
	decoder.EXPECT().Decode(gomock.Any(), []byte("invalid")).Return("", decodeErr)
	client.EXPECT().CommitRecords(gomock.Any(), record).DoAndReturn(
		func(_ context.Context, _ ...*kgo.Record) error {
			cancel()
			return nil
		},
	)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 1}
	cfg.OnReject = RejectDrop
	observer := &recordingObserver{}
	cfg.Observer = observer
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
	require.Len(t, observer.dispositions, 1)
	require.Equal(t, DispositionDropped, observer.dispositions[0].Kind)
	require.ErrorIs(t, observer.dispositions[0].Cause, decodeErr)
}

func TestConsumerSendsRejectedMessageToDLQBeforeCommit(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{
		Topic: "invoices", Partition: 3, Offset: 42,
		Key: []byte("invoice-42"), Value: []byte("invalid"),
		Headers: []kgo.RecordHeader{{Key: "tenant", Value: []byte("acme")}},
	}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.EXPECT().PollRecords(gomock.Any(), 1).Return(fetchedInvoiceRecords(3, record))
	decoder.EXPECT().Decode(gomock.Any(), []byte("invalid")).Return("decoded", nil)
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, batch *Batch[string]) error {
			return batch.Reject(0, errors.New("sensitive business reason"))
		},
	)
	client.EXPECT().ProduceSync(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, records ...*kgo.Record) kgo.ProduceResults {
			require.Len(t, records, 1)
			dlqRecord := records[0]
			require.Equal(t, "invoices.dlq", dlqRecord.Topic)
			require.Equal(t, record.Key, dlqRecord.Key)
			require.Equal(t, record.Value, dlqRecord.Value)
			require.Equal(t, []kgo.RecordHeader{
				{Key: "tenant", Value: []byte("acme")},
				{Key: DLQHeaderOriginalTopic, Value: []byte("invoices")},
				{Key: DLQHeaderOriginalPartition, Value: []byte("3")},
				{Key: DLQHeaderOriginalOffset, Value: []byte("42")},
				{Key: DLQHeaderFailureKind, Value: []byte("handler")},
			}, dlqRecord.Headers)
			return kgo.ProduceResults{{Record: dlqRecord}}
		},
	)
	client.EXPECT().CommitRecords(gomock.Any(), record).DoAndReturn(
		func(_ context.Context, _ ...*kgo.Record) error {
			cancel()
			return nil
		},
	)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 1}
	cfg.OnReject = RejectDLQ
	cfg.DLQ = &DLQConfig{Topic: "invoices.dlq", OnFailure: DLQFailureStop}
	observer := &recordingObserver{}
	cfg.Observer = observer
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
	require.Equal(t, []RecordDisposition{{
		Record: RecordMetadata{Topic: "invoices", Partition: 3, Offset: 42},
		Kind:   DispositionDLQ,
		Cause:  errors.New("sensitive business reason"),
	}}, observer.dispositions)
}

func TestConsumerRetriesCommitWithoutRerunningHandler(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{Topic: "invoices", Partition: 3, Offset: 42, Value: []byte("valid")}
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.EXPECT().PollRecords(gomock.Any(), 1).Return(fetchedInvoiceRecords(3, record))
	decoder.EXPECT().Decode(gomock.Any(), []byte("valid")).Return("decoded", nil).Times(1)
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).Return(nil).Times(1)
	gomock.InOrder(
		client.EXPECT().CommitRecords(gomock.Any(), record).Return(errors.New("coordinator unavailable")),
		client.EXPECT().CommitRecords(gomock.Any(), record).DoAndReturn(
			func(_ context.Context, _ ...*kgo.Record) error {
				cancel()
				return nil
			},
		),
	)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 1}
	cfg.Retry = RetryConfig{Policy: immediatePolicy{}, MaxAttempts: 2}
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
}

func TestConsumerReportsDropWhenDLQDeliveryFailureIsConfiguredToDrop(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	client := NewMockconsumerClient(ctrl)
	decoder := NewMockDecoder[string](ctrl)
	handler := NewMockBatchHandler[string](ctrl)
	record := &kgo.Record{Topic: "invoices", Partition: 3, Offset: 42, Value: []byte("invalid")}
	rejectionCause := errors.New("invalid invoice")
	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client.EXPECT().PollRecords(gomock.Any(), 1).Return(fetchedInvoiceRecords(3, record))
	decoder.EXPECT().Decode(gomock.Any(), record.Value).Return("decoded", nil)
	handler.EXPECT().Handle(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, batch *Batch[string]) error { return batch.Reject(0, rejectionCause) },
	)
	client.EXPECT().ProduceSync(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, records ...*kgo.Record) kgo.ProduceResults {
			return kgo.ProduceResults{{Record: records[0], Err: errors.New("DLQ unavailable")}}
		},
	)
	client.EXPECT().CommitRecords(gomock.Any(), record).DoAndReturn(
		func(_ context.Context, _ ...*kgo.Record) error { cancel(); return nil },
	)
	client.EXPECT().AllowRebalance()
	client.EXPECT().Close()

	observer := &recordingObserver{}
	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 1}
	cfg.Retry = RetryConfig{MaxAttempts: 1}
	cfg.OnReject = RejectDLQ
	cfg.DLQ = &DLQConfig{Topic: "invoices.dlq", OnFailure: DLQFailureDrop}
	cfg.Observer = observer
	consumer := newConsumerWithClient(cfg, client, decoder, handler)

	require.NoError(t, consumer.Run(runCtx))
	require.Len(t, observer.dispositions, 1)
	require.Equal(t, DispositionDropped, observer.dispositions[0].Kind)
	require.Equal(t, []BatchResult{{Size: 1, Rejected: 1, Dropped: 1}}, observer.batches)
}

type immediatePolicy struct{}

func (immediatePolicy) Delay(uint) time.Duration { return 0 }

type recordingObserver struct {
	phases       []AttemptPhase
	retries      []AttemptPhase
	batches      []BatchResult
	dispositions []RecordDisposition
}

func (observer *recordingObserver) StartAttempt(
	ctx context.Context,
	attempt Attempt,
) (context.Context, AttemptDone) {
	observer.phases = append(observer.phases, attempt.Phase)
	return ctx, func(error) {}
}

func (observer *recordingObserver) Retry(_ context.Context, event RetryEvent) {
	observer.retries = append(observer.retries, event.Phase)
}

func (observer *recordingObserver) BatchCompleted(_ context.Context, result BatchResult) {
	observer.batches = append(observer.batches, result)
}

func (observer *recordingObserver) RecordDisposition(_ context.Context, disposition RecordDisposition) {
	disposition.Record.Context = nil
	observer.dispositions = append(observer.dispositions, disposition)
}

func fetchedInvoiceRecords(partition int32, records ...*kgo.Record) kgo.Fetches {
	return kgo.Fetches{{Topics: []kgo.FetchTopic{{
		Topic: "invoices",
		Partitions: []kgo.FetchPartition{{
			Partition: partition,
			Records:   records,
		}},
	}}}}
}
