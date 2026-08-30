package oidcsessionredis_test

import (
	"testing"
	"time"

	"github.com/devctllabs/go-libs/oidcsession/mocks"
	"github.com/devctllabs/go-libs/oidcsessionredis"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewBackendValidatesLifetimePolicy(t *testing.T) {
	t.Parallel()
	controller := gomock.NewController(t)
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})
	t.Cleanup(func() { require.NoError(t, client.Close()) })
	tokens := mocks.NewMockTokenService(controller)
	encryptor := mocks.NewMockEncryptor(controller)

	tests := []oidcsessionredis.BackendConfig{
		{},
		{KeyPrefix: "sessions:", IdleTimeout: time.Hour},
		{KeyPrefix: "sessions:", AbsoluteLifetime: time.Hour},
		{KeyPrefix: "sessions:", IdleTimeout: 2 * time.Hour, AbsoluteLifetime: time.Hour},
		{KeyPrefix: "sessions:", IdleTimeout: time.Hour, AbsoluteLifetime: 2 * time.Hour, RefreshWindow: -time.Second},
	}
	for _, config := range tests {
		backend, err := oidcsessionredis.NewBackend(client, config, tokens, encryptor)
		require.Nil(t, backend)
		require.Error(t, err)
	}
}
