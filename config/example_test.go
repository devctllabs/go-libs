package config_test

import (
	"context"
	"log"
	"os"
	"testing"
	"testing/fstest"

	"github.com/devctllabs/go-libs/config"
	"github.com/stretchr/testify/require"
)

type exampleConfig struct {
	Host  string `default:"default" yaml:"host" env:"EXAMPLE_HOST"`
	Port  int    `default:"1000" yaml:"port" env:"EXAMPLE_PORT"`
	Debug bool   `default:"true" yaml:"debug" env:"EXAMPLE_DEBUG"`
}

type exampleCLILoader struct {
	port  *int
	debug *bool
}

func (l exampleCLILoader) Load(_ context.Context, target *exampleConfig) error {
	if l.port != nil {
		target.Port = *l.port
	}
	if l.debug != nil {
		target.Debug = *l.debug
	}
	return nil
}

func TestTypedCLIOverridePreservesAbsentAndAppliesExplicitZero(t *testing.T) {
	t.Parallel()

	target := exampleConfig{Port: 3000, Debug: true}
	port := 0

	err := config.Typed(exampleCLILoader{port: &port}).Load(context.Background(), &target)

	require.NoError(t, err)
	require.Zero(t, target.Port)
	require.True(t, target.Debug)
}

func ExampleChain() {
	filesystem := fstest.MapFS{
		"config.yaml": {Data: []byte("host: yaml\nport: 2000\n")},
		"local.env":   {Data: []byte("EXAMPLE_PORT=3000\n")},
	}
	cliPort := 4000
	cliDebug := false
	var target exampleConfig
	loader := config.Chain(
		config.Defaults(),
		config.YAML(config.FromFS(filesystem, "config.yaml")),
		config.DotEnv(config.FromFS(filesystem, "local.env")),
		config.EnvMap(map[string]string{"EXAMPLE_HOST": "environment"}),
		config.Typed(exampleCLILoader{port: &cliPort, debug: &cliDebug}),
	)

	if err := loader.Load(context.Background(), &target); err != nil {
		log.Fatal(err)
	}

	logger := log.New(os.Stdout, "", 0)
	logger.Printf("host=%s port=%d debug=%t", target.Host, target.Port, target.Debug)
	// Output: host=environment port=4000 debug=false
}
