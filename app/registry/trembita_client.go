package registry

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func (a *App) prepareTrembitaClientConfig(ctx *gin.Context, r *registry, values map[string]interface{},
	secrets map[string]map[string]interface{}) error {
	if r.TrembitaClientEnabled != "" {
		trembita, ok := values[trembitaValuesKey]
		if !ok {
			trembita = make(map[string]interface{})
		}
		trembitaDict := trembita.(map[string]interface{})

		consumer := make(map[string]interface{})
		consumer["user-id"] = r.TrembitaClientID
		consumer["protocol-version"] = r.TrembitaClientProtocolVersion

		consumerClient := make(map[string]interface{})
		consumerClient["x-road-instance"] = r.TrembitaClientXRoadInstance
		consumerClient["member-class"] = r.TrembitaClientMemberClass
		consumerClient["member-code"] = r.TrembitaClientMemberCode
		consumerClient["subsystem-code"] = r.TrembitaClientSubsystemCode

		consumer["client"] = consumerClient
		trembitaDict["consumer"] = consumer

		_, ok = trembitaDict[trembitaRegistriesKey]
		if !ok {
			trembitaClientRegistries := make(map[string]interface{})

			trembitaClientDefaultRegistries := strings.Split(a.Config.TrembitaClientDefaultRegistries, ",")
			for _, registry := range trembitaClientDefaultRegistries {
				regParts := strings.Split(registry, ":")
				if len(regParts) < 2 {
					continue
				}

				trembitaClientRegistries[regParts[0]] = map[string]string{"type": regParts[1]}
			}

			if len(trembitaClientRegistries) > 0 {
				trembitaDict[trembitaRegistriesKey] = trembitaClientRegistries
			}
		}

		values[trembitaValuesKey] = trembitaDict
	}

	return nil
}
