package codexapp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/codexapp"
	"github.com/devctllabs/go-libs/codexapp/mocks"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/mock/gomock"
)

func TestOpenRejectsMissingClientIdentity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		ctx     context.Context
		config  codexapp.Config
		message string
	}{
		{
			name:    "name",
			ctx:     context.Background(),
			config:  codexapp.Config{},
			message: "client info name",
		},
		{
			name: "version",
			ctx:  context.Background(),
			config: codexapp.Config{ClientInfo: codexapp.ClientInfo{
				Name: "test-client",
			}},
			message: "client info version",
		},
		{
			name: "context",
			config: codexapp.Config{ClientInfo: codexapp.ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			}},
			message: "context",
		},
		{
			name: "event buffer",
			ctx:  context.Background(),
			config: codexapp.Config{
				ClientInfo:      codexapp.ClientInfo{Name: "test-client", Version: "1.0.0"},
				TurnEventBuffer: -1,
			},
			message: "turn event buffer",
		},
		{
			name: "message size",
			ctx:  context.Background(),
			config: codexapp.Config{
				ClientInfo:      codexapp.ClientInfo{Name: "test-client", Version: "1.0.0"},
				MaxMessageBytes: -1,
			},
			message: "max message bytes",
		},
		{
			name: "project directory",
			ctx:  context.Background(),
			config: codexapp.Config{
				ProjectDir: filepath.Join(t.TempDir(), "missing"),
				ClientInfo: codexapp.ClientInfo{Name: "test-client", Version: "1.0.0"},
			},
			message: "project directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client, err := codexapp.Open(tt.ctx, tt.config)

			require.Nil(t, client)
			require.ErrorContains(t, err, tt.message)
		})
	}
}

func TestOpenInitializesAndOwnsAppServerProcess(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	require.NoError(t, err)

	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath: executable,
		Env:       append(os.Environ(), "GO_WANT_CODEXAPP_HELPER=initialize"),
		ClientInfo: codexapp.ClientInfo{
			Name:    "test-client",
			Title:   "Test Client",
			Version: "1.2.3",
		},
	})
	require.NoError(t, err)
	require.Equal(t, codexapp.ServerInfo{
		CodexHome:      "/tmp/codex-home",
		PlatformFamily: "unix",
		PlatformOS:     "test",
		UserAgent:      "codex-test/1",
	}, client.ServerInfo())
	require.NoError(t, client.Close(context.Background()))
}

func TestClientListsModels(t *testing.T) {
	t.Parallel()
	client := openHelperClient(t, "models")

	models, err := client.ListModels(context.Background(), codexapp.ListModelsRequest{
		Limit:         2,
		IncludeHidden: true,
	})

	require.NoError(t, err)
	require.Equal(t, codexapp.ModelList{
		Models: []codexapp.Model{{
			ID:                     "gpt-test",
			Model:                  "gpt-test",
			DisplayName:            "GPT Test",
			Description:            "test model",
			DefaultReasoningEffort: "high",
			SupportedEfforts: []codexapp.ReasoningEffortOption{{
				Effort:      "high",
				Description: "Thorough",
			}},
			IsDefault:           true,
			SupportsPersonality: true,
		}},
		NextCursor: "next-page",
	}, models)
}

func TestClientRunsTurnAndStreamsCorrelatedEvents(t *testing.T) {
	t.Parallel()
	client := openHelperClient(t, "turn")
	thread, err := client.StartThread(context.Background(), codexapp.StartThreadRequest{
		Settings: codexapp.ThreadSettings{Model: "gpt-test", Cwd: "/workspace"},
	})
	require.NoError(t, err)
	require.Equal(t, codexapp.Thread{
		ID:        "thread-1",
		Model:     "gpt-test",
		CreatedAt: 10,
		UpdatedAt: 11,
	}, thread)

	handle, err := client.StartTurn(context.Background(), codexapp.StartTurnRequest{
		ThreadID: thread.ID,
		Input:    []codexapp.Input{codexapp.Text("hello")},
	})
	require.NoError(t, err)

	event, err := handle.Events().Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, codexapp.EventAgentMessageDelta, event.Type)
	require.Equal(t, "thread-1", event.ThreadID)
	require.Equal(t, "turn-1", event.TurnID)
	require.Equal(t, &codexapp.TextDelta{Text: "hello"}, event.TextDelta)

	event, err = handle.Events().Next(context.Background())
	require.NoError(t, err)
	require.Equal(t, codexapp.EventTurnCompleted, event.Type)
	require.Equal(t, &codexapp.Turn{ID: "turn-1", Status: codexapp.TurnStatusCompleted}, event.Turn)

	_, err = handle.Events().Next(context.Background())
	require.ErrorIs(t, err, io.EOF)
	result, err := handle.Wait(context.Background())
	require.NoError(t, err)
	require.Equal(t, codexapp.TurnResult{
		ThreadID: "thread-1",
		Turn:     codexapp.Turn{ID: "turn-1", Status: codexapp.TurnStatusCompleted},
	}, result)
}

