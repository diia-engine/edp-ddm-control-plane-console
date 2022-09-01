package keycloak

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func Make(s *runtime.Scheme, k8sConfig *rest.Config, namespace string) (*Service, error) {
	builder := pkgScheme.Builder{GroupVersion: schema.GroupVersion{Group: "v1.edp.epam.com", Version: "v1alpha1"}}
	builder.Register(&KeycloakRealmUser{}, &KeycloakRealmUserList{})

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
		UserConfig: service.UserConfig{
			RestConfig: k8sConfig,
		},
		restConfig: k8sConfig,
	}, nil
}

func (s *Service) GetUsers(ctx context.Context) ([]KeycloakRealmUser, error) {
	var lst KeycloakRealmUserList

	if err := s.k8sClient.List(ctx, &lst, &client.ListOptions{
		Namespace: s.namespace,
	}); err != nil {
		return nil, errors.Wrap(err, "unable to list users")
	}

	return lst.Items, nil
}

func (s *Service) GetUsersByRealm(ctx context.Context, realmName string) ([]KeycloakRealmUser, error) {
	usrs, err := s.GetUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get users")
	}

	var filteredUsers []KeycloakRealmUser
	for _, u := range usrs {
		if u.Spec.Realm == realmName {
			filteredUsers = append(filteredUsers, u)
		}
	}

	return filteredUsers, nil
}

func (s *Service) CreateUser(ctx context.Context, user *KeycloakRealmUser) error {
	if err := s.k8sClient.Create(ctx, user); err != nil {
		return errors.Wrap(err, "unable to create realm user")
	}

	return nil
}

func (s *Service) UpdateUser(ctx context.Context, user *KeycloakRealmUser) error {
	if err := s.k8sClient.Update(ctx, user); err != nil {
		return errors.Wrap(err, "unable to update user")
	}

	return nil
}

func (s *Service) DeleteUser(ctx context.Context, user *KeycloakRealmUser) error {
	if err := s.k8sClient.Delete(ctx, user); err != nil {
		return errors.Wrap(err, "unable to delete user")
	}

	return nil
}
