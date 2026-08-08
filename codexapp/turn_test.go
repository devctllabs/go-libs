package codexapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStartTurnClassifiesConcurrentStart(t *testing.T) {
	t.Parallel()
	client := &Client{
		starting:    map[string]*TurnHandle{"thread": {}},
		eventBuffer: 1,
	}
	_, err := client.StartTurn(context.Background(), StartTurnRequest{
		ThreadID: "thread",
		Input:    []Input{Text("hello")},
	})
	require.ErrorIs(t, err, ErrTurnStartInProgress)
}

func TestStartTurnRejectsDifferentPermissionProfile(t *testing.T) {
	t.Parallel()
	bound, err := buildPermissionProfile(Permissions{ReadRoots: []string{"/workspace"}})
	require.NoError(t, err)
	client := &Client{
		starting:    map[string]*TurnHandle{},
		profiles:    map[string]normalizedPermissionProfile{"thread": bound},
		eventBuffer: 1,
	}
	_, err = client.StartTurn(context.Background(), StartTurnRequest{
		ThreadID:    "thread",
		Input:       []Input{Text("hello")},
		Permissions: Permissions{WriteRoots: []string{"/workspace"}},
	})
	require.ErrorIs(t, err, ErrPermissionProfileMismatch)
}
