package txmanager_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devctllabs/go-libs/txmanager"
	"github.com/devctllabs/go-libs/txmanager/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewCoordinatorRejectsNilBackend(t *testing.T) {
	t.Parallel()
	coordinator, err := txmanager.NewCoordinator[struct{}](nil)

	require.Nil(t, coordinator)
	require.Error(t, err)
}

func TestWriterWithinTxCommitsSuccessfulCallback(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	backend.EXPECT().Begin(gomock.Any(), txmanager.BeginSpec{Role: txmanager.RoleWriter}).Return("tx", nil)
	backend.EXPECT().Commit(gomock.Any(), "tx").Return(nil)

	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	called := false
	err = coordinator.Writer().WithinTx(context.Background(), func(ctx context.Context) error {
		called = true
		require.NotNil(t, ctx)
		return nil
	})

	require.NoError(t, err)
	require.True(t, called)
}

func TestWithinTxJoinsCallbackAndRollbackErrors(t *testing.T) {
	t.Parallel()
	callbackErr := errors.New("callback failed")
	rollbackErr := errors.New("rollback failed")
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	backend.EXPECT().Begin(gomock.Any(), txmanager.BeginSpec{Role: txmanager.RoleWriter}).Return("tx", nil)
	backend.EXPECT().Rollback(gomock.Any(), "tx").Return(rollbackErr)

	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	err = coordinator.Writer().WithinTx(context.Background(), func(context.Context) error {
		return callbackErr
	})

	require.ErrorIs(t, err, callbackErr)
	require.ErrorIs(t, err, rollbackErr)
}

func TestWithinTxRollsBackBeforeRepanicking(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	gomock.InOrder(
		backend.EXPECT().Begin(gomock.Any(), txmanager.BeginSpec{Role: txmanager.RoleWriter}).Return("tx", nil),
		backend.EXPECT().Rollback(gomock.Any(), "tx").Return(nil),
	)

	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	require.PanicsWithValue(t, "boom", func() {
		_ = coordinator.Writer().WithinTx(context.Background(), func(context.Context) error {
			panic("boom")
		})
	})
}

func TestCurrentReturnsTransactionOnlyInsideCallback(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	backend.EXPECT().Begin(gomock.Any(), txmanager.BeginSpec{Role: txmanager.RoleWriter}).Return("tx", nil)
	backend.EXPECT().Commit(gomock.Any(), "tx").Return(nil)

	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	_, ok := coordinator.Current(context.Background())
	require.False(t, ok)
	err = coordinator.Writer().WithinTx(context.Background(), func(ctx context.Context) error {
		current, currentOK := coordinator.Current(ctx)
		require.True(t, currentOK)
		require.Equal(t, "tx", current)
		return nil
	})
	require.NoError(t, err)
}

func TestReaderInsideWriterReusesActiveTransaction(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	backend.EXPECT().Begin(gomock.Any(), txmanager.BeginSpec{Role: txmanager.RoleWriter}).Return("tx", nil)
	backend.EXPECT().Commit(gomock.Any(), "tx").Return(nil)

	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	err = coordinator.Writer().WithinTx(context.Background(), func(ctx context.Context) error {
		return coordinator.Reader().WithinTx(ctx, func(nestedCtx context.Context) error {
			current, ok := coordinator.Current(nestedCtx)
			require.True(t, ok)
			require.Equal(t, "tx", current)
			return nil
		})
	})
	require.NoError(t, err)
}

func TestWriterInsideReaderReturnsReadOnlyEscalation(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	backend.EXPECT().Begin(gomock.Any(), txmanager.BeginSpec{Role: txmanager.RoleReader}).Return("tx", nil)
	backend.EXPECT().Rollback(gomock.Any(), "tx").Return(nil)

	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	nestedCalled := false
	err = coordinator.Reader().WithinTx(context.Background(), func(ctx context.Context) error {
		return coordinator.Writer().WithinTx(ctx, func(context.Context) error {
			nestedCalled = true
			return nil
		})
	})

	require.ErrorIs(t, err, txmanager.ErrReadOnlyEscalation)
	require.False(t, nestedCalled)
}

func TestNestedDifferentIsolationReturnsMismatch(t *testing.T) {
	t.Parallel()
	outerIsolation := txmanager.Serializable
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	backend.EXPECT().Begin(gomock.Any(), txmanager.BeginSpec{
		Role:      txmanager.RoleWriter,
		Isolation: &outerIsolation,
	}).Return("tx", nil)
	backend.EXPECT().Rollback(gomock.Any(), "tx").Return(nil)

	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	err = coordinator.Writer().WithinTx(
		context.Background(),
		func(ctx context.Context) error {
			return coordinator.Reader().WithinTx(
				ctx,
				func(context.Context) error { return nil },
				txmanager.WithIsolation(txmanager.ReadCommitted),
			)
		},
		txmanager.WithIsolation(txmanager.Serializable),
	)

	require.ErrorIs(t, err, txmanager.ErrIsolationMismatch)
}

func TestManagersReturnsConfiguredRoles(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	backend := mocks.NewMockBackend[string](ctrl)
	coordinator, err := txmanager.NewCoordinator[string](backend)
	require.NoError(t, err)

	managers, err := txmanager.NewManagers(coordinator.Reader(), coordinator.Writer())
	require.NoError(t, err)
	require.Same(t, coordinator.Reader(), managers.Reader())
	require.Same(t, coordinator.Writer(), managers.Writer())
}
