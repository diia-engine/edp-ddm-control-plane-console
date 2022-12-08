package registry

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authTypeNoAuth          = "NO_AUTH"
	authTypeAuthToken       = "AUTH_TOKEN"
	authTypeBearer          = "BEARER"
	authTypeBasic           = "BASIC"
	authTypeAuthTokenBearer = "AUTH_TOKEN+BEARER"
)

type RegistryExternalSystemForm struct {
	RegistryName        string `form:"external-system-registry-name"`
	URL                 string `form:"external-system-url"`
	Protocol            string `form:"external-system-protocol"`
	AuthType            string `form:"external-system-auth-type"`
	AuthURI             string `form:"external-system-auth-uri"`
	AccessTokenJSONPath string `form:"external-system-auth-access-token-json-path"`
	AuthSecret          string `form:"external-system-auth-secret"`
	AuthUsername        string `form:"external-system-auth-username"`
}

func (f RegistryExternalSystemForm) ToNestedForm(vaultRegistryPath string) ExternalSystem {
	es := ExternalSystem{
		URL:      f.URL,
		Protocol: f.Protocol,
		Auth: map[string]string{
			"type": f.AuthType,
		},
	}

	if f.AuthType != authTypeNoAuth {
		es.Auth["secret"] = fmt.Sprintf("%s/external-systems/%s", vaultRegistryPath,
			f.RegistryName)
	}

	if f.AuthType == authTypeAuthTokenBearer {
		es.Auth["auth-uri"] = f.AuthURI
		es.Auth["access-token-json-path"] = f.AccessTokenJSONPath
	}

	return es
}

func (a *App) prepareRegistryExternalSystemsConfig(ctx *gin.Context, r *registry, values map[string]interface{},
	secrets map[string]map[string]interface{}) error {

	registryExternalSystems := strings.Split(a.Config.RegistryDefaultExternalSystems, ",")
	if len(registryExternalSystems) == 0 {
		return nil
	}

	_, ok := values[externalSystemsKey]
	if ok {
		return nil
	}

	externalSystems := make(map[string]interface{})

	for _, res := range registryExternalSystems {
		resParts := strings.Split(res, ":")
		if len(resParts) < 2 {
			continue
		}

		externalSystems[resParts[0]] = map[string]string{
			"type":     resParts[1],
			"protocol": "REST",
		}
	}

	if len(externalSystems) > 0 {
		values[externalSystemsKey] = externalSystems
	}

	return nil
}
