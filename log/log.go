package log

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type config struct {
	output   io.Writer
	encoding Encoding
}

// Encoding selects the zap encoder used for log entries.
type Encoding string

const (
	// EncodingJSON writes one JSON object per log entry.
	EncodingJSON Encoding = "json"
	// EncodingConsole writes production fields in zap's console layout.
	EncodingConsole Encoding = "console"
)

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

// WithEncoding selects the log entry encoding. Unsupported values use JSON.
func WithEncoding(encoding Encoding) Option {
	return func(cfg *config) {
		switch encoding {
		case EncodingConsole:
			cfg.encoding = EncodingConsole
		default:
			cfg.encoding = EncodingJSON
		}
	}
}

// New constructs a logger that emits entries at level and above. It defaults
// to JSON, uses ISO 8601 timestamps, and writes to stderr unless WithOutput
// overrides the destination. When stacktrace is true, Error-level and higher
// entries include a stacktrace. New ignores nil options and does not replace
// zap's global loggers.
func New(level zapcore.Level, stacktrace bool, options ...Option) *zap.Logger {
	cfg := config{output: os.Stderr, encoding: EncodingJSON}
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var encoder zapcore.Encoder
	if cfg.encoding == EncodingConsole {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(zapcore.AddSync(cfg.output)),
		level,
	)

	if stacktrace {
		return zap.New(core, zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return zap.New(core)
}
