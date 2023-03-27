package edpcomponent

import "context"

type ServiceInterface interface {
	GetAll(ctx context.Context, onlyVisible bool) ([]EDPComponent, error)
	GetAllNamespace(ctx context.Context, ns string, onlyVisible bool) ([]EDPComponent, error)
	Get(ctx context.Context, name string) (*EDPComponent, error)
	GetAllCategory(ctx context.Context, ns string) (map[string][]EDPComponentItem, error)
}
