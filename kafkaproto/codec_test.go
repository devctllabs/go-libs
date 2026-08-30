package kafkaproto

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCodecRoundTrip(t *testing.T) {
	t.Parallel()

	encoder := NewEncoder[wrapperspb.StringValue, *wrapperspb.StringValue]()
	decoder := NewDecoder[wrapperspb.StringValue, *wrapperspb.StringValue]()
	want := wrapperspb.String("invoice-42")

	wire, err := encoder.Encode(context.Background(), want)
	require.NoError(t, err)
	got, err := decoder.Decode(context.Background(), wire)
	require.NoError(t, err)
	require.True(t, proto.Equal(want, got))
}

func TestDecoderRejectsInvalidWireFormat(t *testing.T) {
	t.Parallel()

	decoder := NewDecoder[wrapperspb.StringValue, *wrapperspb.StringValue]()

	_, err := decoder.Decode(context.Background(), []byte{0xff})
	require.Error(t, err)
}

func TestMessageEncoderMarshalsProtoMessage(t *testing.T) {
	t.Parallel()
	encoder := NewMessageEncoder[*wrapperspb.StringValue]()
	want := wrapperspb.String("invoice-42")

	wire, err := encoder.Encode(context.Background(), want)
	require.NoError(t, err)
	decoded := &wrapperspb.StringValue{}
	require.NoError(t, proto.Unmarshal(wire, decoded))
	require.True(t, proto.Equal(want, decoded))
}
