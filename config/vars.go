package config

import (
	"runtime"

	"github.com/carlmjohnson/versioninfo"
)

// Automatically set at build time
var (
	Version   string
	BuildTime string
	Revision  string
	OS        string
)

func GetVersion() string {
	if Version == "" {
		return versioninfo.Version
	}
	return Version
}

func GetBuildTime() string {
	if BuildTime == "" {
		return versioninfo.LastCommit.Local().Format("2006-01-02T15:04+00:00")
	}
	return BuildTime
}

func GetRevision() string {
	if Revision == "" {
		return versioninfo.Revision
	}
	return Revision
}

func GetOS() string {
	if OS == "" {
		return runtime.GOOS
	}
	return OS
}
