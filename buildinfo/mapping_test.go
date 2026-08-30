package buildinfo

import (
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFromBuildInfoReturnsZeroForMissingMetadata(t *testing.T) {
	t.Parallel()

	require.Equal(t, Info{}, fromBuildInfo(nil))
}

func TestFromBuildInfoMapsModuleAndVCSMetadata(t *testing.T) {
	t.Parallel()
	wantTime := time.Date(2026, time.August, 20, 11, 20, 30, 123456789, time.UTC)
	raw := &debug.BuildInfo{
		GoVersion: "go1.25.0",
		Main: debug.Module{
			Path:    "example.com/orders",
			Version: "v1.4.0+dirty",
		},
		Settings: []debug.BuildSetting{
			{Key: revisionSetting, Value: "8f31c2a76f1234567890"},
			{Key: revisionTimeSetting, Value: "2026-08-20T11:20:30.123456789Z"},
			{Key: "unrelated", Value: "ignored"},
		},
	}

	require.Equal(t, Info{
		ModulePath:   "example.com/orders",
		Version:      "v1.4.0+dirty",
		Revision:     "8f31c2a76f1234567890",
		RevisionTime: wantTime,
		GoVersion:    "go1.25.0",
	}, fromBuildInfo(raw))
}

func TestFromBuildInfoNormalizesVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "empty", raw: "", want: ""},
		{name: "development sentinel", raw: "(devel)", want: ""},
		{name: "tag", raw: "v1.4.0", want: "v1.4.0"},
		{name: "pseudo-version", raw: "v1.4.1-0.20260820112030-8f31c2a76f12", want: "v1.4.1-0.20260820112030-8f31c2a76f12"},
		{name: "dirty suffix", raw: "v1.4.0+dirty", want: "v1.4.0+dirty"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			raw := &debug.BuildInfo{Main: debug.Module{Version: test.raw}}

			require.Equal(t, test.want, fromBuildInfo(raw).Version)
		})
	}
}

func TestFromBuildInfoIgnoresInvalidRevisionTime(t *testing.T) {
	t.Parallel()
	raw := &debug.BuildInfo{Settings: []debug.BuildSetting{
		{Key: revisionSetting, Value: "8f31c2a76f1234567890"},
		{Key: revisionTimeSetting, Value: "not-a-time"},
	}}

	info := fromBuildInfo(raw)
	require.Equal(t, "8f31c2a76f1234567890", info.Revision)
	require.True(t, info.RevisionTime.IsZero())
}
