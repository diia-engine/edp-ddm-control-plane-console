package vault

import (
	"context"
	"ddm-admin-console/service/k8s"
	"errors"
	"fmt"
	"strings"

	hashiVault "github.com/hashicorp/vault/api"
)

type ServiceInterface interface {
	ReadRaw(path string) (*hashiVault.Secret, error)
	Read(path string) (map[string]interface{}, error)
	WriteRaw(path string, data map[string]interface{}) (*hashiVault.Secret, error)
	Write(path string, data map[string]interface{}) (*hashiVault.Secret, error)
}

var ErrSecretIsNil = errors.New("secret is nil")

type Config struct {
	SecretNamespace string
	SecretName      string
	SecretTokenKey  string
	KVEngineName    string
	APIAddr         string
}

type Service struct {
	l *hashiVault.Logical
}

func Make(cnf Config, k8s k8s.ServiceInterface) (*Service, error) {
	config := hashiVault.DefaultConfig()
	config.Address = cnf.APIAddr

	client, err := hashiVault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create vault client, %w", err)
	}

	token, err := k8s.GetSecretKey(context.Background(), cnf.SecretNamespace, cnf.SecretName, cnf.SecretTokenKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get token from secret, %w", err)
	}

	client.SetToken(token)

	return &Service{
		l: client.Logical(),
	}, nil
}

func ModifyVaultPath(path string) string {
	if strings.Contains(path, "/data/") {
		return path
	}

	pathParts := strings.Split(path, "/")
	pathParts = append(pathParts[:1], append([]string{"data"}, pathParts[1:]...)...)
	return strings.Join(pathParts, "/")
}

func (s *Service) ReadRaw(path string) (*hashiVault.Secret, error) {
	return s.l.Read(path)
}

func (s *Service) Read(path string) (map[string]interface{}, error) {
	sec, err := s.l.Read(ModifyVaultPath(path))
	if err != nil {
		return nil, fmt.Errorf("unable to read secret, %w", err)
	}

	if sec == nil {
		return nil, ErrSecretIsNil
	}

	currentSecretData, ok := sec.Data["data"]
	if !ok {
		return nil, ErrSecretIsNil
	}

	return currentSecretData.(map[string]interface{}), nil
}

func (s *Service) Write(path string, data map[string]interface{}) (*hashiVault.Secret, error) {
	return s.l.Write(ModifyVaultPath(path), map[string]interface{}{
		"data": data,
	})
}

func (s *Service) WriteRaw(path string, data map[string]interface{}) (*hashiVault.Secret, error) {
	return s.l.Write(path, data)
}
