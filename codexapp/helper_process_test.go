package codexapp_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if mode := os.Getenv("GO_WANT_CODEXAPP_HELPER"); mode != "" {
		os.Exit(runCodexAppServerHelper(mode))
	}
	os.Exit(m.Run())
}

func runCodexAppServerHelper(mode string) int {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return 2
	}
	var request struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			Capabilities struct {
				ExperimentalAPI bool `json:"experimentalApi"`
			} `json:"capabilities"`
			ClientInfo struct {
				Name    string `json:"name"`
				Title   string `json:"title"`
				Version string `json:"version"`
			} `json:"clientInfo"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
		return 3
	}
	if request.Method != "initialize" || request.Params.ClientInfo.Name != "test-client" ||
		request.Params.ClientInfo.Title != "Test Client" || request.Params.ClientInfo.Version != "1.2.3" {
		return 4
	}
	if mode == "permissions" && !request.Params.Capabilities.ExperimentalAPI {
		return 19
	}
	response := fmt.Sprintf(
		`{"jsonrpc":"2.0","id":%d,"result":{"codexHome":"/tmp/codex-home","platformFamily":"unix","platformOs":"test","userAgent":"codex-test/1"}}`,
		request.ID,
	)
	if _, err := fmt.Fprintln(os.Stdout, response); err != nil {
		return 5
	}
	if !scanner.Scan() {
		return 6
	}
	var notification struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &notification); err != nil || notification.Method != "initialized" {
		return 7
	}
	if code := runCodexAppServerMode(mode, scanner); code != 0 {
		return code
	}
	for scanner.Scan() {
	}
	return 0
}

func runCodexAppServerMode(mode string, scanner *bufio.Scanner) int {
	switch mode {
	case "permissions":
		return runPermissionsHelper(scanner)
	case "unsupported_permissions":
		return runUnsupportedPermissionsHelper(scanner)
	case "models":
		return runModelsHelper(scanner)
	case "telemetry":
		return runTelemetryHelper(scanner)
	case "cancel":
		return runCancelHelper(scanner)
	case "supervisor":
		return runSupervisorHelper(scanner)
	case "crash_turn":
		return runCrashTurnHelper(scanner)
	case "thread_lifecycle":
		return runThreadLifecycleHelper(scanner)
	case "rpc_error":
		return runRPCErrorHelper(scanner)
	case "stderr_exit":
		return runStderrExitHelper(scanner)
	case "settings":
		return runSettingsHelper(scanner)
	case "server_request":
		return runServerRequestHelper(scanner)
	case "turn":
		return runTurnHelper(scanner)
	default:
		return 0
	}
}

func runPermissionsHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 20
	}
	var probe struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &probe); err != nil || probe.Method != "permissionProfile/list" {
		return 21
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"data":[],"nextCursor":null}}`+"\n", probe.ID); err != nil {
		return 22
	}
	if !scanner.Scan() {
		return 23
	}
	var start struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			Permissions           string   `json:"permissions"`
			RuntimeWorkspaceRoots []string `json:"runtimeWorkspaceRoots"`
			Config                struct {
				Permissions map[string]struct {
					Filesystem map[string]string `json:"filesystem"`
					Network    struct {
						Enabled bool `json:"enabled"`
					} `json:"network"`
				} `json:"permissions"`
			} `json:"config"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &start); err != nil || start.Method != "thread/start" {
		return 24
	}
	profile, ok := start.Params.Config.Permissions[start.Params.Permissions]
	if !ok || start.Params.Permissions == "" || profile.Filesystem[":root"] != "deny" ||
		profile.Filesystem[":minimal"] != "read" || profile.Filesystem[os.Getenv("CODEXAPP_READ_ROOT")] != "read" ||
		profile.Filesystem[os.Getenv("CODEXAPP_WRITE_ROOT")] != "write" || !profile.Network.Enabled ||
		len(start.Params.RuntimeWorkspaceRoots) != 1 || start.Params.RuntimeWorkspaceRoots[0] != os.Getenv("CODEXAPP_WRITE_ROOT") {
		return 25
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"thread":{"id":"thread-permissions"}}}`+"\n", start.ID); err != nil {
		return 26
	}
	return 0
}

func runUnsupportedPermissionsHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 68
	}
	var probe struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &probe); err != nil || probe.Method != "permissionProfile/list" {
		return 69
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"error":{"code":-32601,"message":"method not found"}}`+"\n", probe.ID); err != nil {
		return 70
	}
	return 0
}

func runModelsHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 8
	}
	var modelRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			IncludeHidden bool `json:"includeHidden"`
			Limit         int  `json:"limit"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &modelRequest); err != nil ||
		modelRequest.Method != "model/list" || !modelRequest.Params.IncludeHidden || modelRequest.Params.Limit != 2 {
		return 9
	}
	models := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"data":[{`+
		`"id":"gpt-test","model":"gpt-test","displayName":"GPT Test","description":"test model",`+
		`"defaultReasoningEffort":"high","supportedReasoningEfforts":[{"reasoningEffort":"high","description":"Thorough"}],`+
		`"isDefault":true,"hidden":false,"supportsPersonality":true}],"nextCursor":"next-page"}}`, modelRequest.ID)
	if _, err := fmt.Fprintln(os.Stdout, models); err != nil {
		return 10
	}
	return 0
}

func runTelemetryHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 36
	}
	var tracedRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Trace  struct {
			Traceparent string `json:"traceparent"`
		} `json:"trace"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &tracedRequest); err != nil ||
		tracedRequest.Method != "model/list" || tracedRequest.Trace.Traceparent == "" {
		return 37
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"data":[],"nextCursor":null}}`+"\n", tracedRequest.ID); err != nil {
		return 38
	}
	return 0
}

func runCancelHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 39
	}
	var request struct {
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &request); err != nil || request.Method != "model/list" {
		return 40
	}
	return 0
}

func runSupervisorHelper(scanner *bufio.Scanner) int {
	statePath := os.Getenv("CODEXAPP_SUPERVISOR_STATE")
	_, statErr := os.Stat(statePath)
	firstGeneration := os.IsNotExist(statErr)
	if firstGeneration {
		if err := os.WriteFile(statePath, []byte("started"), 0o600); err != nil {
			return 41
		}
	}
	for scanner.Scan() {
		var request struct {
			ID     int64  `json:"id"`
			Method string `json:"method"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil || request.Method != "model/list" {
			return 42
		}
		if firstGeneration {
			return 43
		}
		if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"data":[],"nextCursor":null}}`+"\n", request.ID); err != nil {
			return 44
		}
	}
	return 0
}

func runCrashTurnHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 45
	}
	var threadRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &threadRequest); err != nil || threadRequest.Method != "thread/start" {
		return 46
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"thread":{"id":"thread-crash"}}}`+"\n", threadRequest.ID); err != nil {
		return 47
	}
	if !scanner.Scan() {
		return 48
	}
	var turnRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &turnRequest); err != nil || turnRequest.Method != "turn/start" {
		return 49
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"turn":{"id":"turn-crash","status":"inProgress"}}}`+"\n", turnRequest.ID); err != nil {
		return 50
	}
	time.Sleep(10 * time.Millisecond)
	return 51
}

func runThreadLifecycleHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 52
	}
	var resume struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			ThreadID string `json:"threadId"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &resume); err != nil || resume.Method != "thread/resume" || resume.Params.ThreadID != "thread-existing" {
		return 53
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"model":"gpt-test","thread":{"id":"thread-existing","turns":[]}}}`+"\n", resume.ID); err != nil {
		return 54
	}
	if !scanner.Scan() {
		return 55
	}
	var read struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			ThreadID     string `json:"threadId"`
			IncludeTurns bool   `json:"includeTurns"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &read); err != nil || read.Method != "thread/read" ||
		read.Params.ThreadID != "thread-existing" || !read.Params.IncludeTurns {
		return 56
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"thread":{"id":"thread-existing","turns":[{"id":"turn-old","status":"completed"}]}}}`+"\n", read.ID); err != nil {
		return 57
	}
	if !scanner.Scan() {
		return 58
	}
	var start struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &start); err != nil || start.Method != "turn/start" {
		return 59
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"turn":{"id":"turn-active","status":"inProgress"}}}`+"\n", start.ID); err != nil {
		return 60
	}
	if !scanner.Scan() {
		return 61
	}
	var interrupt struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			ThreadID string `json:"threadId"`
			TurnID   string `json:"turnId"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &interrupt); err != nil || interrupt.Method != "turn/interrupt" ||
		interrupt.Params.ThreadID != "thread-existing" || interrupt.Params.TurnID != "turn-active" {
		return 62
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{}}`+"\n", interrupt.ID); err != nil {
		return 63
	}
	if _, err := fmt.Fprintln(os.Stdout, `{"jsonrpc":"2.0","method":"turn/completed","params":{"threadId":"thread-existing","turn":{"id":"turn-active","status":"interrupted"}}}`); err != nil {
		return 64
	}
	return 0
}

func runRPCErrorHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 65
	}
	var request struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &request); err != nil || request.Method != "model/list" {
		return 66
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"error":{"code":-32000,"message":"unavailable","data":{"retry":false}}}`+"\n", request.ID); err != nil {
		return 67
	}
	return 0
}

func runStderrExitHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 71
	}
	_, _ = fmt.Fprintln(os.Stderr, "app server exploded")
	return 72
}

func runSettingsHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 73
	}
	var startThread struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			ModelProvider         string `json:"modelProvider"`
			ApprovalPolicy        string `json:"approvalPolicy"`
			BaseInstructions      string `json:"baseInstructions"`
			DeveloperInstructions string `json:"developerInstructions"`
			Personality           string `json:"personality"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &startThread); err != nil || startThread.Method != "thread/start" ||
		startThread.Params.ModelProvider != "openai" || startThread.Params.ApprovalPolicy != "on-request" ||
		startThread.Params.BaseInstructions != "base" || startThread.Params.DeveloperInstructions != "developer" ||
		startThread.Params.Personality != "pragmatic" {
		return 74
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"thread":{"id":"thread-settings"}}}`+"\n", startThread.ID); err != nil {
		return 75
	}
	if !scanner.Scan() {
		return 76
	}
	var startTurn struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			Model        string          `json:"model"`
			Cwd          string          `json:"cwd"`
			Effort       string          `json:"effort"`
			Summary      string          `json:"summary"`
			Personality  string          `json:"personality"`
			OutputSchema json.RawMessage `json:"outputSchema"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &startTurn); err != nil || startTurn.Method != "turn/start" ||
		startTurn.Params.Model != "gpt-test" || startTurn.Params.Cwd != "/workspace" || startTurn.Params.Effort != "high" ||
		startTurn.Params.Summary != "concise" || startTurn.Params.Personality != "friendly" || len(startTurn.Params.OutputSchema) == 0 {
		return 77
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"turn":{"id":"turn-settings","status":"inProgress"}}}`+"\n", startTurn.ID); err != nil {
		return 78
	}
	if _, err := fmt.Fprintln(os.Stdout, `{"jsonrpc":"2.0","method":"turn/completed","params":{"threadId":"thread-settings","turn":{"id":"turn-settings","status":"completed"}}}`); err != nil {
		return 79
	}
	return 0
}

func runServerRequestHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 27
	}
	var firstModelRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &firstModelRequest); err != nil || firstModelRequest.Method != "model/list" {
		return 28
	}
	if _, err := fmt.Fprintln(os.Stdout, `{"jsonrpc":"2.0","id":"approval-1","method":"item/commandExecution/requestApproval","trace":{"traceparent":"00-0123456789abcdef0123456789abcdef-0123456789abcdef-01"},"params":{"threadId":"thread-1","turnId":"turn-1","itemId":"item-1","command":"go test ./...","cwd":"/workspace","reason":"run tests","availableDecisions":["acceptForSession","decline"]}}`); err != nil {
		return 29
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"data":[],"nextCursor":null}}`+"\n", firstModelRequest.ID); err != nil {
		return 30
	}
	var secondModelID int64
	approvalReceived := false
	for secondModelID == 0 || !approvalReceived {
		if !scanner.Scan() {
			return 31
		}
		var message struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Result struct {
				Decision string `json:"decision"`
			} `json:"result"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &message); err != nil {
			return 32
		}
		if message.Method == "model/list" {
			if err := json.Unmarshal(message.ID, &secondModelID); err != nil {
				return 33
			}
		}
		if string(message.ID) == `"approval-1"` {
			if message.Result.Decision != "acceptForSession" {
				return 34
			}
			approvalReceived = true
		}
	}
	if _, err := fmt.Fprintf(os.Stdout, `{"jsonrpc":"2.0","id":%d,"result":{"data":[],"nextCursor":null}}`+"\n", secondModelID); err != nil {
		return 35
	}
	return 0
}

func runTurnHelper(scanner *bufio.Scanner) int {
	if !scanner.Scan() {
		return 11
	}
	var threadRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			Model string `json:"model"`
			Cwd   string `json:"cwd"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &threadRequest); err != nil ||
		threadRequest.Method != "thread/start" || threadRequest.Params.Model != "gpt-test" || threadRequest.Params.Cwd != "/workspace" {
		return 12
	}
	thread := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"model":"gpt-test","thread":{`+
		`"id":"thread-1","preview":"","createdAt":10,"updatedAt":11}}}`, threadRequest.ID)
	if _, err := fmt.Fprintln(os.Stdout, thread); err != nil {
		return 13
	}
	if !scanner.Scan() {
		return 14
	}
	var turnRequest struct {
		ID     int64  `json:"id"`
		Method string `json:"method"`
		Params struct {
			ThreadID string `json:"threadId"`
			Input    []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"input"`
		} `json:"params"`
	}
	if err := json.Unmarshal(scanner.Bytes(), &turnRequest); err != nil || turnRequest.Method != "turn/start" ||
		turnRequest.Params.ThreadID != "thread-1" || len(turnRequest.Params.Input) != 1 ||
		turnRequest.Params.Input[0].Type != "text" || turnRequest.Params.Input[0].Text != "hello" {
		return 15
	}
	started := fmt.Sprintf(`{"jsonrpc":"2.0","id":%d,"result":{"turn":{"id":"turn-1","status":"inProgress"}}}`, turnRequest.ID)
	if _, err := fmt.Fprintln(os.Stdout, started); err != nil {
		return 16
	}
	if _, err := fmt.Fprintln(os.Stdout, `{"jsonrpc":"2.0","method":"item/agentMessage/delta","params":{"threadId":"thread-1","turnId":"turn-1","itemId":"item-1","delta":"hello"}}`); err != nil {
		return 17
	}
	if _, err := fmt.Fprintln(os.Stdout, `{"jsonrpc":"2.0","method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1","status":"completed"}}}`); err != nil {
		return 18
	}
	return 0
}
