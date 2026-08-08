package codexapp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// ThreadSettings contains sticky settings for a new thread.
type ThreadSettings struct {
	Model                 string
	ModelProvider         string
	Cwd                   string
	ApprovalPolicy        ApprovalPolicy
	BaseInstructions      string
	DeveloperInstructions string
	Personality           Personality
}

type ApprovalPolicy string

const (
	ApprovalPolicyUntrusted ApprovalPolicy = "untrusted"
	ApprovalPolicyOnRequest ApprovalPolicy = "on-request"
	ApprovalPolicyNever     ApprovalPolicy = "never"
)

type Personality string

const (
	PersonalityFriendly  Personality = "friendly"
	PersonalityPragmatic Personality = "pragmatic"
	PersonalityNone      Personality = "none"
)

type ReasoningSummary string

const (
	ReasoningSummaryAuto     ReasoningSummary = "auto"
	ReasoningSummaryConcise  ReasoningSummary = "concise"
	ReasoningSummaryDetailed ReasoningSummary = "detailed"
	ReasoningSummaryNone     ReasoningSummary = "none"
)

// StartThreadRequest creates a thread with the supplied settings.
type StartThreadRequest struct {
	Settings    ThreadSettings
	Permissions Permissions
}

// ResumeThreadRequest loads an existing thread and optionally reapplies sticky settings.
type ResumeThreadRequest struct {
	ThreadID    string
	Settings    ThreadSettings
	Permissions Permissions
}

// ReadThreadRequest reads persisted thread state without resuming it.
type ReadThreadRequest struct {
	ThreadID     string
	IncludeTurns bool
}

// Permissions defines exact local roots and network access for a generated beta profile.
type Permissions struct {
	ReadRoots      []string
	WriteRoots     []string
	NetworkEnabled bool
}

// Thread is the stable thread projection returned by this package.
type Thread struct {
	ID        string
	Model     string
	Preview   string
	CreatedAt int64
	UpdatedAt int64
	Turns     []Turn
}

// Input is one user input item. Construct it with Text and the other typed constructors.
type Input struct {
	kind string
	text string
	url  string
	path string
	name string
}

const (
	inputTypeText       = "text"
	inputTypeImage      = "image"
	inputTypeLocalImage = "localImage"
	inputTypeAudio      = "audio"
	inputTypeLocalAudio = "localAudio"
	inputTypeSkill      = "skill"
	inputTypeMention    = "mention"
)

// Text constructs a text input item.
func Text(value string) Input {
	return Input{kind: inputTypeText, text: value}
}

// Image constructs a remote image input.
func Image(url string) Input { return Input{kind: inputTypeImage, url: url} }

// LocalImage constructs a local image input.
func LocalImage(path string) Input { return Input{kind: inputTypeLocalImage, path: path} }

// Audio constructs a remote audio input.
func Audio(url string) Input { return Input{kind: inputTypeAudio, url: url} }

// LocalAudio constructs a local audio input.
func LocalAudio(path string) Input { return Input{kind: inputTypeLocalAudio, path: path} }

// Skill constructs an explicit skill reference.
func Skill(name string, path string) Input {
	return Input{kind: inputTypeSkill, name: name, path: path}
}

// Mention constructs a named file or capability mention.
func Mention(name string, path string) Input {
	return Input{kind: inputTypeMention, name: name, path: path}
}

// StartTurnRequest starts a turn in an existing thread.
type StartTurnRequest struct {
	ThreadID     string
	Input        []Input
	Permissions  Permissions
	Model        string
	Cwd          string
	Effort       string
	Summary      ReasoningSummary
	Personality  Personality
	OutputSchema json.RawMessage
}

// InterruptTurnRequest identifies an active turn to interrupt.
type InterruptTurnRequest struct {
	ThreadID string
	TurnID   string
}

// TurnStatus is the App Server lifecycle status of a turn.
type TurnStatus string

const (
	// TurnStatusInProgress indicates that Codex is still working.
	TurnStatusInProgress TurnStatus = "inProgress"
	// TurnStatusCompleted indicates successful completion.
	TurnStatusCompleted TurnStatus = "completed"
	// TurnStatusFailed indicates terminal failure.
	TurnStatusFailed TurnStatus = "failed"
	// TurnStatusInterrupted indicates caller interruption.
	TurnStatusInterrupted TurnStatus = "interrupted"
)

