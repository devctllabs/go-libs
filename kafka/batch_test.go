package kafka

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBatchClassifiesMessages(t *testing.T) {
	t.Parallel()

	rejected := errors.New("invalid invoice")
	batch := newBatch([]Message[string]{
		{Topic: "invoices", Partition: 1, Offset: 10, Value: "processed"},
		{Topic: "invoices", Partition: 1, Offset: 11, Value: "skipped"},
		{Topic: "invoices", Partition: 1, Offset: 12, Value: "rejected"},
	})

	require.Len(t, batch.Messages(), 3)
	require.NoError(t, batch.Skip(1))
	require.NoError(t, batch.Reject(2, rejected))
	require.Equal(t, []messageDecision{
		{outcome: messageProcessed},
		{outcome: messageSkipped},
		{outcome: messageRejected, err: rejected},
	}, batch.decisions())
}

func TestBatchRejectsInvalidDecisions(t *testing.T) {
	t.Parallel()

	batch := newBatch([]Message[string]{{Value: "invoice"}})

	var decisionErr *BatchDecisionError
	require.ErrorAs(t, batch.Skip(1), &decisionErr)
	require.ErrorAs(t, batch.Reject(0, nil), &decisionErr)
	require.NoError(t, batch.Skip(0))
	require.ErrorAs(t, batch.Reject(0, errors.New("conflict")), &decisionErr)
}

func TestBatchResetClearsAttemptDecisions(t *testing.T) {
	t.Parallel()

	batch := newBatch([]Message[string]{{Value: "invoice"}})
	require.NoError(t, batch.Skip(0))

	batch.reset()

	require.Equal(t, []messageDecision{{outcome: messageProcessed}}, batch.decisions())
}
