package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratedSchemaNameAddsGeneratedSuffix(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"CommandExecutionRequestApprovalParams.json": "CommandExecutionRequestApprovalParams.gen.json",
		"v1/InitializeParams.json":                   "v1/InitializeParams.gen.json",
		"v2/TurnStartResponse.json":                  "v2/TurnStartResponse.gen.json",
	}

	for source, expected := range tests {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, expected, generatedSchemaName(source))
		})
	}
}

func TestGeneratedGoNameHasGeneratedSuffix(t *testing.T) {
	t.Parallel()
	require.Equal(t, "protocol.gen.go", generatedGoArtifactName)
}
