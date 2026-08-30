package kafkaoutboxzap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/kafkaoutbox"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestObserverLogsRetriesFencingAndTopologyChanges(t *testing.T) {
	t.Parallel()
	core, logs := observer.New(zap.DebugLevel)
	adapter, err := New(zap.New(core))
	require.NoError(t, err)
	cause := errors.New("broker unavailable")

	adapter.Retry(context.Background(), kafkaoutbox.RetryEvent{
		Generation: 2, ShardID: 3, Attempt: 4, NextDelay: time.Second, Err: cause,
	})
	adapter.FencingConflict(context.Background(), kafkaoutbox.FencingEvent{
		Phase: kafkaoutbox.OperationFinalize, Generation: 2, ShardID: 3,
	})
	adapter.TopologyReconciled(context.Background(), kafkaoutbox.TopologyResult{
		Revision: 2, Generation: 2, ShardCount: 8, Changed: true,
	})
	adapter.TopologyReconciled(context.Background(), kafkaoutbox.TopologyResult{
		Revision: 2, Generation: 2, ShardCount: 8,
	})

	require.Equal(t, []zapcore.Level{zap.WarnLevel, zap.WarnLevel, zap.InfoLevel}, []zapcore.Level{
		logs.All()[0].Level, logs.All()[1].Level, logs.All()[2].Level,
	})
	require.Equal(t, "kafka outbox publish retry scheduled", logs.All()[0].Message)
	require.Equal(t, "kafka outbox shard claim lost", logs.All()[1].Message)
	require.Equal(t, "kafka outbox topology changed", logs.All()[2].Message)
}

func TestNewRejectsNilLogger(t *testing.T) {
	t.Parallel()
	_, err := New(nil)
	require.Error(t, err)
}
