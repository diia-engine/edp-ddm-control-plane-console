package vault

import (
	"context"
	"ddm-admin-console/service/k8s"

	hashiVault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
)

type ServiceInterface interface {
	Read(path string) (*hashiVault.Secret, error)
	Write(path string, data map[string]interface{}) (*hashiVault.Secret, error)
}

type Config struct {
	SecretNamespace string
	SecretName      string
	SecretTokenKey  string
	KVEngineName    string
	APIAddr         string
}

func Make(cnf Config, k8s k8s.ServiceInterface) (*hashiVault.Logical, error) {
	config := hashiVault.DefaultConfig()
	config.Address = cnf.APIAddr

	client, err := hashiVault.NewClient(config)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create vault client")
	}

	token, err := k8s.GetSecretKey(context.Background(), cnf.SecretNamespace, cnf.SecretName, cnf.SecretTokenKey)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get token from secret")
	}

	client.SetToken(token)

	return client.Logical(), nil
}