func TestPermissionProfilesAreNegotiatedAndAppliedToThread(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	readRoot := filepath.Join(root, "read")
	writeRoot := filepath.Join(root, "write")
	executable, err := os.Executable()
	require.NoError(t, err)
	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath: executable,
		Env: append(os.Environ(),
			"GO_WANT_CODEXAPP_HELPER=permissions",
			"CODEXAPP_READ_ROOT="+readRoot,
			"CODEXAPP_WRITE_ROOT="+writeRoot,
		),
		ClientInfo:           codexapp.ClientInfo{Name: "test-client", Title: "Test Client", Version: "1.2.3"},
		RequiredCapabilities: []codexapp.Capability{codexapp.CapabilityPermissionProfiles},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })

	thread, err := client.StartThread(context.Background(), codexapp.StartThreadRequest{
		Permissions: codexapp.Permissions{
			ReadRoots:      []string{writeRoot, readRoot},
			WriteRoots:     []string{writeRoot},
			NetworkEnabled: true,
		},
	})
	require.NoError(t, err)
	require.Equal(t, "thread-permissions", thread.ID)
}

func TestServerRequestHandlerDoesNotBlockProtocolReader(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	handler := mocks.NewMockServerRequestHandler(ctrl)
	release := make(chan struct{})
	invoked := make(chan struct{})
	inboundSpan := make(chan oteltrace.SpanContext, 1)
	handler.EXPECT().HandleServerRequest(gomock.Any(), codexapp.ServerRequest{
		Type:     codexapp.ServerRequestCommandApproval,
		ThreadID: "thread-1",
		TurnID:   "turn-1",
		ItemID:   "item-1",
		CommandApproval: &codexapp.CommandApprovalRequest{
			Command:            "go test ./...",
			Cwd:                "/workspace",
			Reason:             "run tests",
			AvailableDecisions: []string{"acceptForSession", "decline"},
		},
	}).DoAndReturn(func(ctx context.Context, _ codexapp.ServerRequest) (codexapp.ServerResponse, error) {
		inboundSpan <- oteltrace.SpanContextFromContext(ctx)
		close(invoked)
		<-release
		return codexapp.AcceptCommandForSession(), nil
	})

	executable, err := os.Executable()
	require.NoError(t, err)
	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath:            executable,
		Env:                  append(os.Environ(), "GO_WANT_CODEXAPP_HELPER=server_request"),
		ClientInfo:           codexapp.ClientInfo{Name: "test-client", Title: "Test Client", Version: "1.2.3"},
		ServerRequestHandler: handler,
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = client.ListModels(ctx, codexapp.ListModelsRequest{})
	require.NoError(t, err, "a blocked handler must not block an unrelated response")
	select {
	case <-invoked:
	case <-ctx.Done():
		require.NoError(t, ctx.Err())
	}
	spanContext := <-inboundSpan
	close(release)
	require.Equal(t, "0123456789abcdef0123456789abcdef", spanContext.TraceID().String())
	_, err = client.ListModels(ctx, codexapp.ListModelsRequest{})
	require.NoError(t, err)
}

