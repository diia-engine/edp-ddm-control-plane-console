package openshift

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"ddm-admin-console/service/k8s"
	"net/http"

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

	if restConfig.TLSClientConfig.Insecure {
		svc.restyClient.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	} else {
		cm, err := k8sService.GetConfigMap(context.Background(), "openshift-global-ca", "openshift-controller-manager")
		if err != nil {
			return nil, errors.Wrap(err, "unable to get openshift ca config map")
		}

		ca, ok := cm.Data["ca-bundle.crt"]
		if !ok {
			return nil, errors.New("no service ca found in config map")
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(ca)) {
			return nil, errors.New("unable to append certs from PEM")
		}

		svc.restyClient.GetClient().Transport = &http.Transport{TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		}}
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
