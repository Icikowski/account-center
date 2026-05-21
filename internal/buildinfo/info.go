package buildinfo

import (
	"runtime"
	"time"
)

// BuildInfo represents the build information of the application.
type BuildInfo struct {
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	BuildTime time.Time `json:"build_time"`
	GoVersion string    `json:"go_version"`
}

// Get returns the [BuildInfo].
func Get() BuildInfo {
	var buildTime time.Time
	if parsed, err := time.Parse(time.RFC3339, timestamp); err == nil {
		buildTime = parsed
	}
	return BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
		GoVersion: runtime.Version(),
	}
}
