package config

import (
	"context"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

type tomlLoader struct {
	input Input
}

// TOML creates a loader for TOML data from input. For struct targets, "toml"
// tags select fields, present nested struct fields are overlaid, and maps and
// slices are replaced as whole values.
func TOML(input Input) Loader {
	return tomlLoader{input: input}
}

func (l tomlLoader) Load(ctx context.Context, target any) error {
	data, err := readInput(ctx, l.input, "TOML")
	if err != nil {
		return err
	}

	if !isStructPointer(target) {
		if err := toml.Unmarshal(data, target); err != nil {
			return fmt.Errorf("config: decode TOML: %w", err)
		}
		return nil
	}

	var values map[string]any
	if err := toml.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("config: parse TOML: %w", err)
	}

	err = decodeStructOverlay(target, values, fieldNames{
		tag:         "toml",
		defaultName: func(name string) string { return name },
		fold:        true,
	}, func(target any) error {
		return toml.Unmarshal(data, target)
	})
	if err != nil {
		return fmt.Errorf("config: decode TOML: %w", err)
	}

	return nil
}
