package kafkazap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/kafka"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestObserverLogsRetriesAndTerminalDispositions(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zap.DebugLevel)
	adapter, err := New(zap.New(core))
	require.NoError(t, err)
	cause := errors.New("sensitive reason")

	adapter.Retry(context.Background(), kafka.RetryEvent{
		Phase: kafka.AttemptHandler, Attempt: 2, NextDelay: time.Second, Err: cause,
	})
	adapter.RecordDisposition(context.Background(), kafka.RecordDisposition{
		Record: kafka.RecordMetadata{Topic: "invoices", Partition: 1, Offset: 10},
		Kind:   kafka.DispositionDLQ, Cause: cause,
	})
	adapter.RecordDisposition(context.Background(), kafka.RecordDisposition{
		Record: kafka.RecordMetadata{Topic: "invoices", Partition: 1, Offset: 11},
		Kind:   kafka.DispositionDropped, Cause: cause,
	})

	require.Equal(t, []zapcore.Level{zap.WarnLevel, zap.InfoLevel, zap.WarnLevel}, []zapcore.Level{
		logs.All()[0].Level, logs.All()[1].Level, logs.All()[2].Level,
	})
	require.Equal(t, "kafka operation retry", logs.All()[0].Message)
	require.Equal(t, "kafka record sent to DLQ", logs.All()[1].Message)
	require.Equal(t, "kafka record dropped", logs.All()[2].Message)
}

func TestNewRejectsNilLogger(t *testing.T) {
	t.Parallel()

	_, err := New(nil)
	require.Error(t, err)
}
