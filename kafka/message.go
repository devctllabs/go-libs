package kafka

import "time"

// Header is one Kafka record header. Value is borrowed for the duration of the
// operation that supplied it.
type Header struct {
	Key   string
	Value []byte
}

// Message is a decoded Kafka record. Handlers must treat Message and all data
// reachable from it as immutable and must not retain it after Handle returns.
type Message[T any] struct {
	Topic     string
	Partition int32
	Offset    int64
	Timestamp time.Time
	Key       []byte
	Headers   []Header
	Value     T
}
