// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    initializeParams, err := UnmarshalInitializeParams(bytes)
//    bytes, err = initializeParams.Marshal()
//
//    initializeResponse, err := UnmarshalInitializeResponse(bytes)
//    bytes, err = initializeResponse.Marshal()
//
//    modelListParams, err := UnmarshalModelListParams(bytes)
//    bytes, err = modelListParams.Marshal()
//
//    modelListResponse, err := UnmarshalModelListResponse(bytes)
//    bytes, err = modelListResponse.Marshal()
//
//    permissionProfileListParams, err := UnmarshalPermissionProfileListParams(bytes)
//    bytes, err = permissionProfileListParams.Marshal()
//
//    permissionProfileListResponse, err := UnmarshalPermissionProfileListResponse(bytes)
//    bytes, err = permissionProfileListResponse.Marshal()
//
//    threadStartParams, err := UnmarshalThreadStartParams(bytes)
//    bytes, err = threadStartParams.Marshal()
//
//    threadStartResponse, err := UnmarshalThreadStartResponse(bytes)
//    bytes, err = threadStartResponse.Marshal()
//
//    threadResumeParams, err := UnmarshalThreadResumeParams(bytes)
//    bytes, err = threadResumeParams.Marshal()
//
//    threadResumeResponse, err := UnmarshalThreadResumeResponse(bytes)
//    bytes, err = threadResumeResponse.Marshal()
//
//    threadReadParams, err := UnmarshalThreadReadParams(bytes)
//    bytes, err = threadReadParams.Marshal()
//
//    threadReadResponse, err := UnmarshalThreadReadResponse(bytes)
//    bytes, err = threadReadResponse.Marshal()
//
//    turnStartParams, err := UnmarshalTurnStartParams(bytes)
//    bytes, err = turnStartParams.Marshal()
//
//    turnStartResponse, err := UnmarshalTurnStartResponse(bytes)
//    bytes, err = turnStartResponse.Marshal()
//
//    turnInterruptParams, err := UnmarshalTurnInterruptParams(bytes)
//    bytes, err = turnInterruptParams.Marshal()
//
//    turnInterruptResponse, err := UnmarshalTurnInterruptResponse(bytes)
//    bytes, err = turnInterruptResponse.Marshal()
//
//    turnStartedNotification, err := UnmarshalTurnStartedNotification(bytes)
//    bytes, err = turnStartedNotification.Marshal()
//
//    turnCompletedNotification, err := UnmarshalTurnCompletedNotification(bytes)
//    bytes, err = turnCompletedNotification.Marshal()
//
//    itemStartedNotification, err := UnmarshalItemStartedNotification(bytes)
//    bytes, err = itemStartedNotification.Marshal()
//
//    itemCompletedNotification, err := UnmarshalItemCompletedNotification(bytes)
//    bytes, err = itemCompletedNotification.Marshal()
//
//    agentMessageDeltaNotification, err := UnmarshalAgentMessageDeltaNotification(bytes)
//    bytes, err = agentMessageDeltaNotification.Marshal()
//
//    planDeltaNotification, err := UnmarshalPlanDeltaNotification(bytes)
//    bytes, err = planDeltaNotification.Marshal()
//
//    reasoningSummaryTextDeltaNotification, err := UnmarshalReasoningSummaryTextDeltaNotification(bytes)
//    bytes, err = reasoningSummaryTextDeltaNotification.Marshal()
//
//    reasoningSummaryPartAddedNotification, err := UnmarshalReasoningSummaryPartAddedNotification(bytes)
//    bytes, err = reasoningSummaryPartAddedNotification.Marshal()
//
//    reasoningTextDeltaNotification, err := UnmarshalReasoningTextDeltaNotification(bytes)
//    bytes, err = reasoningTextDeltaNotification.Marshal()
//
//    commandExecutionOutputDeltaNotification, err := UnmarshalCommandExecutionOutputDeltaNotification(bytes)
//    bytes, err = commandExecutionOutputDeltaNotification.Marshal()
//
//    turnDiffUpdatedNotification, err := UnmarshalTurnDiffUpdatedNotification(bytes)
//    bytes, err = turnDiffUpdatedNotification.Marshal()
//
//    turnPlanUpdatedNotification, err := UnmarshalTurnPlanUpdatedNotification(bytes)
//    bytes, err = turnPlanUpdatedNotification.Marshal()
//
//    errorNotification, err := UnmarshalErrorNotification(bytes)
//    bytes, err = errorNotification.Marshal()
//
//    warningNotification, err := UnmarshalWarningNotification(bytes)
//    bytes, err = warningNotification.Marshal()
//
//    threadTokenUsageUpdatedNotification, err := UnmarshalThreadTokenUsageUpdatedNotification(bytes)
//    bytes, err = threadTokenUsageUpdatedNotification.Marshal()
//
//    serverRequestResolvedNotification, err := UnmarshalServerRequestResolvedNotification(bytes)
//    bytes, err = serverRequestResolvedNotification.Marshal()
//
//    commandExecutionRequestApprovalParams, err := UnmarshalCommandExecutionRequestApprovalParams(bytes)
//    bytes, err = commandExecutionRequestApprovalParams.Marshal()
//
//    commandExecutionRequestApprovalResponse, err := UnmarshalCommandExecutionRequestApprovalResponse(bytes)
//    bytes, err = commandExecutionRequestApprovalResponse.Marshal()
//
//    fileChangeRequestApprovalParams, err := UnmarshalFileChangeRequestApprovalParams(bytes)
//    bytes, err = fileChangeRequestApprovalParams.Marshal()
//
//    fileChangeRequestApprovalResponse, err := UnmarshalFileChangeRequestApprovalResponse(bytes)
//    bytes, err = fileChangeRequestApprovalResponse.Marshal()
//
//    permissionsRequestApprovalParams, err := UnmarshalPermissionsRequestApprovalParams(bytes)
//    bytes, err = permissionsRequestApprovalParams.Marshal()
//
//    permissionsRequestApprovalResponse, err := UnmarshalPermissionsRequestApprovalResponse(bytes)
//    bytes, err = permissionsRequestApprovalResponse.Marshal()
//
//    toolRequestUserInputParams, err := UnmarshalToolRequestUserInputParams(bytes)
//    bytes, err = toolRequestUserInputParams.Marshal()
//
//    toolRequestUserInputResponse, err := UnmarshalToolRequestUserInputResponse(bytes)
//    bytes, err = toolRequestUserInputResponse.Marshal()
//
//    mCPServerElicitationRequestParams, err := UnmarshalMCPServerElicitationRequestParams(bytes)
//    bytes, err = mCPServerElicitationRequestParams.Marshal()
//
//    mCPServerElicitationRequestResponse, err := UnmarshalMCPServerElicitationRequestResponse(bytes)
//    bytes, err = mCPServerElicitationRequestResponse.Marshal()

package protocol

import "bytes"
import "errors"

import "encoding/json"

