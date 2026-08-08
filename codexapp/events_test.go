package codexapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEventStreamReportsOverflowAfterDrainingQueuedEvents(t *testing.T) {
	t.Parallel()
	stream := newEventStream(1)
	first := Event{Type: EventAgentMessageDelta, TextDelta: &TextDelta{Text: "first"}}
	require.NoError(t, stream.push(first))

	err := stream.push(Event{Type: EventAgentMessageDelta, TextDelta: &TextDelta{Text: "lost"}})
	require.ErrorIs(t, err, ErrEventOverflow)
	event, err := stream.Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, first, event)
	_, err = stream.Next(context.Background())
	require.ErrorIs(t, err, ErrEventOverflow)
}

func TestEventStreamCloseDoesNotPretendToBeNormalCompletion(t *testing.T) {
	t.Parallel()
	stream := newEventStream(1)
	require.NoError(t, stream.Close())
	_, err := stream.Next(context.Background())
	require.ErrorIs(t, err, ErrClosed)
}
