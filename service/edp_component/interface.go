package edpcomponent

import "context"

type ServiceInterface interface {
	GetAll(ctx context.Context) ([]EDPComponent, error)
	GetAllNamespace(ctx context.Context, ns string, onlyVisible bool) ([]EDPComponent, error)
	Get(ctx context.Context, name string) (*EDPComponent, error)
}
