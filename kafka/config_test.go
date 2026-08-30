package kafka

import (
	"testing"
	"time"

	libretry "github.com/devctllabs/go-libs/retry"
	"github.com/stretchr/testify/require"
)

func TestConsumerConfigAcceptsSingleMessageMode(t *testing.T) {
	t.Parallel()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 1}

	require.NoError(t, cfg.validate())
}

func TestConsumerConfigRequiresFlushIntervalForBatches(t *testing.T) {
	t.Parallel()

	cfg := validConsumerConfig(t)
	cfg.Batch = BatchConfig{MaxSize: 2}

	require.ErrorContains(t, cfg.validate(), "flush interval")
}

func TestConsumerConfigRequiresBoundedRetry(t *testing.T) {
	t.Parallel()

	cfg := validConsumerConfig(t)
	cfg.Retry = RetryConfig{Policy: cfg.Retry.Policy}

	require.ErrorContains(t, cfg.validate(), "retry")
}

func TestConsumerConfigValidatesRetryOverrides(t *testing.T) {
	t.Parallel()

	cfg := validConsumerConfig(t)
	cfg.CommitRetry = &RetryConfig{}
	require.ErrorContains(t, cfg.validate(), "commit retry")

	cfg = validConsumerConfig(t)
	cfg.OnReject = RejectDLQ
	cfg.DLQ = &DLQConfig{
		Topic: "invoices.dlq", OnFailure: DLQFailureStop,
		Retry: &RetryConfig{},
	}
	require.ErrorContains(t, cfg.validate(), "DLQ retry")
}

func TestConsumerConfigValidatesRejectPolicy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		action  RejectAction
		dlq     *DLQConfig
		wantErr string
	}{
		{name: "unspecified", action: RejectUnspecified, wantErr: "reject action"},
		{name: "stop with dlq", action: RejectStop, dlq: &DLQConfig{Topic: "invalid"}, wantErr: "DLQ"},
		{name: "drop with dlq", action: RejectDrop, dlq: &DLQConfig{Topic: "invalid"}, wantErr: "DLQ"},
		{name: "dlq missing config", action: RejectDLQ, wantErr: "DLQ"},
		{name: "dlq missing failure action", action: RejectDLQ, dlq: &DLQConfig{Topic: "invoices.dlq"}, wantErr: "failure action"},
		{name: "stop", action: RejectStop},
		{name: "drop", action: RejectDrop},
		{name: "dlq stop", action: RejectDLQ, dlq: &DLQConfig{Topic: "invoices.dlq", OnFailure: DLQFailureStop}},
		{name: "dlq drop", action: RejectDLQ, dlq: &DLQConfig{Topic: "invoices.dlq", OnFailure: DLQFailureDrop}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			cfg := validConsumerConfig(t)
			cfg.OnReject = test.action
			cfg.DLQ = test.dlq

			err := cfg.validate()
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, test.wantErr)
		})
	}
}

func validConsumerConfig(t *testing.T) ConsumerConfig {
	t.Helper()

	policy, err := libretry.NewExponential(libretry.ExponentialConfig{
		InitialDelay: time.Millisecond,
		MaxDelay:     time.Second,
		Multiplier:   2,
	})
	require.NoError(t, err)
	return ConsumerConfig{
		Brokers:               []string{"localhost:9092"},
		Group:                 "invoice-consumer",
		Topics:                []string{"invoices"},
		Batch:                 BatchConfig{MaxSize: 100, FlushInterval: time.Second},
		Retry:                 RetryConfig{Policy: policy, MaxAttempts: 3},
		OnReject:              RejectStop,
		RebalanceTimeout:      time.Minute,
		RebalanceDrainTimeout: 30 * time.Second,
		ShutdownTimeout:       time.Minute,
	}
}
