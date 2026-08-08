package config_test

import (
	"context"
	"testing"

	"github.com/devctllabs/go-libs/config"
	"github.com/stretchr/testify/require"
)

func TestDefaultsSetsZeroValuesAndPreservesExistingValues(t *testing.T) {
	t.Parallel()

	target := struct {
		Host string `default:"localhost"`
		Port int    `default:"8080"`
	}{
		Port: 9090,
	}

	err := config.Defaults().Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "localhost", target.Host)
	require.Equal(t, 9090, target.Port)
}

func TestDefaultsReturnsMalformedTagError(t *testing.T) {
	t.Parallel()

	target := struct {
		Values []string `default:"not-json"`
	}{}

	err := config.Defaults().Load(context.Background(), &target)

	require.Error(t, err)
}
