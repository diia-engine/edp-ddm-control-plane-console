package gerrit

import "context"

type ServiceInterface interface {
	GetProjects(ctx context.Context) ([]GerritProject, error)
	GetProject(ctx context.Context, name string) (*GerritProject, error)
	GetMergeRequest(ctx context.Context, name string) (*GerritMergeRequest, error)
	CreateMergeRequest(ctx context.Context, mr *MergeRequest) error
	GetMergeRequestByProject(ctx context.Context, projectName string) ([]GerritMergeRequest, error)
}
