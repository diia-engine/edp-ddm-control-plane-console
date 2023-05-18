package registry

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	keycloakIndex = "keycloak"
)

func (a *App) prepareCitizenAuthSettings(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
	if r.RegistryCitizenAuth != "" {
		var citizenAuthFlow KeycloakAuthFlowsCitizenAuthFlow
		if err := json.Unmarshal([]byte(r.RegistryCitizenAuth), &citizenAuthFlow); err != nil {
			return fmt.Errorf("unable to decode citizen auth settings %w", err)
		}
		values.Keycloak.CitizenAuthFlow = citizenAuthFlow
		values.OriginalYaml[keycloakIndex] = values.Keycloak

	}
	return nil
}
