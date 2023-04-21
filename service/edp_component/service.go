package edpcomponent

import (
	"context"
	"sort"

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

const (
	RegistryOperationalZone    = "registry-operational-zone"
	RegistryAdministrationZone = "registry-administration-zone"
	PlatformOperationalZone    = "platform-operational-zone"
	PlatformAdministrationZone = "platform-administration-zone"
	CPCDisplayName             = "control-plane-console/display-name"
	CPCDescription             = "control-plane-console/description"
	CPCDisplayVisible          = "control-plane-console/display-visible"
	CPCDisplayOrder            = "control-plane-console/display-order"
	CPCOperationalZone         = "control-plane-console/operational-zone"
)

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

func PrepareComponentItem(component EDPComponent) EDPComponentItem {
	return EDPComponentItem{
		Type:        component.Spec.Type,
		Url:         component.Spec.Url,
		Icon:        component.Spec.Icon,
		Title:       component.ObjectMeta.Annotations[CPCDisplayName],
		Description: component.ObjectMeta.Annotations[CPCDescription],
		Visible:     component.ObjectMeta.Annotations[CPCDisplayVisible],
	}
}

func (s *Service) SortComponents(components []EDPComponent) []EDPComponent {
	sort.Slice(components, func(i, j int) bool {
		return components[i].ObjectMeta.Annotations[CPCDisplayOrder] < components[j].ObjectMeta.Annotations[CPCDisplayOrder]
	})
	return components
}

func (s *Service) GetAll(ctx context.Context, onlyVisible bool) ([]EDPComponent, error) {
	var lst EDPComponentList
	if err := s.k8sClient.List(ctx, &lst, &client.ListOptions{Namespace: s.namespace}); err != nil {
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

func (s *Service) GetAllCategory(ctx context.Context, ns string) (map[string][]EDPComponentItem, error) {
	categories := make(map[string][]EDPComponentItem)
	platformComponents, err := s.GetAll(ctx, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list platform edp component")
	}

	registryComponents, err := s.GetAllNamespace(ctx, ns, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list edp component")
	}

	for _, val := range s.SortComponents(registryComponents) {
		var objectMetaAnnotations = val.ObjectMeta.Annotations
		if objectMetaAnnotations[CPCOperationalZone] == RegistryOperationalZone {
			categories[RegistryOperationalZone] = append(categories[RegistryOperationalZone], PrepareComponentItem(val))
		}

		if objectMetaAnnotations[CPCOperationalZone] == RegistryAdministrationZone {
			categories[RegistryAdministrationZone] = append(categories[RegistryAdministrationZone], PrepareComponentItem(val))
		}
	}

	for _, val := range s.SortComponents(platformComponents) {
		var objectMetaAnnotations = val.ObjectMeta.Annotations
		if objectMetaAnnotations[CPCOperationalZone] == PlatformOperationalZone {
			categories[PlatformOperationalZone] = append(categories[PlatformOperationalZone], PrepareComponentItem(val))
		}

		if objectMetaAnnotations[CPCOperationalZone] == PlatformAdministrationZone {
			categories[PlatformAdministrationZone] = append(categories[PlatformAdministrationZone], PrepareComponentItem(val))
		}
	}

	return categories, nil
}

func (s *Service) Get(ctx context.Context, name string) (*EDPComponent, error) {
	var comp EDPComponent
	if err := s.k8sClient.Get(ctx, types.NamespacedName{
		Name: name, Namespace: s.namespace}, &comp); err != nil {
		return nil, errors.Wrapf(err, "unable to get edp component by name: %s", name)
	}

	return &comp, nil
}
