package codexapp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInputConstructorsMapToProtocolVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    Input
		expected string
	}{
		{"text", Text("hello"), `{"type":"text","text":"hello"}`},
		{"image", Image("https://example.test/image.png"), `{"type":"image","url":"https://example.test/image.png"}`},
		{"local image", LocalImage("/tmp/image.png"), `{"type":"localImage","path":"/tmp/image.png"}`},
		{"audio", Audio("https://example.test/audio.wav"), `{"type":"audio","url":"https://example.test/audio.wav"}`},
		{"local audio", LocalAudio("/tmp/audio.wav"), `{"type":"localAudio","path":"/tmp/audio.wav"}`},
		{"skill", Skill("review", "/skills/review/SKILL.md"), `{"type":"skill","name":"review","path":"/skills/review/SKILL.md"}`},
		{"mention", Mention("README", "/workspace/README.md"), `{"type":"mention","name":"README","path":"/workspace/README.md"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wire, err := mapInput(tt.input)
			require.NoError(t, err)
			encoded, err := json.Marshal(wire)
			require.NoError(t, err)
			require.JSONEq(t, tt.expected, string(encoded))
		})
	}
}
