package openshift

import (
	"context"
	"ddm-admin-console/service"

	openshiftV1 "github.com/openshift/api/user/v1"
	userV1Client "github.com/openshift/client-go/user/clientset/versioned/typed/user/v1"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

const (
	UserTokenKey = "access-token"
)

type Service struct {
	service.UserConfig
	userV1Client *userV1Client.UserV1Client
	namespace    string
}

func Make(restConfig *rest.Config, namespace string) (*Service, error) {
	svc := Service{
		UserConfig: service.UserConfig{
			RestConfig: restConfig,
		},
		namespace: namespace,
	}

	userCl, err := userV1Client.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init oc user client")
	}

	svc.userV1Client = userCl

	return &svc, nil
}

func (s *Service) serviceForContext(ctx context.Context) (*Service, error) {
	userConfig, changed := s.UserConfig.CreateConfig(ctx)
	if !changed {
		return s, nil
	}

	svc, err := Make(userConfig, s.namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create service for context")
	}

	return svc, nil
}

func (s *Service) ServiceForContext(ctx context.Context) (ServiceInterface, error) {
	return s.serviceForContext(ctx)
}

func (s *Service) GetMe(ctx context.Context) (*openshiftV1.User, error) {
	svc, err := s.serviceForContext(ctx)
	if err != nil {
		return nil, err
	}

	usr, err := svc.userV1Client.Users().Get(ctx, "~", v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user")
	}

	return usr, nil
}
