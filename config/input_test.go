package config_test

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/devctllabs/go-libs/config"
	"github.com/devctllabs/go-libs/config/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPathOpensFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("value: file"), 0o600))

	reader, err := config.Path(path).Open(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reader.Close()) })

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, "value: file", string(data))
}

func TestPathPreservesNotExist(t *testing.T) {
	t.Parallel()

	reader, err := config.Path(filepath.Join(t.TempDir(), "missing.yaml")).Open(context.Background())

	require.Nil(t, reader)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestFromFSOpensFile(t *testing.T) {
	t.Parallel()

	filesystem := fstest.MapFS{
		"config.toml": {Data: []byte(`value = "fs"`)},
	}

	reader, err := config.FromFS(filesystem, "config.toml").Open(context.Background())
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, reader.Close()) })

	data, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, `value = "fs"`, string(data))
}

func TestInputHonorsCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reader, err := config.FromFS(fstest.MapFS{}, "config.yaml").Open(ctx)

	require.Nil(t, reader)
	require.ErrorIs(t, err, context.Canceled)
}

type closeErrorReader struct {
	io.Reader
	closed bool
}

func (r *closeErrorReader) Close() error {
	r.closed = true
	return errors.New("close failed")
}

func TestFormatLoadersCloseInputAndIgnoreCloseError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		loader  func(config.Input) config.Loader
		target  any
	}{
		{
			name:    "YAML",
			content: "value: decoded\n",
			loader:  config.YAML,
			target:  &map[string]string{},
		},
		{
			name:    "dotenv",
			content: "TEST_APP_HOST=decoded\n",
			loader:  config.DotEnv,
			target:  &envConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			input := mocks.NewMockInput(ctrl)
			reader := &closeErrorReader{Reader: strings.NewReader(tt.content)}
			input.EXPECT().Open(gomock.Any()).Return(reader, nil)

			err := tt.loader(input).Load(context.Background(), tt.target)

			require.NoError(t, err)
			require.True(t, reader.closed)
		})
	}
}

func TestFormatLoaderRejectsTypedNilInput(t *testing.T) {
	t.Parallel()

	var input *mocks.MockInput

	err := config.YAML(input).Load(context.Background(), &struct{}{})

	require.ErrorContains(t, err, "nil YAML input")
}
