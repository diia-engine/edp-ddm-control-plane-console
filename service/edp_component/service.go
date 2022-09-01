package edpcomponent

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgScheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

type Service struct {
	k8sClient client.Client
	scheme    *runtime.Scheme
	namespace string
}

func Make(s *runtime.Scheme, k8sConfig *rest.Config, namespace string) (*Service, error) {
	builder := pkgScheme.Builder{GroupVersion: schema.GroupVersion{Group: "v1.edp.epam.com", Version: "v1alpha1"}}
	builder.Register(&EDPComponent{}, &EDPComponentList{})

	if err := builder.AddToScheme(s); err != nil {
		return nil, errors.Wrap(err, "error during builder add to scheme")
	}

	cl, err := client.New(k8sConfig, client.Options{
		Scheme: s,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s jenkins client")
	}

	return &Service{
		k8sClient: cl,
		scheme:    s,
		namespace: namespace,
	}, nil
}

func (s *Service) GetAll(ctx context.Context) ([]EDPComponent, error) {
	var lst EDPComponentList
	if err := s.k8sClient.List(ctx, &lst, &client.ListOptions{Namespace: s.namespace}); err != nil {
		return nil, errors.Wrap(err, "unable to list edp component")
	}

	return lst.Items, nil
}

func (s *Service) GetAllNamespace(ctx context.Context, ns string, onlyVisible bool) ([]EDPComponent, error) {
	var lst EDPComponentList
	if err := s.k8sClient.List(ctx, &lst, &client.ListOptions{Namespace: ns}); err != nil {
		return nil, errors.Wrap(err, "unable to list edp component")
	}

	if !onlyVisible {
		return lst.Items, nil
	}

	items := make([]EDPComponent, 0, len(lst.Items))
	for _, v := range lst.Items {
		if v.Spec.Visible {
			items = append(items, v)
		}
	}

	return items, nil
}

func (s *Service) Get(ctx context.Context, name string) (*EDPComponent, error) {
	var comp EDPComponent
	if err := s.k8sClient.Get(ctx, types.NamespacedName{
		Name: name, Namespace: s.namespace}, &comp); err != nil {
		return nil, errors.Wrapf(err, "unable to get edp component by name: %s", name)
	}

	return &comp, nil
}
