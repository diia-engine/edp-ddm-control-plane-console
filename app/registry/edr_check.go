package registry

import "github.com/gin-gonic/gin"

const (
	keycloakIndex = "keycloak"
)

func (a *App) prepareEDRCheck(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
	values.Keycloak.CitizenAuthFlow.EDRCheck = r.EDRCheckEnabled != ""
	values.OriginalYaml[keycloakIndex] = values.Keycloak
	return nil
}
