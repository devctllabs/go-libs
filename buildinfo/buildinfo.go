package buildinfo

import (
	"runtime/debug"
	"time"
)

const (
	revisionSetting     = "vcs.revision"
	revisionTimeSetting = "vcs.time"
)

// Info identifies the main module and source revision of the current executable.
// Fields contain zero values when the corresponding metadata was not embedded.
// RevisionTime is the time associated with Revision, not the build time.
type Info struct {
	ModulePath   string
	Version      string
	Revision     string
	RevisionTime time.Time
	GoVersion    string
}

// Read returns build metadata embedded in the current executable.
//
// Read performs no filesystem, environment, network, or VCS access. Version is empty when Go
// reports no useful main-module version or "(devel)". Revision is returned without shortening;
// version suffixes such as "+dirty" are preserved.
func Read() Info {
	raw, ok := debug.ReadBuildInfo()
	if !ok {
		return Info{}
	}
	return fromBuildInfo(raw)
}

func fromBuildInfo(raw *debug.BuildInfo) Info {
	if raw == nil {
		return Info{}
	}

	info := Info{
		ModulePath: raw.Main.Path,
		Version:    raw.Main.Version,
		GoVersion:  raw.GoVersion,
	}
	if info.Version == "(devel)" {
		info.Version = ""
	}

	for _, setting := range raw.Settings {
		switch setting.Key {
		case revisionSetting:
			info.Revision = setting.Value
		case revisionTimeSetting:
			if revisionTime, err := time.Parse(time.RFC3339, setting.Value); err == nil {
				info.RevisionTime = revisionTime
			}
		}
	}
	return info
}
