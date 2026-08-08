package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type config struct {
	output io.Writer
}

// Option configures a logger created by New. Callers use the provided With...
// functions rather than implementing Option directly.
type Option func(*config)

// WithOutput directs log entries to output. A nil output leaves the current
// destination unchanged. When supplied more than once, the last non-nil output
// wins. The logger serializes writes but does not close output.
func WithOutput(output io.Writer) Option {
	return func(cfg *config) {
		if output != nil {
			cfg.output = output
		}
	}
}

// New constructs a JSON logger that emits entries at level and above. It uses
// ISO 8601 timestamps and writes to stderr unless WithOutput overrides the
// destination. When stacktrace is true, Error-level and higher entries include
// a stacktrace. New ignores nil options and does not replace zap's global
// loggers.
func New(level zapcore.Level, stacktrace bool, options ...Option) *zap.Logger {
	cfg := config{output: os.Stderr}
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(zapcore.AddSync(cfg.output)),
		level,
	)

	if stacktrace {
		return zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return zap.New(core)
}
