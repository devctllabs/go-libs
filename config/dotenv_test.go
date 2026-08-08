package config_test

import (
	"context"
	"io/fs"
	"os"
	"testing"
	"testing/fstest"

	"github.com/devctllabs/go-libs/config"
	"github.com/stretchr/testify/require"
)

func TestDotEnvLoadsWithoutMutatingProcessEnvironment(t *testing.T) {
	t.Setenv("TEST_APP_HOST", "from-os")
	filesystem := fstest.MapFS{
		"app.env": {Data: []byte("TEST_APP_HOST=from-file\nTEST_APP_PORT=8080\n")},
	}
	var target envConfig

	err := config.DotEnv(config.FromFS(filesystem, "app.env")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "from-file", target.Host)
	require.Equal(t, 8080, target.Port)
	require.Equal(t, "from-os", os.Getenv("TEST_APP_HOST"))
}

func TestMultipleDotEnvFilesAndEnvMapUseChainOrder(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"base.env":  {Data: []byte("TEST_APP_HOST=base\nTEST_APP_PORT=1000\n")},
		"local.env": {Data: []byte("TEST_APP_PORT=2000\n")},
	}
	var target envConfig
	loader := config.Chain(
		config.DotEnv(config.FromFS(filesystem, "base.env")),
		config.DotEnv(config.FromFS(filesystem, "local.env")),
		config.EnvMap(map[string]string{"TEST_APP_PORT": "3000"}),
	)

	err := loader.Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "base", target.Host)
	require.Equal(t, 3000, target.Port)
}

func TestOptionalDotEnvIgnoresMissingInput(t *testing.T) {
	t.Parallel()

	var target envConfig
	loader := config.Optional(config.DotEnv(config.FromFS(fstest.MapFS{}, "missing.env")))

	err := loader.Load(context.Background(), &target)

	require.NoError(t, err)
}

func TestDotEnvPreservesMissingInputError(t *testing.T) {
	t.Parallel()

	var target envConfig

	err := config.DotEnv(config.FromFS(fstest.MapFS{}, "missing.env")).Load(context.Background(), &target)

	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestDotEnvReturnsParseError(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"broken.env": {Data: []byte("BROKEN='unterminated\n")},
	}
	var target envConfig

	err := config.DotEnv(config.FromFS(filesystem, "broken.env")).Load(context.Background(), &target)

	require.Error(t, err)
}
