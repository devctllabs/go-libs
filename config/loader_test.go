package config_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/devctllabs/go-libs/config"
	"github.com/devctllabs/go-libs/config/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestEmptyChainDoesNothing(t *testing.T) {
	t.Parallel()

	target := struct{ Value string }{Value: "unchanged"}

	err := config.Chain().Load(context.Background(), &target)

	require.NoError(t, err)
	require.Equal(t, "unchanged", target.Value)
}

func TestChainLoadsInOrder(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	first := mocks.NewMockLoader(ctrl)
	second := mocks.NewMockLoader(ctrl)
	target := &struct{}{}

	gomock.InOrder(
		first.EXPECT().Load(gomock.Any(), target).Return(nil),
		second.EXPECT().Load(gomock.Any(), target).Return(nil),
	)

	err := config.Chain(first, second).Load(context.Background(), target)

	require.NoError(t, err)
}

func TestChainStopsAtFirstError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	first := mocks.NewMockLoader(ctrl)
	second := mocks.NewMockLoader(ctrl)
	third := mocks.NewMockLoader(ctrl)
	target := &struct{}{}
	wantErr := errors.New("load failed")

	gomock.InOrder(
		first.EXPECT().Load(gomock.Any(), target).Return(nil),
		second.EXPECT().Load(gomock.Any(), target).Return(wantErr),
	)

	err := config.Chain(first, second, third).Load(context.Background(), target)

	require.ErrorIs(t, err, wantErr)
	require.ErrorContains(t, err, "loader 2")
}

func TestOptionalIgnoresNotExist(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	loader := mocks.NewMockLoader(ctrl)
	target := &struct{}{}
	loader.EXPECT().Load(gomock.Any(), target).Return(fmt.Errorf("open config: %w", fs.ErrNotExist))

	err := config.Optional(loader).Load(context.Background(), target)

	require.NoError(t, err)
}

func TestOptionalPreservesOtherErrors(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	loader := mocks.NewMockLoader(ctrl)
	target := &struct{}{}
	wantErr := fs.ErrPermission
	loader.EXPECT().Load(gomock.Any(), target).Return(wantErr)

	err := config.Optional(loader).Load(context.Background(), target)

	require.ErrorIs(t, err, wantErr)
}

type typedTarget struct {
	Value string
}

type typedTargetLoader struct{}

func (typedTargetLoader) Load(_ context.Context, target *typedTarget) error {
	target.Value = "loaded"
	return nil
}

func TestTypedAdaptsConcreteTarget(t *testing.T) {
	t.Parallel()

	target := &typedTarget{}

	err := config.Typed(typedTargetLoader{}).Load(context.Background(), target)

	require.NoError(t, err)
	require.Equal(t, "loaded", target.Value)
}

func TestTypedRejectsDifferentTarget(t *testing.T) {
	t.Parallel()

	err := config.Typed(typedTargetLoader{}).Load(context.Background(), &struct{}{})

	require.ErrorContains(t, err, "expects *config_test.typedTarget")
}

func TestChainRejectsNilLoader(t *testing.T) {
	t.Parallel()

	err := config.Chain(nil).Load(context.Background(), &struct{}{})

	require.ErrorContains(t, err, "nil loader")
}

func TestOptionalRejectsNilLoader(t *testing.T) {
	t.Parallel()

	err := config.Optional(nil).Load(context.Background(), &struct{}{})

	require.ErrorContains(t, err, "nil loader")
}

func TestChainOwnsLoaderSlice(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	first := mocks.NewMockLoader(ctrl)
	second := mocks.NewMockLoader(ctrl)
	target := &struct{}{}
	loaders := []config.Loader{first}
	loader := config.Chain(loaders...)
	loaders[0] = second
	first.EXPECT().Load(gomock.Any(), target).Return(nil)

	err := loader.Load(context.Background(), target)

	require.NoError(t, err)
}

func TestChainRejectsTypedNilLoader(t *testing.T) {
	t.Parallel()

	var loader *mocks.MockLoader

	err := config.Chain(loader).Load(context.Background(), &struct{}{})

	require.ErrorContains(t, err, "nil loader")
}

type nilTypedLoader struct{}

func (*nilTypedLoader) Load(_ context.Context, _ *typedTarget) error {
	panic("must not be called")
}

func TestTypedRejectsTypedNilLoader(t *testing.T) {
	t.Parallel()

	var loader *nilTypedLoader

	err := config.Typed(loader).Load(context.Background(), &typedTarget{})

	require.ErrorContains(t, err, "nil loader")
}
