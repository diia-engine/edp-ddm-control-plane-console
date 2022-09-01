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

	//if err := setupKVEngine(token, &cnf); err != nil {
	//	return nil, errors.Wrap(err, "unable to setup kv engine")
	//}

	return client.Logical(), nil
}

//func setupKVEngine(token string, cnf *Config) error {
//	cl := resty.New().SetHeader("X-Vault-Token", token).SetHostURL(cnf.APIAddr)
//
//	var data map[string]interface{}
//	rsp, err := cl.NewRequest().SetResult(&data).Get("v1/sys/mounts")
//	if err != nil {
//		return errors.Wrap(err, "unable to get secret config")
//	}
//
//	if rsp.StatusCode() >= 300 {
//		log.Println(rsp.String(), rsp.StatusCode())
//		return errors.Errorf("wrong status: %d, body: %s", rsp.StatusCode(), rsp.String())
//	}
//
//	if _, ok := data[fmt.Sprintf("%s/", cnf.KVEngineName)]; !ok {
//		rsp, err = cl.NewRequest().SetBody(map[string]interface{}{
//			"type": "kv",
//		}).Post(fmt.Sprintf("v1/sys/mounts/%s", url.PathEscape(cnf.KVEngineName)))
//
//		if err != nil {
//			return errors.Wrap(err, "unable to create kv engine")
//		}
//
//		if rsp.StatusCode() >= 300 {
//			return errors.Errorf("wrong status: %d, body: %s", rsp.StatusCode(), rsp.String())
//		}
//	}
//
//	return nil
//}