// Turn is the stable turn projection returned by this package.
type Turn struct {
	ID     string
	Status TurnStatus
}

// TurnResult is the terminal value returned by TurnHandle.Wait.
type TurnResult struct {
	ThreadID string
	Turn     Turn
}

// TurnHandle owns the event stream and terminal result for one turn.
type TurnHandle struct {
	client   *Client
	threadID string
	stream   *EventStream
	done     chan struct{}
	doneOnce sync.Once
	mu       sync.Mutex
	turn     Turn
	result   TurnResult
	err      error
	terminal bool
}

// Events returns the bounded stream of notifications correlated to the turn.
func (h *TurnHandle) Events() *EventStream {
	return h.stream
}

// Wait blocks until the turn completes or ctx is canceled.
func (h *TurnHandle) Wait(ctx context.Context) (TurnResult, error) {
	if ctx == nil {
		return TurnResult{}, fmt.Errorf("codexapp: context is required")
	}
	select {
	case <-ctx.Done():
		return TurnResult{}, ctx.Err()
	case <-h.done:
		h.mu.Lock()
		defer h.mu.Unlock()
		return h.result, h.err
	}
}

// Interrupt requests interruption of this turn.
func (h *TurnHandle) Interrupt(ctx context.Context) error {
	h.mu.Lock()
	turnID := h.turn.ID
	h.mu.Unlock()
	if turnID == "" {
		return fmt.Errorf("codexapp: turn id is not available")
	}
	return h.client.InterruptTurn(ctx, InterruptTurnRequest{ThreadID: h.threadID, TurnID: turnID})
}

// StartThread starts a new Codex thread.
func (c *Client) StartThread(ctx context.Context, request StartThreadRequest) (Thread, error) {
	params := threadStartParams{}
	var boundProfile *normalizedPermissionProfile
	if request.Settings.Model != "" {
		params.Model = &request.Settings.Model
	}
	if request.Settings.Cwd != "" {
		params.Cwd = &request.Settings.Cwd
	}
	applyThreadSettings(&params, request.Settings)
	if !request.Permissions.empty() {
		if !c.capabilities[CapabilityPermissionProfiles] {
			return Thread{}, fmt.Errorf("codexapp: permission profiles capability was not requested")
		}
		profile, err := buildPermissionProfile(request.Permissions)
		if err != nil {
			return Thread{}, err
		}
		params.Permissions = &profile.ID
		params.RuntimeWorkspaceRoots = profile.WriteRoots
		params.Config = map[string]any{
			"permissions": map[string]permissionProfileConfig{profile.ID: profile.Config},
		}
		boundProfile = &profile
	}
	var response threadStartResponse
	if err := c.call(ctx, "thread/start", params, &response); err != nil {
		return Thread{}, fmt.Errorf("codexapp: start thread: %w", err)
	}
	thread := Thread{
		ID:        response.Thread.ID,
		Model:     response.Model,
		Preview:   response.Thread.Preview,
		CreatedAt: response.Thread.CreatedAt,
		UpdatedAt: response.Thread.UpdatedAt,
	}
	if boundProfile != nil {
		c.turnsMu.Lock()
		c.profiles[thread.ID] = *boundProfile
		c.turnsMu.Unlock()
	}
	return thread, nil
}

// ResumeThread loads an existing thread into the current App Server session.
func (c *Client) ResumeThread(ctx context.Context, request ResumeThreadRequest) (Thread, error) {
	if strings.TrimSpace(request.ThreadID) == "" {
		return Thread{}, fmt.Errorf("codexapp: thread id is required")
	}
	params := threadResumeParams{ThreadID: request.ThreadID}
	if request.Settings.Model != "" {
		params.Model = &request.Settings.Model
	}
	if request.Settings.Cwd != "" {
		params.Cwd = &request.Settings.Cwd
	}
	var response threadStartResponse
	if err := c.call(ctx, "thread/resume", params, &response); err != nil {
		return Thread{}, fmt.Errorf("codexapp: resume thread: %w", err)
	}
	return mapThread(response.Thread, response.Model), nil
}

