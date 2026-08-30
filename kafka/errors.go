package kafka

import "fmt"

// DecodeError reports a record value that could not be decoded. Index refers
// to the raw batch passed through the consumer.
type DecodeError struct {
	Index int
	Err   error
}

// Error implements error.
func (err *DecodeError) Error() string {
	return fmt.Sprintf("kafka: decode record %d: %v", err.Index, err.Err)
}

// Unwrap returns the decoder failure.
func (err *DecodeError) Unwrap() error {
	return err.Err
}

// RejectedMessageError reports a permanently rejected Kafka record for which
// the configured policy stopped consumption.
type RejectedMessageError struct {
	Topic     string
	Partition int32
	Offset    int64
	Cause     error
}

// Error implements error.
func (err *RejectedMessageError) Error() string {
	return fmt.Sprintf("kafka: rejected message %s/%d at offset %d: %v", err.Topic, err.Partition, err.Offset, err.Cause)
}

// Unwrap returns the handler-provided rejection cause.
func (err *RejectedMessageError) Unwrap() error {
	return err.Cause
}
