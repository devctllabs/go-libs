package codexapp

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/devctllabs/go-libs/codexapp/internal/protocol"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// SchemaCodexVersion is the Codex CLI version that produced the checked-in schema snapshot.
const SchemaCodexVersion = protocol.CodexVersion

// Config defines how the Codex App Server process is opened.
type Config struct {
	CodexPath            string
	ProjectDir           string
	Env                  []string
	Stderr               io.Writer
	ClientInfo           ClientInfo
	RequiredCapabilities []Capability
	ServerRequestHandler ServerRequestHandler
	Telemetry            Telemetry
	TurnEventBuffer      int
	MaxMessageBytes      int
}

// Telemetry supplies instance-scoped OpenTelemetry dependencies. Nil fields use no-op providers
// and W3C Trace Context plus Baggage propagation; package globals are never read or changed.
type Telemetry struct {
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider
	Propagator     propagation.TextMapPropagator
}

// Capability identifies an optional App Server protocol feature.
type Capability string

const (
	// CapabilityPermissionProfiles enables beta named permission profiles.
	CapabilityPermissionProfiles Capability = "permissionProfiles"
)

// ClientInfo identifies the caller during the App Server handshake.
type ClientInfo struct {
	Name    string `json:"name"`
	Title   string `json:"title,omitempty"`
	Version string `json:"version"`
}

// ServerInfo describes the initialized Codex App Server process.
type ServerInfo struct {
	CodexHome      string
	PlatformFamily string
	PlatformOS     string
	UserAgent      string
}

// ListModelsRequest controls model catalog pagination.
type ListModelsRequest struct {
	Cursor        string
	Limit         int
	IncludeHidden bool
}

// ModelList is one page of models available from Codex.
type ModelList struct {
	Models     []Model
	NextCursor string
}

// Model is the stable subset of model catalog metadata exposed by this package.
type Model struct {
	ID                     string
	Model                  string
	DisplayName            string
	Description            string
	DefaultReasoningEffort string
	SupportedEfforts       []ReasoningEffortOption
	IsDefault              bool
	Hidden                 bool
	SupportsPersonality    bool
}

// ReasoningEffortOption describes one reasoning effort supported by a model.
type ReasoningEffortOption struct {
	Effort      string
	Description string
}

// Client is a connection to one Codex App Server process.
type Client struct {
	info                 ServerInfo
	command              *exec.Cmd
	stdin                io.WriteCloser
	writeMu              sync.Mutex
	nextID               atomic.Int64
	pendingMu            sync.Mutex
	pending              map[int64]chan rpcResponse
	closeOnce            sync.Once
	doneOnce             sync.Once
	done                 chan struct{}
	terminalMu           sync.Mutex
	terminalErr          error
	processDone          chan struct{}
	waitMu               sync.Mutex
	waitErr              error
	turnsMu              sync.Mutex
	turns                map[string]*TurnHandle
	starting             map[string]*TurnHandle
	capabilities         map[Capability]bool
	serverRequestHandler ServerRequestHandler
	handlerCtx           context.Context
	handlerCancel        context.CancelFunc
	instrumentation      *instrumentation
	generation           int64
	eventBuffer          int
	maxMessageBytes      int
	stderrTail           *tailWriter
	profiles             map[string]normalizedPermissionProfile
}

// ServerInfo returns immutable metadata from the initialize response.
func (c *Client) ServerInfo() ServerInfo {
	return c.info
}

// Generation identifies the process generation. Standalone clients use generation one.
func (c *Client) Generation() int64 { return c.generation }

