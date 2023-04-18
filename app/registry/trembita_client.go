package registry

import (
	"ddm-admin-console/router"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	MRLabelTargetTrembitaRegistryUpdate = "trembita-registry-update"
	MRLabelTrembitaRegsitryName         = "trembita-registry-name"
)

type TrembitaClientRegistryForm struct {
	TrembitaClientProtocolVersion string `form:"trembita-client-protocol-version" binding:"required"`
	TrembitaClientURL             string `form:"trembita-client-url"`
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
	TrembitaServiceServiceCode    string `form:"trembita-service-service-code"`
	TrembitaServiceServiceVersion string `form:"trembita-service-service-version"`
	TrembitaServiceAuthType       string `form:"trembita-service-auth-type" binding:"required"`
	TrembitaServiceAuthSecret     string `form:"trembita-service-auth-secret"`
}

func (tf TrembitaClientRegistryForm) ToNestedStruct(wiremockAddr string) TrembitaRegistry {
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
			MemberCode:     tf.TrembitaServiceMemberCode,
			MemberClass:    tf.TrembitaServiceMemberClass,
			XRoadInstance:  tf.TrembitaServiceXRoadInstance,
			SubsystemCode:  tf.TrembitaServiceSubsystemCode,
			ServiceCode:    tf.TrembitaServiceServiceCode,
			ServiceVersion: tf.TrembitaServiceServiceVersion,
		},
		Auth: map[string]string{
			"type": tf.TrembitaServiceAuthType,
		},
	}

	if tr.URL == "" {
		tr.Mock = true
		tr.URL = wiremockAddr
	}

	return tr
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

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	trembitaRegistryFromValues, ok := values.Trembita.Registries[tf.TrembitaClientRegitryName]
	if !ok {
		return nil, errors.New("wrong registry name")
	}

	trembitaRegistry := tf.ToNestedStruct(a.Config.WiremockAddr)
	trembitaRegistry.Type = trembitaRegistryFromValues.Type
	trembitaRegistry.Protocol = trembitaRegistryFromValues.Protocol

	trembita, ok := values.OriginalYaml[trembitaValuesKey]
	if !ok {
		return nil, errors.New("no trembita config in values")
	}
	trembitaDict := trembita.(map[string]interface{})
	registriesDict := trembitaDict[trembitaRegistriesValuesKet].(map[string]interface{})

	//TODO: change path to single secret vault:secret/<registry>/trembita-registries
	//TODO: check if keys rewrited or keep
	if tf.TrembitaServiceAuthType == authTypeAuthToken && tf.TrembitaServiceAuthSecret != "" {
		vaultPath := fmt.Sprintf("%s/trembita-registries", a.vaultRegistryPath(registryName))
		prefixedPath := fmt.Sprintf("vault:%s", vaultPath)

		if tf.TrembitaServiceAuthSecret != prefixedPath {
			if err := CreateVaultSecrets(a.Vault, map[string]map[string]interface{}{
				vaultPath: {
					fmt.Sprintf("trembita.registries.%s.auth.secret.token", tf.TrembitaClientRegitryName): tf.TrembitaServiceAuthSecret,
				},
			}, true); err != nil {
				return nil, errors.Wrap(err, "unable to create auth token secret")
			}
		}

		//todo: maybe move to nested struct converter
		trembitaRegistry.Auth["secret"] = prefixedPath
	}

	registriesDict[tf.TrembitaClientRegitryName] = trembitaRegistry
	trembitaDict[trembitaRegistriesKey] = registriesDict
	values.OriginalYaml[trembitaValuesKey] = trembitaDict

	if err := CreateEditMergeRequest(ctx, registryName, values.OriginalYaml, a.Gerrit,
		[]string{}, MRLabel{Key: MRLabelApprove, Value: MRLabelApproveAuto}, MRLabel{Key: MRLabelTarget, Value: MRLabelTargetTrembitaRegistryUpdate},
		MRLabel{Key: MRLabelTrembitaRegsitryName, Value: tf.TrembitaClientRegitryName}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}
	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) createTrembitaClientRegistry(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	var tf TrembitaClientRegistryForm
	if err := ctx.ShouldBind(&tf); err != nil {
		return nil, errors.Wrap(err, "unable to parse form")
	}
	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	_, ok := values.Trembita.Registries[tf.TrembitaClientRegitryName]

	if ok {
		return nil, errors.Wrap(err, "trembita client already exists")
	}

	trembitaRegistry := tf.ToNestedStruct(a.Config.WiremockAddr)
	trembitaRegistry.Type = externalSystemDeletableType
	trembitaRegistry.Protocol = tf.TrembitaClientProtocol

	if tf.TrembitaServiceAuthType == authTypeAuthToken && tf.TrembitaServiceAuthSecret != "" {
		vaultPath := fmt.Sprintf("%s/trembita-registries", a.vaultRegistryPath(registryName))
		prefixedPath := fmt.Sprintf("vault:%s", vaultPath)

		if tf.TrembitaServiceAuthSecret != prefixedPath {
			if err := CreateVaultSecrets(a.Vault, map[string]map[string]interface{}{
				vaultPath: {
					fmt.Sprintf("trembita.registries.%s.auth.secret.token", tf.TrembitaClientRegitryName): tf.TrembitaServiceAuthSecret,
				},
			}, true); err != nil {
				return nil, errors.Wrap(err, "unable to create auth token secret")
			}
		}

		trembitaRegistry.Auth["secret"] = prefixedPath
	}
	values.Trembita.Registries[tf.TrembitaClientRegitryName] = trembitaRegistry
	values.OriginalYaml[trembitaValuesKey] = values.Trembita
	if err := CreateEditMergeRequest(ctx, registryName, values.OriginalYaml, a.Gerrit,
		[]string{}, MRLabel{Key: MRLabelApprove, Value: MRLabelApproveAuto}, MRLabel{Key: MRLabelTarget, Value: MRLabelTargetTrembitaRegistryUpdate},
		MRLabel{Key: MRLabelTrembitaRegsitryName, Value: tf.TrembitaClientRegitryName}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}
	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) deleteTrembitaClient(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	trembitaClientName := ctx.Query("trembita-client")

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	trembita, ok := values.Trembita.Registries[trembitaClientName]
	if !ok {
		return nil, errors.New("trembita client does not exists")
	}

	if trembita.Type == "platform" {
		return nil, errors.New("trembita client is unavailable to delete")
	}

	delete(values.Trembita.Registries, trembitaClientName)
	values.OriginalYaml[trembitaValuesKey] = values.Trembita
	if err := CreateEditMergeRequest(ctx, registryName, values.OriginalYaml, a.Gerrit, []string{}, MRLabel{Key: MRLabelApprove, Value: MRLabelApproveAuto}); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) checkTrembitaClientExists(ctx *gin.Context) (rsp router.Response, retErr error) {
	registryName := ctx.Param("name")

	_, err := a.Codebase.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to find registry")
	}

	trembitaClientName := ctx.Query("trembita-client")

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values")
	}

	_, ok := values.Trembita.Registries[trembitaClientName]
	if ok {
		return router.MakeStatusResponse(http.StatusOK), nil
	}

	return router.MakeStatusResponse(http.StatusNotFound), nil
}

func (a *App) prepareTrembitaIPList(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
	if r.TrembitaIPList != "" {
		var ipList []string
		if err := json.Unmarshal([]byte(r.TrembitaIPList), &ipList); err != nil {
			return fmt.Errorf("unable to decode trembita ip list %w", err)
		}

		values.Trembita.IPList = ipList
		values.OriginalYaml[trembitaValuesKey] = values.Trembita
	}

	return nil
}
