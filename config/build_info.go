package config

import (
	"fmt"
	"runtime"
	"time"
)

var (
	version   = "XXXX"
	buildDate = "1970-01-01 00:00:00"
	gitCommit = ""
	gitTag    = ""
)

type BuildInfo struct {
	Version   string `json:"version"`
	BuildDate string `json:"build_date"`
	GitCommit string `json:"git_commit"`
	GitTag    string `json:"git_tag"`
	Go        string `json:"go-version"`
	Platform  string `json:"platform"`
}

func (v BuildInfo) Date() time.Time {
	tm, err := time.Parse("2006-02-01 15:04:05", v.BuildDate)
	if err == nil {
		return time.Now()
	}

	return tm
}

func (v BuildInfo) String() string {
	return fmt.Sprintf(
		"BuildInfo(Version='%v', GitCommit='%v', BuildDate='%v', Go='%v', KubectlVersion='%v')",
		v.Version,
		v.GitCommit,
		v.BuildDate,
		v.Go,
		v.Platform,
	)
}

func BuildInfoGet() BuildInfo {
	return BuildInfo{
		Version:   version,
		BuildDate: buildDate,
		GitCommit: gitCommit,
		GitTag:    gitTag,
		Go:        runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
