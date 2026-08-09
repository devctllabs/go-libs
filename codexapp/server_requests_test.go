package codexapp

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecodeServerRequestVariants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		method     ServerRequestType
		params     string
		assertions func(*testing.T, ServerRequest)
	}{
		{
			name:   "file change",
			method: ServerRequestFileChangeApproval,
			params: `{"threadId":"thread","turnId":"turn","itemId":"item","reason":"write","grantRoot":"/workspace","startedAtMs":12}`,
			assertions: func(t *testing.T, request ServerRequest) {
				require.Equal(t, &FileChangeApprovalRequest{Reason: "write", GrantRoot: "/workspace", StartedAtMS: 12}, request.FileChangeApproval)
			},
		},
		{
			name:   "permissions",
			method: ServerRequestPermissionApproval,
			params: `{"threadId":"thread","turnId":"turn","itemId":"item","cwd":"/workspace","reason":"network","startedAtMs":13,"permissions":{"network":{"enabled":true}}}`,
			assertions: func(t *testing.T, request ServerRequest) {
				require.Equal(t, "/workspace", request.PermissionApproval.Cwd)
				require.JSONEq(t, `{"network":{"enabled":true}}`, string(request.PermissionApproval.Requested))
			},
		},
		{
			name:   "user input",
			method: ServerRequestUserInput,
			params: `{"threadId":"thread","turnId":"turn","itemId":"item","autoResolutionMs":1000,"questions":[{"id":"choice","header":"Choice","question":"Pick","options":[{"label":"A","description":"first"}]}]}`,
			assertions: func(t *testing.T, request ServerRequest) {
				require.Equal(t, "choice", request.UserInput.Questions[0].ID)
				require.EqualValues(t, 1000, *request.UserInput.AutoResolutionMS)
			},
		},
		{
			name:   "mcp elicitation",
			method: ServerRequestMCPElicitation,
			params: `{"threadId":"thread","turnId":"turn","serverName":"docs","mode":"form","message":"Credentials","requestedSchema":{"type":"object"}}`,
			assertions: func(t *testing.T, request ServerRequest) {
				require.Equal(t, "docs", request.MCPElicitation.ServerName)
				require.JSONEq(t, `{"type":"object"}`, string(request.MCPElicitation.RequestedSchema))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			request, err := decodeServerRequest(string(tt.method), json.RawMessage(tt.params))
			require.NoError(t, err)
			require.Equal(t, tt.method, request.Type)
			require.Equal(t, "thread", request.ThreadID)
			require.Equal(t, "turn", request.TurnID)
			tt.assertions(t, request)
		})
	}
}
