package kafka

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBytesEncoderReturnsBorrowedWireBytes(t *testing.T) {
	t.Parallel()
	encoder := NewBytesEncoder()
	value := []byte("event")

	wire, err := encoder.Encode(context.Background(), value)
	require.NoError(t, err)
	require.Equal(t, value, wire)
	wire[0] = 'E'
	require.Equal(t, []byte("Event"), value)

	wire, err = encoder.Encode(context.Background(), nil)
	require.NoError(t, err)
	require.Nil(t, wire)
}

func TestJSONCodecRoundTrip(t *testing.T) {
	t.Parallel()
	encoder := NewJSONEncoder[jsonEvent]()
	decoder := NewJSONDecoder[jsonEvent]()
	want := jsonEvent{ID: "event-42"}

	wire, err := encoder.Encode(context.Background(), want)
	require.NoError(t, err)
	require.JSONEq(t, `{"id":"event-42"}`, string(wire))
	got, err := decoder.Decode(context.Background(), wire)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestJSONDecoderAllocatesFreshPointer(t *testing.T) {
	t.Parallel()
	decoder := NewJSONDecoder[*jsonEvent]()

	first, err := decoder.Decode(context.Background(), []byte(`{"id":"first"}`))
	require.NoError(t, err)
	second, err := decoder.Decode(context.Background(), []byte(`{"id":"second"}`))
	require.NoError(t, err)

	require.NotSame(t, first, second)
	require.Equal(t, "first", first.ID)
	require.Equal(t, "second", second.ID)
}

func TestJSONCodecReturnsSerializationErrors(t *testing.T) {
	t.Parallel()
	encoder := NewJSONEncoder[func()]()
	decoder := NewJSONDecoder[jsonEvent]()

	_, err := encoder.Encode(context.Background(), func() {})
	require.Error(t, err)
	_, err = decoder.Decode(context.Background(), []byte(`{"id":`))
	require.Error(t, err)
}

type jsonEvent struct {
	ID string `json:"id"`
}
