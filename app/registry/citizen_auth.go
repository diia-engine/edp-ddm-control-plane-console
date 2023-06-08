package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	keycloakIndex                = "keycloak"
	RegistryCitizenIdGovUaSecret = "RegistryCitizenIdGovUaSecret"
	citizenPortalIndex           = "citizenPortal"
)

func (a *App) prepareCitizenAuthSettings(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
	if r.RegistryCitizenAuth != "" {
		var citizenAuthSettings struct {
			KeycloakAuthFlowsCitizenAuthFlow
			CitizenPortal CitizenPortal `json:"citizenPortal"`
		}
		if err := json.Unmarshal([]byte(r.RegistryCitizenAuth), &citizenAuthSettings); err != nil {
			return false, fmt.Errorf("unable to decode citizen auth settings %w", err)
		}
		if citizenAuthSettings.RegistryIdGovUa.ClientSecret != "" {
			vaultPath := a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", "registry-id-gov-ua-secret", time.Now().Format("20060201T150405Z")))
			citizenAuthSettings.RegistryIdGovUa.ClientSecret = vaultPath
			secrets[vaultPath] = map[string]interface{}{
				RegistryCitizenIdGovUaSecret: values.Keycloak.CitizenAuthFlow.RegistryIdGovUa.ClientSecret,
			}
		} else {
			citizenAuthSettings.RegistryIdGovUa.ClientSecret = values.Keycloak.CitizenAuthFlow.RegistryIdGovUa.ClientSecret
		}
		values.Keycloak.CitizenAuthFlow = KeycloakAuthFlowsCitizenAuthFlow{
			EDRCheck:        citizenAuthSettings.EDRCheck,
			AuthType:        citizenAuthSettings.AuthType,
			Widget:          citizenAuthSettings.Widget,
			RegistryIdGovUa: citizenAuthSettings.RegistryIdGovUa,
		}
		values.CitizenPortal = citizenAuthSettings.CitizenPortal
		values.OriginalYaml[keycloakIndex] = values.Keycloak
		values.OriginalYaml[citizenPortalIndex] = values.CitizenPortal
		return true, nil
	}
	return nil
}
