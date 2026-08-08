package codexapp

import (
	"context"
	"encoding/json"
	"fmt"
)

//go:generate go tool mockgen -destination mocks/server_requests.gen.go -package mocks . ServerRequestHandler

// ServerRequestHandler handles a request initiated by App Server.
type ServerRequestHandler interface {
	// HandleServerRequest decides a server-initiated request. Implementations may block; the
	// protocol reader continues independently and the context ends with the client session.
	HandleServerRequest(ctx context.Context, request ServerRequest) (ServerResponse, error)
}

// ServerRequestType identifies the payload populated on ServerRequest.
type ServerRequestType string

const (
	// ServerRequestCommandApproval asks whether a command may execute.
	ServerRequestCommandApproval ServerRequestType = "item/commandExecution/requestApproval"
	// ServerRequestFileChangeApproval asks whether file modifications may be applied.
	ServerRequestFileChangeApproval ServerRequestType = "item/fileChange/requestApproval"
	// ServerRequestPermissionApproval asks for additional sandbox permissions.
	ServerRequestPermissionApproval ServerRequestType = "item/permissions/requestApproval"
	// ServerRequestUserInput asks the host to answer structured questions.
	ServerRequestUserInput ServerRequestType = "item/tool/requestUserInput"
	// ServerRequestMCPElicitation forwards an MCP server elicitation.
	ServerRequestMCPElicitation ServerRequestType = "mcpServer/elicitation/request"
)

// ServerRequest is a tagged union for requests initiated by App Server.
type ServerRequest struct {
	Type               ServerRequestType
	ThreadID           string
	TurnID             string
	ItemID             string
	CommandApproval    *CommandApprovalRequest
	FileChangeApproval *FileChangeApprovalRequest
	PermissionApproval *PermissionApprovalRequest
	UserInput          *UserInputRequest
	MCPElicitation     *MCPElicitationRequest
}

// CommandApprovalRequest describes a command execution approval request.
type CommandApprovalRequest struct {
	Command            string
	Cwd                string
	Reason             string
	AvailableDecisions []string
	StartedAtMS        int64
}

// FileChangeApprovalRequest describes a file-change approval request.
type FileChangeApprovalRequest struct {
	Reason      string
	GrantRoot   string
	StartedAtMS int64
}

// PermissionApprovalRequest preserves the requested profile as schema-versioned JSON.
type PermissionApprovalRequest struct {
	Cwd         string
	Reason      string
	StartedAtMS int64
	Requested   json.RawMessage
}

// UserInputRequest contains structured questions from a tool call.
type UserInputRequest struct {
	Questions        []UserInputQuestion
	AutoResolutionMS *int64
}

// UserInputQuestion describes one structured question.
type UserInputQuestion struct {
	ID       string
	Header   string
	Question string
	Options  []UserInputOption
	IsOther  bool
	IsSecret bool
}

// UserInputOption is one selectable answer.
type UserInputOption struct {
	Label       string
	Description string
}

// MCPElicitationRequest describes a downstream MCP elicitation.
type MCPElicitationRequest struct {
	ServerName      string
	Mode            string
	Message         string
	URL             string
	RequestedSchema json.RawMessage
}

// ServerResponse is a typed response to a ServerRequest.
type ServerResponse struct {
	result   any
	decision string
}

// AcceptCommandForSession approves the command and equivalent requests for this session.
func AcceptCommandForSession() ServerResponse {
	return decisionResponse("acceptForSession")
}

// AcceptCommand approves one command.
func AcceptCommand() ServerResponse { return decisionResponse("accept") }

// DeclineCommand declines one command.
func DeclineCommand() ServerResponse { return decisionResponse("decline") }

// CancelCommand cancels the approval flow.
func CancelCommand() ServerResponse { return decisionResponse("cancel") }

// AnswerUserInput responds to structured questions by question id.
func AnswerUserInput(answers map[string][]string) ServerResponse {
	wire := make(map[string]struct {
		Answers []string `json:"answers"`
	}, len(answers))
	for id, values := range answers {
		wire[id] = struct {
			Answers []string `json:"answers"`
		}{Answers: append([]string(nil), values...)}
	}
	return ServerResponse{result: struct {
		Answers any `json:"answers"`
	}{Answers: wire}}
}

// DeclineMCPElicitation declines an MCP elicitation without content.
func DeclineMCPElicitation() ServerResponse {
	return ServerResponse{result: struct {
		Action  string `json:"action"`
		Content any    `json:"content"`
	}{Action: "decline"}}
}

func decisionResponse(decision string) ServerResponse {
	return ServerResponse{decision: decision, result: struct {
		Decision string `json:"decision"`
	}{Decision: decision}}
}

func decodeServerRequest(method string, params json.RawMessage) (ServerRequest, error) {
	var common struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		ItemID   string `json:"itemId"`
	}
	if err := json.Unmarshal(params, &common); err != nil {
		return ServerRequest{}, fmt.Errorf("decode %s params: %w", method, err)
	}
	request := ServerRequest{Type: ServerRequestType(method), ThreadID: common.ThreadID, TurnID: common.TurnID, ItemID: common.ItemID}
	var err error
	switch request.Type {
	case ServerRequestCommandApproval:
		request.CommandApproval, err = decodeCommandApproval(params)
	case ServerRequestFileChangeApproval:
		request.FileChangeApproval, err = decodeFileChangeApproval(params)
	case ServerRequestPermissionApproval:
		request.PermissionApproval, err = decodePermissionApproval(params)
	case ServerRequestUserInput:
		request.UserInput, err = decodeUserInput(params)
	case ServerRequestMCPElicitation:
		request.MCPElicitation, err = decodeMCPElicitation(params)
	default:
		return ServerRequest{}, fmt.Errorf("unsupported server request method %q", method)
	}
	if err != nil {
		return ServerRequest{}, err
	}
	return request, nil
}

