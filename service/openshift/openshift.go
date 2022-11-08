package openshift

import (
	"context"
	"crypto/tls"

	"github.com/pkg/errors"
	"gopkg.in/resty.v1"
	"k8s.io/client-go/rest"
)

type Service struct {
	restConfig  *rest.Config
	restyClient *resty.Client
}

func Make(restConfig *rest.Config) (*Service, error) {
	svc := Service{
		restConfig: restConfig,
		restyClient: resty.New().SetHostURL(restConfig.Host).
			SetHeader("Accept", "application/json"),
	}

	if restConfig.TLSClientConfig.Insecure {
		svc.restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	}

	return &svc, nil
}

type User struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	FullName string `json:"fullName"`
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
