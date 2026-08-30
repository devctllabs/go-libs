package log_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/log"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewWritesJSONToConfiguredOutput(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer

	logger := log.New(zapcore.InfoLevel, false, log.WithOutput(&output))
	logger.Info("started", zap.String("service", "api"))

	entries := decodeEntries(t, output.Bytes())
	require.Len(t, entries, 1)
	entry := entries[0]
	require.Equal(t, "info", entry["level"])
	require.Equal(t, "started", entry["msg"])
	require.Equal(t, "api", entry["service"])

	timestamp, ok := entry["ts"].(string)
	require.True(t, ok)
	_, err := time.Parse("2006-01-02T15:04:05.000Z0700", timestamp)
	require.NoError(t, err)
}

func TestNewWritesConfiguredEncoding(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		encoding log.Encoding
		assert   func(*testing.T, []byte)
	}{
		{
			name:     "json",
			encoding: log.EncodingJSON,
			assert: func(t *testing.T, output []byte) {
				t.Helper()
				entries := decodeEntries(t, output)
				require.Len(t, entries, 1)
				require.Equal(t, "started", entries[0]["msg"])
			},
		},
		{
			name:     "console",
			encoding: log.EncodingConsole,
			assert: func(t *testing.T, output []byte) {
				t.Helper()
				require.Contains(t, string(output), "\tinfo\tstarted\t")
				require.Contains(t, string(output), `"service": "api"`)
			},
		},
		{
			name:     "unsupported falls back to json",
			encoding: log.Encoding("yaml"),
			assert: func(t *testing.T, output []byte) {
				t.Helper()
				entries := decodeEntries(t, output)
				require.Len(t, entries, 1)
				require.Equal(t, "started", entries[0]["msg"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer

			logger := log.New(
				zapcore.InfoLevel,
				false,
				log.WithOutput(&output),
				log.WithEncoding(tt.encoding),
			)
			logger.Info("started", zap.String("service", "api"))

			tt.assert(t, output.Bytes())
		})
	}
}

func TestNewFiltersEntriesBelowLevel(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer

	logger := log.New(zapcore.ErrorLevel, false, log.WithOutput(&output))
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	entries := decodeEntries(t, output.Bytes())
	require.Len(t, entries, 1)
	require.Equal(t, "error", entries[0]["level"])
	require.Equal(t, "error", entries[0]["msg"])
}

func TestNewConfiguresStacktraces(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stacktrace bool
		entryLevel zapcore.Level
		wantStack  bool
	}{
		{
			name:       "disabled for error",
			stacktrace: false,
			entryLevel: zapcore.ErrorLevel,
			wantStack:  false,
		},
		{
			name:       "enabled below error",
			stacktrace: true,
			entryLevel: zapcore.InfoLevel,
			wantStack:  false,
		},
		{
			name:       "enabled for error",
			stacktrace: true,
			entryLevel: zapcore.ErrorLevel,
			wantStack:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var output bytes.Buffer

			logger := log.New(zapcore.DebugLevel, tt.stacktrace, log.WithOutput(&output))
			logger.Log(tt.entryLevel, "message")

			entries := decodeEntries(t, output.Bytes())
			require.Len(t, entries, 1)
			if tt.wantStack {
				require.NotEmpty(t, entries[0]["stacktrace"])
				return
			}
			require.NotContains(t, entries[0], "stacktrace")
		})
	}
}

//nolint:paralleltest // The test temporarily replaces process-global os.Stderr.
func TestNewWritesToStderrByDefault(t *testing.T) {
	originalStderr := os.Stderr
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		os.Stderr = originalStderr
		_ = reader.Close()
		_ = writer.Close()
	})

	os.Stderr = writer
	logger := log.New(zapcore.InfoLevel, false)
	os.Stderr = originalStderr

	logger.Info("message")
	require.NoError(t, writer.Close())
	output, err := io.ReadAll(reader)
	require.NoError(t, err)

	entries := decodeEntries(t, output)
	require.Len(t, entries, 1)
	require.Equal(t, "message", entries[0]["msg"])
}

func TestWithOutputIgnoresNil(t *testing.T) {
	t.Parallel()
	var output bytes.Buffer

	logger := log.New(
		zapcore.InfoLevel,
		false,
		log.WithOutput(&output),
		log.WithOutput(nil),
	)
	logger.Info("message")

	entries := decodeEntries(t, output.Bytes())
	require.Len(t, entries, 1)
	require.Equal(t, "message", entries[0]["msg"])
}

func TestWithOutputUsesLastNonNilOutput(t *testing.T) {
	t.Parallel()
	var firstOutput bytes.Buffer
	var secondOutput bytes.Buffer

	logger := log.New(
		zapcore.InfoLevel,
		false,
		log.WithOutput(&firstOutput),
		log.WithOutput(&secondOutput),
	)
	logger.Info("message")

	require.Empty(t, firstOutput.Bytes())
	entries := decodeEntries(t, secondOutput.Bytes())
	require.Len(t, entries, 1)
	require.Equal(t, "message", entries[0]["msg"])
}

func TestNewDoesNotReplaceGlobalLoggers(t *testing.T) {
	t.Parallel()
	globalLogger := zap.L()
	globalSugar := zap.S()

	_ = log.New(zapcore.InfoLevel, false, log.WithOutput(io.Discard))

	require.Same(t, globalLogger, zap.L())
	require.Same(t, globalSugar, zap.S())
}

func decodeEntries(t *testing.T, output []byte) []map[string]any {
	t.Helper()

	lines := bytes.Split(bytes.TrimSpace(output), []byte{'\n'})
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var entry map[string]any
		require.NoError(t, json.Unmarshal(line, &entry))
		entries = append(entries, entry)
	}

	return entries
}
