package log_test

import (
	"bytes"
	"encoding/json"
	stdlog "log"
	"os"

	applog "github.com/devctllabs/go-libs/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ExampleNew() {
	var output bytes.Buffer
	logger := applog.New(
		zapcore.InfoLevel,
		false,
		applog.WithOutput(&output),
	)
	logger.Info("started", zap.String("service", "api"))

	var entry struct {
		Level   string `json:"level"`
		Message string `json:"msg"`
		Service string `json:"service"`
	}
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		stdlog.Fatal(err)
	}

	outputLogger := stdlog.New(os.Stdout, "", 0)
	outputLogger.Printf("level=%s message=%s service=%s", entry.Level, entry.Message, entry.Service)
	// Output: level=info message=started service=api
}

func ExampleWithEncoding() {
	var output bytes.Buffer
	logger := applog.New(
		zapcore.InfoLevel,
		false,
		applog.WithEncoding(applog.EncodingConsole),
		applog.WithOutput(&output),
	)
	logger.Info("started")

	fields := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte{'\t'})
	outputLogger := stdlog.New(os.Stdout, "", 0)
	outputLogger.Printf("level=%s message=%s", fields[1], fields[2])
	// Output: level=info message=started
}
