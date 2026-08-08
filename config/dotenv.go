package config

import (
	"context"
	"errors"
	"fmt"

	"github.com/joho/godotenv"
)

type dotEnvLoader struct {
	input Input
}

// DotEnv creates a loader that parses dotenv data from input and decodes it
// through "env" struct tags. It does not modify or otherwise read the process
// environment.
func DotEnv(input Input) Loader {
	return dotEnvLoader{input: input}
}

func (l dotEnvLoader) Load(ctx context.Context, target any) error {
	if isNilValue(l.input) {
		return errors.New("config: nil dotenv input")
	}

	reader, err := l.input.Open(ctx)
	if err != nil {
		return fmt.Errorf("config: open dotenv input: %w", err)
	}
	if reader == nil {
		return errors.New("config: dotenv input returned a nil reader")
	}
	defer func() { _ = reader.Close() }()

	values, err := godotenv.Parse(reader)
	if err != nil {
		return fmt.Errorf("config: parse dotenv: %w", err)
	}

	return loadEnvMap(ctx, target, values)
}