func UnmarshalInitializeParams(data []byte) (InitializeParams, error) {
	var r InitializeParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *InitializeParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalInitializeResponse(data []byte) (InitializeResponse, error) {
	var r InitializeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *InitializeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalModelListParams(data []byte) (ModelListParams, error) {
	var r ModelListParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ModelListParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalModelListResponse(data []byte) (ModelListResponse, error) {
	var r ModelListResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ModelListResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPermissionProfileListParams(data []byte) (PermissionProfileListParams, error) {
	var r PermissionProfileListParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PermissionProfileListParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPermissionProfileListResponse(data []byte) (PermissionProfileListResponse, error) {
	var r PermissionProfileListResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PermissionProfileListResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalThreadStartParams(data []byte) (ThreadStartParams, error) {
	var r ThreadStartParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ThreadStartParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalThreadStartResponse(data []byte) (ThreadStartResponse, error) {
	var r ThreadStartResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ThreadStartResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalThreadResumeParams(data []byte) (ThreadResumeParams, error) {
	var r ThreadResumeParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ThreadResumeParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalThreadResumeResponse(data []byte) (ThreadResumeResponse, error) {
	var r ThreadResumeResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ThreadResumeResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalThreadReadParams(data []byte) (ThreadReadParams, error) {
	var r ThreadReadParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ThreadReadParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalThreadReadResponse(data []byte) (ThreadReadResponse, error) {
	var r ThreadReadResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ThreadReadResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTurnStartParams(data []byte) (TurnStartParams, error) {
	var r TurnStartParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnStartParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTurnStartResponse(data []byte) (TurnStartResponse, error) {
	var r TurnStartResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnStartResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTurnInterruptParams(data []byte) (TurnInterruptParams, error) {
	var r TurnInterruptParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnInterruptParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type TurnInterruptResponse map[string]interface{}

func UnmarshalTurnInterruptResponse(data []byte) (TurnInterruptResponse, error) {
	var r TurnInterruptResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnInterruptResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTurnStartedNotification(data []byte) (TurnStartedNotification, error) {
	var r TurnStartedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnStartedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTurnCompletedNotification(data []byte) (TurnCompletedNotification, error) {
	var r TurnCompletedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnCompletedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalItemStartedNotification(data []byte) (ItemStartedNotification, error) {
	var r ItemStartedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ItemStartedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalItemCompletedNotification(data []byte) (ItemCompletedNotification, error) {
	var r ItemCompletedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ItemCompletedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalAgentMessageDeltaNotification(data []byte) (AgentMessageDeltaNotification, error) {
	var r AgentMessageDeltaNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *AgentMessageDeltaNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPlanDeltaNotification(data []byte) (PlanDeltaNotification, error) {
	var r PlanDeltaNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PlanDeltaNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalReasoningSummaryTextDeltaNotification(data []byte) (ReasoningSummaryTextDeltaNotification, error) {
	var r ReasoningSummaryTextDeltaNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ReasoningSummaryTextDeltaNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalReasoningSummaryPartAddedNotification(data []byte) (ReasoningSummaryPartAddedNotification, error) {
	var r ReasoningSummaryPartAddedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ReasoningSummaryPartAddedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalReasoningTextDeltaNotification(data []byte) (ReasoningTextDeltaNotification, error) {
	var r ReasoningTextDeltaNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ReasoningTextDeltaNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalCommandExecutionOutputDeltaNotification(data []byte) (CommandExecutionOutputDeltaNotification, error) {
	var r CommandExecutionOutputDeltaNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *CommandExecutionOutputDeltaNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTurnDiffUpdatedNotification(data []byte) (TurnDiffUpdatedNotification, error) {
	var r TurnDiffUpdatedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnDiffUpdatedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTurnPlanUpdatedNotification(data []byte) (TurnPlanUpdatedNotification, error) {
	var r TurnPlanUpdatedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *TurnPlanUpdatedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalErrorNotification(data []byte) (ErrorNotification, error) {
	var r ErrorNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ErrorNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalWarningNotification(data []byte) (WarningNotification, error) {
	var r WarningNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *WarningNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalThreadTokenUsageUpdatedNotification(data []byte) (ThreadTokenUsageUpdatedNotification, error) {
	var r ThreadTokenUsageUpdatedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ThreadTokenUsageUpdatedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalServerRequestResolvedNotification(data []byte) (ServerRequestResolvedNotification, error) {
	var r ServerRequestResolvedNotification
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ServerRequestResolvedNotification) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalCommandExecutionRequestApprovalParams(data []byte) (CommandExecutionRequestApprovalParams, error) {
	var r CommandExecutionRequestApprovalParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *CommandExecutionRequestApprovalParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalCommandExecutionRequestApprovalResponse(data []byte) (CommandExecutionRequestApprovalResponse, error) {
	var r CommandExecutionRequestApprovalResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *CommandExecutionRequestApprovalResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalFileChangeRequestApprovalParams(data []byte) (FileChangeRequestApprovalParams, error) {
	var r FileChangeRequestApprovalParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *FileChangeRequestApprovalParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalFileChangeRequestApprovalResponse(data []byte) (FileChangeRequestApprovalResponse, error) {
	var r FileChangeRequestApprovalResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *FileChangeRequestApprovalResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPermissionsRequestApprovalParams(data []byte) (PermissionsRequestApprovalParams, error) {
	var r PermissionsRequestApprovalParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PermissionsRequestApprovalParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalPermissionsRequestApprovalResponse(data []byte) (PermissionsRequestApprovalResponse, error) {
	var r PermissionsRequestApprovalResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PermissionsRequestApprovalResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalToolRequestUserInputParams(data []byte) (ToolRequestUserInputParams, error) {
	var r ToolRequestUserInputParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ToolRequestUserInputParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalToolRequestUserInputResponse(data []byte) (ToolRequestUserInputResponse, error) {
	var r ToolRequestUserInputResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *ToolRequestUserInputResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalMCPServerElicitationRequestParams(data []byte) (MCPServerElicitationRequestParams, error) {
	var r MCPServerElicitationRequestParams
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MCPServerElicitationRequestParams) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalMCPServerElicitationRequestResponse(data []byte) (MCPServerElicitationRequestResponse, error) {
	var r MCPServerElicitationRequestResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *MCPServerElicitationRequestResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

type InitializeParams struct {
	Capabilities *InitializeCapabilities `json:"capabilities"`
	ClientInfo   ClientInfo              `json:"clientInfo"`
}

// Client-declared capabilities negotiated during initialize.
type InitializeCapabilities struct {
	// Opt into receiving experimental API methods and fields.
	ExperimentalAPI *bool `json:"experimentalApi,omitempty"`
	// Allow downstream MCP servers to request OpenAI extended form elicitations.
	MCPServerOpenaiFormElicitation *bool `json:"mcpServerOpenaiFormElicitation,omitempty"`
	// Exact notification method names that should be suppressed for this connection (for
	// example `thread/started`).
	OptOutNotificationMethods []string `json:"optOutNotificationMethods"`
	// Opt into `attestation/generate` requests for upstream `x-oai-attestation`.
	RequestAttestation *bool `json:"requestAttestation,omitempty"`
}

type ClientInfo struct {
	Name    string  `json:"name"`
	Title   *string `json:"title"`
	Version string  `json:"version"`
}

type InitializeResponse struct {
	// Absolute path to the server's $CODEX_HOME directory.
	CodexHome string `json:"codexHome"`
	// Platform family for the running app-server target, for example `"unix"` or `"windows"`.
	PlatformFamily string `json:"platformFamily"`
	// Operating system for the running app-server target, for example `"macos"`, `"linux"`, or
	// `"windows"`.
	PlatformOS string `json:"platformOs"`
	UserAgent  string `json:"userAgent"`
}

type ModelListParams struct {
	// Opaque pagination cursor returned by a previous call.
	Cursor *string `json:"cursor"`
	// When true, include models that are hidden from the default picker list.
	IncludeHidden *bool `json:"includeHidden"`
	// Optional page size; defaults to a reasonable server-side value.
	Limit *int64 `json:"limit"`
}

type ModelListResponse struct {
	Data []Model `json:"data"`
	// Opaque cursor to pass to the next call to continue after the last item. If None, there
	// are no more items to return.
	NextCursor *string `json:"nextCursor"`
}

type Model struct {
	// Deprecated: use `serviceTiers` instead.
	AdditionalSpeedTiers   []string              `json:"additionalSpeedTiers,omitempty"`
	AvailabilityNux        *ModelAvailabilityNux `json:"availabilityNux"`
	DefaultReasoningEffort string                `json:"defaultReasoningEffort"`
	// Catalog default service tier id for this model, when one is configured.
	DefaultServiceTier        *string                 `json:"defaultServiceTier"`
	Description               string                  `json:"description"`
	DisplayName               string                  `json:"displayName"`
	Hidden                    bool                    `json:"hidden"`
	ID                        string                  `json:"id"`
	InputModalities           []InputModality         `json:"inputModalities,omitempty"`
	IsDefault                 bool                    `json:"isDefault"`
	Model                     string                  `json:"model"`
	ServiceTiers              []ModelServiceTier      `json:"serviceTiers,omitempty"`
	SupportedReasoningEfforts []ReasoningEffortOption `json:"supportedReasoningEfforts"`
	SupportsPersonality       *bool                   `json:"supportsPersonality,omitempty"`
	Upgrade                   *string                 `json:"upgrade"`
	UpgradeInfo               *ModelUpgradeInfo       `json:"upgradeInfo"`
}

type ModelAvailabilityNux struct {
	Message string `json:"message"`
}

type ModelServiceTier struct {
	Description string `json:"description"`
	ID          string `json:"id"`
	Name        string `json:"name"`
}

type ReasoningEffortOption struct {
	Description     string `json:"description"`
	ReasoningEffort string `json:"reasoningEffort"`
}

type ModelUpgradeInfo struct {
	MigrationMarkdown *string `json:"migrationMarkdown"`
	Model             string  `json:"model"`
	ModelLink         *string `json:"modelLink"`
	UpgradeCopy       *string `json:"upgradeCopy"`
}

type PermissionProfileListParams struct {
	// Opaque pagination cursor returned by a previous call.
	Cursor *string `json:"cursor"`
	// Optional working directory to resolve project config layers.
	Cwd *string `json:"cwd"`
	// Optional page size; defaults to the full result set.
	Limit *int64 `json:"limit"`
}

type PermissionProfileListResponse struct {
	Data []PermissionProfileSummary `json:"data"`
	// Opaque cursor to pass to the next call to continue after the last item. If None, there
	// are no more items to return.
	NextCursor *string `json:"nextCursor"`
}

type PermissionProfileSummary struct {
	// Whether the effective requirements allow selecting this profile.
	Allowed bool `json:"allowed"`
	// Optional user-facing description for display in clients.
	Description *string `json:"description"`
	// Available permission profile identifier.
	ID string `json:"id"`
}

type ThreadStartParams struct {
	// Allow a provider with an authoritative static model catalog to replace an unavailable
	// requested model with its default.
	AllowProviderModelFallback *bool                            `json:"allowProviderModelFallback,omitempty"`
	ApprovalPolicy             *ThreadStartParamsApprovalPolicy `json:"approvalPolicy"`
	// Override where approval requests are routed for review on this thread and subsequent
	// turns.
	ApprovalsReviewer     *ApprovalsReviewer     `json:"approvalsReviewer"`
	BaseInstructions      *string                `json:"baseInstructions"`
	Config                map[string]interface{} `json:"config"`
	Cwd                   *string                `json:"cwd"`
	DeveloperInstructions *string                `json:"developerInstructions"`
	DynamicTools          []DynamicToolSpec      `json:"dynamicTools"`
	// Optional sticky environments for this thread.
	//
	// Omitted selects the default environment when environment access is enabled. Empty
	// disables environment access for turns that do not provide a turn override. Non-empty
	// selects the first environment as the current turn environment.
	Environments []ThreadStartParamsEnvironment `json:"environments"`
	Ephemeral    *bool                          `json:"ephemeral"`
	// If true, opt into emitting raw Responses API items on the event stream. This is for
	// internal use only (e.g. Codex Cloud).
	ExperimentalRawEvents *bool `json:"experimentalRawEvents,omitempty"`
	// Persisted thread history contract to use for this new thread.
	HistoryMode *ThreadHistoryMode `json:"historyMode"`
	// Test-only experimental field used to validate experimental gating and schema filtering
	// behavior in a stable way.
	MockExperimentalField *string `json:"mockExperimentalField"`
	Model                 *string `json:"model"`
	ModelProvider         *string `json:"modelProvider"`
	// @deprecated Ignored. Use Ultra reasoning effort for proactive multi-agent behavior.
	MultiAgentMode *ThreadStartParamsMultiAgentMode `json:"multiAgentMode"`
	// Named profile id for this thread. Cannot be combined with `sandbox`.
	Permissions *string      `json:"permissions"`
	Personality *Personality `json:"personality"`
	// Replace the thread's runtime workspace roots. Paths must be absolute.
	RuntimeWorkspaceRoots []string     `json:"runtimeWorkspaceRoots"`
	Sandbox               *SandboxMode `json:"sandbox"`
	// Capability roots selected for this thread by the hosting platform.
	SelectedCapabilityRoots []SelectedCapabilityRoot `json:"selectedCapabilityRoots"`
	ServiceName             *string                  `json:"serviceName"`
	ServiceTier             *string                  `json:"serviceTier"`
	SessionStartSource      *ThreadStartSource       `json:"sessionStartSource"`
	// Optional client-supplied analytics source classification for this thread.
	ThreadSource *string `json:"threadSource"`
}

type PurpleGranularAskForApproval struct {
	Granular PurpleGranular `json:"granular"`
}

type PurpleGranular struct {
	MCPElicitations    bool  `json:"mcp_elicitations"`
	RequestPermissions *bool `json:"request_permissions,omitempty"`
	Rules              bool  `json:"rules"`
	SandboxApproval    bool  `json:"sandbox_approval"`
	SkillApproval      *bool `json:"skill_approval,omitempty"`
}

type DynamicToolSpec struct {
	DeferLoading *bool                      `json:"deferLoading,omitempty"`
	Description  string                     `json:"description"`
	InputSchema  interface{}                `json:"inputSchema"`
	Name         string                     `json:"name"`
	Type         DynamicToolSpecType        `json:"type"`
	Tools        []DynamicToolNamespaceTool `json:"tools,omitempty"`
}

type DynamicToolNamespaceTool struct {
	DeferLoading *bool                                `json:"deferLoading,omitempty"`
	Description  string                               `json:"description"`
	InputSchema  interface{}                          `json:"inputSchema"`
	Name         string                               `json:"name"`
	Type         FunctionDynamicToolNamespaceToolType `json:"type"`
}

type ThreadStartParamsEnvironment struct {
	Cwd           string `json:"cwd"`
	EnvironmentID string `json:"environmentId"`
	// Environment-native runtime workspace roots. Omitted defaults to `cwd`.
	RuntimeWorkspaceRoots []string `json:"runtimeWorkspaceRoots"`
}

type PurpleCustomMultiAgentMode struct {
	Custom string `json:"custom"`
}

// A user-selected root that can expose one or more runtime capabilities.
type SelectedCapabilityRoot struct {
	// Stable identifier supplied by the capability selection platform.
	ID string `json:"id"`
	// Where the selected root can be resolved.
	Location CapabilityRootLocation `json:"location"`
}

// Where the selected root can be resolved.
//
// Location used to resolve a selected capability root.
//
// A path owned by an execution environment.
type CapabilityRootLocation struct {
	EnvironmentID string `json:"environmentId"`
	// Absolute path for the root in the selected environment.
	Path string                                `json:"path"`
	Type EnvironmentCapabilityRootLocationType `json:"type"`
}

type ThreadStartResponse struct {
	// Named or implicit built-in profile that produced the active permissions, when known.
	ActivePermissionProfile *ThreadStartResponseActivePermissionProfile `json:"activePermissionProfile"`
	ApprovalPolicy          *ThreadStartResponseAskForApproval          `json:"approvalPolicy"`
	// Reviewer currently used for approval requests on this thread.
	ApprovalsReviewer ApprovalsReviewer `json:"approvalsReviewer"`
	Cwd               string            `json:"cwd"`
	// Environment-native paths to instruction source files currently loaded for this thread.
	InstructionSources []string `json:"instructionSources,omitempty"`
	Model              string   `json:"model"`
	ModelProvider      string   `json:"modelProvider"`
	// @deprecated Always `explicitRequestOnly`. Use `reasoningEffort` for Ultra behavior.
	MultiAgentMode  *ThreadStartResponseMultiAgentMode `json:"multiAgentMode"`
	ReasoningEffort *string                            `json:"reasoningEffort"`
	// Thread-scoped runtime workspace roots used to materialize `:workspace_roots`.
	RuntimeWorkspaceRoots []string `json:"runtimeWorkspaceRoots,omitempty"`
	// Legacy sandbox policy retained for compatibility. Experimental clients should prefer
	// `activePermissionProfile` for profile provenance.
	Sandbox     ThreadStartResponseSandboxPolicy `json:"sandbox"`
	ServiceTier *string                          `json:"serviceTier"`
	Thread      ThreadStartResponseThread        `json:"thread"`
}

type ThreadStartResponseActivePermissionProfile struct {
	// Parent profile identifier from the selected permissions profile's `extends` setting, when
	// present.
	Extends *string `json:"extends"`
	// Identifier from `default_permissions` or the implicit built-in default, such as
	// `:workspace` or a user-defined `[permissions.<id>]` profile.
	ID string `json:"id"`
}

type FluffyGranularAskForApproval struct {
	Granular FluffyGranular `json:"granular"`
}

type FluffyGranular struct {
	MCPElicitations    bool  `json:"mcp_elicitations"`
	RequestPermissions *bool `json:"request_permissions,omitempty"`
	Rules              bool  `json:"rules"`
	SandboxApproval    bool  `json:"sandbox_approval"`
	SkillApproval      *bool `json:"skill_approval,omitempty"`
}

type FluffyCustomMultiAgentMode struct {
	Custom string `json:"custom"`
}

// Legacy sandbox policy retained for compatibility. Experimental clients should prefer
// `activePermissionProfile` for profile provenance.
type ThreadStartResponseSandboxPolicy struct {
	Type                SandboxPolicyType   `json:"type"`
	NetworkAccess       *NetworkAccessUnion `json:"networkAccess"`
	ExcludeSlashTmp     *bool               `json:"excludeSlashTmp,omitempty"`
	ExcludeTmpdirEnvVar *bool               `json:"excludeTmpdirEnvVar,omitempty"`
	WritableRoots       []string            `json:"writableRoots,omitempty"`
}

type ThreadStartResponseThread struct {
	// Optional random unique nickname assigned to an AgentControl-spawned sub-agent.
	AgentNickname *string `json:"agentNickname"`
	// Optional role (agent_role) assigned to an AgentControl-spawned sub-agent.
	AgentRole *string `json:"agentRole"`
	// Whether the app server accepts direct turn input for this loaded thread. `None` means the
	// capability is unavailable, such as for an unloaded stored thread.
	CanAcceptDirectInput *bool `json:"canAcceptDirectInput"`
	// Version of the CLI that created the thread.
	CLIVersion string `json:"cliVersion"`
	// Unix timestamp (in seconds) when the thread was created.
	CreatedAt int64 `json:"createdAt"`
	// Working directory captured for the thread.
	Cwd string `json:"cwd"`
	// Whether the thread is ephemeral and should not be materialized on disk.
	Ephemeral bool `json:"ephemeral"`
	// Optional implementation-specific thread data.
	Extra map[string]interface{} `json:"extra"`
	// Source thread id when this thread was created by forking another thread.
	ForkedFromID *string `json:"forkedFromId"`
	// Optional Git metadata captured when the thread was created.
	GitInfo *PurpleGitInfo `json:"gitInfo"`
	// Persisted thread history contract selected when this thread was created.
	HistoryMode *ThreadHistoryMode `json:"historyMode,omitempty"`
	// Identifier for this thread. Codex-generated thread IDs are UUIDv7.
	ID string `json:"id"`
	// Whether the thread has been pinned by the user.
	IsPinned *bool `json:"isPinned,omitempty"`
	// Model provider used for this thread (for example, 'openai').
	ModelProvider string `json:"modelProvider"`
	// Optional user-facing thread title.
	Name *string `json:"name"`
	// The ID of the parent thread. This will only be set if this thread is a subagent.
	ParentThreadID *string `json:"parentThreadId"`
	// [UNSTABLE] Path to the thread on disk.
	Path *string `json:"path"`
	// Usually the first user message in the thread, if available.
	Preview string `json:"preview"`
	// Unix timestamp (in seconds) used for thread recency ordering.
	RecencyAt *int64 `json:"recencyAt"`
	// Session id shared by threads that belong to the same session tree.
	SessionID string `json:"sessionId"`
	// Origin of the thread (CLI, VSCode, codex exec, codex app-server, etc.).
	Source *StickySessionSource `json:"source"`
	// Current runtime status for the thread.
	Status PurpleThreadStatus `json:"status"`
	// Optional analytics source classification for this thread.
	ThreadSource *string `json:"threadSource"`
	// Only populated on `thread/resume`, `thread/rollback`, `thread/fork`, and `thread/read`
	// (when `includeTurns` is true) responses. For all other responses and notifications
	// returning a Thread, the turns field will be an empty list.
	Turns []PurpleTurn `json:"turns"`
	// Unix timestamp (in seconds) when the thread was last updated.
	UpdatedAt int64 `json:"updatedAt"`
}

type PurpleGitInfo struct {
	Branch    *string `json:"branch"`
	OriginURL *string `json:"originUrl"`
	SHA       *string `json:"sha"`
}

type PurpleSessionSource struct {
	Custom   *string               `json:"custom,omitempty"`
	SubAgent *StickySubAgentSource `json:"subAgent"`
}

type PurpleSubAgentSource struct {
	ThreadSpawn *PurpleThreadSpawn `json:"thread_spawn,omitempty"`
	Other       *string            `json:"other,omitempty"`
}

type PurpleThreadSpawn struct {
	AgentNickname  *string `json:"agent_nickname"`
	AgentPath      *string `json:"agent_path"`
	AgentRole      *string `json:"agent_role"`
	Depth          int64   `json:"depth"`
	ParentThreadID string  `json:"parent_thread_id"`
}

// Current runtime status for the thread.
type PurpleThreadStatus struct {
	Type        ThreadStatusType   `json:"type"`
	ActiveFlags []ThreadActiveFlag `json:"activeFlags,omitempty"`
}

type PurpleTurn struct {
	// Unix timestamp (in seconds) when the turn completed.
	CompletedAt *int64 `json:"completedAt"`
	// Duration between turn start and completion in milliseconds, if known.
	DurationMS *int64 `json:"durationMs"`
	// Only populated when the Turn's status is failed.
	Error *PurpleTurnError `json:"error"`
	// Identifier for this turn. Codex-generated turn IDs are UUIDv7.
	ID string `json:"id"`
	// Thread items currently included in this turn payload.
	Items []PurpleThreadItem `json:"items"`
	// Describes how much of `items` has been loaded for this turn.
	ItemsView *TurnItemsView `json:"itemsView,omitempty"`
	// Unix timestamp (in seconds) when the turn started.
	StartedAt *int64     `json:"startedAt"`
	Status    TurnStatus `json:"status"`
}

type PurpleTurnError struct {
	AdditionalDetails *string                  `json:"additionalDetails"`
	CodexErrorInfo    *AmbitiousCodexErrorInfo `json:"codexErrorInfo"`
	Message           string                   `json:"message"`
}

// Failed to connect to the response SSE stream.
//
// The response SSE stream disconnected in the middle of a turn before completion.
//
// Reached the retry limit for responses.
//
// Returned when `turn/start` or `turn/steer` is submitted while the current active turn
// cannot accept same-turn steering, for example `/review` or manual `/compact`.
type PurpleCodexErrorInfo struct {
	HTTPConnectionFailed           *PurpleHTTPConnectionFailed           `json:"httpConnectionFailed,omitempty"`
	ResponseStreamConnectionFailed *PurpleResponseStreamConnectionFailed `json:"responseStreamConnectionFailed,omitempty"`
	ResponseStreamDisconnected     *PurpleResponseStreamDisconnected     `json:"responseStreamDisconnected,omitempty"`
	ResponseTooManyFailedAttempts  *PurpleResponseTooManyFailedAttempts  `json:"responseTooManyFailedAttempts,omitempty"`
	ActiveTurnNotSteerable         *PurpleActiveTurnNotSteerable         `json:"activeTurnNotSteerable,omitempty"`
}

type PurpleActiveTurnNotSteerable struct {
	TurnKind NonSteerableTurnKind `json:"turnKind"`
}

type PurpleHTTPConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type PurpleResponseStreamConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type PurpleResponseStreamDisconnected struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type PurpleResponseTooManyFailedAttempts struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type PurpleThreadItem struct {
	ClientID *string            `json:"clientId"`
	Content  []CunningUserInput `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                     `json:"id"`
	Type           ThreadItemType             `json:"type"`
	Fragments      []PurpleHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *PurpleMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase              `json:"phase"`
	Text           *string                    `json:"text,omitempty"`
	Summary        []string                   `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []PurpleCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                      `json:"status,omitempty"`
	Changes    []PurpleFileUpdateChange     `json:"changes,omitempty"`
	AppContext *PurpleMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                  `json:"arguments"`
	Error      *PurpleMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string       `json:"mcpAppResourceUri"`
	Result            *PurpleResult `json:"result"`
	Server            *string       `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                  `json:"tool,omitempty"`
	ContentItems []PurpleDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                  `json:"namespace"`
	Success      *bool                                    `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]PurpleCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                `json:"senderThreadId,omitempty"`
	AgentPath      *string                `json:"agentPath,omitempty"`
	AgentThreadID  *string                `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind  `json:"kind,omitempty"`
	Action         *PurpleWebSearchAction `json:"action"`
	Query          *string                `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type PurpleWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type PurpleCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type PurpleMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type PurpleFileUpdateChange struct {
	Diff string                `json:"diff"`
	Kind PurplePatchChangeKind `json:"kind"`
	Path string                `json:"path"`
}

type PurplePatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type PurpleCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type PurpleUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []PurpleTextElement `json:"text_elements,omitempty"`
	Type         UserInputType       `json:"type"`
	Detail       *ImageDetail        `json:"detail"`
	URL          *string             `json:"url,omitempty"`
	Path         *string             `json:"path,omitempty"`
	Name         *string             `json:"name,omitempty"`
}

type PurpleTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange PurpleByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type PurpleByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type PurpleDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type PurpleMCPToolCallError struct {
	Message string `json:"message"`
}

type PurpleHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type PurpleMemoryCitation struct {
	Entries   []PurpleMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                    `json:"threadIds"`
}

type PurpleMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type PurpleMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

// There are three ways to resume a thread: 1. By thread_id: load the thread from disk by
// thread_id and resume it. 2. By history: instantiate the thread from memory and resume it.
// 3. By path: load the thread from disk by path and resume it.
//
// For non-running threads, the precedence is: history > non-empty path > thread_id. If
// using history or a non-empty path for a non-running thread, the thread_id param will be
// ignored.
//
// If thread_id identifies a running thread, app-server rejoins that thread and treats a
// non-empty path as a consistency check against the active rollout path. Empty string path
// values are treated as absent.
//
// Prefer using thread_id whenever possible.
type ThreadResumeParams struct {
	ApprovalPolicy *ThreadResumeParamsApprovalPolicy `json:"approvalPolicy"`
	// Override where approval requests are routed for review on this thread and subsequent
	// turns.
	ApprovalsReviewer     *ApprovalsReviewer     `json:"approvalsReviewer"`
	BaseInstructions      *string                `json:"baseInstructions"`
	Config                map[string]interface{} `json:"config"`
	Cwd                   *string                `json:"cwd"`
	DeveloperInstructions *string                `json:"developerInstructions"`
	// When true, return only thread metadata and live-resume state without populating
	// `thread.turns`. This is useful when the client plans to call `thread/turns/list`
	// immediately after resuming.
	ExcludeTurns *bool `json:"excludeTurns,omitempty"`
	// [UNSTABLE] FOR CODEX CLOUD - DO NOT USE. If specified, the thread will be resumed with
	// the provided history instead of loaded from disk.
	History []ResponseItem `json:"history"`
	// When present, include a `thread/turns/list` page in the resume response so clients can
	// bootstrap recent turns without a second request.
	InitialTurnsPage *ThreadResumeInitialTurnsPageParams `json:"initialTurnsPage"`
	// Configuration overrides for the resumed thread, if any.
	Model         *string `json:"model"`
	ModelProvider *string `json:"modelProvider"`
	// [UNSTABLE] Specify the rollout path to resume from. If specified for a non-running
	// thread, the thread_id param will be ignored. If thread_id identifies a running thread,
	// the path must match the active rollout path.
	Path *string `json:"path"`
	// Named profile id for the resumed thread. Cannot be combined with `sandbox`.
	Permissions *string      `json:"permissions"`
	Personality *Personality `json:"personality"`
	// Replace the thread's runtime workspace roots. Paths must be absolute.
	RuntimeWorkspaceRoots []string     `json:"runtimeWorkspaceRoots"`
	Sandbox               *SandboxMode `json:"sandbox"`
	ServiceTier           *string      `json:"serviceTier"`
	ThreadID              string       `json:"threadId"`
}

type TentacledGranularAskForApproval struct {
	Granular TentacledGranular `json:"granular"`
}

type TentacledGranular struct {
	MCPElicitations    bool  `json:"mcp_elicitations"`
	RequestPermissions *bool `json:"request_permissions,omitempty"`
	Rules              bool  `json:"rules"`
	SandboxApproval    bool  `json:"sandbox_approval"`
	SkillApproval      *bool `json:"skill_approval,omitempty"`
}

type ResponseItem struct {
	Content []ContentItem `json:"content"`
	// Legacy id field retained for compatibility with older payloads.
	ID                                     *string                                 `json:"id"`
	InternalChatMessageMetadataPassthrough *InternalChatMessageMetadataPassthrough `json:"internal_chat_message_metadata_passthrough"`
	Phase                                  *MessagePhase                           `json:"phase"`
	Role                                   *string                                 `json:"role,omitempty"`
	Type                                   ResponseItemType                        `json:"type"`
	Author                                 *string                                 `json:"author,omitempty"`
	Recipient                              *string                                 `json:"recipient,omitempty"`
	EncryptedContent                       *string                                 `json:"encrypted_content"`
	Summary                                []ReasoningItemReasoningSummary         `json:"summary,omitempty"`
	Action                                 *Action                                 `json:"action"`
	// Set when using the Responses API.
	CallID        *string                 `json:"call_id"`
	Status        *string                 `json:"status"`
	Arguments     interface{}             `json:"arguments"`
	Name          *string                 `json:"name"`
	Namespace     *string                 `json:"namespace"`
	Execution     *string                 `json:"execution,omitempty"`
	Output        *FunctionCallOutputBody `json:"output"`
	Input         *string                 `json:"input,omitempty"`
	Tools         []interface{}           `json:"tools,omitempty"`
	Result        *string                 `json:"result,omitempty"`
	RevisedPrompt *string                 `json:"revised_prompt"`
}

type Action struct {
	Command          []string          `json:"command,omitempty"`
	Env              map[string]string `json:"env"`
	TimeoutMS        *int64            `json:"timeout_ms"`
	Type             ActionType        `json:"type"`
	User             *string           `json:"user"`
	WorkingDirectory *string           `json:"working_directory"`
	Queries          []string          `json:"queries"`
	Query            *string           `json:"query"`
	URL              *string           `json:"url"`
	Pattern          *string           `json:"pattern"`
}

type ContentItem struct {
	Text             *string      `json:"text,omitempty"`
	Type             Type         `json:"type"`
	Detail           *ImageDetail `json:"detail"`
	ImageURL         *string      `json:"image_url,omitempty"`
	AudioURL         *string      `json:"audio_url,omitempty"`
	EncryptedContent *string      `json:"encrypted_content,omitempty"`
}

// Internal Responses API passthrough metadata copied into underlying chat messages.
//
// Responses API strongly types this payload. Do not modify it without first getting API
// approval and making the corresponding Responses API change.
type InternalChatMessageMetadataPassthrough struct {
	TurnID *string `json:"turn_id"`
}

// Responses API compatible content items that can be returned by a tool call. This is a
// subset of ContentItem with the types we support as function call outputs.
type FunctionCallOutputContentItem struct {
	Text             *string                           `json:"text,omitempty"`
	Type             FunctionCallOutputContentItemType `json:"type"`
	Detail           *ImageDetail                      `json:"detail"`
	ImageURL         *string                           `json:"image_url,omitempty"`
	AudioURL         *string                           `json:"audio_url,omitempty"`
	EncryptedContent *string                           `json:"encrypted_content,omitempty"`
}

type ReasoningItemReasoningSummary struct {
	Text string                                       `json:"text"`
	Type SummaryTextReasoningItemReasoningSummaryType `json:"type"`
}

type ThreadResumeInitialTurnsPageParams struct {
	// How much item detail to include for each returned turn; defaults to summary.
	ItemsView *TurnItemsView `json:"itemsView"`
	// Optional turn page size.
	Limit *int64 `json:"limit"`
	// Optional turn pagination direction; defaults to descending.
	SortDirection *SortDirection `json:"sortDirection"`
}

type ThreadResumeResponse struct {
	// Named or implicit built-in profile that produced the active permissions, when known.
	ActivePermissionProfile *ThreadResumeResponseActivePermissionProfile `json:"activePermissionProfile"`
	ApprovalPolicy          *ThreadResumeResponseAskForApproval          `json:"approvalPolicy"`
	// Reviewer currently used for approval requests on this thread.
	ApprovalsReviewer ApprovalsReviewer `json:"approvalsReviewer"`
	Cwd               string            `json:"cwd"`
	// `thread/turns/list` page returned when requested by `initialTurnsPage`.
	InitialTurnsPage *TurnsPage `json:"initialTurnsPage"`
	// Environment-native paths to instruction source files currently loaded for this thread.
	InstructionSources []string `json:"instructionSources,omitempty"`
	// Opaque head cursor for hydrating paginated items backwards.
	//
	// Pass this as `cursor` to `thread/items/list` with `sortDirection: "desc"`. The first page
	// includes the cursor's head item.
	ItemsBackwardsCursor *string `json:"itemsBackwardsCursor"`
	Model                string  `json:"model"`
	ModelProvider        string  `json:"modelProvider"`
	// @deprecated Always `explicitRequestOnly`. Use `reasoningEffort` for Ultra behavior.
	MultiAgentMode  *ThreadResumeResponseMultiAgentMode `json:"multiAgentMode"`
	ReasoningEffort *string                             `json:"reasoningEffort"`
	// Thread-scoped runtime workspace roots used to materialize `:workspace_roots`.
	RuntimeWorkspaceRoots []string `json:"runtimeWorkspaceRoots,omitempty"`
	// Legacy sandbox policy retained for compatibility. Experimental clients should prefer
	// `activePermissionProfile` for profile provenance.
	Sandbox     ThreadResumeResponseSandboxPolicy `json:"sandbox"`
	ServiceTier *string                           `json:"serviceTier"`
	Thread      ThreadResumeResponseThread        `json:"thread"`
	// Opaque head cursor for hydrating paginated turns backwards.
	//
	// Pass this as `cursor` to `thread/turns/list` with `sortDirection: "desc"`. The first page
	// includes the cursor's head turn.
	TurnsBackwardsCursor *string `json:"turnsBackwardsCursor"`
}

type ThreadResumeResponseActivePermissionProfile struct {
	// Parent profile identifier from the selected permissions profile's `extends` setting, when
	// present.
	Extends *string `json:"extends"`
	// Identifier from `default_permissions` or the implicit built-in default, such as
	// `:workspace` or a user-defined `[permissions.<id>]` profile.
	ID string `json:"id"`
}

type StickyGranularAskForApproval struct {
	Granular StickyGranular `json:"granular"`
}

type StickyGranular struct {
	MCPElicitations    bool  `json:"mcp_elicitations"`
	RequestPermissions *bool `json:"request_permissions,omitempty"`
	Rules              bool  `json:"rules"`
	SandboxApproval    bool  `json:"sandbox_approval"`
	SkillApproval      *bool `json:"skill_approval,omitempty"`
}

type TurnsPage struct {
	BackwardsCursor *string        `json:"backwardsCursor"`
	Data            []DatumElement `json:"data"`
	NextCursor      *string        `json:"nextCursor"`
}

type DatumElement struct {
	// Unix timestamp (in seconds) when the turn completed.
	CompletedAt *int64 `json:"completedAt"`
	// Duration between turn start and completion in milliseconds, if known.
	DurationMS *int64 `json:"durationMs"`
	// Only populated when the Turn's status is failed.
	Error *DatumTurnError `json:"error"`
	// Identifier for this turn. Codex-generated turn IDs are UUIDv7.
	ID string `json:"id"`
	// Thread items currently included in this turn payload.
	Items []DatumThreadItem `json:"items"`
	// Describes how much of `items` has been loaded for this turn.
	ItemsView *TurnItemsView `json:"itemsView,omitempty"`
	// Unix timestamp (in seconds) when the turn started.
	StartedAt *int64     `json:"startedAt"`
	Status    TurnStatus `json:"status"`
}

type DatumTurnError struct {
	AdditionalDetails *string                `json:"additionalDetails"`
	CodexErrorInfo    *CunningCodexErrorInfo `json:"codexErrorInfo"`
	Message           string                 `json:"message"`
}

// Failed to connect to the response SSE stream.
//
// The response SSE stream disconnected in the middle of a turn before completion.
//
// Reached the retry limit for responses.
//
// Returned when `turn/start` or `turn/steer` is submitted while the current active turn
// cannot accept same-turn steering, for example `/review` or manual `/compact`.
type FluffyCodexErrorInfo struct {
	HTTPConnectionFailed           *FluffyHTTPConnectionFailed           `json:"httpConnectionFailed,omitempty"`
	ResponseStreamConnectionFailed *FluffyResponseStreamConnectionFailed `json:"responseStreamConnectionFailed,omitempty"`
	ResponseStreamDisconnected     *FluffyResponseStreamDisconnected     `json:"responseStreamDisconnected,omitempty"`
	ResponseTooManyFailedAttempts  *FluffyResponseTooManyFailedAttempts  `json:"responseTooManyFailedAttempts,omitempty"`
	ActiveTurnNotSteerable         *FluffyActiveTurnNotSteerable         `json:"activeTurnNotSteerable,omitempty"`
}

type FluffyActiveTurnNotSteerable struct {
	TurnKind NonSteerableTurnKind `json:"turnKind"`
}

type FluffyHTTPConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type FluffyResponseStreamConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type FluffyResponseStreamDisconnected struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type FluffyResponseTooManyFailedAttempts struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type DatumThreadItem struct {
	ClientID *string            `json:"clientId"`
	Content  []MagentaUserInput `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                     `json:"id"`
	Type           ThreadItemType             `json:"type"`
	Fragments      []FluffyHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *FluffyMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase              `json:"phase"`
	Text           *string                    `json:"text,omitempty"`
	Summary        []string                   `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []FluffyCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                      `json:"status,omitempty"`
	Changes    []FluffyFileUpdateChange     `json:"changes,omitempty"`
	AppContext *FluffyMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                  `json:"arguments"`
	Error      *FluffyMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string       `json:"mcpAppResourceUri"`
	Result            *FluffyResult `json:"result"`
	Server            *string       `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                  `json:"tool,omitempty"`
	ContentItems []FluffyDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                  `json:"namespace"`
	Success      *bool                                    `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]FluffyCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                `json:"senderThreadId,omitempty"`
	AgentPath      *string                `json:"agentPath,omitempty"`
	AgentThreadID  *string                `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind  `json:"kind,omitempty"`
	Action         *FluffyWebSearchAction `json:"action"`
	Query          *string                `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type FluffyWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type FluffyCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type FluffyMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type FluffyFileUpdateChange struct {
	Diff string                `json:"diff"`
	Kind FluffyPatchChangeKind `json:"kind"`
	Path string                `json:"path"`
}

type FluffyPatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type FluffyCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type FluffyUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []FluffyTextElement `json:"text_elements,omitempty"`
	Type         UserInputType       `json:"type"`
	Detail       *ImageDetail        `json:"detail"`
	URL          *string             `json:"url,omitempty"`
	Path         *string             `json:"path,omitempty"`
	Name         *string             `json:"name,omitempty"`
}

type FluffyTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange FluffyByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type FluffyByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type FluffyDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type FluffyMCPToolCallError struct {
	Message string `json:"message"`
}

type FluffyHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type FluffyMemoryCitation struct {
	Entries   []FluffyMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                    `json:"threadIds"`
}

type FluffyMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type FluffyMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

type TentacledCustomMultiAgentMode struct {
	Custom string `json:"custom"`
}

// Legacy sandbox policy retained for compatibility. Experimental clients should prefer
// `activePermissionProfile` for profile provenance.
type ThreadResumeResponseSandboxPolicy struct {
	Type                SandboxPolicyType   `json:"type"`
	NetworkAccess       *NetworkAccessUnion `json:"networkAccess"`
	ExcludeSlashTmp     *bool               `json:"excludeSlashTmp,omitempty"`
	ExcludeTmpdirEnvVar *bool               `json:"excludeTmpdirEnvVar,omitempty"`
	WritableRoots       []string            `json:"writableRoots,omitempty"`
}

type ThreadResumeResponseThread struct {
	// Optional random unique nickname assigned to an AgentControl-spawned sub-agent.
	AgentNickname *string `json:"agentNickname"`
	// Optional role (agent_role) assigned to an AgentControl-spawned sub-agent.
	AgentRole *string `json:"agentRole"`
	// Whether the app server accepts direct turn input for this loaded thread. `None` means the
	// capability is unavailable, such as for an unloaded stored thread.
	CanAcceptDirectInput *bool `json:"canAcceptDirectInput"`
	// Version of the CLI that created the thread.
	CLIVersion string `json:"cliVersion"`
	// Unix timestamp (in seconds) when the thread was created.
	CreatedAt int64 `json:"createdAt"`
	// Working directory captured for the thread.
	Cwd string `json:"cwd"`
	// Whether the thread is ephemeral and should not be materialized on disk.
	Ephemeral bool `json:"ephemeral"`
	// Optional implementation-specific thread data.
	Extra map[string]interface{} `json:"extra"`
	// Source thread id when this thread was created by forking another thread.
	ForkedFromID *string `json:"forkedFromId"`
	// Optional Git metadata captured when the thread was created.
	GitInfo *FluffyGitInfo `json:"gitInfo"`
	// Persisted thread history contract selected when this thread was created.
	HistoryMode *ThreadHistoryMode `json:"historyMode,omitempty"`
	// Identifier for this thread. Codex-generated thread IDs are UUIDv7.
	ID string `json:"id"`
	// Whether the thread has been pinned by the user.
	IsPinned *bool `json:"isPinned,omitempty"`
	// Model provider used for this thread (for example, 'openai').
	ModelProvider string `json:"modelProvider"`
	// Optional user-facing thread title.
	Name *string `json:"name"`
	// The ID of the parent thread. This will only be set if this thread is a subagent.
	ParentThreadID *string `json:"parentThreadId"`
	// [UNSTABLE] Path to the thread on disk.
	Path *string `json:"path"`
	// Usually the first user message in the thread, if available.
	Preview string `json:"preview"`
	// Unix timestamp (in seconds) used for thread recency ordering.
	RecencyAt *int64 `json:"recencyAt"`
	// Session id shared by threads that belong to the same session tree.
	SessionID string `json:"sessionId"`
	// Origin of the thread (CLI, VSCode, codex exec, codex app-server, etc.).
	Source *IndigoSessionSource `json:"source"`
	// Current runtime status for the thread.
	Status FluffyThreadStatus `json:"status"`
	// Optional analytics source classification for this thread.
	ThreadSource *string `json:"threadSource"`
	// Only populated on `thread/resume`, `thread/rollback`, `thread/fork`, and `thread/read`
	// (when `includeTurns` is true) responses. For all other responses and notifications
	// returning a Thread, the turns field will be an empty list.
	Turns []DatumElement `json:"turns"`
	// Unix timestamp (in seconds) when the thread was last updated.
	UpdatedAt int64 `json:"updatedAt"`
}

type FluffyGitInfo struct {
	Branch    *string `json:"branch"`
	OriginURL *string `json:"originUrl"`
	SHA       *string `json:"sha"`
}

type FluffySessionSource struct {
	Custom   *string               `json:"custom,omitempty"`
	SubAgent *IndigoSubAgentSource `json:"subAgent"`
}

type FluffySubAgentSource struct {
	ThreadSpawn *FluffyThreadSpawn `json:"thread_spawn,omitempty"`
	Other       *string            `json:"other,omitempty"`
}

type FluffyThreadSpawn struct {
	AgentNickname  *string `json:"agent_nickname"`
	AgentPath      *string `json:"agent_path"`
	AgentRole      *string `json:"agent_role"`
	Depth          int64   `json:"depth"`
	ParentThreadID string  `json:"parent_thread_id"`
}

// Current runtime status for the thread.
type FluffyThreadStatus struct {
	Type        ThreadStatusType   `json:"type"`
	ActiveFlags []ThreadActiveFlag `json:"activeFlags,omitempty"`
}

type ThreadReadParams struct {
	// When true, include turns and their items from rollout history.
	IncludeTurns *bool  `json:"includeTurns,omitempty"`
	ThreadID     string `json:"threadId"`
}

type ThreadReadResponse struct {
	Thread ThreadReadResponseThread `json:"thread"`
}

type ThreadReadResponseThread struct {
	// Optional random unique nickname assigned to an AgentControl-spawned sub-agent.
	AgentNickname *string `json:"agentNickname"`
	// Optional role (agent_role) assigned to an AgentControl-spawned sub-agent.
	AgentRole *string `json:"agentRole"`
	// Whether the app server accepts direct turn input for this loaded thread. `None` means the
	// capability is unavailable, such as for an unloaded stored thread.
	CanAcceptDirectInput *bool `json:"canAcceptDirectInput"`
	// Version of the CLI that created the thread.
	CLIVersion string `json:"cliVersion"`
	// Unix timestamp (in seconds) when the thread was created.
	CreatedAt int64 `json:"createdAt"`
	// Working directory captured for the thread.
	Cwd string `json:"cwd"`
	// Whether the thread is ephemeral and should not be materialized on disk.
	Ephemeral bool `json:"ephemeral"`
	// Optional implementation-specific thread data.
	Extra map[string]interface{} `json:"extra"`
	// Source thread id when this thread was created by forking another thread.
	ForkedFromID *string `json:"forkedFromId"`
	// Optional Git metadata captured when the thread was created.
	GitInfo *TentacledGitInfo `json:"gitInfo"`
	// Persisted thread history contract selected when this thread was created.
	HistoryMode *ThreadHistoryMode `json:"historyMode,omitempty"`
	// Identifier for this thread. Codex-generated thread IDs are UUIDv7.
	ID string `json:"id"`
	// Whether the thread has been pinned by the user.
	IsPinned *bool `json:"isPinned,omitempty"`
	// Model provider used for this thread (for example, 'openai').
	ModelProvider string `json:"modelProvider"`
	// Optional user-facing thread title.
	Name *string `json:"name"`
	// The ID of the parent thread. This will only be set if this thread is a subagent.
	ParentThreadID *string `json:"parentThreadId"`
	// [UNSTABLE] Path to the thread on disk.
	Path *string `json:"path"`
	// Usually the first user message in the thread, if available.
	Preview string `json:"preview"`
	// Unix timestamp (in seconds) used for thread recency ordering.
	RecencyAt *int64 `json:"recencyAt"`
	// Session id shared by threads that belong to the same session tree.
	SessionID string `json:"sessionId"`
	// Origin of the thread (CLI, VSCode, codex exec, codex app-server, etc.).
	Source *IndecentSessionSource `json:"source"`
	// Current runtime status for the thread.
	Status TentacledThreadStatus `json:"status"`
	// Optional analytics source classification for this thread.
	ThreadSource *string `json:"threadSource"`
	// Only populated on `thread/resume`, `thread/rollback`, `thread/fork`, and `thread/read`
	// (when `includeTurns` is true) responses. For all other responses and notifications
	// returning a Thread, the turns field will be an empty list.
	Turns []FluffyTurn `json:"turns"`
	// Unix timestamp (in seconds) when the thread was last updated.
	UpdatedAt int64 `json:"updatedAt"`
}

type TentacledGitInfo struct {
	Branch    *string `json:"branch"`
	OriginURL *string `json:"originUrl"`
	SHA       *string `json:"sha"`
}

type TentacledSessionSource struct {
	Custom   *string                 `json:"custom,omitempty"`
	SubAgent *IndecentSubAgentSource `json:"subAgent"`
}

type TentacledSubAgentSource struct {
	ThreadSpawn *TentacledThreadSpawn `json:"thread_spawn,omitempty"`
	Other       *string               `json:"other,omitempty"`
}

type TentacledThreadSpawn struct {
	AgentNickname  *string `json:"agent_nickname"`
	AgentPath      *string `json:"agent_path"`
	AgentRole      *string `json:"agent_role"`
	Depth          int64   `json:"depth"`
	ParentThreadID string  `json:"parent_thread_id"`
}

// Current runtime status for the thread.
type TentacledThreadStatus struct {
	Type        ThreadStatusType   `json:"type"`
	ActiveFlags []ThreadActiveFlag `json:"activeFlags,omitempty"`
}

type FluffyTurn struct {
	// Unix timestamp (in seconds) when the turn completed.
	CompletedAt *int64 `json:"completedAt"`
	// Duration between turn start and completion in milliseconds, if known.
	DurationMS *int64 `json:"durationMs"`
	// Only populated when the Turn's status is failed.
	Error *FluffyTurnError `json:"error"`
	// Identifier for this turn. Codex-generated turn IDs are UUIDv7.
	ID string `json:"id"`
	// Thread items currently included in this turn payload.
	Items []FluffyThreadItem `json:"items"`
	// Describes how much of `items` has been loaded for this turn.
	ItemsView *TurnItemsView `json:"itemsView,omitempty"`
	// Unix timestamp (in seconds) when the turn started.
	StartedAt *int64     `json:"startedAt"`
	Status    TurnStatus `json:"status"`
}

type FluffyTurnError struct {
	AdditionalDetails *string                `json:"additionalDetails"`
	CodexErrorInfo    *MagentaCodexErrorInfo `json:"codexErrorInfo"`
	Message           string                 `json:"message"`
}

// Failed to connect to the response SSE stream.
//
// The response SSE stream disconnected in the middle of a turn before completion.
//
// Reached the retry limit for responses.
//
// Returned when `turn/start` or `turn/steer` is submitted while the current active turn
// cannot accept same-turn steering, for example `/review` or manual `/compact`.
type TentacledCodexErrorInfo struct {
	HTTPConnectionFailed           *TentacledHTTPConnectionFailed           `json:"httpConnectionFailed,omitempty"`
	ResponseStreamConnectionFailed *TentacledResponseStreamConnectionFailed `json:"responseStreamConnectionFailed,omitempty"`
	ResponseStreamDisconnected     *TentacledResponseStreamDisconnected     `json:"responseStreamDisconnected,omitempty"`
	ResponseTooManyFailedAttempts  *TentacledResponseTooManyFailedAttempts  `json:"responseTooManyFailedAttempts,omitempty"`
	ActiveTurnNotSteerable         *TentacledActiveTurnNotSteerable         `json:"activeTurnNotSteerable,omitempty"`
}

type TentacledActiveTurnNotSteerable struct {
	TurnKind NonSteerableTurnKind `json:"turnKind"`
}

type TentacledHTTPConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type TentacledResponseStreamConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type TentacledResponseStreamDisconnected struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type TentacledResponseTooManyFailedAttempts struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type FluffyThreadItem struct {
	ClientID *string           `json:"clientId"`
	Content  []FriskyUserInput `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                        `json:"id"`
	Type           ThreadItemType                `json:"type"`
	Fragments      []TentacledHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *TentacledMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase                 `json:"phase"`
	Text           *string                       `json:"text,omitempty"`
	Summary        []string                      `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []TentacledCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                         `json:"status,omitempty"`
	Changes    []TentacledFileUpdateChange     `json:"changes,omitempty"`
	AppContext *TentacledMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                     `json:"arguments"`
	Error      *TentacledMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string          `json:"mcpAppResourceUri"`
	Result            *TentacledResult `json:"result"`
	Server            *string          `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                     `json:"tool,omitempty"`
	ContentItems []TentacledDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                     `json:"namespace"`
	Success      *bool                                       `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]TentacledCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                   `json:"senderThreadId,omitempty"`
	AgentPath      *string                   `json:"agentPath,omitempty"`
	AgentThreadID  *string                   `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind     `json:"kind,omitempty"`
	Action         *TentacledWebSearchAction `json:"action"`
	Query          *string                   `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type TentacledWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type TentacledCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type TentacledMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type TentacledFileUpdateChange struct {
	Diff string                   `json:"diff"`
	Kind TentacledPatchChangeKind `json:"kind"`
	Path string                   `json:"path"`
}

type TentacledPatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type TentacledCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type TentacledUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []TentacledTextElement `json:"text_elements,omitempty"`
	Type         UserInputType          `json:"type"`
	Detail       *ImageDetail           `json:"detail"`
	URL          *string                `json:"url,omitempty"`
	Path         *string                `json:"path,omitempty"`
	Name         *string                `json:"name,omitempty"`
}

type TentacledTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange TentacledByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type TentacledByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type TentacledDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type TentacledMCPToolCallError struct {
	Message string `json:"message"`
}

type TentacledHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type TentacledMemoryCitation struct {
	Entries   []TentacledMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                       `json:"threadIds"`
}

type TentacledMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type TentacledMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

type TurnStartParams struct {
	// Optional client-provided context fragments keyed by an opaque source identifier.
	AdditionalContext map[string]AdditionalContextEntry `json:"additionalContext"`
	// Override the approval policy for this turn and subsequent turns.
	ApprovalPolicy *TurnStartParamsApprovalPolicy `json:"approvalPolicy"`
	// Override where approval requests are routed for review on this turn and subsequent turns.
	ApprovalsReviewer   *ApprovalsReviewer `json:"approvalsReviewer"`
	ClientUserMessageID *string            `json:"clientUserMessageId"`
	// EXPERIMENTAL - Set a pre-set collaboration mode. Takes precedence over model,
	// reasoning_effort, and developer instructions if set.
	//
	// For `collaboration_mode.settings.developer_instructions`, `null` means "use the built-in
	// instructions for the selected mode".
	CollaborationMode *CollaborationMode `json:"collaborationMode"`
	// Override the working directory for this turn and subsequent turns.
	Cwd *string `json:"cwd"`
	// Override the reasoning effort for this turn and subsequent turns.
	Effort *string `json:"effort"`
	// Optional environments for this turn and subsequent turns.
	//
	// Omitted uses the thread sticky environments. Empty disables environment access for this
	// turn. Non-empty selects the first environment as the current turn environment for this
	// turn.
	Environments []TurnStartParamsEnvironment `json:"environments"`
	Input        []UserInput                  `json:"input"`
	// Override the model for this turn and subsequent turns.
	Model *string `json:"model"`
	// @deprecated Ignored. Use `effort: "ultra"` for proactive multi-agent behavior.
	MultiAgentMode *TurnStartParamsMultiAgentMode `json:"multiAgentMode"`
	// Optional JSON Schema used to constrain the final assistant message for this turn.
	OutputSchema interface{} `json:"outputSchema"`
	// Select a named permissions profile id for this turn and subsequent turns. Cannot be
	// combined with `sandboxPolicy`.
	Permissions *string `json:"permissions"`
	// Override the personality for this turn and subsequent turns.
	Personality *Personality `json:"personality"`
	// Optional metadata to enrich Codex's ResponsesAPI turn metadata.
	//
	// Entries are flattened into the JSON string sent as
	// `client_metadata["x-codex-turn-metadata"]` on ResponsesAPI HTTP and websocket requests.
	//
	// They are not sent as top-level ResponsesAPI `client_metadata` keys, and reserved keys
	// such as `session_id`, `thread_id`, `turn_id`, and `window_id` cannot be overridden.
	ResponsesapiClientMetadata map[string]string `json:"responsesapiClientMetadata"`
	// Replace the thread's runtime workspace roots for this turn and subsequent turns. Paths
	// must be absolute.
	RuntimeWorkspaceRoots []string `json:"runtimeWorkspaceRoots"`
	// Override the sandbox policy for this turn and subsequent turns.
	SandboxPolicy *SandboxPolicy `json:"sandboxPolicy"`
	// Override the service tier for this turn and subsequent turns.
	ServiceTier *string `json:"serviceTier"`
	// Override the reasoning summary for this turn and subsequent turns.
	Summary  *ReasoningSummary `json:"summary"`
	ThreadID string            `json:"threadId"`
}

type AdditionalContextEntry struct {
	Kind  AdditionalContextKind `json:"kind"`
	Value string                `json:"value"`
}

type IndigoGranularAskForApproval struct {
	Granular IndigoGranular `json:"granular"`
}

type IndigoGranular struct {
	MCPElicitations    bool  `json:"mcp_elicitations"`
	RequestPermissions *bool `json:"request_permissions,omitempty"`
	Rules              bool  `json:"rules"`
	SandboxApproval    bool  `json:"sandbox_approval"`
	SkillApproval      *bool `json:"skill_approval,omitempty"`
}

// Collaboration mode for a Codex session.
type CollaborationMode struct {
	Mode     ModeKind `json:"mode"`
	Settings Settings `json:"settings"`
}

// Settings for a collaboration mode.
type Settings struct {
	DeveloperInstructions *string `json:"developer_instructions"`
	Model                 string  `json:"model"`
	ReasoningEffort       *string `json:"reasoning_effort"`
}

type TurnStartParamsEnvironment struct {
	Cwd           string `json:"cwd"`
	EnvironmentID string `json:"environmentId"`
	// Environment-native runtime workspace roots. Omitted defaults to `cwd`.
	RuntimeWorkspaceRoots []string `json:"runtimeWorkspaceRoots"`
}

type UserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []UserInputTextElement `json:"text_elements,omitempty"`
	Type         UserInputType          `json:"type"`
	Detail       *ImageDetail           `json:"detail"`
	URL          *string                `json:"url,omitempty"`
	Path         *string                `json:"path,omitempty"`
	Name         *string                `json:"name,omitempty"`
}

type UserInputTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange StickyByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type StickyByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type StickyCustomMultiAgentMode struct {
	Custom string `json:"custom"`
}

type SandboxPolicy struct {
	Type                SandboxPolicyType   `json:"type"`
	NetworkAccess       *NetworkAccessUnion `json:"networkAccess"`
	ExcludeSlashTmp     *bool               `json:"excludeSlashTmp,omitempty"`
	ExcludeTmpdirEnvVar *bool               `json:"excludeTmpdirEnvVar,omitempty"`
	WritableRoots       []string            `json:"writableRoots,omitempty"`
}

type TurnStartResponse struct {
	Turn TurnStartResponseTurn `json:"turn"`
}

type TurnStartResponseTurn struct {
	// Unix timestamp (in seconds) when the turn completed.
	CompletedAt *int64 `json:"completedAt"`
	// Duration between turn start and completion in milliseconds, if known.
	DurationMS *int64 `json:"durationMs"`
	// Only populated when the Turn's status is failed.
	Error *TentacledTurnError `json:"error"`
	// Identifier for this turn. Codex-generated turn IDs are UUIDv7.
	ID string `json:"id"`
	// Thread items currently included in this turn payload.
	Items []TentacledThreadItem `json:"items"`
	// Describes how much of `items` has been loaded for this turn.
	ItemsView *TurnItemsView `json:"itemsView,omitempty"`
	// Unix timestamp (in seconds) when the turn started.
	StartedAt *int64     `json:"startedAt"`
	Status    TurnStatus `json:"status"`
}

type TentacledTurnError struct {
	AdditionalDetails *string               `json:"additionalDetails"`
	CodexErrorInfo    *FriskyCodexErrorInfo `json:"codexErrorInfo"`
	Message           string                `json:"message"`
}

// Failed to connect to the response SSE stream.
//
// The response SSE stream disconnected in the middle of a turn before completion.
//
// Reached the retry limit for responses.
//
// Returned when `turn/start` or `turn/steer` is submitted while the current active turn
// cannot accept same-turn steering, for example `/review` or manual `/compact`.
type StickyCodexErrorInfo struct {
	HTTPConnectionFailed           *StickyHTTPConnectionFailed           `json:"httpConnectionFailed,omitempty"`
	ResponseStreamConnectionFailed *StickyResponseStreamConnectionFailed `json:"responseStreamConnectionFailed,omitempty"`
	ResponseStreamDisconnected     *StickyResponseStreamDisconnected     `json:"responseStreamDisconnected,omitempty"`
	ResponseTooManyFailedAttempts  *StickyResponseTooManyFailedAttempts  `json:"responseTooManyFailedAttempts,omitempty"`
	ActiveTurnNotSteerable         *StickyActiveTurnNotSteerable         `json:"activeTurnNotSteerable,omitempty"`
}

type StickyActiveTurnNotSteerable struct {
	TurnKind NonSteerableTurnKind `json:"turnKind"`
}

type StickyHTTPConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type StickyResponseStreamConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type StickyResponseStreamDisconnected struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type StickyResponseTooManyFailedAttempts struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type TentacledThreadItem struct {
	ClientID *string                `json:"clientId"`
	Content  []MischievousUserInput `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                     `json:"id"`
	Type           ThreadItemType             `json:"type"`
	Fragments      []StickyHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *StickyMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase              `json:"phase"`
	Text           *string                    `json:"text,omitempty"`
	Summary        []string                   `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []StickyCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                      `json:"status,omitempty"`
	Changes    []StickyFileUpdateChange     `json:"changes,omitempty"`
	AppContext *StickyMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                  `json:"arguments"`
	Error      *StickyMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string       `json:"mcpAppResourceUri"`
	Result            *StickyResult `json:"result"`
	Server            *string       `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                  `json:"tool,omitempty"`
	ContentItems []StickyDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                  `json:"namespace"`
	Success      *bool                                    `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]StickyCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                `json:"senderThreadId,omitempty"`
	AgentPath      *string                `json:"agentPath,omitempty"`
	AgentThreadID  *string                `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind  `json:"kind,omitempty"`
	Action         *StickyWebSearchAction `json:"action"`
	Query          *string                `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type StickyWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type StickyCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type StickyMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type StickyFileUpdateChange struct {
	Diff string                `json:"diff"`
	Kind StickyPatchChangeKind `json:"kind"`
	Path string                `json:"path"`
}

type StickyPatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type StickyCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type StickyUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []StickyTextElement `json:"text_elements,omitempty"`
	Type         UserInputType       `json:"type"`
	Detail       *ImageDetail        `json:"detail"`
	URL          *string             `json:"url,omitempty"`
	Path         *string             `json:"path,omitempty"`
	Name         *string             `json:"name,omitempty"`
}

type StickyTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange IndigoByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type IndigoByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type StickyDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type StickyMCPToolCallError struct {
	Message string `json:"message"`
}

type StickyHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type StickyMemoryCitation struct {
	Entries   []StickyMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                    `json:"threadIds"`
}

type StickyMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type StickyMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

type TurnInterruptParams struct {
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

type TurnStartedNotification struct {
	ThreadID string                      `json:"threadId"`
	Turn     TurnStartedNotificationTurn `json:"turn"`
}

type TurnStartedNotificationTurn struct {
	// Unix timestamp (in seconds) when the turn completed.
	CompletedAt *int64 `json:"completedAt"`
	// Duration between turn start and completion in milliseconds, if known.
	DurationMS *int64 `json:"durationMs"`
	// Only populated when the Turn's status is failed.
	Error *StickyTurnError `json:"error"`
	// Identifier for this turn. Codex-generated turn IDs are UUIDv7.
	ID string `json:"id"`
	// Thread items currently included in this turn payload.
	Items []StickyThreadItem `json:"items"`
	// Describes how much of `items` has been loaded for this turn.
	ItemsView *TurnItemsView `json:"itemsView,omitempty"`
	// Unix timestamp (in seconds) when the turn started.
	StartedAt *int64     `json:"startedAt"`
	Status    TurnStatus `json:"status"`
}

type StickyTurnError struct {
	AdditionalDetails *string                    `json:"additionalDetails"`
	CodexErrorInfo    *MischievousCodexErrorInfo `json:"codexErrorInfo"`
	Message           string                     `json:"message"`
}

// Failed to connect to the response SSE stream.
//
// The response SSE stream disconnected in the middle of a turn before completion.
//
// Reached the retry limit for responses.
//
// Returned when `turn/start` or `turn/steer` is submitted while the current active turn
// cannot accept same-turn steering, for example `/review` or manual `/compact`.
type IndigoCodexErrorInfo struct {
	HTTPConnectionFailed           *IndigoHTTPConnectionFailed           `json:"httpConnectionFailed,omitempty"`
	ResponseStreamConnectionFailed *IndigoResponseStreamConnectionFailed `json:"responseStreamConnectionFailed,omitempty"`
	ResponseStreamDisconnected     *IndigoResponseStreamDisconnected     `json:"responseStreamDisconnected,omitempty"`
	ResponseTooManyFailedAttempts  *IndigoResponseTooManyFailedAttempts  `json:"responseTooManyFailedAttempts,omitempty"`
	ActiveTurnNotSteerable         *IndigoActiveTurnNotSteerable         `json:"activeTurnNotSteerable,omitempty"`
}

type IndigoActiveTurnNotSteerable struct {
	TurnKind NonSteerableTurnKind `json:"turnKind"`
}

type IndigoHTTPConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type IndigoResponseStreamConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type IndigoResponseStreamDisconnected struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type IndigoResponseTooManyFailedAttempts struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type StickyThreadItem struct {
	ClientID *string                  `json:"clientId"`
	Content  []BraggadociousUserInput `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                     `json:"id"`
	Type           ThreadItemType             `json:"type"`
	Fragments      []IndigoHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *IndigoMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase              `json:"phase"`
	Text           *string                    `json:"text,omitempty"`
	Summary        []string                   `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []IndigoCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                      `json:"status,omitempty"`
	Changes    []IndigoFileUpdateChange     `json:"changes,omitempty"`
	AppContext *IndigoMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                  `json:"arguments"`
	Error      *IndigoMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string       `json:"mcpAppResourceUri"`
	Result            *IndigoResult `json:"result"`
	Server            *string       `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                  `json:"tool,omitempty"`
	ContentItems []IndigoDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                  `json:"namespace"`
	Success      *bool                                    `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]IndigoCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                `json:"senderThreadId,omitempty"`
	AgentPath      *string                `json:"agentPath,omitempty"`
	AgentThreadID  *string                `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind  `json:"kind,omitempty"`
	Action         *IndigoWebSearchAction `json:"action"`
	Query          *string                `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type IndigoWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type IndigoCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type IndigoMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type IndigoFileUpdateChange struct {
	Diff string                `json:"diff"`
	Kind IndigoPatchChangeKind `json:"kind"`
	Path string                `json:"path"`
}

type IndigoPatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type IndigoCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type IndigoUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []IndigoTextElement `json:"text_elements,omitempty"`
	Type         UserInputType       `json:"type"`
	Detail       *ImageDetail        `json:"detail"`
	URL          *string             `json:"url,omitempty"`
	Path         *string             `json:"path,omitempty"`
	Name         *string             `json:"name,omitempty"`
}

type IndigoTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange IndecentByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type IndecentByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type IndigoDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type IndigoMCPToolCallError struct {
	Message string `json:"message"`
}

type IndigoHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type IndigoMemoryCitation struct {
	Entries   []IndigoMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                    `json:"threadIds"`
}

type IndigoMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type IndigoMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

type TurnCompletedNotification struct {
	ThreadID string                        `json:"threadId"`
	Turn     TurnCompletedNotificationTurn `json:"turn"`
}

type TurnCompletedNotificationTurn struct {
	// Unix timestamp (in seconds) when the turn completed.
	CompletedAt *int64 `json:"completedAt"`
	// Duration between turn start and completion in milliseconds, if known.
	DurationMS *int64 `json:"durationMs"`
	// Only populated when the Turn's status is failed.
	Error *IndigoTurnError `json:"error"`
	// Identifier for this turn. Codex-generated turn IDs are UUIDv7.
	ID string `json:"id"`
	// Thread items currently included in this turn payload.
	Items []IndigoThreadItem `json:"items"`
	// Describes how much of `items` has been loaded for this turn.
	ItemsView *TurnItemsView `json:"itemsView,omitempty"`
	// Unix timestamp (in seconds) when the turn started.
	StartedAt *int64     `json:"startedAt"`
	Status    TurnStatus `json:"status"`
}

type IndigoTurnError struct {
	AdditionalDetails *string                      `json:"additionalDetails"`
	CodexErrorInfo    *BraggadociousCodexErrorInfo `json:"codexErrorInfo"`
	Message           string                       `json:"message"`
}

// Failed to connect to the response SSE stream.
//
// The response SSE stream disconnected in the middle of a turn before completion.
//
// Reached the retry limit for responses.
//
// Returned when `turn/start` or `turn/steer` is submitted while the current active turn
// cannot accept same-turn steering, for example `/review` or manual `/compact`.
type IndecentCodexErrorInfo struct {
	HTTPConnectionFailed           *IndecentHTTPConnectionFailed           `json:"httpConnectionFailed,omitempty"`
	ResponseStreamConnectionFailed *IndecentResponseStreamConnectionFailed `json:"responseStreamConnectionFailed,omitempty"`
	ResponseStreamDisconnected     *IndecentResponseStreamDisconnected     `json:"responseStreamDisconnected,omitempty"`
	ResponseTooManyFailedAttempts  *IndecentResponseTooManyFailedAttempts  `json:"responseTooManyFailedAttempts,omitempty"`
	ActiveTurnNotSteerable         *IndecentActiveTurnNotSteerable         `json:"activeTurnNotSteerable,omitempty"`
}

type IndecentActiveTurnNotSteerable struct {
	TurnKind NonSteerableTurnKind `json:"turnKind"`
}

type IndecentHTTPConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type IndecentResponseStreamConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type IndecentResponseStreamDisconnected struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type IndecentResponseTooManyFailedAttempts struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type IndigoThreadItem struct {
	ClientID *string      `json:"clientId"`
	Content  []UserInput1 `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                       `json:"id"`
	Type           ThreadItemType               `json:"type"`
	Fragments      []IndecentHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *IndecentMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase                `json:"phase"`
	Text           *string                      `json:"text,omitempty"`
	Summary        []string                     `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []IndecentCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                        `json:"status,omitempty"`
	Changes    []IndecentFileUpdateChange     `json:"changes,omitempty"`
	AppContext *IndecentMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                    `json:"arguments"`
	Error      *IndecentMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string         `json:"mcpAppResourceUri"`
	Result            *IndecentResult `json:"result"`
	Server            *string         `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                    `json:"tool,omitempty"`
	ContentItems []IndecentDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                    `json:"namespace"`
	Success      *bool                                      `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]IndecentCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                  `json:"senderThreadId,omitempty"`
	AgentPath      *string                  `json:"agentPath,omitempty"`
	AgentThreadID  *string                  `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind    `json:"kind,omitempty"`
	Action         *IndecentWebSearchAction `json:"action"`
	Query          *string                  `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type IndecentWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type IndecentCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type IndecentMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type IndecentFileUpdateChange struct {
	Diff string                  `json:"diff"`
	Kind IndecentPatchChangeKind `json:"kind"`
	Path string                  `json:"path"`
}

type IndecentPatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type IndecentCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type IndecentUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []IndecentTextElement `json:"text_elements,omitempty"`
	Type         UserInputType         `json:"type"`
	Detail       *ImageDetail          `json:"detail"`
	URL          *string               `json:"url,omitempty"`
	Path         *string               `json:"path,omitempty"`
	Name         *string               `json:"name,omitempty"`
}

type IndecentTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange HilariousByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type HilariousByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type IndecentDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type IndecentMCPToolCallError struct {
	Message string `json:"message"`
}

type IndecentHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type IndecentMemoryCitation struct {
	Entries   []IndecentMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                      `json:"threadIds"`
}

type IndecentMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type IndecentMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

type ItemStartedNotification struct {
	Item ItemStartedNotificationThreadItem `json:"item"`
	// Unix timestamp (in milliseconds) when this item lifecycle started.
	StartedAtMS int64  `json:"startedAtMs"`
	ThreadID    string `json:"threadId"`
	TurnID      string `json:"turnId"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type ItemStartedNotificationThreadItem struct {
	ClientID *string      `json:"clientId"`
	Content  []UserInput2 `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                        `json:"id"`
	Type           ThreadItemType                `json:"type"`
	Fragments      []HilariousHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *HilariousMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase                 `json:"phase"`
	Text           *string                       `json:"text,omitempty"`
	Summary        []string                      `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []HilariousCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                         `json:"status,omitempty"`
	Changes    []HilariousFileUpdateChange     `json:"changes,omitempty"`
	AppContext *HilariousMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                     `json:"arguments"`
	Error      *HilariousMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string          `json:"mcpAppResourceUri"`
	Result            *HilariousResult `json:"result"`
	Server            *string          `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                     `json:"tool,omitempty"`
	ContentItems []HilariousDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                     `json:"namespace"`
	Success      *bool                                       `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]HilariousCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                   `json:"senderThreadId,omitempty"`
	AgentPath      *string                   `json:"agentPath,omitempty"`
	AgentThreadID  *string                   `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind     `json:"kind,omitempty"`
	Action         *HilariousWebSearchAction `json:"action"`
	Query          *string                   `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type HilariousWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type HilariousCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type HilariousMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type HilariousFileUpdateChange struct {
	Diff string                   `json:"diff"`
	Kind HilariousPatchChangeKind `json:"kind"`
	Path string                   `json:"path"`
}

type HilariousPatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type HilariousCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type HilariousUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []HilariousTextElement `json:"text_elements,omitempty"`
	Type         UserInputType          `json:"type"`
	Detail       *ImageDetail           `json:"detail"`
	URL          *string                `json:"url,omitempty"`
	Path         *string                `json:"path,omitempty"`
	Name         *string                `json:"name,omitempty"`
}

type HilariousTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange AmbitiousByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type AmbitiousByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type HilariousDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type HilariousMCPToolCallError struct {
	Message string `json:"message"`
}

type HilariousHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type HilariousMemoryCitation struct {
	Entries   []HilariousMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                       `json:"threadIds"`
}

type HilariousMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type HilariousMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

type ItemCompletedNotification struct {
	// Unix timestamp (in milliseconds) when this item lifecycle completed.
	CompletedAtMS int64                               `json:"completedAtMs"`
	Item          ItemCompletedNotificationThreadItem `json:"item"`
	ThreadID      string                              `json:"threadId"`
	TurnID        string                              `json:"turnId"`
}

// EXPERIMENTAL - proposed plan item content. The completed plan item is authoritative and
// may not match the concatenation of `PlanDelta` text.
//
// Display item emitted by the interruptible `clock.sleep` tool.
type ItemCompletedNotificationThreadItem struct {
	ClientID *string      `json:"clientId"`
	Content  []UserInput3 `json:"content,omitempty"`
	// Unique identifier for this collab tool call.
	ID             string                        `json:"id"`
	Type           ThreadItemType                `json:"type"`
	Fragments      []AmbitiousHookPromptFragment `json:"fragments,omitempty"`
	MemoryCitation *AmbitiousMemoryCitation      `json:"memoryCitation"`
	Phase          *MessagePhase                 `json:"phase"`
	Text           *string                       `json:"text,omitempty"`
	Summary        []string                      `json:"summary,omitempty"`
	// The command's output, aggregated from stdout and stderr.
	AggregatedOutput *string `json:"aggregatedOutput"`
	// The command to be executed.
	Command *string `json:"command,omitempty"`
	// A best-effort parsing of the command to understand the action(s) it will perform. This
	// returns a list of CommandAction objects because a single shell command may be composed of
	// many commands piped together.
	CommandActions []AmbitiousCommandAction `json:"commandActions,omitempty"`
	// The command's working directory.
	Cwd *string `json:"cwd,omitempty"`
	// The duration of the command execution in milliseconds.
	//
	// The duration of the MCP tool call in milliseconds.
	//
	// The duration of the dynamic tool call in milliseconds.
	DurationMS *int64 `json:"durationMs"`
	// The command's exit code.
	ExitCode *int64 `json:"exitCode"`
	// Trusted first-party plugin id when this command resolves to one plugin script.
	PluginID *string `json:"pluginId"`
	// Identifier for the underlying PTY process (when available).
	ProcessID *string `json:"processId"`
	// Safe plugin-relative path when this command resolves to one plugin script.
	ScriptPath *string                 `json:"scriptPath"`
	Source     *CommandExecutionSource `json:"source,omitempty"`
	// Current status of the collab tool call.
	Status     *string                         `json:"status,omitempty"`
	Changes    []AmbitiousFileUpdateChange     `json:"changes,omitempty"`
	AppContext *AmbitiousMCPToolCallAppContext `json:"appContext"`
	Arguments  interface{}                     `json:"arguments"`
	Error      *AmbitiousMCPToolCallError      `json:"error"`
	// Deprecated: use `appContext.resourceUri` instead.
	MCPAppResourceURI *string          `json:"mcpAppResourceUri"`
	Result            *AmbitiousResult `json:"result"`
	Server            *string          `json:"server,omitempty"`
	// Name of the collab tool that was invoked.
	Tool         *string                                     `json:"tool,omitempty"`
	ContentItems []AmbitiousDynamicToolCallOutputContentItem `json:"contentItems"`
	Namespace    *string                                     `json:"namespace"`
	Success      *bool                                       `json:"success"`
	// Last known status of the target agents, when available.
	AgentsStates map[string]AmbitiousCollabAgentState `json:"agentsStates,omitempty"`
	// Model requested for the spawned agent, when applicable.
	Model *string `json:"model"`
	// Prompt text sent as part of the collab tool call, when available.
	Prompt *string `json:"prompt"`
	// Reasoning effort requested for the spawned agent, when applicable.
	ReasoningEffort *string `json:"reasoningEffort"`
	// Thread ID of the receiving agent, when applicable. In case of spawn operation, this
	// corresponds to the newly spawned agent.
	ReceiverThreadIDS []string `json:"receiverThreadIds,omitempty"`
	// Thread ID of the agent issuing the collab request.
	SenderThreadID *string                   `json:"senderThreadId,omitempty"`
	AgentPath      *string                   `json:"agentPath,omitempty"`
	AgentThreadID  *string                   `json:"agentThreadId,omitempty"`
	Kind           *SubAgentActivityKind     `json:"kind,omitempty"`
	Action         *AmbitiousWebSearchAction `json:"action"`
	Query          *string                   `json:"query,omitempty"`
	// Structured search results returned out-of-band by standalone web search.
	//
	// These stay as opaque JSON at the extension/app-server boundary so new result fields and
	// result types can pass through without a Codex release.
	Results       []interface{} `json:"results"`
	Path          *string       `json:"path,omitempty"`
	RevisedPrompt *string       `json:"revisedPrompt"`
	SavedPath     *string       `json:"savedPath"`
	Review        *string       `json:"review,omitempty"`
}

type AmbitiousWebSearchAction struct {
	Queries []string            `json:"queries"`
	Query   *string             `json:"query"`
	Type    WebSearchActionType `json:"type"`
	URL     *string             `json:"url"`
	Pattern *string             `json:"pattern"`
}

type AmbitiousCollabAgentState struct {
	Message *string           `json:"message"`
	Status  CollabAgentStatus `json:"status"`
}

type AmbitiousMCPToolCallAppContext struct {
	ActionName  *string `json:"actionName"`
	AppName     *string `json:"appName"`
	ConnectorID string  `json:"connectorId"`
	LinkID      *string `json:"linkId"`
	ResourceURI *string `json:"resourceUri"`
}

type AmbitiousFileUpdateChange struct {
	Diff string                   `json:"diff"`
	Kind AmbitiousPatchChangeKind `json:"kind"`
	Path string                   `json:"path"`
}

type AmbitiousPatchChangeKind struct {
	Type     PatchChangeKindType `json:"type"`
	MovePath *string             `json:"move_path"`
}

type AmbitiousCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type AmbitiousUserInput struct {
	Text *string `json:"text,omitempty"`
	// UI-defined spans within `text` used to render or persist special elements.
	TextElements []AmbitiousTextElement `json:"text_elements,omitempty"`
	Type         UserInputType          `json:"type"`
	Detail       *ImageDetail           `json:"detail"`
	URL          *string                `json:"url,omitempty"`
	Path         *string                `json:"path,omitempty"`
	Name         *string                `json:"name,omitempty"`
}

type AmbitiousTextElement struct {
	// Byte range in the parent `text` buffer that this element occupies.
	ByteRange CunningByteRange `json:"byteRange"`
	// Optional human-readable placeholder for the element, displayed in the UI.
	Placeholder *string `json:"placeholder"`
}

// Byte range in the parent `text` buffer that this element occupies.
type CunningByteRange struct {
	End   int64 `json:"end"`
	Start int64 `json:"start"`
}

type AmbitiousDynamicToolCallOutputContentItem struct {
	Text     *string                                   `json:"text,omitempty"`
	Type     InputDynamicToolCallOutputContentItemType `json:"type"`
	ImageURL *string                                   `json:"imageUrl,omitempty"`
	AudioURL *string                                   `json:"audioUrl,omitempty"`
}

type AmbitiousMCPToolCallError struct {
	Message string `json:"message"`
}

type AmbitiousHookPromptFragment struct {
	HookRunID string `json:"hookRunId"`
	Text      string `json:"text"`
}

type AmbitiousMemoryCitation struct {
	Entries   []AmbitiousMemoryCitationEntry `json:"entries"`
	ThreadIDS []string                       `json:"threadIds"`
}

type AmbitiousMemoryCitationEntry struct {
	LineEnd   int64  `json:"lineEnd"`
	LineStart int64  `json:"lineStart"`
	Note      string `json:"note"`
	Path      string `json:"path"`
}

type AmbitiousMCPToolCallResult struct {
	Meta              interface{}   `json:"_meta"`
	Content           []interface{} `json:"content"`
	StructuredContent interface{}   `json:"structuredContent"`
}

type AgentMessageDeltaNotification struct {
	Delta    string `json:"delta"`
	ItemID   string `json:"itemId"`
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

// EXPERIMENTAL - proposed plan streaming deltas for plan items. Clients should not assume
// concatenated deltas match the completed plan item content.
type PlanDeltaNotification struct {
	Delta    string `json:"delta"`
	ItemID   string `json:"itemId"`
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

type ReasoningSummaryTextDeltaNotification struct {
	Delta        string `json:"delta"`
	ItemID       string `json:"itemId"`
	SummaryIndex int64  `json:"summaryIndex"`
	ThreadID     string `json:"threadId"`
	TurnID       string `json:"turnId"`
}

type ReasoningSummaryPartAddedNotification struct {
	ItemID       string `json:"itemId"`
	SummaryIndex int64  `json:"summaryIndex"`
	ThreadID     string `json:"threadId"`
	TurnID       string `json:"turnId"`
}

type ReasoningTextDeltaNotification struct {
	ContentIndex int64  `json:"contentIndex"`
	Delta        string `json:"delta"`
	ItemID       string `json:"itemId"`
	ThreadID     string `json:"threadId"`
	TurnID       string `json:"turnId"`
}

type CommandExecutionOutputDeltaNotification struct {
	Delta    string `json:"delta"`
	ItemID   string `json:"itemId"`
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

// Notification that the turn-level unified diff has changed. Contains the latest aggregated
// diff across all file changes in the turn.
type TurnDiffUpdatedNotification struct {
	Diff     string `json:"diff"`
	ThreadID string `json:"threadId"`
	TurnID   string `json:"turnId"`
}

type TurnPlanUpdatedNotification struct {
	Explanation *string        `json:"explanation"`
	Plan        []TurnPlanStep `json:"plan"`
	ThreadID    string         `json:"threadId"`
	TurnID      string         `json:"turnId"`
}

type TurnPlanStep struct {
	Status TurnPlanStepStatus `json:"status"`
	Step   string             `json:"step"`
}

type ErrorNotification struct {
	Error     TurnError `json:"error"`
	ThreadID  string    `json:"threadId"`
	TurnID    string    `json:"turnId"`
	WillRetry bool      `json:"willRetry"`
}

type TurnError struct {
	AdditionalDetails *string              `json:"additionalDetails"`
	CodexErrorInfo    *ErrorCodexErrorInfo `json:"codexErrorInfo"`
	Message           string               `json:"message"`
}

// Failed to connect to the response SSE stream.
//
// The response SSE stream disconnected in the middle of a turn before completion.
//
// Reached the retry limit for responses.
//
// Returned when `turn/start` or `turn/steer` is submitted while the current active turn
// cannot accept same-turn steering, for example `/review` or manual `/compact`.
type HilariousCodexErrorInfo struct {
	HTTPConnectionFailed           *HilariousHTTPConnectionFailed           `json:"httpConnectionFailed,omitempty"`
	ResponseStreamConnectionFailed *HilariousResponseStreamConnectionFailed `json:"responseStreamConnectionFailed,omitempty"`
	ResponseStreamDisconnected     *HilariousResponseStreamDisconnected     `json:"responseStreamDisconnected,omitempty"`
	ResponseTooManyFailedAttempts  *HilariousResponseTooManyFailedAttempts  `json:"responseTooManyFailedAttempts,omitempty"`
	ActiveTurnNotSteerable         *HilariousActiveTurnNotSteerable         `json:"activeTurnNotSteerable,omitempty"`
}

type HilariousActiveTurnNotSteerable struct {
	TurnKind NonSteerableTurnKind `json:"turnKind"`
}

type HilariousHTTPConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type HilariousResponseStreamConnectionFailed struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type HilariousResponseStreamDisconnected struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type HilariousResponseTooManyFailedAttempts struct {
	HTTPStatusCode *int64 `json:"httpStatusCode"`
}

type WarningNotification struct {
	// Concise warning message for the user.
	Message string `json:"message"`
	// Optional thread target when the warning applies to a specific thread.
	ThreadID *string `json:"threadId"`
}

type ThreadTokenUsageUpdatedNotification struct {
	ThreadID   string           `json:"threadId"`
	TokenUsage ThreadTokenUsage `json:"tokenUsage"`
	TurnID     string           `json:"turnId"`
}

type ThreadTokenUsage struct {
	Last               TokenUsageBreakdown `json:"last"`
	ModelContextWindow *int64              `json:"modelContextWindow"`
	Total              TokenUsageBreakdown `json:"total"`
}

type TokenUsageBreakdown struct {
	CachedInputTokens     int64  `json:"cachedInputTokens"`
	CacheWriteInputTokens *int64 `json:"cacheWriteInputTokens,omitempty"`
	InputTokens           int64  `json:"inputTokens"`
	OutputTokens          int64  `json:"outputTokens"`
	ReasoningOutputTokens int64  `json:"reasoningOutputTokens"`
	TotalTokens           int64  `json:"totalTokens"`
}

type ServerRequestResolvedNotification struct {
	RequestID *RequestID `json:"requestId"`
	ThreadID  string     `json:"threadId"`
}

type CommandExecutionRequestApprovalParams struct {
	// Optional additional permissions requested for this command.
	AdditionalPermissions *AdditionalPermissionProfile `json:"additionalPermissions"`
	// Unique identifier for this specific approval callback.
	//
	// For regular shell/unified_exec approvals, this is null.
	//
	// For zsh-exec-bridge subcommand approvals, multiple callbacks can belong to one parent
	// `itemId`, so `approvalId` is a distinct opaque callback id (a UUID) used to disambiguate
	// routing.
	ApprovalID *string `json:"approvalId"`
	// Ordered list of decisions the client may present for this prompt.
	AvailableDecisions []CommandExecutionApprovalDecisionElement `json:"availableDecisions"`
	// The command to be executed.
	Command *string `json:"command"`
	// Best-effort parsed command actions for friendly display.
	CommandActions []CommandExecutionRequestApprovalParamsCommandAction `json:"commandActions"`
	// The command's working directory.
	Cwd *string `json:"cwd"`
	// Environment in which the command will run.
	EnvironmentID *string `json:"environmentId"`
	ItemID        string  `json:"itemId"`
	// Optional context for a managed-network approval prompt.
	NetworkApprovalContext *NetworkApprovalContext `json:"networkApprovalContext"`
	// Optional proposed execpolicy amendment to allow similar commands without prompting.
	ProposedExecpolicyAmendment []string `json:"proposedExecpolicyAmendment"`
	// Optional proposed network policy amendments (allow/deny host) for future requests.
	ProposedNetworkPolicyAmendments []ProposedNetworkPolicyAmendmentElement `json:"proposedNetworkPolicyAmendments"`
	// Optional explanatory reason (e.g. request for network access).
	Reason *string `json:"reason"`
	// Unix timestamp (in milliseconds) when this approval request started.
	StartedAtMS int64  `json:"startedAtMs"`
	ThreadID    string `json:"threadId"`
	TurnID      string `json:"turnId"`
}

type AdditionalPermissionProfile struct {
	FileSystem *AdditionalPermissionProfileAdditionalFileSystemPermissions `json:"fileSystem"`
	// Partial overlay used for per-command permission requests.
	Network *AdditionalPermissionProfileAdditionalNetworkPermissions `json:"network"`
}

type AdditionalPermissionProfileAdditionalFileSystemPermissions struct {
	Entries          []PurpleFileSystemSandboxEntry `json:"entries"`
	GlobScanMaxDepth *int64                         `json:"globScanMaxDepth"`
	// This will be removed in favor of `entries`.
	Read []string `json:"read"`
	// This will be removed in favor of `entries`.
	Write []string `json:"write"`
}

type PurpleFileSystemSandboxEntry struct {
	Access FileSystemAccessMode `json:"access"`
	Path   PurpleFileSystemPath `json:"path"`
}

type PurpleFileSystemPath struct {
	Path    *string                      `json:"path,omitempty"`
	Type    FileSystemPathType           `json:"type"`
	Pattern *string                      `json:"pattern,omitempty"`
	Value   *PurpleFileSystemSpecialPath `json:"value,omitempty"`
}

type PurpleFileSystemSpecialPath struct {
	Kind    Kind    `json:"kind"`
	Subpath *string `json:"subpath"`
	Path    *string `json:"path,omitempty"`
}

type AdditionalPermissionProfileAdditionalNetworkPermissions struct {
	Enabled *bool `json:"enabled"`
}

// User approved the command, and wants to apply the proposed execpolicy amendment so future
// matching commands can run without prompting.
//
// User chose a persistent network policy rule (allow/deny) for this host.
type PurplePolicyAmendmentCommandExecutionApprovalDecision struct {
	AcceptWithExecpolicyAmendment *PurpleAcceptWithExecpolicyAmendment `json:"acceptWithExecpolicyAmendment,omitempty"`
	ApplyNetworkPolicyAmendment   *PurpleApplyNetworkPolicyAmendment   `json:"applyNetworkPolicyAmendment,omitempty"`
}

type PurpleAcceptWithExecpolicyAmendment struct {
	ExecpolicyAmendment []string `json:"execpolicy_amendment"`
}

type PurpleApplyNetworkPolicyAmendment struct {
	NetworkPolicyAmendment ProposedNetworkPolicyAmendmentElement `json:"network_policy_amendment"`
}

type ProposedNetworkPolicyAmendmentElement struct {
	Action NetworkPolicyRuleAction `json:"action"`
	Host   string                  `json:"host"`
}

type CommandExecutionRequestApprovalParamsCommandAction struct {
	Command string            `json:"command"`
	Name    *string           `json:"name,omitempty"`
	Path    *string           `json:"path"`
	Type    CommandActionType `json:"type"`
	Query   *string           `json:"query"`
}

type NetworkApprovalContext struct {
	Host     string                  `json:"host"`
	Protocol NetworkApprovalProtocol `json:"protocol"`
}

type CommandExecutionRequestApprovalResponse struct {
	Decision *CommandExecutionRequestApprovalResponseCommandExecutionApprovalDecision `json:"decision"`
}

// User approved the command, and wants to apply the proposed execpolicy amendment so future
// matching commands can run without prompting.
//
// User chose a persistent network policy rule (allow/deny) for this host.
type FluffyPolicyAmendmentCommandExecutionApprovalDecision struct {
	AcceptWithExecpolicyAmendment *FluffyAcceptWithExecpolicyAmendment `json:"acceptWithExecpolicyAmendment,omitempty"`
	ApplyNetworkPolicyAmendment   *FluffyApplyNetworkPolicyAmendment   `json:"applyNetworkPolicyAmendment,omitempty"`
}

type FluffyAcceptWithExecpolicyAmendment struct {
	ExecpolicyAmendment []string `json:"execpolicy_amendment"`
}

type FluffyApplyNetworkPolicyAmendment struct {
	NetworkPolicyAmendment PurpleNetworkPolicyAmendment `json:"network_policy_amendment"`
}

type PurpleNetworkPolicyAmendment struct {
	Action NetworkPolicyRuleAction `json:"action"`
	Host   string                  `json:"host"`
}

type FileChangeRequestApprovalParams struct {
	// [UNSTABLE] When set, the agent is asking the user to allow writes under this root for the
	// remainder of the session (unclear if this is honored today).
	GrantRoot *string `json:"grantRoot"`
	ItemID    string  `json:"itemId"`
	// Optional explanatory reason (e.g. request for extra write access).
	Reason *string `json:"reason"`
	// Unix timestamp (in milliseconds) when this approval request started.
	StartedAtMS int64  `json:"startedAtMs"`
	ThreadID    string `json:"threadId"`
	TurnID      string `json:"turnId"`
}

type FileChangeRequestApprovalResponse struct {
	Decision FileChangeApprovalDecision `json:"decision"`
}

type PermissionsRequestApprovalParams struct {
	Cwd           string                   `json:"cwd"`
	EnvironmentID *string                  `json:"environmentId"`
	ItemID        string                   `json:"itemId"`
	Permissions   RequestPermissionProfile `json:"permissions"`
	Reason        *string                  `json:"reason"`
	// Unix timestamp (in milliseconds) when this approval request started.
	StartedAtMS int64  `json:"startedAtMs"`
	ThreadID    string `json:"threadId"`
	TurnID      string `json:"turnId"`
}

type RequestPermissionProfile struct {
	FileSystem *PurpleAdditionalFileSystemPermissions `json:"fileSystem"`
	Network    *PurpleAdditionalNetworkPermissions    `json:"network"`
}

type PurpleAdditionalFileSystemPermissions struct {
	Entries          []FluffyFileSystemSandboxEntry `json:"entries"`
	GlobScanMaxDepth *int64                         `json:"globScanMaxDepth"`
	// This will be removed in favor of `entries`.
	Read []string `json:"read"`
	// This will be removed in favor of `entries`.
	Write []string `json:"write"`
}

type FluffyFileSystemSandboxEntry struct {
	Access FileSystemAccessMode `json:"access"`
	Path   FluffyFileSystemPath `json:"path"`
}

type FluffyFileSystemPath struct {
	Path    *string                      `json:"path,omitempty"`
	Type    FileSystemPathType           `json:"type"`
	Pattern *string                      `json:"pattern,omitempty"`
	Value   *FluffyFileSystemSpecialPath `json:"value,omitempty"`
}

type FluffyFileSystemSpecialPath struct {
	Kind    Kind    `json:"kind"`
	Subpath *string `json:"subpath"`
	Path    *string `json:"path,omitempty"`
}

type PurpleAdditionalNetworkPermissions struct {
	Enabled *bool `json:"enabled"`
}

type PermissionsRequestApprovalResponse struct {
	Permissions GrantedPermissionProfile `json:"permissions"`
	Scope       *PermissionGrantScope    `json:"scope,omitempty"`
	// Review every subsequent command in this turn before normal sandboxed execution.
	StrictAutoReview *bool `json:"strictAutoReview"`
}

type GrantedPermissionProfile struct {
	FileSystem *FluffyAdditionalFileSystemPermissions `json:"fileSystem"`
	Network    *FluffyAdditionalNetworkPermissions    `json:"network"`
}

type FluffyAdditionalFileSystemPermissions struct {
	Entries          []TentacledFileSystemSandboxEntry `json:"entries"`
	GlobScanMaxDepth *int64                            `json:"globScanMaxDepth"`
	// This will be removed in favor of `entries`.
	Read []string `json:"read"`
	// This will be removed in favor of `entries`.
	Write []string `json:"write"`
}

type TentacledFileSystemSandboxEntry struct {
	Access FileSystemAccessMode    `json:"access"`
	Path   TentacledFileSystemPath `json:"path"`
}

type TentacledFileSystemPath struct {
	Path    *string                         `json:"path,omitempty"`
	Type    FileSystemPathType              `json:"type"`
	Pattern *string                         `json:"pattern,omitempty"`
	Value   *TentacledFileSystemSpecialPath `json:"value,omitempty"`
}

type TentacledFileSystemSpecialPath struct {
	Kind    Kind    `json:"kind"`
	Subpath *string `json:"subpath"`
	Path    *string `json:"path,omitempty"`
}

type FluffyAdditionalNetworkPermissions struct {
	Enabled *bool `json:"enabled"`
}

// EXPERIMENTAL. Params sent with a request_user_input event.
type ToolRequestUserInputParams struct {
	AutoResolutionMS *int64                         `json:"autoResolutionMs"`
	ItemID           string                         `json:"itemId"`
	Questions        []ToolRequestUserInputQuestion `json:"questions"`
	ThreadID         string                         `json:"threadId"`
	TurnID           string                         `json:"turnId"`
}

// EXPERIMENTAL. Represents one request_user_input question and its required options.
type ToolRequestUserInputQuestion struct {
	Header   string                       `json:"header"`
	ID       string                       `json:"id"`
	IsOther  *bool                        `json:"isOther,omitempty"`
	IsSecret *bool                        `json:"isSecret,omitempty"`
	Options  []ToolRequestUserInputOption `json:"options"`
	Question string                       `json:"question"`
}

// EXPERIMENTAL. Defines a single selectable option for request_user_input.
type ToolRequestUserInputOption struct {
	Description string `json:"description"`
	Label       string `json:"label"`
}

// EXPERIMENTAL. Response payload mapping question ids to answers.
type ToolRequestUserInputResponse struct {
	Answers map[string]ToolRequestUserInputAnswer `json:"answers"`
}

// EXPERIMENTAL. Captures a user's answer to a request_user_input question.
type ToolRequestUserInputAnswer struct {
	Answers []string `json:"answers"`
}

type MCPServerElicitationRequestParams struct {
	ServerName string `json:"serverName"`
	ThreadID   string `json:"threadId"`
	// Active Codex turn when this elicitation was observed, if app-server could correlate one.
	//
	// This is nullable because MCP models elicitation as a standalone server-to-client request
	// identified by the MCP server request id. It may be triggered during a turn, but turn
	// context is app-server correlation rather than part of the protocol identity of the
	// elicitation itself.
	TurnID          *string     `json:"turnId"`
	Meta            interface{} `json:"_meta"`
	Message         string      `json:"message"`
	Mode            Mode        `json:"mode"`
	RequestedSchema interface{} `json:"requestedSchema"`
	ElicitationID   *string     `json:"elicitationId,omitempty"`
	URL             *string     `json:"url,omitempty"`
}

type MCPServerElicitationRequestResponse struct {
	// Optional client metadata for form-mode action handling.
	Meta   interface{}                `json:"_meta"`
	Action MCPServerElicitationAction `json:"action"`
	// Structured user input for accepted elicitations, mirroring RMCP
	// `CreateElicitationResult`.
	//
	// This is nullable because decline/cancel responses have no content.
	Content interface{} `json:"content"`
}

// Canonical user-input modality tags advertised by a model.
//
// Plain text turns and tool payloads.
//
// Image attachments included in user turns.
//
// Audio attachments included in user turns.
type InputModality string

const (
	InputModalityAudio InputModality = "audio"
	InputModalityImage InputModality = "image"
	InputModalityText  InputModality = "text"
)

type ApprovalPolicyEnum string

const (
	AskForApprovalUntrusted ApprovalPolicyEnum = "untrusted"
	Never                   ApprovalPolicyEnum = "never"
	OnRequest               ApprovalPolicyEnum = "on-request"
)

// Configures who approval requests are routed to for review. Examples include sandbox
// escapes, blocked network access, MCP approval prompts, and ARC escalations. Defaults to
// `user`. `auto_review` uses a carefully prompted subagent to gather relevant context and
// apply a risk-based decision framework before approving or denying the request. The legacy
// value `guardian_subagent` is accepted for compatibility.
//
// Reviewer currently used for approval requests on this thread.
type ApprovalsReviewer string

const (
	AutoReview       ApprovalsReviewer = "auto_review"
	GuardianSubagent ApprovalsReviewer = "guardian_subagent"
	User             ApprovalsReviewer = "user"
)

type FunctionDynamicToolNamespaceToolType string

const (
	FunctionDynamicToolNamespaceToolTypeFunction FunctionDynamicToolNamespaceToolType = "function"
)

type DynamicToolSpecType string

const (
	DynamicToolSpecTypeFunction DynamicToolSpecType = "function"
	Namespace                   DynamicToolSpecType = "namespace"
)

// Persisted thread history contract selected when this thread was created.
type ThreadHistoryMode string

const (
	Legacy    ThreadHistoryMode = "legacy"
	Paginated ThreadHistoryMode = "paginated"
)

type MultiAgentModeEnum string

const (
	ExplicitRequestOnly MultiAgentModeEnum = "explicitRequestOnly"
	Proactive           MultiAgentModeEnum = "proactive"
)

type Personality string

const (
	Friendly        Personality = "friendly"
	PersonalityNone Personality = "none"
	Pragmatic       Personality = "pragmatic"
)

type SandboxMode string

const (
	DangerFullAccess SandboxMode = "danger-full-access"
	ReadOnly         SandboxMode = "read-only"
	WorkspaceWrite   SandboxMode = "workspace-write"
)

type EnvironmentCapabilityRootLocationType string

const (
	Environment EnvironmentCapabilityRootLocationType = "environment"
)

type ThreadStartSource string

const (
	Clear   ThreadStartSource = "clear"
	Startup ThreadStartSource = "startup"
)

type NetworkAccess string

const (
	Enabled    NetworkAccess = "enabled"
	Restricted NetworkAccess = "restricted"
)

type SandboxPolicyType string

const (
	ExternalSandbox                   SandboxPolicyType = "externalSandbox"
	SandboxPolicyTypeDangerFullAccess SandboxPolicyType = "dangerFullAccess"
	SandboxPolicyTypeReadOnly         SandboxPolicyType = "readOnly"
	SandboxPolicyTypeWorkspaceWrite   SandboxPolicyType = "workspaceWrite"
)

type SubAgentSource string

const (
	MemoryConsolidation   SubAgentSource = "memory_consolidation"
	SubAgentSourceCompact SubAgentSource = "compact"
	SubAgentSourceReview  SubAgentSource = "review"
)

type SessionSource string

const (
	AppServer            SessionSource = "appServer"
	CLI                  SessionSource = "cli"
	SessionSourceExec    SessionSource = "exec"
	SessionSourceUnknown SessionSource = "unknown"
	Vscode               SessionSource = "vscode"
)

type ThreadActiveFlag string

const (
	WaitingOnApproval  ThreadActiveFlag = "waitingOnApproval"
	WaitingOnUserInput ThreadActiveFlag = "waitingOnUserInput"
)

type ThreadStatusType string

const (
	Active                    ThreadStatusType = "active"
	Idle                      ThreadStatusType = "idle"
	SystemError               ThreadStatusType = "systemError"
	ThreadStatusTypeNotLoaded ThreadStatusType = "notLoaded"
)

type NonSteerableTurnKind string

const (
	NonSteerableTurnKindCompact NonSteerableTurnKind = "compact"
	NonSteerableTurnKindReview  NonSteerableTurnKind = "review"
)

type CodexErrorInfoEnum string

const (
	BadRequest            CodexErrorInfoEnum = "badRequest"
	CodexErrorInfoOther   CodexErrorInfoEnum = "other"
	ContextWindowExceeded CodexErrorInfoEnum = "contextWindowExceeded"
	CyberPolicy           CodexErrorInfoEnum = "cyberPolicy"
	InternalServerError   CodexErrorInfoEnum = "internalServerError"
	SandboxError          CodexErrorInfoEnum = "sandboxError"
	ServerOverloaded      CodexErrorInfoEnum = "serverOverloaded"
	SessionBudgetExceeded CodexErrorInfoEnum = "sessionBudgetExceeded"
	ThreadRollbackFailed  CodexErrorInfoEnum = "threadRollbackFailed"
	Unauthorized          CodexErrorInfoEnum = "unauthorized"
	UsageLimitExceeded    CodexErrorInfoEnum = "usageLimitExceeded"
)

type WebSearchActionType string

const (
	FindInPage                WebSearchActionType = "findInPage"
	OpenPage                  WebSearchActionType = "openPage"
	WebSearchActionTypeOther  WebSearchActionType = "other"
	WebSearchActionTypeSearch WebSearchActionType = "search"
)

type CollabAgentStatus string

const (
	CollabAgentStatusCompleted   CollabAgentStatus = "completed"
	CollabAgentStatusInterrupted CollabAgentStatus = "interrupted"
	Errored                      CollabAgentStatus = "errored"
	NotFound                     CollabAgentStatus = "notFound"
	PendingInit                  CollabAgentStatus = "pendingInit"
	Running                      CollabAgentStatus = "running"
	Shutdown                     CollabAgentStatus = "shutdown"
)

type PatchChangeKindType string

const (
	Add    PatchChangeKindType = "add"
	Delete PatchChangeKindType = "delete"
	Update PatchChangeKindType = "update"
)

type CommandActionType string

const (
	CommandActionTypeRead    CommandActionType = "read"
	CommandActionTypeSearch  CommandActionType = "search"
	CommandActionTypeUnknown CommandActionType = "unknown"
	ListFiles                CommandActionType = "listFiles"
)

type ImageDetail string

const (
	High            ImageDetail = "high"
	ImageDetailAuto ImageDetail = "auto"
	Low             ImageDetail = "low"
	Original        ImageDetail = "original"
)

type UserInputType string

const (
	LocalAudio         UserInputType = "localAudio"
	LocalImage         UserInputType = "localImage"
	Mention            UserInputType = "mention"
	Skill              UserInputType = "skill"
	UserInputTypeAudio UserInputType = "audio"
	UserInputTypeImage UserInputType = "image"
	UserInputTypeText  UserInputType = "text"
)

type InputDynamicToolCallOutputContentItemType string

const (
	InputAudio InputDynamicToolCallOutputContentItemType = "inputAudio"
	InputImage InputDynamicToolCallOutputContentItemType = "inputImage"
	InputText  InputDynamicToolCallOutputContentItemType = "inputText"
)

type SubAgentActivityKind string

const (
	Interacted                      SubAgentActivityKind = "interacted"
	Started                         SubAgentActivityKind = "started"
	SubAgentActivityKindInterrupted SubAgentActivityKind = "interrupted"
)

// Mid-turn assistant text (for example preamble/progress narration).
//
// Additional tool calls or assistant output may follow before turn completion.
//
// The assistant's terminal answer text for the current turn.
type MessagePhase string

const (
	Commentary  MessagePhase = "commentary"
	FinalAnswer MessagePhase = "final_answer"
)

type CommandExecutionSource string

const (
	Agent                  CommandExecutionSource = "agent"
	UnifiedExecInteraction CommandExecutionSource = "unifiedExecInteraction"
	UnifiedExecStartup     CommandExecutionSource = "unifiedExecStartup"
	UserShell              CommandExecutionSource = "userShell"
)

type ThreadItemType string

const (
	AgentMessage            ThreadItemType = "agentMessage"
	CollabAgentToolCall     ThreadItemType = "collabAgentToolCall"
	CommandExecution        ThreadItemType = "commandExecution"
	ContextCompaction       ThreadItemType = "contextCompaction"
	DynamicToolCall         ThreadItemType = "dynamicToolCall"
	EnteredReviewMode       ThreadItemType = "enteredReviewMode"
	ExitedReviewMode        ThreadItemType = "exitedReviewMode"
	FileChange              ThreadItemType = "fileChange"
	HookPrompt              ThreadItemType = "hookPrompt"
	ImageGeneration         ThreadItemType = "imageGeneration"
	ImageView               ThreadItemType = "imageView"
	MCPToolCall             ThreadItemType = "mcpToolCall"
	Sleep                   ThreadItemType = "sleep"
	SubAgentActivity        ThreadItemType = "subAgentActivity"
	ThreadItemTypePlan      ThreadItemType = "plan"
	ThreadItemTypeReasoning ThreadItemType = "reasoning"
	UserMessage             ThreadItemType = "userMessage"
	WebSearch               ThreadItemType = "webSearch"
)

// Describes how much of `items` has been loaded for this turn.
//
// `items` was not loaded for this turn. The field is intentionally empty.
//
// `items` contains only a display summary for this turn.
//
// `items` contains every ThreadItem available from persisted app-server history for this
// turn.
type TurnItemsView string

const (
	Full                   TurnItemsView = "full"
	Summary                TurnItemsView = "summary"
	TurnItemsViewNotLoaded TurnItemsView = "notLoaded"
)

type TurnStatus string

const (
	Failed                TurnStatus = "failed"
	TurnStatusCompleted   TurnStatus = "completed"
	TurnStatusInProgress  TurnStatus = "inProgress"
	TurnStatusInterrupted TurnStatus = "interrupted"
)

type ActionType string

const (
	ActionTypeExec       ActionType = "exec"
	ActionTypeFindInPage ActionType = "find_in_page"
	ActionTypeOpenPage   ActionType = "open_page"
	ActionTypeOther      ActionType = "other"
	ActionTypeSearch     ActionType = "search"
)

type Type string

const (
	OutputText           Type = "output_text"
	ReasoningText        Type = "reasoning_text"
	TypeEncryptedContent Type = "encrypted_content"
	TypeInputAudio       Type = "input_audio"
	TypeInputImage       Type = "input_image"
	TypeInputText        Type = "input_text"
	TypeText             Type = "text"
)

type FunctionCallOutputContentItemType string

const (
	FunctionCallOutputContentItemTypeEncryptedContent FunctionCallOutputContentItemType = "encrypted_content"
	FunctionCallOutputContentItemTypeInputAudio       FunctionCallOutputContentItemType = "input_audio"
	FunctionCallOutputContentItemTypeInputImage       FunctionCallOutputContentItemType = "input_image"
	FunctionCallOutputContentItemTypeInputText        FunctionCallOutputContentItemType = "input_text"
)

type SummaryTextReasoningItemReasoningSummaryType string

const (
	SummaryText SummaryTextReasoningItemReasoningSummaryType = "summary_text"
)

type ResponseItemType string

const (
	Compaction                        ResponseItemType = "compaction"
	CompactionTrigger                 ResponseItemType = "compaction_trigger"
	CustomToolCall                    ResponseItemType = "custom_tool_call"
	CustomToolCallOutput              ResponseItemType = "custom_tool_call_output"
	FunctionCall                      ResponseItemType = "function_call"
	FunctionCallOutput                ResponseItemType = "function_call_output"
	ImageGenerationCall               ResponseItemType = "image_generation_call"
	LocalShellCall                    ResponseItemType = "local_shell_call"
	Message                           ResponseItemType = "message"
	ResponseItemTypeAgentMessage      ResponseItemType = "agent_message"
	ResponseItemTypeContextCompaction ResponseItemType = "context_compaction"
	ResponseItemTypeOther             ResponseItemType = "other"
	ResponseItemTypeReasoning         ResponseItemType = "reasoning"
	ToolSearchCall                    ResponseItemType = "tool_search_call"
	ToolSearchOutput                  ResponseItemType = "tool_search_output"
	WebSearchCall                     ResponseItemType = "web_search_call"
)

type SortDirection string

const (
	Asc  SortDirection = "asc"
	Desc SortDirection = "desc"
)

type AdditionalContextKind string

const (
	AdditionalContextKindUntrusted AdditionalContextKind = "untrusted"
	Application                    AdditionalContextKind = "application"
)

// Initial collaboration mode to use when the TUI starts.
type ModeKind string

const (
	Default      ModeKind = "default"
	ModeKindPlan ModeKind = "plan"
)

// Option to disable reasoning summaries.
type ReasoningSummary string

const (
	Concise              ReasoningSummary = "concise"
	Detailed             ReasoningSummary = "detailed"
	ReasoningSummaryAuto ReasoningSummary = "auto"
	ReasoningSummaryNone ReasoningSummary = "none"
)

type TurnPlanStepStatus string

const (
	Pending                      TurnPlanStepStatus = "pending"
	TurnPlanStepStatusCompleted  TurnPlanStepStatus = "completed"
	TurnPlanStepStatusInProgress TurnPlanStepStatus = "inProgress"
)

type FileSystemAccessMode string

const (
	FileSystemAccessModeDeny FileSystemAccessMode = "deny"
	FileSystemAccessModeRead FileSystemAccessMode = "read"
	Write                    FileSystemAccessMode = "write"
)

type FileSystemPathType string

const (
	GlobPattern FileSystemPathType = "glob_pattern"
	Path        FileSystemPathType = "path"
	Special     FileSystemPathType = "special"
)

type Kind string

const (
	KindUnknown  Kind = "unknown"
	Minimal      Kind = "minimal"
	ProjectRoots Kind = "project_roots"
	Root         Kind = "root"
	SlashTmp     Kind = "slash_tmp"
	Tmpdir       Kind = "tmpdir"
)

type NetworkPolicyRuleAction string

const (
	Allow                       NetworkPolicyRuleAction = "allow"
	NetworkPolicyRuleActionDeny NetworkPolicyRuleAction = "deny"
)

// User approved the command.
//
// User approved the command and future prompts in the same session-scoped approval cache
// should run without prompting.
//
// User denied the command. The agent will continue the turn.
//
// User denied the command. The turn will also be immediately interrupted.
//
// User approved the file changes.
//
// User approved the file changes and future changes to the same files should run without
// prompting.
//
// User denied the file changes. The agent will continue the turn.
//
// User denied the file changes. The turn will also be immediately interrupted.
type FileChangeApprovalDecision string

const (
	AcceptForSession                  FileChangeApprovalDecision = "acceptForSession"
	FileChangeApprovalDecisionAccept  FileChangeApprovalDecision = "accept"
	FileChangeApprovalDecisionCancel  FileChangeApprovalDecision = "cancel"
	FileChangeApprovalDecisionDecline FileChangeApprovalDecision = "decline"
)

type NetworkApprovalProtocol string

const (
	HTTP      NetworkApprovalProtocol = "http"
	HTTPS     NetworkApprovalProtocol = "https"
	Socks5TCP NetworkApprovalProtocol = "socks5Tcp"
	Socks5UDP NetworkApprovalProtocol = "socks5Udp"
)

type PermissionGrantScope string

const (
	Session PermissionGrantScope = "session"
	Turn    PermissionGrantScope = "turn"
)

type Mode string

const (
	Form       Mode = "form"
	OpenaiForm Mode = "openai/form"
	URL        Mode = "url"
)

type MCPServerElicitationAction string

const (
	MCPServerElicitationActionAccept  MCPServerElicitationAction = "accept"
	MCPServerElicitationActionCancel  MCPServerElicitationAction = "cancel"
	MCPServerElicitationActionDecline MCPServerElicitationAction = "decline"
)

type ThreadStartParamsApprovalPolicy struct {
	Enum                         *ApprovalPolicyEnum
	PurpleGranularAskForApproval *PurpleGranularAskForApproval
}

func (x *ThreadStartParamsApprovalPolicy) UnmarshalJSON(data []byte) error {
	x.PurpleGranularAskForApproval = nil
	x.Enum = nil
	var c PurpleGranularAskForApproval
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.PurpleGranularAskForApproval = &c
	}
	return nil
}

func (x *ThreadStartParamsApprovalPolicy) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.PurpleGranularAskForApproval != nil, x.PurpleGranularAskForApproval, false, nil, x.Enum != nil, x.Enum, true)
}

// @deprecated Ignored. Use Ultra reasoning effort for proactive multi-agent behavior.
type ThreadStartParamsMultiAgentMode struct {
	Enum                       *MultiAgentModeEnum
	PurpleCustomMultiAgentMode *PurpleCustomMultiAgentMode
}

func (x *ThreadStartParamsMultiAgentMode) UnmarshalJSON(data []byte) error {
	x.PurpleCustomMultiAgentMode = nil
	x.Enum = nil
	var c PurpleCustomMultiAgentMode
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.PurpleCustomMultiAgentMode = &c
	}
	return nil
}

func (x *ThreadStartParamsMultiAgentMode) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.PurpleCustomMultiAgentMode != nil, x.PurpleCustomMultiAgentMode, false, nil, x.Enum != nil, x.Enum, true)
}

type ThreadStartResponseAskForApproval struct {
	Enum                         *ApprovalPolicyEnum
	FluffyGranularAskForApproval *FluffyGranularAskForApproval
}

func (x *ThreadStartResponseAskForApproval) UnmarshalJSON(data []byte) error {
	x.FluffyGranularAskForApproval = nil
	x.Enum = nil
	var c FluffyGranularAskForApproval
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.FluffyGranularAskForApproval = &c
	}
	return nil
}

func (x *ThreadStartResponseAskForApproval) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.FluffyGranularAskForApproval != nil, x.FluffyGranularAskForApproval, false, nil, x.Enum != nil, x.Enum, false)
}

// @deprecated Always `explicitRequestOnly`. Use `reasoningEffort` for Ultra behavior.
//
// Controls the effective multi-agent delegation instructions for a turn. `custom` means the
// configured mode hint defines the policy instead of a built-in policy.
type ThreadStartResponseMultiAgentMode struct {
	Enum                       *MultiAgentModeEnum
	FluffyCustomMultiAgentMode *FluffyCustomMultiAgentMode
}

func (x *ThreadStartResponseMultiAgentMode) UnmarshalJSON(data []byte) error {
	x.FluffyCustomMultiAgentMode = nil
	x.Enum = nil
	var c FluffyCustomMultiAgentMode
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.FluffyCustomMultiAgentMode = &c
	}
	return nil
}

func (x *ThreadStartResponseMultiAgentMode) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.FluffyCustomMultiAgentMode != nil, x.FluffyCustomMultiAgentMode, false, nil, x.Enum != nil, x.Enum, false)
}

type NetworkAccessUnion struct {
	Bool *bool
	Enum *NetworkAccess
}

func (x *NetworkAccessUnion) UnmarshalJSON(data []byte) error {
	x.Enum = nil
	object, err := unmarshalUnion(data, nil, nil, &x.Bool, nil, false, nil, false, nil, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

func (x *NetworkAccessUnion) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, x.Bool, nil, false, nil, false, nil, false, nil, x.Enum != nil, x.Enum, false)
}

// Origin of the thread (CLI, VSCode, codex exec, codex app-server, etc.).
type StickySessionSource struct {
	Enum                *SessionSource
	PurpleSessionSource *PurpleSessionSource
}

func (x *StickySessionSource) UnmarshalJSON(data []byte) error {
	x.PurpleSessionSource = nil
	x.Enum = nil
	var c PurpleSessionSource
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.PurpleSessionSource = &c
	}
	return nil
}

func (x *StickySessionSource) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.PurpleSessionSource != nil, x.PurpleSessionSource, false, nil, x.Enum != nil, x.Enum, false)
}

type StickySubAgentSource struct {
	Enum                 *SubAgentSource
	PurpleSubAgentSource *PurpleSubAgentSource
}

func (x *StickySubAgentSource) UnmarshalJSON(data []byte) error {
	x.PurpleSubAgentSource = nil
	x.Enum = nil
	var c PurpleSubAgentSource
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.PurpleSubAgentSource = &c
	}
	return nil
}

func (x *StickySubAgentSource) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.PurpleSubAgentSource != nil, x.PurpleSubAgentSource, false, nil, x.Enum != nil, x.Enum, false)
}

type AmbitiousCodexErrorInfo struct {
	Enum                 *CodexErrorInfoEnum
	PurpleCodexErrorInfo *PurpleCodexErrorInfo
}

func (x *AmbitiousCodexErrorInfo) UnmarshalJSON(data []byte) error {
	x.PurpleCodexErrorInfo = nil
	x.Enum = nil
	var c PurpleCodexErrorInfo
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.PurpleCodexErrorInfo = &c
	}
	return nil
}

func (x *AmbitiousCodexErrorInfo) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.PurpleCodexErrorInfo != nil, x.PurpleCodexErrorInfo, false, nil, x.Enum != nil, x.Enum, true)
}

type CunningUserInput struct {
	PurpleUserInput *PurpleUserInput
	String          *string
}

func (x *CunningUserInput) UnmarshalJSON(data []byte) error {
	x.PurpleUserInput = nil
	var c PurpleUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.PurpleUserInput = &c
	}
	return nil
}

func (x *CunningUserInput) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.PurpleUserInput != nil, x.PurpleUserInput, false, nil, false, nil, false)
}

type PurpleResult struct {
	PurpleMCPToolCallResult *PurpleMCPToolCallResult
	String                  *string
}

func (x *PurpleResult) UnmarshalJSON(data []byte) error {
	x.PurpleMCPToolCallResult = nil
	var c PurpleMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.PurpleMCPToolCallResult = &c
	}
	return nil
}

func (x *PurpleResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.PurpleMCPToolCallResult != nil, x.PurpleMCPToolCallResult, false, nil, false, nil, true)
}

type ThreadResumeParamsApprovalPolicy struct {
	Enum                            *ApprovalPolicyEnum
	TentacledGranularAskForApproval *TentacledGranularAskForApproval
}

func (x *ThreadResumeParamsApprovalPolicy) UnmarshalJSON(data []byte) error {
	x.TentacledGranularAskForApproval = nil
	x.Enum = nil
	var c TentacledGranularAskForApproval
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.TentacledGranularAskForApproval = &c
	}
	return nil
}

func (x *ThreadResumeParamsApprovalPolicy) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.TentacledGranularAskForApproval != nil, x.TentacledGranularAskForApproval, false, nil, x.Enum != nil, x.Enum, true)
}

type FunctionCallOutputBody struct {
	FunctionCallOutputContentItemArray []FunctionCallOutputContentItem
	String                             *string
}

func (x *FunctionCallOutputBody) UnmarshalJSON(data []byte) error {
	x.FunctionCallOutputContentItemArray = nil
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, true, &x.FunctionCallOutputContentItemArray, false, nil, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

func (x *FunctionCallOutputBody) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, x.FunctionCallOutputContentItemArray != nil, x.FunctionCallOutputContentItemArray, false, nil, false, nil, false, nil, false)
}

type ThreadResumeResponseAskForApproval struct {
	Enum                         *ApprovalPolicyEnum
	StickyGranularAskForApproval *StickyGranularAskForApproval
}

func (x *ThreadResumeResponseAskForApproval) UnmarshalJSON(data []byte) error {
	x.StickyGranularAskForApproval = nil
	x.Enum = nil
	var c StickyGranularAskForApproval
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.StickyGranularAskForApproval = &c
	}
	return nil
}

func (x *ThreadResumeResponseAskForApproval) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.StickyGranularAskForApproval != nil, x.StickyGranularAskForApproval, false, nil, x.Enum != nil, x.Enum, false)
}

type CunningCodexErrorInfo struct {
	Enum                 *CodexErrorInfoEnum
	FluffyCodexErrorInfo *FluffyCodexErrorInfo
}

func (x *CunningCodexErrorInfo) UnmarshalJSON(data []byte) error {
	x.FluffyCodexErrorInfo = nil
	x.Enum = nil
	var c FluffyCodexErrorInfo
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.FluffyCodexErrorInfo = &c
	}
	return nil
}

func (x *CunningCodexErrorInfo) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.FluffyCodexErrorInfo != nil, x.FluffyCodexErrorInfo, false, nil, x.Enum != nil, x.Enum, true)
}

type MagentaUserInput struct {
	FluffyUserInput *FluffyUserInput
	String          *string
}

func (x *MagentaUserInput) UnmarshalJSON(data []byte) error {
	x.FluffyUserInput = nil
	var c FluffyUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.FluffyUserInput = &c
	}
	return nil
}

func (x *MagentaUserInput) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.FluffyUserInput != nil, x.FluffyUserInput, false, nil, false, nil, false)
}

type FluffyResult struct {
	FluffyMCPToolCallResult *FluffyMCPToolCallResult
	String                  *string
}

func (x *FluffyResult) UnmarshalJSON(data []byte) error {
	x.FluffyMCPToolCallResult = nil
	var c FluffyMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.FluffyMCPToolCallResult = &c
	}
	return nil
}

func (x *FluffyResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.FluffyMCPToolCallResult != nil, x.FluffyMCPToolCallResult, false, nil, false, nil, true)
}

// @deprecated Always `explicitRequestOnly`. Use `reasoningEffort` for Ultra behavior.
//
// Controls the effective multi-agent delegation instructions for a turn. `custom` means the
// configured mode hint defines the policy instead of a built-in policy.
type ThreadResumeResponseMultiAgentMode struct {
	Enum                          *MultiAgentModeEnum
	TentacledCustomMultiAgentMode *TentacledCustomMultiAgentMode
}

func (x *ThreadResumeResponseMultiAgentMode) UnmarshalJSON(data []byte) error {
	x.TentacledCustomMultiAgentMode = nil
	x.Enum = nil
	var c TentacledCustomMultiAgentMode
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.TentacledCustomMultiAgentMode = &c
	}
	return nil
}

func (x *ThreadResumeResponseMultiAgentMode) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.TentacledCustomMultiAgentMode != nil, x.TentacledCustomMultiAgentMode, false, nil, x.Enum != nil, x.Enum, false)
}

// Origin of the thread (CLI, VSCode, codex exec, codex app-server, etc.).
type IndigoSessionSource struct {
	Enum                *SessionSource
	FluffySessionSource *FluffySessionSource
}

func (x *IndigoSessionSource) UnmarshalJSON(data []byte) error {
	x.FluffySessionSource = nil
	x.Enum = nil
	var c FluffySessionSource
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.FluffySessionSource = &c
	}
	return nil
}

func (x *IndigoSessionSource) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.FluffySessionSource != nil, x.FluffySessionSource, false, nil, x.Enum != nil, x.Enum, false)
}

type IndigoSubAgentSource struct {
	Enum                 *SubAgentSource
	FluffySubAgentSource *FluffySubAgentSource
}

func (x *IndigoSubAgentSource) UnmarshalJSON(data []byte) error {
	x.FluffySubAgentSource = nil
	x.Enum = nil
	var c FluffySubAgentSource
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.FluffySubAgentSource = &c
	}
	return nil
}

func (x *IndigoSubAgentSource) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.FluffySubAgentSource != nil, x.FluffySubAgentSource, false, nil, x.Enum != nil, x.Enum, false)
}

// Origin of the thread (CLI, VSCode, codex exec, codex app-server, etc.).
type IndecentSessionSource struct {
	Enum                   *SessionSource
	TentacledSessionSource *TentacledSessionSource
}

func (x *IndecentSessionSource) UnmarshalJSON(data []byte) error {
	x.TentacledSessionSource = nil
	x.Enum = nil
	var c TentacledSessionSource
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.TentacledSessionSource = &c
	}
	return nil
}

func (x *IndecentSessionSource) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.TentacledSessionSource != nil, x.TentacledSessionSource, false, nil, x.Enum != nil, x.Enum, false)
}

type IndecentSubAgentSource struct {
	Enum                    *SubAgentSource
	TentacledSubAgentSource *TentacledSubAgentSource
}

func (x *IndecentSubAgentSource) UnmarshalJSON(data []byte) error {
	x.TentacledSubAgentSource = nil
	x.Enum = nil
	var c TentacledSubAgentSource
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.TentacledSubAgentSource = &c
	}
	return nil
}

func (x *IndecentSubAgentSource) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.TentacledSubAgentSource != nil, x.TentacledSubAgentSource, false, nil, x.Enum != nil, x.Enum, false)
}

type MagentaCodexErrorInfo struct {
	Enum                    *CodexErrorInfoEnum
	TentacledCodexErrorInfo *TentacledCodexErrorInfo
}

func (x *MagentaCodexErrorInfo) UnmarshalJSON(data []byte) error {
	x.TentacledCodexErrorInfo = nil
	x.Enum = nil
	var c TentacledCodexErrorInfo
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.TentacledCodexErrorInfo = &c
	}
	return nil
}

func (x *MagentaCodexErrorInfo) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.TentacledCodexErrorInfo != nil, x.TentacledCodexErrorInfo, false, nil, x.Enum != nil, x.Enum, true)
}

type FriskyUserInput struct {
	String             *string
	TentacledUserInput *TentacledUserInput
}

func (x *FriskyUserInput) UnmarshalJSON(data []byte) error {
	x.TentacledUserInput = nil
	var c TentacledUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.TentacledUserInput = &c
	}
	return nil
}

func (x *FriskyUserInput) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.TentacledUserInput != nil, x.TentacledUserInput, false, nil, false, nil, false)
}

type TentacledResult struct {
	String                     *string
	TentacledMCPToolCallResult *TentacledMCPToolCallResult
}

func (x *TentacledResult) UnmarshalJSON(data []byte) error {
	x.TentacledMCPToolCallResult = nil
	var c TentacledMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.TentacledMCPToolCallResult = &c
	}
	return nil
}

func (x *TentacledResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.TentacledMCPToolCallResult != nil, x.TentacledMCPToolCallResult, false, nil, false, nil, true)
}

// Override the approval policy for this turn and subsequent turns.
type TurnStartParamsApprovalPolicy struct {
	Enum                         *ApprovalPolicyEnum
	IndigoGranularAskForApproval *IndigoGranularAskForApproval
}

func (x *TurnStartParamsApprovalPolicy) UnmarshalJSON(data []byte) error {
	x.IndigoGranularAskForApproval = nil
	x.Enum = nil
	var c IndigoGranularAskForApproval
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.IndigoGranularAskForApproval = &c
	}
	return nil
}

func (x *TurnStartParamsApprovalPolicy) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.IndigoGranularAskForApproval != nil, x.IndigoGranularAskForApproval, false, nil, x.Enum != nil, x.Enum, true)
}

// @deprecated Ignored. Use `effort: "ultra"` for proactive multi-agent behavior.
type TurnStartParamsMultiAgentMode struct {
	Enum                       *MultiAgentModeEnum
	StickyCustomMultiAgentMode *StickyCustomMultiAgentMode
}

func (x *TurnStartParamsMultiAgentMode) UnmarshalJSON(data []byte) error {
	x.StickyCustomMultiAgentMode = nil
	x.Enum = nil
	var c StickyCustomMultiAgentMode
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.StickyCustomMultiAgentMode = &c
	}
	return nil
}

func (x *TurnStartParamsMultiAgentMode) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.StickyCustomMultiAgentMode != nil, x.StickyCustomMultiAgentMode, false, nil, x.Enum != nil, x.Enum, true)
}

type FriskyCodexErrorInfo struct {
	Enum                 *CodexErrorInfoEnum
	StickyCodexErrorInfo *StickyCodexErrorInfo
}

func (x *FriskyCodexErrorInfo) UnmarshalJSON(data []byte) error {
	x.StickyCodexErrorInfo = nil
	x.Enum = nil
	var c StickyCodexErrorInfo
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.StickyCodexErrorInfo = &c
	}
	return nil
}

func (x *FriskyCodexErrorInfo) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.StickyCodexErrorInfo != nil, x.StickyCodexErrorInfo, false, nil, x.Enum != nil, x.Enum, true)
}

type MischievousUserInput struct {
	StickyUserInput *StickyUserInput
	String          *string
}

func (x *MischievousUserInput) UnmarshalJSON(data []byte) error {
	x.StickyUserInput = nil
	var c StickyUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.StickyUserInput = &c
	}
	return nil
}

func (x *MischievousUserInput) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.StickyUserInput != nil, x.StickyUserInput, false, nil, false, nil, false)
}

type StickyResult struct {
	StickyMCPToolCallResult *StickyMCPToolCallResult
	String                  *string
}

func (x *StickyResult) UnmarshalJSON(data []byte) error {
	x.StickyMCPToolCallResult = nil
	var c StickyMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.StickyMCPToolCallResult = &c
	}
	return nil
}

func (x *StickyResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.StickyMCPToolCallResult != nil, x.StickyMCPToolCallResult, false, nil, false, nil, true)
}

type MischievousCodexErrorInfo struct {
	Enum                 *CodexErrorInfoEnum
	IndigoCodexErrorInfo *IndigoCodexErrorInfo
}

func (x *MischievousCodexErrorInfo) UnmarshalJSON(data []byte) error {
	x.IndigoCodexErrorInfo = nil
	x.Enum = nil
	var c IndigoCodexErrorInfo
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.IndigoCodexErrorInfo = &c
	}
	return nil
}

func (x *MischievousCodexErrorInfo) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.IndigoCodexErrorInfo != nil, x.IndigoCodexErrorInfo, false, nil, x.Enum != nil, x.Enum, true)
}

type BraggadociousUserInput struct {
	IndigoUserInput *IndigoUserInput
	String          *string
}

func (x *BraggadociousUserInput) UnmarshalJSON(data []byte) error {
	x.IndigoUserInput = nil
	var c IndigoUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.IndigoUserInput = &c
	}
	return nil
}

func (x *BraggadociousUserInput) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.IndigoUserInput != nil, x.IndigoUserInput, false, nil, false, nil, false)
}

type IndigoResult struct {
	IndigoMCPToolCallResult *IndigoMCPToolCallResult
	String                  *string
}

func (x *IndigoResult) UnmarshalJSON(data []byte) error {
	x.IndigoMCPToolCallResult = nil
	var c IndigoMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.IndigoMCPToolCallResult = &c
	}
	return nil
}

func (x *IndigoResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.IndigoMCPToolCallResult != nil, x.IndigoMCPToolCallResult, false, nil, false, nil, true)
}

type BraggadociousCodexErrorInfo struct {
	Enum                   *CodexErrorInfoEnum
	IndecentCodexErrorInfo *IndecentCodexErrorInfo
}

func (x *BraggadociousCodexErrorInfo) UnmarshalJSON(data []byte) error {
	x.IndecentCodexErrorInfo = nil
	x.Enum = nil
	var c IndecentCodexErrorInfo
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.IndecentCodexErrorInfo = &c
	}
	return nil
}

func (x *BraggadociousCodexErrorInfo) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.IndecentCodexErrorInfo != nil, x.IndecentCodexErrorInfo, false, nil, x.Enum != nil, x.Enum, true)
}

type UserInput1 struct {
	IndecentUserInput *IndecentUserInput
	String            *string
}

func (x *UserInput1) UnmarshalJSON(data []byte) error {
	x.IndecentUserInput = nil
	var c IndecentUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.IndecentUserInput = &c
	}
	return nil
}

func (x *UserInput1) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.IndecentUserInput != nil, x.IndecentUserInput, false, nil, false, nil, false)
}

type IndecentResult struct {
	IndecentMCPToolCallResult *IndecentMCPToolCallResult
	String                    *string
}

func (x *IndecentResult) UnmarshalJSON(data []byte) error {
	x.IndecentMCPToolCallResult = nil
	var c IndecentMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.IndecentMCPToolCallResult = &c
	}
	return nil
}

func (x *IndecentResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.IndecentMCPToolCallResult != nil, x.IndecentMCPToolCallResult, false, nil, false, nil, true)
}

type UserInput2 struct {
	HilariousUserInput *HilariousUserInput
	String             *string
}

func (x *UserInput2) UnmarshalJSON(data []byte) error {
	x.HilariousUserInput = nil
	var c HilariousUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.HilariousUserInput = &c
	}
	return nil
}

func (x *UserInput2) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.HilariousUserInput != nil, x.HilariousUserInput, false, nil, false, nil, false)
}

type HilariousResult struct {
	HilariousMCPToolCallResult *HilariousMCPToolCallResult
	String                     *string
}

func (x *HilariousResult) UnmarshalJSON(data []byte) error {
	x.HilariousMCPToolCallResult = nil
	var c HilariousMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.HilariousMCPToolCallResult = &c
	}
	return nil
}

func (x *HilariousResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.HilariousMCPToolCallResult != nil, x.HilariousMCPToolCallResult, false, nil, false, nil, true)
}

type UserInput3 struct {
	AmbitiousUserInput *AmbitiousUserInput
	String             *string
}

func (x *UserInput3) UnmarshalJSON(data []byte) error {
	x.AmbitiousUserInput = nil
	var c AmbitiousUserInput
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
		x.AmbitiousUserInput = &c
	}
	return nil
}

func (x *UserInput3) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.AmbitiousUserInput != nil, x.AmbitiousUserInput, false, nil, false, nil, false)
}

type AmbitiousResult struct {
	AmbitiousMCPToolCallResult *AmbitiousMCPToolCallResult
	String                     *string
}

func (x *AmbitiousResult) UnmarshalJSON(data []byte) error {
	x.AmbitiousMCPToolCallResult = nil
	var c AmbitiousMCPToolCallResult
	object, err := unmarshalUnion(data, nil, nil, nil, &x.String, false, nil, true, &c, false, nil, false, nil, true)
	if err != nil {
		return err
	}
	if object {
		x.AmbitiousMCPToolCallResult = &c
	}
	return nil
}

func (x *AmbitiousResult) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, x.String, false, nil, x.AmbitiousMCPToolCallResult != nil, x.AmbitiousMCPToolCallResult, false, nil, false, nil, true)
}

type ErrorCodexErrorInfo struct {
	Enum                    *CodexErrorInfoEnum
	HilariousCodexErrorInfo *HilariousCodexErrorInfo
}

func (x *ErrorCodexErrorInfo) UnmarshalJSON(data []byte) error {
	x.HilariousCodexErrorInfo = nil
	x.Enum = nil
	var c HilariousCodexErrorInfo
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, true)
	if err != nil {
		return err
	}
	if object {
		x.HilariousCodexErrorInfo = &c
	}
	return nil
}

func (x *ErrorCodexErrorInfo) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.HilariousCodexErrorInfo != nil, x.HilariousCodexErrorInfo, false, nil, x.Enum != nil, x.Enum, true)
}

type RequestID struct {
	Integer *int64
	String  *string
}

func (x *RequestID) UnmarshalJSON(data []byte) error {
	object, err := unmarshalUnion(data, &x.Integer, nil, nil, &x.String, false, nil, false, nil, false, nil, false, nil, false)
	if err != nil {
		return err
	}
	if object {
	}
	return nil
}

func (x *RequestID) MarshalJSON() ([]byte, error) {
	return marshalUnion(x.Integer, nil, nil, x.String, false, nil, false, nil, false, nil, false, nil, false)
}

type CommandExecutionApprovalDecisionElement struct {
	Enum                                                  *FileChangeApprovalDecision
	PurplePolicyAmendmentCommandExecutionApprovalDecision *PurplePolicyAmendmentCommandExecutionApprovalDecision
}

func (x *CommandExecutionApprovalDecisionElement) UnmarshalJSON(data []byte) error {
	x.PurplePolicyAmendmentCommandExecutionApprovalDecision = nil
	x.Enum = nil
	var c PurplePolicyAmendmentCommandExecutionApprovalDecision
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.PurplePolicyAmendmentCommandExecutionApprovalDecision = &c
	}
	return nil
}

func (x *CommandExecutionApprovalDecisionElement) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.PurplePolicyAmendmentCommandExecutionApprovalDecision != nil, x.PurplePolicyAmendmentCommandExecutionApprovalDecision, false, nil, x.Enum != nil, x.Enum, false)
}

type CommandExecutionRequestApprovalResponseCommandExecutionApprovalDecision struct {
	Enum                                                  *FileChangeApprovalDecision
	FluffyPolicyAmendmentCommandExecutionApprovalDecision *FluffyPolicyAmendmentCommandExecutionApprovalDecision
}

func (x *CommandExecutionRequestApprovalResponseCommandExecutionApprovalDecision) UnmarshalJSON(data []byte) error {
	x.FluffyPolicyAmendmentCommandExecutionApprovalDecision = nil
	x.Enum = nil
	var c FluffyPolicyAmendmentCommandExecutionApprovalDecision
	object, err := unmarshalUnion(data, nil, nil, nil, nil, false, nil, true, &c, false, nil, true, &x.Enum, false)
	if err != nil {
		return err
	}
	if object {
		x.FluffyPolicyAmendmentCommandExecutionApprovalDecision = &c
	}
	return nil
}

func (x *CommandExecutionRequestApprovalResponseCommandExecutionApprovalDecision) MarshalJSON() ([]byte, error) {
	return marshalUnion(nil, nil, nil, nil, false, nil, x.FluffyPolicyAmendmentCommandExecutionApprovalDecision != nil, x.FluffyPolicyAmendmentCommandExecutionApprovalDecision, false, nil, x.Enum != nil, x.Enum, false)
}

func unmarshalUnion(data []byte, pi **int64, pf **float64, pb **bool, ps **string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) (bool, error) {
	if pi != nil {
		*pi = nil
	}
	if pf != nil {
		*pf = nil
	}
	if pb != nil {
		*pb = nil
	}
	if ps != nil {
		*ps = nil
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return false, err
	}

	switch v := tok.(type) {
	case json.Number:
		if pi != nil {
			i, err := v.Int64()
			if err == nil {
				*pi = &i
				return false, nil
			}
		}
		if pf != nil {
			f, err := v.Float64()
			if err == nil {
				*pf = &f
				return false, nil
			}
			return false, errors.New("Unparsable number")
		}
		return false, errors.New("Union does not contain number")
	case float64:
		return false, errors.New("Decoder should not return float64")
	case bool:
		if pb != nil {
			*pb = &v
			return false, nil
		}
		return false, errors.New("Union does not contain bool")
	case string:
		if haveEnum {
			return false, json.Unmarshal(data, pe)
		}
		if ps != nil {
			*ps = &v
			return false, nil
		}
		return false, errors.New("Union does not contain string")
	case nil:
		if nullable {
			return false, nil
		}
		return false, errors.New("Union does not contain null")
	case json.Delim:
		if v == '{' {
			if haveObject {
				return true, json.Unmarshal(data, pc)
			}
			if haveMap {
				return false, json.Unmarshal(data, pm)
			}
			return false, errors.New("Union does not contain object")
		}
		if v == '[' {
			if haveArray {
				return false, json.Unmarshal(data, pa)
			}
			return false, errors.New("Union does not contain array")
		}
		return false, errors.New("Cannot handle delimiter")
	}
	return false, errors.New("Cannot unmarshal union")
}

func marshalUnion(pi *int64, pf *float64, pb *bool, ps *string, haveArray bool, pa interface{}, haveObject bool, pc interface{}, haveMap bool, pm interface{}, haveEnum bool, pe interface{}, nullable bool) ([]byte, error) {
	if pi != nil {
		return json.Marshal(*pi)
	}
	if pf != nil {
		return json.Marshal(*pf)
	}
	if pb != nil {
		return json.Marshal(*pb)
	}
	if ps != nil {
		return json.Marshal(*ps)
	}
	if haveArray {
		return json.Marshal(pa)
	}
	if haveObject {
		return json.Marshal(pc)
	}
	if haveMap {
		return json.Marshal(pm)
	}
	if haveEnum {
		return json.Marshal(pe)
	}
	if nullable {
		return json.Marshal(nil)
	}
	return nil, errors.New("Union must not be null")
}
