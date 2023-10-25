package registry

import (
	"ddm-admin-console/service/vault"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	keycloakIndex                    = "keycloak"
	registryCitizenClientIdIndex     = "clientId"
	registryCitizenClientSecretIndex = "clientSecret"
	portalsIndex                     = "portals"
)

func (a *App) prepareCitizenAuthSettings(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {
	valuesChanged := false

	if r.RegistryCitizenAuth != "" {
		var requestDataCitizenAuthSettings struct {
			KeycloakAuthFlowsCitizenAuthFlow
			Portals Portals `json:"portals"`
		}
		requestDataCitizenAuthSettings.Portals = values.Portals
		if err := json.Unmarshal([]byte(r.RegistryCitizenAuth), &requestDataCitizenAuthSettings); err != nil {
			return false, fmt.Errorf("unable to decode citizen auth settings %w", err)
		}
		if requestDataCitizenAuthSettings.RegistryIdGovUa.ClientSecret != "" || requestDataCitizenAuthSettings.RegistryIdGovUa.ClientId != "" {
			idGovUaCredsChanged, oldSecret, err := a.citizenIdGovUASecretChanged(
				values.Keycloak.CitizenAuthFlow.RegistryIdGovUa.ClientSecret,
				requestDataCitizenAuthSettings.RegistryIdGovUa.ClientId,
				requestDataCitizenAuthSettings.RegistryIdGovUa.ClientSecret,
			)
			if err != nil {
				return false, fmt.Errorf("unable to get secret, %w", err)
			}
			if idGovUaCredsChanged {
				var clientSecret string
				if requestDataCitizenAuthSettings.RegistryIdGovUa.ClientSecret == emptyClientSecret {
					clientSecret = oldSecret
				} else {
					clientSecret = requestDataCitizenAuthSettings.RegistryIdGovUa.ClientSecret
				}
				vaultPath := a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", "citizen-id-gov-ua-client-info", time.Now().Format("20060201T150405Z")))
				secrets[vaultPath] = map[string]any{
					registryCitizenClientIdIndex:     requestDataCitizenAuthSettings.RegistryIdGovUa.ClientId,
					registryCitizenClientSecretIndex: clientSecret,
				}
				requestDataCitizenAuthSettings.RegistryIdGovUa.ClientSecret = vaultPath
				requestDataCitizenAuthSettings.RegistryIdGovUa.ClientId = vaultPath
				valuesChanged = true
			} else {
				requestDataCitizenAuthSettings.RegistryIdGovUa.ClientSecret = values.Keycloak.CitizenAuthFlow.RegistryIdGovUa.ClientSecret
				requestDataCitizenAuthSettings.RegistryIdGovUa.ClientId = values.Keycloak.CitizenAuthFlow.RegistryIdGovUa.ClientId
			}
		}

		newCitizenAuthFlow := KeycloakAuthFlowsCitizenAuthFlow{
			EDRCheck:        requestDataCitizenAuthSettings.EDRCheck,
			AuthType:        requestDataCitizenAuthSettings.AuthType,
			Widget:          requestDataCitizenAuthSettings.Widget,
			RegistryIdGovUa: requestDataCitizenAuthSettings.RegistryIdGovUa,
		}

		if !reflect.DeepEqual(newCitizenAuthFlow, values.Keycloak.CitizenAuthFlow) {
			values.Keycloak.CitizenAuthFlow = newCitizenAuthFlow
			valuesChanged = true
		}

		if !reflect.DeepEqual(values.Portals.Citizen, requestDataCitizenAuthSettings.Portals.Citizen) {
			values.Portals.Citizen = requestDataCitizenAuthSettings.Portals.Citizen
			valuesChanged = true
		}

		if valuesChanged {
			values.OriginalYaml[keycloakIndex] = values.Keycloak
			values.OriginalYaml[portalsIndex] = values.Portals
			return true, nil
		}
	}

	return false, nil
}

func (a *App) citizenIdGovUASecretChanged(path string, inputClientId string, inputClientSecret string) (bool, string, error) {
	data, err := a.Vault.Read(path)
	if err != nil && !errors.Is(err, vault.ErrSecretIsNil) {
		return false, "", fmt.Errorf("unable to get secret, %w", err)
	}

	if errors.Is(err, vault.ErrSecretIsNil) {
		return true, "", nil
	}

	clID, ok := data[idGovUASecretClientID]
	if !ok {
		return true, "", nil
	}

	clSecret, ok := data[idGovUASecretClientSecret]
	if !ok {
		return true, "", nil
	}
	clSecretString, ok := clSecret.(string)
	if !ok {
		return true, "", nil
	}

	if clID != inputClientId || (clSecretString != inputClientSecret && inputClientSecret != emptyClientSecret) {
		return true, clSecretString, nil
	}

	return false, "", nil
}
