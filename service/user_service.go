package service

import (
	"context"

	"k8s.io/client-go/rest"
)

const (
	UserTokenKey = "access-token"
)

type UserConfig struct {
	RestConfig *rest.Config
}

func (s *UserConfig) CreateConfig(ctx context.Context) (config *rest.Config, changed bool) {
	tok := ctx.Value(UserTokenKey)
	if tok == nil {
		return s.RestConfig, false
	}

	tokString, ok := tok.(string)
	if !ok {
		return s.RestConfig, false
	}

	userConfig := rest.AnonymousClientConfig(s.RestConfig)
	userConfig.BearerToken = tokString

	return userConfig, true
}