// ListModels returns one page of models advertised by App Server.
func (c *Client) ListModels(ctx context.Context, request ListModelsRequest) (ModelList, error) {
	params := modelListParams{IncludeHidden: &request.IncludeHidden}
	if request.Cursor != "" {
		params.Cursor = &request.Cursor
	}
	if request.Limit > 0 {
		limit := int64(request.Limit)
		params.Limit = &limit
	}
	var response protocol.ModelListResponse
	if err := c.call(ctx, "model/list", params, &response); err != nil {
		return ModelList{}, fmt.Errorf("codexapp: list models: %w", err)
	}
	models := make([]Model, 0, len(response.Data))
	for _, model := range response.Data {
		efforts := make([]ReasoningEffortOption, 0, len(model.SupportedReasoningEfforts))
		for _, effort := range model.SupportedReasoningEfforts {
			efforts = append(efforts, ReasoningEffortOption{
				Effort:      effort.ReasoningEffort,
				Description: effort.Description,
			})
		}
		models = append(models, Model{
			ID:                     model.ID,
			Model:                  model.Model,
			DisplayName:            model.DisplayName,
			Description:            model.Description,
			DefaultReasoningEffort: model.DefaultReasoningEffort,
			SupportedEfforts:       efforts,
			IsDefault:              model.IsDefault,
			Hidden:                 model.Hidden,
			SupportsPersonality:    model.SupportsPersonality != nil && *model.SupportsPersonality,
		})
	}
	result := ModelList{Models: models}
	if response.NextCursor != nil {
		result.NextCursor = *response.NextCursor
	}
	return result, nil
}

// Close stops the owned App Server process and waits for it to exit.
func (c *Client) Close(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("codexapp: context is required")
	}
	c.closeOnce.Do(func() {
		_ = c.stdin.Close()
	})
	select {
	case <-c.processDone:
		c.waitMu.Lock()
		defer c.waitMu.Unlock()
		return c.waitErr
	case <-ctx.Done():
		_ = c.command.Process.Kill()
		<-c.processDone
		return ctx.Err()
	}
}

// Open starts and initializes a Codex App Server client.
func Open(ctx context.Context, cfg Config) (*Client, error) {
	if ctx == nil {
		return nil, fmt.Errorf("codexapp: context is required")
	}
	var err error
	cfg, err = normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}
	capabilities, err := requiredCapabilities(cfg.RequiredCapabilities)
	if err != nil {
		return nil, err
	}
	instrumentation, err := newInstrumentation(cfg.Telemetry)
	if err != nil {
		return nil, fmt.Errorf("codexapp: create telemetry instruments: %w", err)
	}
	process, err := startAppServer(cfg)
	if err != nil {
		return nil, err
	}
	client := newClient(cfg, capabilities, instrumentation, process)
	go client.readLoop(process.stdout, client.maxMessageBytes)
	go client.waitForProcess()
	if err := client.initializeOpen(ctx, cfg.ClientInfo); err != nil {
		client.abortOpen()
		return nil, err
	}
	return client, nil
}

type appServerProcess struct {
	command    *exec.Cmd
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderrTail *tailWriter
}

func requiredCapabilities(required []Capability) (map[Capability]bool, error) {
	capabilities := make(map[Capability]bool, len(required))
	for _, capability := range required {
		if capability != CapabilityPermissionProfiles {
			return nil, fmt.Errorf("codexapp: unknown required capability %q", capability)
		}
		capabilities[capability] = true
	}
	return capabilities, nil
}

func startAppServer(cfg Config) (*appServerProcess, error) {
	path := strings.TrimSpace(cfg.CodexPath)
	if path == "" {
		var err error
		path, err = exec.LookPath("codex")
		if err != nil {
			return nil, fmt.Errorf("codexapp: find codex executable: %w", err)
		}
	}
	//nolint:gosec // CodexPath is an explicit caller-selected executable.
	command := exec.Command(path, "app-server", "--listen", "stdio://")
	command.Dir = cfg.ProjectDir
	if cfg.Env != nil {
		command.Env = append([]string(nil), cfg.Env...)
	}
	stderrTail := newTailWriter(64 << 10)
	command.Stderr = stderrTail
	if cfg.Stderr != nil {
		command.Stderr = io.MultiWriter(stderrTail, cfg.Stderr)
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("codexapp: open stdin: %w", err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("codexapp: open stdout: %w", err)
	}
	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("codexapp: start app server: %w", err)
	}
	return &appServerProcess{command: command, stdin: stdin, stdout: stdout, stderrTail: stderrTail}, nil
}