// ReadThread reads thread state without making it active.
func (c *Client) ReadThread(ctx context.Context, request ReadThreadRequest) (Thread, error) {
	if strings.TrimSpace(request.ThreadID) == "" {
		return Thread{}, fmt.Errorf("codexapp: thread id is required")
	}
	var response struct {
		Thread threadWire `json:"thread"`
	}
	if err := c.call(ctx, "thread/read", threadReadParams(request), &response); err != nil {
		return Thread{}, fmt.Errorf("codexapp: read thread: %w", err)
	}
	return mapThread(response.Thread, ""), nil
}

// InterruptTurn requests interruption of an active turn.
func (c *Client) InterruptTurn(ctx context.Context, request InterruptTurnRequest) error {
	if strings.TrimSpace(request.ThreadID) == "" || strings.TrimSpace(request.TurnID) == "" {
		return fmt.Errorf("codexapp: thread id and turn id are required")
	}
	if err := c.call(ctx, "turn/interrupt", turnInterruptParams(request), nil); err != nil {
		return fmt.Errorf("codexapp: interrupt turn: %w", err)
	}
	return nil
}

// StartTurn starts a turn and returns its independently consumable handle.
func (c *Client) StartTurn(ctx context.Context, request StartTurnRequest) (*TurnHandle, error) {
	if strings.TrimSpace(request.ThreadID) == "" {
		return nil, fmt.Errorf("codexapp: thread id is required")
	}
	params, err := c.buildTurnStartParams(request)
	if err != nil {
		return nil, err
	}
	handle, err := c.beginTurnStart(request.ThreadID)
	if err != nil {
		return nil, err
	}

	var response turnStartResponse
	if err := c.call(ctx, "turn/start", params, &response); err != nil {
		c.abortTurnStart(request.ThreadID)
		return nil, fmt.Errorf("codexapp: start turn: %w", err)
	}
	c.finishTurnStart(request.ThreadID, handle, response.Turn)
	return handle, nil
}

func (c *Client) buildTurnStartParams(request StartTurnRequest) (turnStartParams, error) {
	inputs := make([]inputWire, 0, len(request.Input))
	for _, input := range request.Input {
		wire, err := mapInput(input)
		if err != nil {
			return turnStartParams{}, err
		}
		inputs = append(inputs, wire)
	}
	params := turnStartParams{ThreadID: request.ThreadID, Input: inputs}
	params.Model = optionalString(request.Model)
	params.Cwd = optionalString(request.Cwd)
	params.Effort = optionalString(request.Effort)
	params.Summary = optionalReasoningSummary(request.Summary)
	params.Personality = optionalPersonality(request.Personality)
	if len(request.OutputSchema) > 0 {
		params.OutputSchema = append(json.RawMessage(nil), request.OutputSchema...)
	}
	if err := c.bindTurnPermissions(request, &params); err != nil {
		return turnStartParams{}, err
	}
	return params, nil
}

func (c *Client) bindTurnPermissions(request StartTurnRequest, params *turnStartParams) error {
	if request.Permissions.empty() {
		return nil
	}
	profile, err := buildPermissionProfile(request.Permissions)
	if err != nil {
		return err
	}
	c.turnsMu.Lock()
	bound, exists := c.profiles[request.ThreadID]
	c.turnsMu.Unlock()
	if !exists || bound.ID != profile.ID {
		return fmt.Errorf("%w %s", ErrPermissionProfileMismatch, request.ThreadID)
	}
	params.Permissions = &bound.ID
	params.RuntimeWorkspaceRoots = append([]string(nil), bound.WriteRoots...)
	return nil
}

func (c *Client) beginTurnStart(threadID string) (*TurnHandle, error) {
	handle := &TurnHandle{
		client:   c,
		threadID: threadID,
		stream:   newEventStream(c.eventBuffer),
		done:     make(chan struct{}),
	}
	c.turnsMu.Lock()
	if _, exists := c.starting[threadID]; exists {
		c.turnsMu.Unlock()
		return nil, fmt.Errorf("%w for thread %s", ErrTurnStartInProgress, threadID)
	}
	c.starting[threadID] = handle
	c.turnsMu.Unlock()
	return handle, nil
}

func (c *Client) abortTurnStart(threadID string) {
	c.turnsMu.Lock()
	delete(c.starting, threadID)
	c.turnsMu.Unlock()
}

func (c *Client) finishTurnStart(threadID string, handle *TurnHandle, turn turnWire) {
	handle.bind(turn)
	c.turnsMu.Lock()
	delete(c.starting, threadID)
	handle.mu.Lock()
	terminal := handle.terminal
	turnID := handle.turn.ID
	handle.mu.Unlock()
	if !terminal {
		c.turns[turnID] = handle
	}
	c.turnsMu.Unlock()
}

