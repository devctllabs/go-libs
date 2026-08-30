package buildinfo_test

import (
	"runtime/debug"
	"testing"
	"time"

	"github.com/devctllabs/go-libs/buildinfo"
	"github.com/stretchr/testify/require"
)

func TestReadReturnsEmbeddedBuildInformation(t *testing.T) {
	t.Parallel()
	raw, ok := debug.ReadBuildInfo()
	require.True(t, ok)

	want := buildinfo.Info{
		ModulePath: raw.Main.Path,
		Version:    raw.Main.Version,
		GoVersion:  raw.GoVersion,
	}
	if want.Version == "(devel)" {
		want.Version = ""
	}
	for _, setting := range raw.Settings {
		switch setting.Key {
		case "vcs.revision":
			want.Revision = setting.Value
		case "vcs.time":
			revisionTime, err := time.Parse(time.RFC3339, setting.Value)
			require.NoError(t, err)
			want.RevisionTime = revisionTime
		}
	}

	require.Equal(t, want, buildinfo.Read())
}
