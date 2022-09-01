package k8s

import (
	"context"

	v1 "k8s.io/api/core/v1"
)

type ServiceInterface interface {
	ServiceForContext(ctx context.Context) (ServiceInterface, error)
	GetSecret(name string) (*v1.Secret, error)
	RecreateSecret(secretName string, data map[string][]byte) error
	CanI(group, resource, verb, name string) (bool, error)
	GetSecretFromNamespace(ctx context.Context, name, namespace string) (*v1.Secret, error)
	GetSecretKey(ctx context.Context, namespace, name, key string) (string, error)
}
