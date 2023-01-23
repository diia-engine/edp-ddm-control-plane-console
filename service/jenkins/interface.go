package jenkins

import "context"

type ServiceInterface interface {
	CreateJobBuildRunRaw(ctx context.Context, jb *JenkinsJobBuildRun) error
	CreateJobBuildRun(ctx context.Context, name, jobPath string, jobParams map[string]string) error
	ServiceForContext(ctx context.Context) (ServiceInterface, error)
	GetJobStatus(ctx context.Context, jobName string) (string, error)
}
