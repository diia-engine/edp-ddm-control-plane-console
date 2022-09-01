package gerrit

import (
	"context"

	goGerrit "github.com/andygrunwald/go-gerrit"
)

type ServiceInterface interface {
	GetProjects(ctx context.Context) ([]GerritProject, error)
	GetProject(ctx context.Context, name string) (*GerritProject, error)
	GetMergeRequest(ctx context.Context, name string) (*GerritMergeRequest, error)
	CreateMergeRequest(ctx context.Context, mr *MergeRequest) error
	GetMergeRequestByProject(ctx context.Context, projectName string) ([]GerritMergeRequest, error)
	CreateProject(ctx context.Context, name string) error
	GetFileContents(ctx context.Context, projectName, branch, filePath string) (string, error)
	CreateMergeRequestWithContents(ctx context.Context, mr *MergeRequest, contents map[string]string) error
	GoGerritClient() *goGerrit.Client
}