func TestRPCUsesExplicitOpenTelemetryProvidersAndW3CPropagation(t *testing.T) {
	t.Parallel()
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	executable, err := os.Executable()
	require.NoError(t, err)
	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath:  executable,
		Env:        append(os.Environ(), "GO_WANT_CODEXAPP_HELPER=telemetry"),
		ClientInfo: codexapp.ClientInfo{Name: "test-client", Title: "Test Client", Version: "1.2.3"},
		Telemetry: codexapp.Telemetry{
			TracerProvider: provider,
			Propagator:     propagation.TraceContext{},
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, client.Close(context.Background())) })

	ctx, parent := provider.Tracer("codexapp-test").Start(context.Background(), "parent")
	_, err = client.ListModels(ctx, codexapp.ListModelsRequest{})
	parent.End()
	require.NoError(t, err)

	var rpcSpan sdktrace.ReadOnlySpan
	for _, span := range recorder.Ended() {
		if span.Name() == "codexapp.model/list" {
			rpcSpan = span
			break
		}
	}
	require.NotNil(t, rpcSpan)
	require.Equal(t, parent.SpanContext().SpanID(), rpcSpan.Parent().SpanID())
}

func TestCanceledWrittenCallReportsUnknownOutcome(t *testing.T) {
	t.Parallel()
	client := openHelperClient(t, "cancel")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	_, err := client.ListModels(ctx, codexapp.ListModelsRequest{})

	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.ErrorIs(t, err, codexapp.ErrOutcomeUnknown)
	var callErr *codexapp.CallError
	require.ErrorAs(t, err, &callErr)
	require.Equal(t, "model/list", callErr.Method)
	require.True(t, callErr.OutcomeUnknown)
}

func TestSupervisorRestartsProcessWithoutReplayingFailedCall(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	require.NoError(t, err)
	statePath := filepath.Join(t.TempDir(), "generation")
	supervisor, err := codexapp.NewSupervisor(codexapp.Config{
		CodexPath: executable,
		Env: append(os.Environ(),
			"GO_WANT_CODEXAPP_HELPER=supervisor",
			"CODEXAPP_SUPERVISOR_STATE="+statePath,
		),
		ClientInfo: codexapp.ClientInfo{Name: "test-client", Title: "Test Client", Version: "1.2.3"},
	}, codexapp.RestartConfig{
		InitialDelay: time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2,
		ResetAfter:   time.Minute,
	})
	require.NoError(t, err)

	runCtx, cancelRun := context.WithCancel(context.Background())
	defer cancelRun()
	runResult := make(chan error, 1)
	go func() { runResult <- supervisor.Run(runCtx) }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	first, err := supervisor.Client(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 1, first.Generation())
	_, err = first.ListModels(ctx, codexapp.ListModelsRequest{})
	require.Error(t, err)

	second, err := supervisor.Client(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 2, second.Generation())
	require.NotSame(t, first, second)
	_, err = second.ListModels(ctx, codexapp.ListModelsRequest{})
	require.NoError(t, err)

	require.NoError(t, supervisor.Shutdown(ctx))
	require.NoError(t, <-runResult)
}

func TestActiveTurnFailsWhenProcessSessionIsLost(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	require.NoError(t, err)
	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath:  executable,
		Env:        append(os.Environ(), "GO_WANT_CODEXAPP_HELPER=crash_turn"),
		ClientInfo: codexapp.ClientInfo{Name: "test-client", Title: "Test Client", Version: "1.2.3"},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	thread, err := client.StartThread(context.Background(), codexapp.StartThreadRequest{})
	require.NoError(t, err)
	handle, err := client.StartTurn(context.Background(), codexapp.StartTurnRequest{
		ThreadID: thread.ID,
		Input:    []codexapp.Input{codexapp.Text("hello")},
	})
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, waitErr := handle.Wait(ctx)
	var sessionErr *codexapp.SessionLostError
	require.ErrorAs(t, waitErr, &sessionErr)
	require.EqualValues(t, 1, sessionErr.Generation)
	var exitErr *codexapp.ProcessExitError
	require.ErrorAs(t, waitErr, &exitErr)
	_, streamErr := handle.Events().Next(ctx)
	require.ErrorAs(t, streamErr, &sessionErr)
}

func TestClientResumesReadsAndInterruptsThread(t *testing.T) {
	t.Parallel()
	client := openHelperClient(t, "thread_lifecycle")
	thread, err := client.ResumeThread(context.Background(), codexapp.ResumeThreadRequest{ThreadID: "thread-existing"})
	require.NoError(t, err)
	require.Equal(t, "thread-existing", thread.ID)
	require.Equal(t, "gpt-test", thread.Model)
	thread, err = client.ReadThread(context.Background(), codexapp.ReadThreadRequest{
		ThreadID: "thread-existing", IncludeTurns: true,
	})
	require.NoError(t, err)
	require.Equal(t, []codexapp.Turn{{ID: "turn-old", Status: codexapp.TurnStatusCompleted}}, thread.Turns)
	handle, err := client.StartTurn(context.Background(), codexapp.StartTurnRequest{
		ThreadID: "thread-existing",
		Input:    []codexapp.Input{codexapp.Text("continue")},
	})
	require.NoError(t, err)
	require.NoError(t, handle.Interrupt(context.Background()))
	result, err := handle.Wait(context.Background())
	require.NoError(t, err)
	require.Equal(t, codexapp.TurnStatusInterrupted, result.Turn.Status)
}

