package config_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/devctllabs/go-libs/config"
	"github.com/stretchr/testify/require"
)

type fileServerConfig struct {
	Host    string            `yaml:"host" toml:"host"`
	Port    int               `yaml:"port" toml:"port"`
	Headers map[string]string `yaml:"headers" toml:"headers"`
	Names   []string          `yaml:"names" toml:"names"`
}

type fileConfig struct {
	Server  *fileServerConfig `yaml:"server" toml:"server"`
	Ignored string            `yaml:"-" toml:"-"`
}

func TestYAMLOverlaysStructAndReplacesCollections(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.yaml": {Data: []byte(`
server:
  port: 2000
  headers:
    new: value
  names: [new]
unknown: ignored
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

	err := config.YAML(config.FromFS(filesystem, "config.yaml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "existing", target.Server.Host)
	require.Equal(t, 2000, target.Server.Port)
	require.Equal(t, map[string]string{"new": "value"}, target.Server.Headers)
	require.Equal(t, []string{"new"}, target.Server.Names)
	require.Equal(t, "existing", target.Ignored)
}

func TestYAMLNullClearsPointer(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.yaml": {Data: []byte("server: null\n")},
	}
	target := fileConfig{Server: &fileServerConfig{Host: "existing"}}

	err := config.YAML(config.FromFS(filesystem, "config.yaml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Nil(t, target.Server)
}

func TestEmptyYAMLDoesNothing(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{"config.yaml": {Data: nil}}
	target := fileConfig{Server: &fileServerConfig{Host: "existing"}}

	err := config.YAML(config.FromFS(filesystem, "config.yaml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "existing", target.Server.Host)
}

func TestYAMLRejectsMultipleDocuments(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.yaml": {Data: []byte("server:\n  port: 1000\n---\nserver:\n  port: 2000\n")},
	}
	var target fileConfig

	err := config.YAML(config.FromFS(filesystem, "config.yaml")).Load(context.Background(), &target)

	require.ErrorContains(t, err, "multiple documents")
}

func TestYAMLDelegatesNonStructTargetToNativeDecoder(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.yaml": {Data: []byte("value: decoded\n")},
	}
	target := map[string]string{}

	err := config.YAML(config.FromFS(filesystem, "config.yaml")).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, map[string]string{"value": "decoded"}, target)
}
