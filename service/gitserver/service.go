package gitserver

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgScheme "sigs.k8s.io/controller-runtime/pkg/scheme"

	"ddm-admin-console/service"
)

type Service struct {
	service.UserConfig
	k8sClient  client.Client
	scheme     *runtime.Scheme
	namespace  string
	restConfig *rest.Config
}

func New(sch *runtime.Scheme, k8sConfig *rest.Config, namespace string) (*Service, error) {
	builder := pkgScheme.Builder{
		GroupVersion: schema.GroupVersion{
			Group:   "v2.edp.epam.com",
			Version: "v1alpha1",
		},
	}

	builder.Register(&GitServer{}, &GitServerList{})

	if err := builder.AddToScheme(sch); err != nil {
		return nil, fmt.Errorf("failed to add builder to scheme: %w", err)
	}

	cl, err := client.New(
		k8sConfig,
		client.Options{
			Scheme: sch,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to build k8s client: %w", err)
	}

	return &Service{
		k8sClient: cl,
		scheme:    sch,
		namespace: namespace,
		UserConfig: service.UserConfig{
			RestConfig: k8sConfig,
		},
		restConfig: k8sConfig,
	}, nil
}

func (s *Service) Get(name string) (*GitServer, error) {
	var gitServer GitServer

	if err := s.k8sClient.Get(
		context.Background(),
		types.NamespacedName{
			Namespace: s.namespace,
			Name:      name,
		},
		&gitServer,
	); err != nil {
		return nil, fmt.Errorf("failed to get codebase %s: %w", name, err)
	}

	return &gitServer, nil
}
