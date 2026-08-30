// Package kafkaproto adapts protobuf messages to kafka encoders and decoders.
package kafkaproto

import (
	"context"

	"github.com/devctllabs/go-libs/kafka"
	"google.golang.org/protobuf/proto"
)

// ProtoPtr constrains PT to the generated protobuf pointer for T.
type ProtoPtr[T any] interface {
	*T
	proto.Message
}

// NewEncoder returns a stateless protobuf value encoder.
func NewEncoder[T any, PT ProtoPtr[T]]() kafka.Encoder[PT] {
	return NewMessageEncoder[PT]()
}

type messageEncoder[T proto.Message] struct{}

// NewMessageEncoder returns a stateless encoder using a single protobuf
// message type parameter, for example NewMessageEncoder[*mypb.Event]().
func NewMessageEncoder[T proto.Message]() kafka.Encoder[T] {
	return messageEncoder[T]{}
}

func (messageEncoder[T]) Encode(_ context.Context, message T) ([]byte, error) {
	return proto.Marshal(message)
}

type decoder[T any, PT ProtoPtr[T]] struct{}

// NewDecoder returns a concurrency-safe decoder that allocates a fresh
// protobuf message for every record.
func NewDecoder[T any, PT ProtoPtr[T]]() kafka.Decoder[PT] {
	return decoder[T, PT]{}
}

func (decoder[T, PT]) Decode(_ context.Context, wire []byte) (PT, error) {
	message := PT(new(T))
	if err := proto.Unmarshal(wire, message); err != nil {
		var zero PT
		return zero, err
	}
	return message, nil
}
