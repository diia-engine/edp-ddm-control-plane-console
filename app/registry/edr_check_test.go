package registry

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrepareEDRCheckTrue(t *testing.T) {
	app := App{}
	ctx := gin.Context{}
	values := Values{
		OriginalYaml: map[string]interface{}{},
	}
	secrets := map[string]map[string]interface{}{}
	var mrActions []string
	r := registry{
		EDRCheckEnabled: "true",
	}

	result := app.prepareEDRCheck(&ctx, &r, &values, secrets, &mrActions)

	assert.Nil(t, result)
	assert.Equal(t, values.Keycloak.CitizenAuthFlow.EDRCheck, true)
	assert.Equal(t, values.OriginalYaml[keycloakIndex], values.Keycloak)
}

func TestPrepareEDRCheckFalse(t *testing.T) {
	app := App{}
	ctx := gin.Context{}
	values := Values{
		OriginalYaml: map[string]interface{}{},
	}
	secrets := map[string]map[string]interface{}{}
	var mrActions []string
	r := registry{
		EDRCheckEnabled: "",
	}

	result := app.prepareEDRCheck(&ctx, &r, &values, secrets, &mrActions)

	assert.Nil(t, result)
	assert.Equal(t, values.Keycloak.CitizenAuthFlow.EDRCheck, false)
	assert.Equal(t, values.OriginalYaml[keycloakIndex], values.Keycloak)
}
