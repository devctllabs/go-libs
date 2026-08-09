package config

import (
	"context"
	"fmt"

	"github.com/creasty/defaults"
)

type defaultsLoader struct{}

// Defaults creates a loader for "default" struct tags. Existing non-zero values
// are preserved according to github.com/creasty/defaults semantics.
func Defaults() Loader {
	return defaultsLoader{}
}

func (defaultsLoader) Load(ctx context.Context, target any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := defaults.Set(target); err != nil {
		return fmt.Errorf("config: load defaults: %w", err)
	}

	return nil
}
