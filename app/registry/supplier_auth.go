package registry

import (
	"ddm-admin-console/service/vault"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	supAuthBrowserFlowWidget  = "dso-officer-auth-flow"
	supAuthBrowserFlowIdGovUa = "id-gov-ua-officer-redirector"
	idGovUASecretPath         = "officer-id-gov-ua-client-info"
	idGovUASecretClientID     = "clientId"
	idGovUASecretClientSecret = "clientSecret"
	emptyClientSecret         = "*****"
)

func (a *App) prepareSupplierAuthConfig(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) (bool, error) {

	if r.SupAuthBrowserFlow == "" {
		return false, nil
	}

	valuesChanged := values.Keycloak.Realms.OfficerPortal.SelfRegistration != (r.SelfRegistration == "on") ||
		values.Keycloak.Realms.OfficerPortal.BrowserFlow != r.SupAuthBrowserFlow

	values.Keycloak.Realms.OfficerPortal.SelfRegistration = r.SelfRegistration == "on"
	values.Keycloak.Realms.OfficerPortal.BrowserFlow = r.SupAuthBrowserFlow

	if r.SupAuthBrowserFlow == supAuthBrowserFlowWidget {
		widgetHeight, err := strconv.ParseInt(r.SupAuthWidgetHeight, 10, 32)
		if err != nil {
			return false, fmt.Errorf("unable to decode int, err: %w", err)
		}

		if !valuesChanged {
			valuesChanged = values.Keycloak.AuthFlows.OfficerAuthFlow.WidgetHeight != int(widgetHeight) ||
				values.SignWidget.URL != r.SupAuthURL
		}

		values.Keycloak.AuthFlows.OfficerAuthFlow.WidgetHeight = int(widgetHeight)
		values.SignWidget.URL = r.SupAuthURL
	} else if r.SupAuthBrowserFlow == supAuthBrowserFlowIdGovUa {
		if !valuesChanged {
			valuesChanged = values.Keycloak.IdentityProviders.IDGovUA.URL != r.SupAuthURL
		}

		values.Keycloak.IdentityProviders.IDGovUA.URL = r.SupAuthURL
		if r.SupAuthClientID != "" && r.SupAuthClientSecret != "" {
			secretPath := a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", idGovUASecretPath,
				time.Now().Format("20060201T150405Z")))

			idGovUaCredsChanged, oldClientSecret, err := a.idGovUASecretChanged(values.Keycloak.IdentityProviders.IDGovUA.SecretKey, r)
			if err != nil {
				return false, fmt.Errorf("unable to get secret, %w", err)
			}

			if idGovUaCredsChanged {
				var clientSecret string
				if r.SupAuthClientSecret == emptyClientSecret {
					clientSecret = oldClientSecret
				} else {
					clientSecret = r.SupAuthClientSecret
				}
				secrets[secretPath] = map[string]interface{}{
					idGovUASecretClientID:     r.SupAuthClientID,
					idGovUASecretClientSecret: clientSecret,
				}

				values.Keycloak.IdentityProviders.IDGovUA.SecretKey = secretPath
				valuesChanged = true
			}
		}

	}

	if !valuesChanged {
		return false, nil
	}

	values.OriginalYaml["signWidget"] = values.SignWidget

	keycloakInterface, ok := values.OriginalYaml["keycloak"]
	if !ok {
		keycloakInterface = map[string]interface{}{}
	}
	keycloakDict := keycloakInterface.(map[string]interface{})

	keycloakDict["realms"] = values.Keycloak.Realms
	keycloakDict["authFlows"] = values.Keycloak.AuthFlows
	keycloakDict["identityProviders"] = values.Keycloak.IdentityProviders

	values.OriginalYaml["keycloak"] = keycloakDict

	return true, nil
}

func (a *App) idGovUASecretChanged(path string, r *registry) (bool, string, error) {
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

	if clID != r.SupAuthClientID || (clSecret != r.SupAuthClientSecret && r.SupAuthClientSecret != emptyClientSecret) {
		return true, clSecret.(string), nil
	}

	return false, "", nil
}
