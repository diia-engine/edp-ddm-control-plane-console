package registry

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	supAuthBrowserFlowWidget  = "dso-officer-auth-flow"
	supAuthBrowserFlowIdGovUa = "id-gov-ua-officer-redirector"
	idGovUASecretPath         = "officer-id-gov-ua-client-info"
	idGovUASecretClientID     = "clientId"
)

func (a *App) prepareSupplierAuthConfig(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {

	values.Keycloak.Realms.OfficerPortal.SelfRegistration = r.SelfRegistration == "on"

	if r.SupAuthBrowserFlow == "" {
		return nil
	}

	values.Keycloak.Realms.OfficerPortal.BrowserFlow = r.SupAuthBrowserFlow

	if r.SupAuthBrowserFlow == supAuthBrowserFlowWidget {
		widgetHeight, err := strconv.ParseInt(r.SupAuthWidgetHeight, 10, 32)
		if err != nil {
			return fmt.Errorf("unable to decode int, err: %w", err)
		}
		values.Keycloak.AuthFlows.OfficerAuthFlow.WidgetHeight = int(widgetHeight)
		values.SignWidget.URL = r.SupAuthURL
	} else if r.SupAuthBrowserFlow == supAuthBrowserFlowIdGovUa {
		values.Keycloak.IdentityProviders.IDGovUA.URL = r.SupAuthURL

		secretPath := a.vaultRegistryPathKey(r.Name, idGovUASecretPath)
		if r.SupAuthClientID != "" && r.SupAuthClientSecret != "" && r.SupAuthClientSecret != "*****" {
			secrets[secretPath] = map[string]interface{}{
				idGovUASecretClientID: r.SupAuthClientID,
				"clientSecret":        r.SupAuthClientSecret,
			}
		}
		values.Keycloak.IdentityProviders.IDGovUA.SecretKey = secretPath
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

	return nil
}