type threadStartParams struct {
	Model                 *string         `json:"model,omitempty"`
	Cwd                   *string         `json:"cwd,omitempty"`
	Config                map[string]any  `json:"config,omitempty"`
	Permissions           *string         `json:"permissions,omitempty"`
	RuntimeWorkspaceRoots []string        `json:"runtimeWorkspaceRoots,omitempty"`
	ModelProvider         *string         `json:"modelProvider,omitempty"`
	ApprovalPolicy        *ApprovalPolicy `json:"approvalPolicy,omitempty"`
	BaseInstructions      *string         `json:"baseInstructions,omitempty"`
	DeveloperInstructions *string         `json:"developerInstructions,omitempty"`
	Personality           *Personality    `json:"personality,omitempty"`
}

type threadStartResponse struct {
	Model  string     `json:"model"`
	Thread threadWire `json:"thread"`
}

type threadWire struct {
	ID        string     `json:"id"`
	Preview   string     `json:"preview"`
	CreatedAt int64      `json:"createdAt"`
	UpdatedAt int64      `json:"updatedAt"`
	Turns     []turnWire `json:"turns"`
}

type threadResumeParams struct {
	ThreadID string  `json:"threadId"`
	Model    *string `json:"model,omitempty"`
	Cwd      *string `json:"cwd,omitempty"`
}

type threadReadParams struct {
	ThreadID     string `json:"threadId"`
	IncludeTurns bool   `json:"includeTurns"`
}

type turnInterruptParams struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

type inputWire struct {
	Type string  `json:"type"`
	Text *string `json:"text,omitempty"`
	URL  *string `json:"url,omitempty"`
	Path *string `json:"path,omitempty"`
	Name *string `json:"name,omitempty"`
}

func mapInput(input Input) (inputWire, error) {
	wire := inputWire{Type: input.kind}
	switch input.kind {
	case inputTypeText:
		wire.Text = &input.text
	case inputTypeImage, inputTypeAudio:
		wire.URL = &input.url
	case inputTypeLocalImage, inputTypeLocalAudio:
		wire.Path = &input.path
	case inputTypeSkill, inputTypeMention:
		wire.Name = &input.name
		wire.Path = &input.path
	default:
		return inputWire{}, fmt.Errorf("codexapp: unsupported input type %q", input.kind)
	}
	return wire, nil
}

type turnStartParams struct {
	ThreadID              string            `json:"threadId"`
	Input                 []inputWire       `json:"input"`
	Permissions           *string           `json:"permissions,omitempty"`
	RuntimeWorkspaceRoots []string          `json:"runtimeWorkspaceRoots,omitempty"`
	Model                 *string           `json:"model,omitempty"`
	Cwd                   *string           `json:"cwd,omitempty"`
	Effort                *string           `json:"effort,omitempty"`
	Summary               *ReasoningSummary `json:"summary,omitempty"`
	Personality           *Personality      `json:"personality,omitempty"`
	OutputSchema          json.RawMessage   `json:"outputSchema,omitempty"`
}

func applyThreadSettings(params *threadStartParams, settings ThreadSettings) {
	params.ModelProvider = optionalString(settings.ModelProvider)
	params.ApprovalPolicy = optionalApprovalPolicy(settings.ApprovalPolicy)
	params.BaseInstructions = optionalString(settings.BaseInstructions)
	params.DeveloperInstructions = optionalString(settings.DeveloperInstructions)
	params.Personality = optionalPersonality(settings.Personality)
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func optionalApprovalPolicy(value ApprovalPolicy) *ApprovalPolicy {
	if value == "" {
		return nil
	}
	return &value
}

func optionalPersonality(value Personality) *Personality {
	if value == "" {
		return nil
	}
	return &value
}

func optionalReasoningSummary(value ReasoningSummary) *ReasoningSummary {
	if value == "" {
		return nil
	}
	return &value
}

type turnStartResponse struct {
	Turn turnWire `json:"turn"`
}

type turnWire struct {
	ID     string     `json:"id"`
	Status TurnStatus `json:"status"`
}

func mapThread(raw threadWire, model string) Thread {
	turns := make([]Turn, 0, len(raw.Turns))
	for _, turn := range raw.Turns {
		turns = append(turns, Turn(turn))
	}
	return Thread{
		ID: raw.ID, Model: model, Preview: raw.Preview,
		CreatedAt: raw.CreatedAt, UpdatedAt: raw.UpdatedAt, Turns: turns,
	}
}

func (h *TurnHandle) bind(turn turnWire) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.turn.ID == "" {
		h.turn = Turn(turn)
	}
}

