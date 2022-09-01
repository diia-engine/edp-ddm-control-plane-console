package openshift

import (
	"context"

	openshiftV1 "github.com/openshift/api/user/v1"
)

type ServiceInterface interface {
	ServiceForContext(ctx context.Context) (ServiceInterface, error)
	GetMe(ctx context.Context) (*openshiftV1.User, error)
}
