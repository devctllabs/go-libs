package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"go.yaml.in/yaml/v3"
)

type yamlLoader struct {
	input Input
}

// YAML creates a loader for one YAML document from input. For struct targets,
// "yaml" tags select fields, present nested struct fields are overlaid, and maps
// and slices are replaced as whole values. Multiple documents are rejected.
func YAML(input Input) Loader {
	return yamlLoader{input: input}
}

func (l yamlLoader) Load(ctx context.Context, target any) error {
	data, err := readInput(ctx, l.input, "YAML")
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var document yaml.Node
	if err := decoder.Decode(&document); errors.Is(err, io.EOF) {
		return nil
	} else if err != nil {
		return fmt.Errorf("config: parse YAML: %w", err)
	}

	var extra yaml.Node
	if err := decoder.Decode(&extra); err == nil {
		return errors.New("config: YAML input contains multiple documents")
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("config: parse YAML: %w", err)
	}

	if !isStructPointer(target) {
		if err := document.Decode(target); err != nil {
			return fmt.Errorf("config: decode YAML: %w", err)
		}
		return nil
	}

	var values map[string]any
	if err := document.Decode(&values); err != nil {
		return fmt.Errorf("config: decode YAML fields: %w", err)
	}

	err = decodeStructOverlay(target, values, fieldNames{
		tag:         "yaml",
		defaultName: strings.ToLower,
	}, document.Decode)
	if err != nil {
		return fmt.Errorf("config: decode YAML: %w", err)
	}

	return nil
}
