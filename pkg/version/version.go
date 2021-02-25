package version

import (
	"fmt"
	"runtime"
)

var (
	GitVersion string
	GitCommit  string
	BuildDate  string
)

type Version struct {
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	GoVersion  string `json:"goVersion"`
	BuildDate  string `json:"buildDate"`
}

// Get returns the codebase version. It's for detecting what code a binary was built from.
func Get() Version {
	return Version{GitVersion, GitCommit, runtime.Version(), BuildDate}
}

func (v Version) String() string {
	return fmt.Sprintf("{GitVersion: %s, GitCommit: %s, Goversion: %s, BuildDate: %s}", v.GitVersion, v.GitCommit, v.GoVersion, v.BuildDate)
}