func decodeCommandApproval(params json.RawMessage) (*CommandApprovalRequest, error) {
	var raw struct {
		Command            string   `json:"command"`
		Cwd                string   `json:"cwd"`
		Reason             string   `json:"reason"`
		AvailableDecisions []string `json:"availableDecisions"`
		StartedAtMS        int64    `json:"startedAtMs"`
	}
	if err := json.Unmarshal(params, &raw); err != nil {
		return nil, err
	}
	return &CommandApprovalRequest{
		Command: raw.Command, Cwd: raw.Cwd, Reason: raw.Reason,
		AvailableDecisions: append([]string(nil), raw.AvailableDecisions...), StartedAtMS: raw.StartedAtMS,
	}, nil
}

func decodeFileChangeApproval(params json.RawMessage) (*FileChangeApprovalRequest, error) {
	var raw struct {
		Reason      string `json:"reason"`
		GrantRoot   string `json:"grantRoot"`
		StartedAtMS int64  `json:"startedAtMs"`
	}
	if err := json.Unmarshal(params, &raw); err != nil {
		return nil, err
	}
	approval := FileChangeApprovalRequest(raw)
	return &approval, nil
}

func decodePermissionApproval(params json.RawMessage) (*PermissionApprovalRequest, error) {
	var raw struct {
		Cwd         string          `json:"cwd"`
		Reason      string          `json:"reason"`
		StartedAtMS int64           `json:"startedAtMs"`
		Permissions json.RawMessage `json:"permissions"`
	}
	if err := json.Unmarshal(params, &raw); err != nil {
		return nil, err
	}
	return &PermissionApprovalRequest{
		Cwd: raw.Cwd, Reason: raw.Reason, StartedAtMS: raw.StartedAtMS,
		Requested: append(json.RawMessage(nil), raw.Permissions...),
	}, nil
}

func decodeUserInput(params json.RawMessage) (*UserInputRequest, error) {
	var raw struct {
		AutoResolutionMS *int64 `json:"autoResolutionMs"`
		Questions        []struct {
			ID       string `json:"id"`
			Header   string `json:"header"`
			Question string `json:"question"`
			IsOther  bool   `json:"isOther"`
			IsSecret bool   `json:"isSecret"`
			Options  []struct {
				Label       string `json:"label"`
				Description string `json:"description"`
			} `json:"options"`
		} `json:"questions"`
	}
	if err := json.Unmarshal(params, &raw); err != nil {
		return nil, err
	}
	questions := make([]UserInputQuestion, 0, len(raw.Questions))
	for _, question := range raw.Questions {
		options := make([]UserInputOption, 0, len(question.Options))
		for _, option := range question.Options {
			options = append(options, UserInputOption(option))
		}
		questions = append(questions, UserInputQuestion{
			ID: question.ID, Header: question.Header, Question: question.Question,
			Options: options, IsOther: question.IsOther, IsSecret: question.IsSecret,
		})
	}
	return &UserInputRequest{Questions: questions, AutoResolutionMS: raw.AutoResolutionMS}, nil
}

func decodeMCPElicitation(params json.RawMessage) (*MCPElicitationRequest, error) {
	var raw struct {
		ServerName      string          `json:"serverName"`
		Mode            string          `json:"mode"`
		Message         string          `json:"message"`
		URL             string          `json:"url"`
		RequestedSchema json.RawMessage `json:"requestedSchema"`
	}
	if err := json.Unmarshal(params, &raw); err != nil {
		return nil, err
	}
	return &MCPElicitationRequest{
		ServerName: raw.ServerName, Mode: raw.Mode, Message: raw.Message, URL: raw.URL,
		RequestedSchema: append(json.RawMessage(nil), raw.RequestedSchema...),
	}, nil
}

func (c *Client) dispatchServerRequest(id json.RawMessage, method string, params json.RawMessage, traceCarrier map[string]string) {
	request, err := decodeServerRequest(method, params)
	if err != nil {
		go func() {
			_ = c.writeJSON(struct {
				JSONRPC string          `json:"jsonrpc"`
				ID      json.RawMessage `json:"id"`
				Error   RPCError        `json:"error"`
			}{"2.0", id, RPCError{Code: -32601, Message: "method not supported"}}, "server request error")
		}()
		return
	}
	requestCtx := c.instrumentation.extract(c.handlerCtx, traceCarrier)
	go c.handleServerRequest(requestCtx, id, request)
}

func (c *Client) handleServerRequest(ctx context.Context, id json.RawMessage, request ServerRequest) {
	response := defaultServerResponse(request.Type)
	if c.serverRequestHandler != nil {
		selected, err := c.serverRequestHandler.HandleServerRequest(ctx, request)
		if err == nil && selected.result != nil {
			response = selected
		}
	}
	_ = c.writeJSON(struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  any             `json:"result"`
	}{JSONRPC: "2.0", ID: id, Result: response.result}, "server request response")
}

func defaultServerResponse(requestType ServerRequestType) ServerResponse {
	switch requestType {
	case ServerRequestCommandApproval, ServerRequestFileChangeApproval:
		return DeclineCommand()
	case ServerRequestPermissionApproval:
		return ServerResponse{result: struct {
			Permissions map[string]any `json:"permissions"`
			Scope       string         `json:"scope"`
		}{Permissions: map[string]any{}, Scope: "turn"}}
	case ServerRequestUserInput:
		return AnswerUserInput(map[string][]string{})
	case ServerRequestMCPElicitation:
		return DeclineMCPElicitation()
	default:
		return DeclineCommand()
	}
}
