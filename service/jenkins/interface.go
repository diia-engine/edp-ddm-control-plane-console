package jenkins

import "context"

type ServiceInterface interface {
	CreateJobBuildRunRaw(jb *JenkinsJobBuildRun) error
	CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error
	ServiceForContext(ctx context.Context) (ServiceInterface, error)
}
