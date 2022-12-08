package registry

import (
	"ddm-admin-console/router"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type TrembitaClientRegistryForm struct {
	TrembitaClientProtocolVersion string `form:"trembita-client-protocol-version" binding:"required"`
	TrembitaClientURL             string `form:"trembita-client-url" binding:"required"`
	TrembitaClientUserID          string `form:"trembita-client-user-id" binding:"required"`
	TrembitaClientXRoadInstance   string `form:"trembita-client-x-road-instance" binding:"required"`
	TrembitaClientMemberClass     string `form:"trembita-client-member-class" binding:"required"`
	TrembitaClientMemberCode      string `form:"trembita-client-member-code" binding:"required"`
	TrembitaClientSubsystemCode   string `form:"trembita-client-subsystem-code" binding:"required"`
	TrembitaClientRegitryName     string `form:"trembita-client-regitry-name" binding:"required"`
	TrembitaClientProtocol        string `form:"trembita-client-protocol" binding:"required"`
	TrembitaServiceXRoadInstance  string `form:"trembita-service-x-road-instance" binding:"required"`
	TrembitaServiceMemberClass    string `form:"trembita-service-member-class" binding:"required"`
	TrembitaServiceMemberCode     string `form:"trembita-service-member-code" binding:"required"`
	TrembitaServiceSubsystemCode  string `form:"trembita-service-subsystem-code" binding:"required"`
	TrembitaServiceAuthType       string `form:"trembita-service-auth-type" binding:"required"`
	TrembitaServiceAuthSecret     string `form:"trembita-service-auth-secret"`
}

func (tf TrembitaClientRegistryForm) ToNestedStruct() TrembitaRegistry {
	tr := TrembitaRegistry{
		URL:             tf.TrembitaClientURL,
		UserID:          tf.TrembitaClientUserID,
		ProtocolVersion: tf.TrembitaClientProtocolVersion,
		Client: TrembitaRegistryClient{
			MemberClass:   tf.TrembitaClientMemberClass,
			MemberCode:    tf.TrembitaClientMemberCode,
			SubsystemCode: tf.TrembitaClientSubsystemCode,
			XRoadInstance: tf.TrembitaClientXRoadInstance,
		},
		Service: TrembitaRegistryService{
			TrembitaRegistryClient: TrembitaRegistryClient{
				MemberCode:    tf.TrembitaServiceMemberCode,
				MemberClass:   tf.TrembitaServiceMemberClass,
				XRoadInstance: tf.TrembitaServiceXRoadInstance,
				SubsystemCode: tf.TrembitaServiceSubsystemCode,
			},
			Auth: map[string]string{
				"type": tf.TrembitaServiceAuthType,
			},
		},
	}

	return tr
}

func (a *App) prepareTrembitaClientConfig(ctx *gin.Context, r *registry, values map[string]interface{},
	secrets map[string]map[string]interface{}) error {

	trembita, ok := values[trembitaValuesKey]
	if !ok {
		trembita = make(map[string]interface{})
	}
	trembitaDict := trembita.(map[string]interface{})

	_, ok = trembitaDict[trembitaRegistriesKey]
	if !ok {
		trembitaClientRegistries := make(map[string]interface{})

		trembitaClientDefaultRegistries := strings.Split(a.Config.TrembitaClientDefaultRegistries, ",")
		for _, registry := range trembitaClientDefaultRegistries {
			regParts := strings.Split(registry, ":")
			if len(regParts) < 2 {
				continue
			}

			trembitaClientRegistries[regParts[0]] = map[string]string{
				"type":     regParts[1],
				"protocol": "SOAP",
			}
		}

		if len(trembitaClientRegistries) > 0 {
			trembitaDict[trembitaRegistriesKey] = trembitaClientRegistries
		}
	}

	values[trembitaValuesKey] = trembitaDict

	return nil
}

func (a *App) setTrembitaClientRegistryData(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")
	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	var tf TrembitaClientRegistryForm
	if err := ctx.ShouldBind(&tf); err != nil {
		return nil, errors.Wrap(err, "unable to parse form")
	}

	values, _, err := GetValuesFromGit(ctx, registryName, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	trembitaRegistryFromValues, ok := values.Trembita.Registries[tf.TrembitaClientRegitryName]
	if !ok {
		return nil, errors.New("wrong registry name")
	}

	trembitaRegistry := tf.ToNestedStruct()
	trembitaRegistry.Type = trembitaRegistryFromValues.Type
	trembitaRegistry.Protocol = trembitaRegistryFromValues.Protocol

	trembita, ok := values.OriginalYaml[trembitaValuesKey]
	if !ok {
		return nil, errors.New("no trembita config in values")
	}
	trembitaDict := trembita.(map[string]interface{})
	registriesDict := trembitaDict[trembitaRegistriesValuesKet].(map[string]interface{})

	//path vault:secret/<registry>/trembita-registries/<trembita-registry-name>
	//key trembita.registries.<registry-name>.auth.secret.token
	//create secret

	if tf.TrembitaServiceAuthType == authTypeAuthToken && tf.TrembitaServiceAuthSecret != "" {
		vaultPath := fmt.Sprintf("%s/trembita-registries/%s", a.vaultRegistryPath(registryName),
			tf.TrembitaClientRegitryName)
		prefixedPath := fmt.Sprintf("vault:%s", vaultPath)

		//secretPath := fmt.Sprintf("vault:secret/%s/trembita-registries/%s", registryName,
		//	tf.TrembitaClientRegitryName)

		if tf.TrembitaServiceAuthSecret != prefixedPath {
			if err := a.createVaultSecrets(map[string]map[string]interface{}{
				vaultPath: {
					fmt.Sprintf("trembita.registries.%s.auth.secret.token", tf.TrembitaClientRegitryName): tf.TrembitaServiceAuthSecret,
				},
			}); err != nil {
				return nil, errors.Wrap(err, "unable to create auth token secret")
			}
		}
		//todo: maybe move to nested struct converter
		trembitaRegistry.Service.Auth["secret"] = prefixedPath
	}

	registriesDict[tf.TrembitaClientRegitryName] = trembitaRegistry
	trembitaDict[trembitaRegistriesKey] = registriesDict
	values.OriginalYaml[trembitaValuesKey] = trembitaDict

	if err := CreateEditMergeRequest(ctx, registryName, values.OriginalYaml, a.Gerrit,
		MRLabel{Key: MRLabelApprove, Value: MRLabelApproveAuto}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}
