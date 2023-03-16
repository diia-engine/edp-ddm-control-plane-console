package jenkins

import (
	"context"
	"ddm-admin-console/service"
	"ddm-admin-console/service/k8s"
	"net/http"

	"github.com/bndr/gojenkins"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgScheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

type Service struct {
	service.UserConfig
	Config
	k8sClient client.Client
	scheme    *runtime.Scheme
	goJenkins *gojenkins.Jenkins
	k8s       k8s.ServiceInterface
}

type Config struct {
	AdminSecretName string
	APIUrl          string
	Namespace       string
}

func Make(s *runtime.Scheme, k8sConfig *rest.Config, k8s k8s.ServiceInterface, cnf Config) (*Service, error) {
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

	svc := Service{
		Config:    cnf,
		k8sClient: cl,
		scheme:    s,
		UserConfig: service.UserConfig{
			RestConfig: k8sConfig,
		},
		k8s: k8s,
	}

	if err := svc.initJenkinsAPIClient(context.Background(), k8s); err != nil {
		return nil, errors.Wrap(err, "unable to ini jenkins api client")
	}

	return &svc, nil
}

func (s *Service) initJenkinsAPIClient(ctx context.Context, k8s k8s.ServiceInterface) error {
	rsp, err := k8s.GetSecretKeys(ctx, s.Namespace, s.AdminSecretName,
		[]string{"username", "password"})
	if err != nil {
		return errors.Wrap(err, "unable to get jenkins admin secret")
	}

	jenkinsClient := gojenkins.CreateJenkins(http.DefaultClient, s.APIUrl, rsp["username"], rsp["password"])
	j, err := jenkinsClient.Init(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to init jenkins client")
	}

	s.goJenkins = j
	return nil
}

func (s *Service) GetJobStatus(ctx context.Context, jobName string) (string, int64, error) {
	j, err := s.goJenkins.GetJob(ctx, jobName)
	if err != nil {
		return "", 0, errors.Wrap(err, "unable to get job")
	}

	lastBuild, err := j.GetLastBuild(ctx)
	if err != nil {
		return "", 0, errors.Wrap(err, "unable to get last build")
	}

	if _, err := lastBuild.Poll(ctx); err != nil {
		return "", 0, errors.Wrap(err, "unable to poll last build")
	}

	return lastBuild.GetResult(), lastBuild.GetBuildNumber(), nil
}

func (s *Service) CreateJobBuildRunRaw(ctx context.Context, jb *JenkinsJobBuildRun) error {
	jb.Namespace = s.Namespace

	if err := s.k8sClient.Create(ctx, jb, &client.CreateOptions{}); err != nil {
		return errors.Wrap(err, "unable to create job build run")
	}

	return nil
}

func (s *Service) CreateJobBuildRun(ctx context.Context, name, jobPath string, jobParams map[string]string) error {
	jbr := JenkinsJobBuildRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: s.Namespace,
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

	if err := s.CreateJobBuildRunRaw(ctx, &jbr); err != nil {
		return errors.Wrap(err, "unableto create jenkins job build run")
	}

	return nil
}

func (s *Service) ServiceForContext(ctx context.Context) (ServiceInterface, error) {
	userConfig, changed := s.UserConfig.CreateConfig(ctx)
	if !changed {
		return s, nil
	}

	svc, err := Make(s.scheme, userConfig, s.k8s, s.Config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create service for context")
	}

	return svc, nil
}
