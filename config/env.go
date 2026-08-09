package config

import (
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

type envMapLoader struct {
	values map[string]string
}

type osEnvLoader struct{}

// EnvMap creates a loader that decodes "env" struct tags from a copied snapshot
// of values. It never reads the process environment, including when values is
// nil.
func EnvMap(values map[string]string) Loader {
	snapshot := make(map[string]string, len(values))
	for key, value := range values {
		snapshot[key] = value
	}
	return envMapLoader{values: snapshot}
}

// OSEnv creates a loader that reads the process environment at Load time and
// decodes values through "env" struct tags.
func OSEnv() Loader {
	return osEnvLoader{}
}

func (l envMapLoader) Load(ctx context.Context, target any) error {
	return loadEnvMap(ctx, target, l.values)
}

func (osEnvLoader) Load(ctx context.Context, target any) error {
	return loadEnvMap(ctx, target, env.ToMap(os.Environ()))
}

func loadEnvMap(ctx context.Context, target any, values map[string]string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := env.ParseWithOptions(target, env.Options{Environment: values}); err != nil {
		return fmt.Errorf("config: load environment: %w", err)
	}

	return nil
}
