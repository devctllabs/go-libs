package oidcsession_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func receiveWithin[T any](t *testing.T, values <-chan T, timeout time.Duration) T {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case value := <-values:
		return value
	case <-timer.C:
		require.FailNow(t, "timed out waiting for channel value")
		var zero T
		return zero
	}
}
