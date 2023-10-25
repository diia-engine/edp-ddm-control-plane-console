package k8s

import (
	"context"

	"ddm-admin-console/service"

	"github.com/pkg/errors"
	authorizationv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Service struct {
	service.UserConfig
	clientSet  *kubernetes.Clientset
	namespace  string
	restConfig *rest.Config
}

func Make(config *rest.Config, namespace string) (*Service, error) {
	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s client")
	}

	return &Service{
		UserConfig: service.UserConfig{
			RestConfig: config,
		},
		clientSet:  cs,
		namespace:  namespace,
		restConfig: config,
	}, nil
}

func (s *Service) RecreateSecret(secretName string, data map[string][]byte) error {
	if _, err := s.clientSet.CoreV1().Secrets(s.namespace).
		Get(context.Background(), secretName, metav1.GetOptions{}); err == nil {
		if err := s.clientSet.CoreV1().Secrets(s.namespace).
			Delete(context.Background(), secretName, metav1.DeleteOptions{}); err != nil {
			return errors.Wrapf(err, "unable to delete secret: %s", secretName)
		}
	}

	sc := v1.Secret{
		Data: data,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: s.namespace,
		},
	}

	if _, err := s.clientSet.CoreV1().Secrets(s.namespace).
		Create(context.Background(), &sc, metav1.CreateOptions{}); err != nil {
		return errors.Wrap(err, "unable to create secret")
	}

	return nil
}

func (s *Service) ServiceForContext(ctx context.Context) (ServiceInterface, error) {
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

func (s *Service) GetSecretFromNamespace(ctx context.Context, name, namespace string) (*v1.Secret, error) {
	secret, err := s.clientSet.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get secret: %s, ns: %s", name, namespace)
	}

	return secret, nil
}

func (s *Service) GetSecret(name string) (*v1.Secret, error) {
	secret, err := s.clientSet.CoreV1().Secrets(s.namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get secret: %s", name)
	}

	return secret, nil
}

func (s *Service) GetConfigMap(ctx context.Context, name, namespace string) (*v1.ConfigMap, error) {
	cm, err := s.clientSet.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get config map")
	}

	return cm, nil
}

func (s *Service) CreateConfigMap(ctx context.Context, cm *v1.ConfigMap, namespace string) error {
	_, err := s.clientSet.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrap(err, "unable to create config map")
	}

	return nil
}

func (s *Service) GetSecretKeys(ctx context.Context, namespace, name string, keys []string) (map[string]string, error) {
	sec, err := s.GetSecretFromNamespace(ctx, name, namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get secret")
	}

	result := make(map[string]string)

	for _, k := range keys {
		val, ok := sec.Data[k]
		if !ok {
			return nil, errors.Errorf("key [%s] from secret [%s] in namespace [%s] not found", k, name, namespace)
		}

		result[k] = string(val)
	}

	return result, nil
}

func (s *Service) GetSecretKey(ctx context.Context, namespace, name, key string) (string, error) {
	sec, err := s.GetSecretFromNamespace(ctx, name, namespace)
	if err != nil {
		return "", errors.Wrap(err, "unable to get secret")
	}

	val, ok := sec.Data[key]
	if !ok {
		return "", errors.Errorf("key [%s] from secret [%s] in namespace [%s] not found", key, name, namespace)
	}

	return string(val), nil
}

func (s *Service) CanI(group, resource, verb, name string) (bool, error) {
	review := authorizationv1.SelfSubjectAccessReview{
		//ObjectMeta: metav1.ObjectMeta{Namespace: s.namespace},
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: s.namespace,
				Verb:      verb,
				Group:     group,
				Resource:  resource,
				Name:      name,
			},
		},
	}

	r, err := s.clientSet.AuthorizationV1().SelfSubjectAccessReviews().Create(context.Background(), &review,
		metav1.CreateOptions{})
	if err != nil {
		return false, errors.Wrap(err, "unable to create self subject review")
	}

	return r.Status.Allowed, nil
}
