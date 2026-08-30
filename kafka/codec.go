package kafka

import (
	"context"
	"encoding/json"
)

//go:generate go tool mockgen -destination=contracts.gen_test.go -package=kafka -typed . Decoder,Encoder,BatchHandler

// Decoder converts one Kafka record value into the handler's immutable type.
type Decoder[T any] interface {
	// Decode converts value into T. Implementations must be concurrency-safe
	// when used by a partitioned consumer.
	Decode(ctx context.Context, value []byte) (T, error)
}

// DecoderFunc adapts a function to Decoder.
type DecoderFunc[T any] func(ctx context.Context, value []byte) (T, error)

// Decode calls decoder(ctx, value).
func (decoder DecoderFunc[T]) Decode(ctx context.Context, value []byte) (T, error) {
	return decoder(ctx, value)
}

// Encoder converts a typed value into one Kafka record value.
type Encoder[T any] interface {
	// Encode serializes value. Implementations must be concurrency-safe.
	Encode(ctx context.Context, value T) ([]byte, error)
}

// EncoderFunc adapts a function to Encoder.
type EncoderFunc[T any] func(ctx context.Context, value T) ([]byte, error)

// Encode calls encoder(ctx, value).
func (encoder EncoderFunc[T]) Encode(ctx context.Context, value T) ([]byte, error) {
	return encoder(ctx, value)
}

// NewBytesEncoder returns a zero-copy encoder for an already serialized value.
// Callers must not mutate value until the operation using the encoded bytes returns.
func NewBytesEncoder() Encoder[[]byte] {
	return EncoderFunc[[]byte](func(_ context.Context, value []byte) ([]byte, error) {
		return value, nil
	})
}

// NewJSONEncoder returns a stateless JSON value encoder.
func NewJSONEncoder[T any]() Encoder[T] {
	return EncoderFunc[T](func(_ context.Context, value T) ([]byte, error) {
		return json.Marshal(value)
	})
}

// NewJSONDecoder returns a concurrency-safe decoder that creates a fresh value for every record.
func NewJSONDecoder[T any]() Decoder[T] {
	return DecoderFunc[T](func(_ context.Context, wire []byte) (T, error) {
		var value T
		err := json.Unmarshal(wire, &value)
		return value, err
	})
}
