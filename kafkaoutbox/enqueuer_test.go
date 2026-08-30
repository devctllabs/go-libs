package kafkaoutbox

import (
	"context"
	"errors"
	"testing"

	"github.com/devctllabs/go-libs/kafka"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEnqueuerEncodesAndAppendsEvent(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	appender := NewMockAppender(ctrl)
	encoder := kafka.EncoderFunc[string](func(_ context.Context, value string) ([]byte, error) {
		return []byte("encoded:" + value), nil
	})
	appender.EXPECT().Append(gomock.Any(), Event[[]byte]{
		Topic:         "orders",
		AggregateType: "Order",
		AggregateID:   "42",
		Type:          "OrderPaid",
		Value:         []byte("encoded:payload"),
	}).Return(nil)

	enqueuer, err := NewEnqueuer(EnqueuerConfig{}, appender, encoder)
	require.NoError(t, err)

	err = enqueuer.Enqueue(context.Background(), Event[string]{
		Topic:         "orders",
		AggregateType: "Order",
		AggregateID:   "42",
		Type:          "OrderPaid",
		Value:         "payload",
	})
	require.NoError(t, err)
}

func TestEnqueuerObservesSuccessfulOperation(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	appender := NewMockAppender(ctrl)
	encoder := NewMockEncoder[string](ctrl)
	observer := NewMockObserver(ctrl)
	ctx := context.Background()
	observedCtx := context.WithValue(ctx, observationContextKey{}, "observed")
	done := make(chan error, 1)
	event := validStringEvent()

	observer.EXPECT().StartOperation(ctx, Operation{Phase: OperationEnqueue}).Return(
		observedCtx,
		OperationDone(func(err error) { done <- err }),
	)
	encoder.EXPECT().Encode(observedCtx, event.Value).Return([]byte("wire"), nil)
	appender.EXPECT().Append(observedCtx, gomock.Any()).Return(nil)
	observer.EXPECT().Enqueued(observedCtx)
	enqueuer, err := NewEnqueuer(EnqueuerConfig{Observer: observer}, appender, encoder)
	require.NoError(t, err)

	require.NoError(t, enqueuer.Enqueue(ctx, event))
	require.NoError(t, <-done)
}

func TestEnqueuerRejectsInvalidEventBeforeEncoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		ctx   context.Context
		event Event[string]
	}{
		{name: "nil context", event: validStringEvent()},
		{name: "blank topic", ctx: context.Background(), event: eventWith(func(event *Event[string]) { event.Topic = " " })},
		{name: "blank aggregate type", ctx: context.Background(), event: eventWith(func(event *Event[string]) { event.AggregateType = "" })},
		{name: "blank aggregate id", ctx: context.Background(), event: eventWith(func(event *Event[string]) { event.AggregateID = "" })},
		{name: "blank event type", ctx: context.Background(), event: eventWith(func(event *Event[string]) { event.Type = "\t" })},
		{name: "reserved delimiter in aggregate type", ctx: context.Background(), event: eventWith(func(event *Event[string]) { event.AggregateType = "Sales::Order" })},
		{name: "reserved delimiter in aggregate id", ctx: context.Background(), event: eventWith(func(event *Event[string]) { event.AggregateID = "tenant::42" })},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			appender := NewMockAppender(ctrl)
			encoder := NewMockEncoder[string](ctrl)
			enqueuer, err := NewEnqueuer(EnqueuerConfig{}, appender, encoder)
			require.NoError(t, err)

			err = enqueuer.Enqueue(test.ctx, test.event)
			require.Error(t, err)
		})
	}
}

func TestEnqueuerPreservesEncoderAndAppenderErrors(t *testing.T) {
	t.Parallel()
	t.Run("encoder", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		appender := NewMockAppender(ctrl)
		encoder := NewMockEncoder[string](ctrl)
		encodeErr := errors.New("encode failed")
		encoder.EXPECT().Encode(gomock.Any(), "payload").Return(nil, encodeErr)
		enqueuer, err := NewEnqueuer(EnqueuerConfig{}, appender, encoder)
		require.NoError(t, err)

		err = enqueuer.Enqueue(context.Background(), validStringEvent())
		require.ErrorIs(t, err, encodeErr)
	})

	t.Run("appender", func(t *testing.T) {
		t.Parallel()
		ctrl := gomock.NewController(t)
		appender := NewMockAppender(ctrl)
		encoder := NewMockEncoder[string](ctrl)
		appendErr := errors.New("append failed")
		encoder.EXPECT().Encode(gomock.Any(), "payload").Return([]byte("wire"), nil)
		appender.EXPECT().Append(gomock.Any(), gomock.Any()).Return(appendErr)
		enqueuer, err := NewEnqueuer(EnqueuerConfig{}, appender, encoder)
		require.NoError(t, err)

		err = enqueuer.Enqueue(context.Background(), validStringEvent())
		require.ErrorIs(t, err, appendErr)
	})
}

func TestNewEnqueuerRequiresDependencies(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	appender := NewMockAppender(ctrl)
	encoder := NewMockEncoder[string](ctrl)

	_, err := NewEnqueuer[string](EnqueuerConfig{}, nil, encoder)
	require.Error(t, err)
	_, err = NewEnqueuer[string](EnqueuerConfig{}, appender, nil)
	require.Error(t, err)
}

func validStringEvent() Event[string] {
	return Event[string]{
		Topic:         "orders",
		AggregateType: "Order",
		AggregateID:   "42",
		Type:          "OrderPaid",
		Value:         "payload",
	}
}

func eventWith(change func(event *Event[string])) Event[string] {
	event := validStringEvent()
	change(&event)
	return event
}

type observationContextKey struct{}
