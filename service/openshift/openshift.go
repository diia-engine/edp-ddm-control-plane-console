package openshift

import (
	"context"
	"crypto/tls"
	"ddm-admin-console/service/k8s"

	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	"k8s.io/client-go/rest"
)

type Service struct {
	restConfig  *rest.Config
	restyClient *resty.Client
	k8sService  k8s.ServiceInterface
}

func Make(restConfig *rest.Config, k8sService k8s.ServiceInterface) (*Service, error) {
	svc := Service{
		restConfig: restConfig,
		restyClient: resty.New().SetHostURL(restConfig.Host).
			SetHeader("Accept", "application/json"),
		k8sService: k8sService,
	}

	svc.restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &svc, nil
}

type Metadata struct {
	Name string `json:"name"`
}

type User struct {
	Metadata Metadata `json:"metadata"`
	FullName string   `json:"fullName"`
}

type Status struct {
	Type string `json:"type"`
}

type ClusterStatus struct {
	PlatformStatus Status `json:"platformStatus"`
}

type ClusterInfrastructure struct {
	Status ClusterStatus `json:"status"`
}

func (s *Service) GetMe(ctx context.Context) (*User, error) {
	accessToken := ctx.Value("access-token")
	if accessToken == nil {
		return nil, errors.New("no access token in context")
	}

	var u User
	rsp, err := s.restyClient.R().SetAuthToken(accessToken.(string)).SetResult(&u).Get("/apis/user.openshift.io/v1/users/~")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get user")
	}

	if rsp.StatusCode() >= 300 {
		return nil, errors.Errorf("code: %d body: %s", rsp.StatusCode(), rsp.String())
	}

	return &u, nil
}

func (s *Service) GetInfrastructureCluster(ctx context.Context) (*ClusterInfrastructure, error) {
	accessToken := ctx.Value("access-token")
	if accessToken == nil {
		return nil, errors.New("no access token in context")
	}

	var result ClusterInfrastructure

	_, err := s.restyClient.R().SetAuthToken(accessToken.(string)).SetResult(&result).Get("/apis/config.openshift.io/v1/infrastructures/cluster")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get infrastructures of cluster")
	}

	return &result, nil
}
