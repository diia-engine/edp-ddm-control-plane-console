package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"ddm-admin-console/service/k8s"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type OAuth2 struct {
	clientID     string
	secret       string
	discoveryURL string
	redirectURL  string
	httpClient   *http.Client
	providerInfo providerInfo
	Config       *oauth2.Config
}

type providerInfo struct {
	Issuer          string   `json:"issuer"`
	AuthURL         string   `json:"authorization_endpoint"`
	TokenURL        string   `json:"token_endpoint"`
	ScopesSupported []string `json:"scopes_supported"`
}

func InitOauth2(clientID, secret, discoveryURL, redirectURL string, httpClient *http.Client) (*OAuth2, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	oa2 := OAuth2{
		clientID:     clientID,
		secret:       secret,
		discoveryURL: discoveryURL,
		redirectURL:  redirectURL,
		httpClient:   httpClient,
	}

	if !strings.Contains(discoveryURL, ".well-known") {
		discoveryURL = strings.TrimSuffix(discoveryURL, "/") + "/.well-known/oauth-authorization-server"
	}

	rsp, err := httpClient.Get(discoveryURL)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get discovery url")
	}
	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, errors.Errorf("unable to read response body: %v", err)
	}

	if rsp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%s: %s", rsp.Status, string(body))
	}

	if err := json.Unmarshal(body, &oa2.providerInfo); err != nil {
		return nil, errors.Wrap(err, "unable to unmarshal discovery body")
	}

	oa2.initConfig()

	return &oa2, nil
}

func (o *OAuth2) initConfig() {
	o.Config = &oauth2.Config{
		ClientID:     o.clientID,
		ClientSecret: o.secret,
		Scopes:       o.providerInfo.ScopesSupported,
		Endpoint: oauth2.Endpoint{
			AuthURL:  o.providerInfo.AuthURL,
			TokenURL: o.providerInfo.TokenURL,
		},
		RedirectURL: o.redirectURL,
	}
}

func (o *OAuth2) UseInternalTokenService(ctx context.Context, serviceHost string, k8sService k8s.ServiceInterface) error {
	tokenUrl, err := url.Parse(o.Config.Endpoint.TokenURL)
	if err != nil {
		return errors.Wrap(err, "unable to parse token url")
	}

	tokenUrl.Host = serviceHost
	o.Config.Endpoint.TokenURL = tokenUrl.String()

	cm, err := k8sService.GetConfigMap(ctx, "openshift-service-ca.crt", "openshift-config-managed")
	if err != nil {
		return errors.Wrap(err, "unable to get openshift ca config map")
	}

	ca, ok := cm.Data["service-ca.crt"]
	if !ok {
		return errors.New("no service ca found in config map")
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(ca)) {
		return errors.New("unable to append certs from PEM")
	}
	o.httpClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{
		RootCAs: caCertPool,
	}}

	return nil
}

func (o *OAuth2) AuthCodeURL() string {
	return o.Config.AuthCodeURL(fmt.Sprintf("state-%d", time.Now().Unix()))
}

func (o *OAuth2) GetTokenClient(ctx context.Context, code string) (token *oauth2.Token, oauthClient *http.Client,
	err error) {

	ctx = context.WithValue(ctx, oauth2.HTTPClient, o.httpClient)

	token, err = o.Config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to get access token")
	}

	oauthClient = o.Config.Client(ctx, token)
	return
}

func (o *OAuth2) GetHTTPClient(ctx context.Context, token *oauth2.Token) *http.Client {
	return o.Config.Client(ctx, token)
}
