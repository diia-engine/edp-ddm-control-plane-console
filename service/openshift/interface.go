package openshift

import (
	"context"
)

type ServiceInterface interface {
	GetMe(ctx context.Context) (*User, error)
	GetInfrastructureCluster(ctx context.Context) (*ClusterInfrastructure, error)
}