func (h *TurnHandle) complete(turn Turn) {
	h.mu.Lock()
	if h.terminal {
		h.mu.Unlock()
		return
	}
	h.turn = turn
	h.result = TurnResult{ThreadID: h.threadID, Turn: turn}
	h.terminal = true
	h.mu.Unlock()
	h.stream.finish(nil)
	h.doneOnce.Do(func() { close(h.done) })
}

func (h *TurnHandle) fail(err error) {
	h.mu.Lock()
	if h.terminal {
		h.mu.Unlock()
		return
	}
	h.err = err
	h.terminal = true
	h.mu.Unlock()
	h.stream.finish(err)
	h.doneOnce.Do(func() { close(h.done) })
}

func (c *Client) failTurns(cause error) {
	sessionErr := &SessionLostError{Generation: c.generation, Cause: cause}
	c.turnsMu.Lock()
	unique := make(map[*TurnHandle]struct{}, len(c.turns)+len(c.starting))
	for _, handle := range c.turns {
		unique[handle] = struct{}{}
	}
	for _, handle := range c.starting {
		unique[handle] = struct{}{}
	}
	c.turns = make(map[string]*TurnHandle)
	c.starting = make(map[string]*TurnHandle)
	c.turnsMu.Unlock()
	for handle := range unique {
		handle.fail(sessionErr)
	}
}

func (c *Client) dispatchNotification(method string, params json.RawMessage) {
	switch method {
	case string(EventAgentMessageDelta):
		c.dispatchAgentMessageDelta(params)
	case string(EventTurnCompleted):
		c.dispatchTurnCompleted(params)
	case string(EventCommandOutputDelta), string(EventPlanDelta), string(EventReasoningSummaryDelta), string(EventReasoningTextDelta):
		c.dispatchDelta(EventType(method), params)
	case string(EventTurnDiffUpdated):
		c.dispatchTurnDiffUpdated(params)
	case string(EventTurnPlanUpdated):
		c.dispatchTurnPlanUpdated(params)
	case string(EventItemStarted), string(EventItemCompleted):
		c.dispatchItem(EventType(method), params)
	case string(EventError):
		c.dispatchError(params)
	case string(EventWarning):
		c.dispatchWarning(params)
	case string(EventTokenUsageUpdated):
		c.dispatchTokenUsageUpdated(params)
	}
}

func (c *Client) dispatchAgentMessageDelta(params json.RawMessage) {
	var notification struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Delta    string `json:"delta"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	handle := c.findTurn(notification.ThreadID, notification.TurnID)
	if handle == nil {
		return
	}
	handle.bind(turnWire{ID: notification.TurnID, Status: TurnStatusInProgress})
	_ = handle.stream.push(Event{
		Type:      EventAgentMessageDelta,
		ThreadID:  notification.ThreadID,
		TurnID:    notification.TurnID,
		TextDelta: &TextDelta{Text: notification.Delta},
	})
}

func (c *Client) dispatchTurnCompleted(params json.RawMessage) {
	var notification struct {
		ThreadID string   `json:"threadId"`
		Turn     turnWire `json:"turn"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	handle := c.findTurn(notification.ThreadID, notification.Turn.ID)
	if handle == nil {
		return
	}
	turn := Turn{ID: notification.Turn.ID, Status: notification.Turn.Status}
	_ = handle.stream.push(Event{
		Type:     EventTurnCompleted,
		ThreadID: notification.ThreadID,
		TurnID:   turn.ID,
		Turn:     &turn,
	})
	handle.complete(turn)
	c.turnsMu.Lock()
	delete(c.turns, turn.ID)
	c.turnsMu.Unlock()
}

