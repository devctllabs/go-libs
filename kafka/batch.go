package kafka

import (
	"fmt"
)

type messageOutcome uint8

const (
	messageProcessed messageOutcome = iota
	messageSkipped
	messageRejected
)

type messageDecision struct {
	outcome messageOutcome
	err     error
}

// Batch is one immutable collection of decoded messages and its mutable
// per-attempt outcome buffer.
type Batch[T any] struct {
	messages        []Message[T]
	decisionsBuffer []messageDecision
}

// BatchDecisionError reports invalid Skip or Reject usage by a handler.
type BatchDecisionError struct {
	Index  int
	Reason string
}

// Error implements error.
func (err *BatchDecisionError) Error() string {
	return fmt.Sprintf("kafka: batch decision for index %d: %s", err.Index, err.Reason)
}

func newBatch[T any](messages []Message[T]) *Batch[T] {
	return &Batch[T]{
		messages:        messages,
		decisionsBuffer: make([]messageDecision, len(messages)),
	}
}

// Messages returns a borrowed immutable view of the decoded messages. The
// returned slice and its elements are reused between handler attempts.
func (batch *Batch[T]) Messages() []Message[T] {
	return batch.messages
}

// Skip marks messages[index] as intentionally acknowledged without business
// processing. A message may be classified at most once per attempt.
func (batch *Batch[T]) Skip(index int) error {
	return batch.mark(index, messageDecision{outcome: messageSkipped})
}

// Reject marks messages[index] as permanently rejected for the configured
// reject policy. cause is diagnostic and must not be nil.
func (batch *Batch[T]) Reject(index int, cause error) error {
	if cause == nil {
		return &BatchDecisionError{Index: index, Reason: "reject cause must not be nil"}
	}
	return batch.mark(index, messageDecision{outcome: messageRejected, err: cause})
}

func (batch *Batch[T]) mark(index int, decision messageDecision) error {
	if index < 0 || index >= len(batch.decisionsBuffer) {
		return &BatchDecisionError{Index: index, Reason: "index is out of range"}
	}
	if batch.decisionsBuffer[index].outcome != messageProcessed || batch.decisionsBuffer[index].err != nil {
		return &BatchDecisionError{Index: index, Reason: "message is already classified"}
	}
	batch.decisionsBuffer[index] = decision
	return nil
}

func (batch *Batch[T]) reset() {
	clear(batch.decisionsBuffer)
}

func (batch *Batch[T]) decisions() []messageDecision {
	return append([]messageDecision(nil), batch.decisionsBuffer...)
}