func newClient(
	cfg Config,
	capabilities map[Capability]bool,
	instrumentation *instrumentation,
	process *appServerProcess,
) *Client {
	handlerCtx, handlerCancel := context.WithCancel(context.Background())
	return &Client{
		command:              process.command,
		stdin:                process.stdin,
		pending:              make(map[int64]chan rpcResponse),
		done:                 make(chan struct{}),
		processDone:          make(chan struct{}),
		turns:                make(map[string]*TurnHandle),
		starting:             make(map[string]*TurnHandle),
		capabilities:         capabilities,
		serverRequestHandler: cfg.ServerRequestHandler,
		handlerCtx:           handlerCtx,
		handlerCancel:        handlerCancel,
		instrumentation:      instrumentation,
		generation:           1,
		eventBuffer:          cfg.TurnEventBuffer,
		maxMessageBytes:      cfg.MaxMessageBytes,
		stderrTail:           process.stderrTail,
		profiles:             make(map[string]normalizedPermissionProfile),
	}
}

func (c *Client) waitForProcess() {
	err := c.command.Wait()
	c.waitMu.Lock()
	c.waitErr = err
	c.waitMu.Unlock()
	terminalErr := io.EOF
	if err != nil {
		terminalErr = &ProcessExitError{Cause: err, StderrTail: c.stderrTail.String()}
	}
	c.finish(terminalErr)
	close(c.processDone)
}

func (c *Client) initializeOpen(ctx context.Context, info ClientInfo) error {
	if err := c.initialize(ctx, info); err != nil {
		return err
	}
	if !c.capabilities[CapabilityPermissionProfiles] {
		return nil
	}
	var profiles protocol.PermissionProfileListResponse
	if err := c.call(ctx, "permissionProfile/list", permissionProfileListParams{}, &profiles); err != nil {
		return &UnsupportedCapabilityError{Capability: CapabilityPermissionProfiles, Cause: err}
	}
	return nil
}

func (c *Client) abortOpen() {
	_ = c.stdin.Close()
	_ = c.command.Process.Kill()
	<-c.processDone
}

func normalizeConfig(cfg Config) (Config, error) {
	if strings.TrimSpace(cfg.ClientInfo.Name) == "" {
		return Config{}, fmt.Errorf("codexapp: client info name is required")
	}
	if strings.TrimSpace(cfg.ClientInfo.Version) == "" {
		return Config{}, fmt.Errorf("codexapp: client info version is required")
	}
	if cfg.TurnEventBuffer < 0 {
		return Config{}, fmt.Errorf("codexapp: turn event buffer must not be negative")
	}
	if cfg.TurnEventBuffer == 0 {
		cfg.TurnEventBuffer = 256
	}
	if cfg.MaxMessageBytes < 0 {
		return Config{}, fmt.Errorf("codexapp: max message bytes must not be negative")
	}
	if cfg.MaxMessageBytes == 0 {
		cfg.MaxMessageBytes = defaultMaxMessageBytes
	}
	if cfg.ProjectDir != "" {
		absolute, err := filepath.Abs(cfg.ProjectDir)
		if err != nil {
			return Config{}, fmt.Errorf("codexapp: resolve project directory: %w", err)
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return Config{}, fmt.Errorf("codexapp: inspect project directory: %w", err)
		}
		if !info.IsDir() {
			return Config{}, fmt.Errorf("codexapp: project directory %q is not a directory", absolute)
		}
		cfg.ProjectDir = filepath.Clean(absolute)
	}
	return cfg, nil
}

func (c *Client) initialize(ctx context.Context, clientInfo ClientInfo) error {
	params := initializeParams{ClientInfo: clientInfo}
	if c.capabilities[CapabilityPermissionProfiles] {
		enabled := true
		params.Capabilities = &initializeCapabilities{ExperimentalAPI: &enabled}
	}
	var response protocol.InitializeResponse
	if err := c.call(ctx, "initialize", params, &response); err != nil {
		return fmt.Errorf("codexapp: initialize: %w", err)
	}
	c.info = ServerInfo{
		CodexHome:      response.CodexHome,
		PlatformFamily: response.PlatformFamily,
		PlatformOS:     response.PlatformOS,
		UserAgent:      response.UserAgent,
	}
	if err := c.notify("initialized", nil); err != nil {
		return fmt.Errorf("codexapp: initialized notification: %w", err)
	}
	return nil
}
