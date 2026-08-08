package config_test

import (
	"context"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/config"
	"github.com/stretchr/testify/require"
)

type envConfig struct {
	Host    string        `env:"TEST_APP_HOST"`
	Port    int           `env:"TEST_APP_PORT"`
	Timeout time.Duration `env:"TEST_APP_TIMEOUT"`
	Labels  []string      `env:"TEST_APP_LABELS"`
}

func TestEnvMapDecodesExplicitTags(t *testing.T) {
	t.Parallel()

	target := envConfig{Host: "existing"}
	loader := config.EnvMap(map[string]string{
		"TEST_APP_PORT":    "8080",
		"TEST_APP_TIMEOUT": "5s",
		"TEST_APP_LABELS":  "one,two",
	})

	err := loader.Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "existing", target.Host)
	require.Equal(t, 8080, target.Port)
	require.Equal(t, 5*time.Second, target.Timeout)
	require.Equal(t, []string{"one", "two"}, target.Labels)
}

func TestEnvMapTakesSnapshot(t *testing.T) {
	t.Parallel()

	values := map[string]string{"TEST_APP_HOST": "before"}
	loader := config.EnvMap(values)
	values["TEST_APP_HOST"] = "after"
	var target envConfig

	err := loader.Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "before", target.Host)
}

func TestEnvMapNilDoesNotReadOSEnvironment(t *testing.T) {
	t.Setenv("TEST_APP_HOST", "from-os")
	target := envConfig{Host: "existing"}

	err := config.EnvMap(nil).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "existing", target.Host)
}

func TestOSEnvReadsEnvironmentAtLoadTime(t *testing.T) {
	loader := config.OSEnv()
	t.Setenv("TEST_APP_HOST", "from-os")
	var target envConfig

	err := loader.Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "from-os", target.Host)
}

func TestEnvMapReturnsConversionError(t *testing.T) {
	t.Parallel()

	var target envConfig

	err := config.EnvMap(map[string]string{"TEST_APP_PORT": "invalid"}).Load(context.Background(), &target)

	require.Error(t, err)
}
