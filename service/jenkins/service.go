package jenkins

import (
	"context"

	"ddm-admin-console/service"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgScheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

type Service struct {
	k8sClient client.Client
	scheme    *runtime.Scheme
	namespace string
	service.UserConfig
}

func Make(s *runtime.Scheme, k8sConfig *rest.Config, namespace string) (*Service, error) {
	builder := pkgScheme.Builder{GroupVersion: schema.GroupVersion{Group: "v2.edp.epam.com", Version: "v1alpha1"}}
	builder.Register(&JenkinsJobBuildRun{}, &JenkinsJobBuildRunList{})

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
	}, nil
}

func (s *Service) CreateJobBuildRunRaw(jb *JenkinsJobBuildRun) error {
	jb.Namespace = s.namespace

	if err := s.k8sClient.Create(context.Background(), jb, &client.CreateOptions{}); err != nil {
		return errors.Wrap(err, "unable to create job build run")
	}

	return nil
}

func (s *Service) CreateJobBuildRun(name, jobPath string, jobParams map[string]string) error {
	jbr := JenkinsJobBuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.namespace,
			Name:      name,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1alpha1",
			Kind:       "JenkinsJobBuildRun",
		},
		Spec: JenkinsJobBuildRunSpec{
			Params:  jobParams,
			JobPath: jobPath,
			Retry:   5,
		},
	}

	if err := s.CreateJobBuildRunRaw(&jbr); err != nil {
		return errors.Wrap(err, "unableto create jenkins job build run")
	}

	return nil
}

func (s *Service) ServiceForContext(ctx context.Context) (ServiceInterface, error) {
	userConfig, changed := s.UserConfig.CreateConfig(ctx)
	if !changed {
		return s, nil
	}

	svc, err := Make(s.scheme, userConfig, s.namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create service for context")
	}

	return svc, nil
}
