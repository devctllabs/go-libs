package config_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/devctllabs/go-libs/config"
	"github.com/stretchr/testify/require"
)

func TestTOMLOverlaysStructAndReplacesCollections(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.toml": {Data: []byte(`
[server]
port = 2000
names = ["new"]
unknown = "ignored"

[server.headers]
new = "value"
`)},
	}
	target := fileConfig{
		Server: &fileServerConfig{
			Host:    "existing",
			Port:    1000,
			Headers: map[string]string{"old": "value"},
			Names:   []string{"old"},
		},
		Ignored: "existing",
	}

	err := config.TOML(config.FromFS(filesystem, "config.toml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "existing", target.Server.Host)
	require.Equal(t, 2000, target.Server.Port)
	require.Equal(t, map[string]string{"new": "value"}, target.Server.Headers)
	require.Equal(t, []string{"new"}, target.Server.Names)
	require.Equal(t, "existing", target.Ignored)
}

func TestTOMLAllocatesNestedPointer(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.toml": {Data: []byte("[server]\nhost = \"created\"\n")},
	}
	var target fileConfig

	err := config.TOML(config.FromFS(filesystem, "config.toml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.NotNil(t, target.Server)
	require.Equal(t, "created", target.Server.Host)
}

func TestEmptyTOMLDoesNothing(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{"config.toml": {Data: nil}}
	target := fileConfig{Server: &fileServerConfig{Host: "existing"}}

	err := config.TOML(config.FromFS(filesystem, "config.toml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "existing", target.Server.Host)
}

func TestTOMLReturnsParseError(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.toml": {Data: []byte("invalid = [\n")},
	}
	var target fileConfig

	err := config.TOML(config.FromFS(filesystem, "config.toml")).Load(context.Background(), &target)

	require.Error(t, err)
}

func TestTOMLDelegatesNonStructTargetToNativeDecoder(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.toml": {Data: []byte(`value = "decoded"`)},
	}
	target := map[string]string{}

	err := config.TOML(config.FromFS(filesystem, "config.toml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, map[string]string{"value": "decoded"}, target)
}