func (c *Client) dispatchDelta(eventType EventType, params json.RawMessage) {
	var notification struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Delta    string `json:"delta"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	handle := c.findTurn(notification.ThreadID, notification.TurnID)
	if handle == nil {
		return
	}
	event := Event{Type: eventType, ThreadID: notification.ThreadID, TurnID: notification.TurnID}
	switch eventType {
	case EventCommandOutputDelta:
		event.CommandOutput = &CommandOutputDelta{Text: notification.Delta}
	case EventPlanDelta:
		event.PlanDelta = &PlanDelta{Text: notification.Delta}
	default:
		event.TextDelta = &TextDelta{Text: notification.Delta}
	}
	_ = handle.stream.push(event)
}

func (c *Client) dispatchTurnDiffUpdated(params json.RawMessage) {
	var notification struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Diff     string `json:"diff"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	if handle := c.findTurn(notification.ThreadID, notification.TurnID); handle != nil {
		_ = handle.stream.push(Event{Type: EventTurnDiffUpdated, ThreadID: notification.ThreadID, TurnID: notification.TurnID, Diff: &DiffUpdate{Diff: notification.Diff}})
	}
}

func (c *Client) dispatchTurnPlanUpdated(params json.RawMessage) {
	var notification struct {
		ThreadID    string     `json:"threadId"`
		TurnID      string     `json:"turnId"`
		Explanation string     `json:"explanation"`
		Plan        []PlanStep `json:"plan"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	if handle := c.findTurn(notification.ThreadID, notification.TurnID); handle != nil {
		_ = handle.stream.push(Event{Type: EventTurnPlanUpdated, ThreadID: notification.ThreadID, TurnID: notification.TurnID, Plan: &PlanUpdate{Explanation: notification.Explanation, Steps: notification.Plan}})
	}
}

func (c *Client) dispatchItem(eventType EventType, params json.RawMessage) {
	var notification struct {
		ThreadID string          `json:"threadId"`
		TurnID   string          `json:"turnId"`
		Item     json.RawMessage `json:"item"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	var identity struct{ ID, Type string }
	_ = json.Unmarshal(notification.Item, &identity)
	if handle := c.findTurn(notification.ThreadID, notification.TurnID); handle != nil {
		_ = handle.stream.push(Event{Type: eventType, ThreadID: notification.ThreadID, TurnID: notification.TurnID, Item: &Item{ID: identity.ID, Type: identity.Type, Raw: append(json.RawMessage(nil), notification.Item...)}})
	}
}

func (c *Client) dispatchError(params json.RawMessage) {
	var notification struct {
		ThreadID  string `json:"threadId"`
		TurnID    string `json:"turnId"`
		WillRetry bool   `json:"willRetry"`
		Error     struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	if handle := c.findTurn(notification.ThreadID, notification.TurnID); handle != nil {
		_ = handle.stream.push(Event{Type: EventError, ThreadID: notification.ThreadID, TurnID: notification.TurnID, Failure: &FailureEvent{Message: notification.Error.Message, WillRetry: notification.WillRetry}})
	}
}

func (c *Client) dispatchWarning(params json.RawMessage) {
	var notification struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Message  string `json:"message"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	if handle := c.findTurn(notification.ThreadID, notification.TurnID); handle != nil {
		_ = handle.stream.push(Event{Type: EventWarning, ThreadID: notification.ThreadID, TurnID: notification.TurnID, Warning: &WarningEvent{Message: notification.Message}})
	}
}

func (c *Client) dispatchTokenUsageUpdated(params json.RawMessage) {
	var notification struct {
		ThreadID string          `json:"threadId"`
		TurnID   string          `json:"turnId"`
		Usage    json.RawMessage `json:"tokenUsage"`
	}
	if json.Unmarshal(params, &notification) != nil {
		return
	}
	if handle := c.findTurn(notification.ThreadID, notification.TurnID); handle != nil {
		_ = handle.stream.push(Event{Type: EventTokenUsageUpdated, ThreadID: notification.ThreadID, TurnID: notification.TurnID, TokenUsage: &TokenUsageEvent{Raw: append(json.RawMessage(nil), notification.Usage...)}})
	}
}

func (c *Client) findTurn(threadID string, turnID string) *TurnHandle {
	c.turnsMu.Lock()
	defer c.turnsMu.Unlock()
	if handle := c.turns[turnID]; handle != nil {
		return handle
	}
	if handle := c.starting[threadID]; handle != nil {
		return handle
	}
	for _, handle := range c.turns {
		if handle.threadID == threadID {
			return handle
		}
	}
	return nil
}
