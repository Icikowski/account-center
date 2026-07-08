package buildinfo

import (
	"runtime"
	"strings"
	"time"
)

// BuildInfo represents the build information of the application.
type BuildInfo struct {
	Version      string    `json:"version"`
	GitReference string    `json:"commit"`
	BuildTime    time.Time `json:"build_time"`
	GoVersion    string    `json:"go_version"`
}

// Get returns the [BuildInfo].
func Get() BuildInfo {
	var buildTime time.Time
	if parsed, err := time.Parse(time.RFC3339, timestamp); err == nil {
		buildTime = parsed
	}
	return BuildInfo{
		Version:      version,
		GitReference: gitref,
		BuildTime:    buildTime,
		GoVersion:    strings.TrimPrefix(runtime.Version(), "go"),
	}
}
