package codebase

import (
	"context"

	"ddm-admin-console/service/k8s"
)

type ServiceInterface interface {
	GetAllByType(tp string) ([]Codebase, error)
	Get(name string) (*Codebase, error)
	GetBranchesByCodebase(ctx context.Context, codebaseName string) ([]CodebaseBranch, error)
	Create(cb *Codebase) error
	CreateBranch(branch *CodebaseBranch) error
	Update(ctx context.Context, cb *Codebase) error
	Delete(name string) error
	CreateDefaultBranch(cb *Codebase) error
	ServiceForContext(ctx context.Context) (ServiceInterface, error)
	CreateTempSecrets(cb *Codebase, k8sService k8s.ServiceInterface, gerritCreatorSecretName string) error
	CheckPermissions(initial []Codebase, k8sService k8s.ServiceInterface) ([]WithPermissions, error)
	CheckIsAllowedToCreate(k8sService k8s.ServiceInterface) (bool, error)
	CheckIsAllowedToUpdate(codebaseName string, k8sService k8s.ServiceInterface) (bool, error)
}
