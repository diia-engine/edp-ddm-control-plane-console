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
	portalsIndex                 = "portals"
)

func (a *App) prepareCitizenAuthSettings(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {
	if r.RegistryCitizenAuth != "" {
		var citizenAuthSettings struct {
			KeycloakAuthFlowsCitizenAuthFlow
			Portals Portals `json:"portals"`
		}
		citizenAuthSettings.Portals = values.Portals
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
		values.Portals.Citizen = citizenAuthSettings.Portals.Citizen
		values.OriginalYaml[keycloakIndex] = values.Keycloak
		values.OriginalYaml[portalsIndex] = values.Portals
		return true, nil
	}
	return false, nil
}