func TestRPCErrorRemainsTypedThroughPublicMethod(t *testing.T) {
	t.Parallel()
	client := openHelperClient(t, "rpc_error")
	_, err := client.ListModels(context.Background(), codexapp.ListModelsRequest{})
	var rpcErr *codexapp.RPCError
	require.ErrorAs(t, err, &rpcErr)
	require.Equal(t, -32000, rpcErr.Code)
	require.Equal(t, "unavailable", rpcErr.Message)
	require.JSONEq(t, `{"retry":false}`, string(rpcErr.Data))
}

func TestOpenClassifiesMissingRequiredCapability(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	require.NoError(t, err)
	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath:            executable,
		Env:                  append(os.Environ(), "GO_WANT_CODEXAPP_HELPER=unsupported_permissions"),
		ClientInfo:           codexapp.ClientInfo{Name: "test-client", Title: "Test Client", Version: "1.2.3"},
		RequiredCapabilities: []codexapp.Capability{codexapp.CapabilityPermissionProfiles},
	})
	require.Nil(t, client)
	var capabilityErr *codexapp.UnsupportedCapabilityError
	require.ErrorAs(t, err, &capabilityErr)
	require.Equal(t, codexapp.CapabilityPermissionProfiles, capabilityErr.Capability)
	var rpcErr *codexapp.RPCError
	require.ErrorAs(t, err, &rpcErr)
	require.Equal(t, -32601, rpcErr.Code)
}

func TestProcessExitIncludesBoundedStderrTailAndForwardsStderr(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	require.NoError(t, err)
	var stderr bytes.Buffer
	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath:  executable,
		Env:        append(os.Environ(), "GO_WANT_CODEXAPP_HELPER=stderr_exit"),
		Stderr:     &stderr,
		ClientInfo: codexapp.ClientInfo{Name: "test-client", Title: "Test Client", Version: "1.2.3"},
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = client.Close(context.Background()) })
	_, err = client.ListModels(context.Background(), codexapp.ListModelsRequest{})
	var exitErr *codexapp.ProcessExitError
	require.ErrorAs(t, err, &exitErr)
	require.Contains(t, exitErr.StderrTail, "app server exploded")
	require.Contains(t, stderr.String(), "app server exploded")
}

func TestThreadAndTurnSettingsMapToProtocol(t *testing.T) {
	t.Parallel()
	client := openHelperClient(t, "settings")
	thread, err := client.StartThread(context.Background(), codexapp.StartThreadRequest{Settings: codexapp.ThreadSettings{
		ModelProvider: "openai", ApprovalPolicy: codexapp.ApprovalPolicyOnRequest,
		BaseInstructions: "base", DeveloperInstructions: "developer", Personality: codexapp.PersonalityPragmatic,
	}})
	require.NoError(t, err)
	handle, err := client.StartTurn(context.Background(), codexapp.StartTurnRequest{
		ThreadID: thread.ID, Input: []codexapp.Input{codexapp.Text("hello")}, Model: "gpt-test", Cwd: "/workspace",
		Effort: "high", Summary: codexapp.ReasoningSummaryConcise, Personality: codexapp.PersonalityFriendly,
		OutputSchema: json.RawMessage(`{"type":"object"}`),
	})
	require.NoError(t, err)
	_, err = handle.Wait(context.Background())
	require.NoError(t, err)
}

func openHelperClient(t *testing.T, mode string) *codexapp.Client {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	client, err := codexapp.Open(context.Background(), codexapp.Config{
		CodexPath: executable,
		Env:       append(os.Environ(), "GO_WANT_CODEXAPP_HELPER="+mode),
		ClientInfo: codexapp.ClientInfo{
			Name:    "test-client",
			Title:   "Test Client",
			Version: "1.2.3",
		},
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, client.Close(context.Background()))
	})
	return client
}
