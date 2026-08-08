package healthserver

import (
	"context"
	"testing"

	"github.com/devctllabs/go-libs/healthserver/internal/generated"
	"github.com/stretchr/testify/require"
)

func TestEmbeddedOpenAPIContractIsValid(t *testing.T) {
	t.Parallel()
	document, err := generated.GetSpec()
	require.NoError(t, err)
	require.NoError(t, document.Validate(context.Background()))

	require.Nil(t, document.Paths.Find("/startupz"))

	liveness := document.Paths.Find("/livez")
	require.NotNil(t, liveness)
	require.NotNil(t, liveness.Get)
	require.NotNil(t, liveness.Get.Responses.Value("200"))
}
