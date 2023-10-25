package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
)

const (
	registryIdGovUaCitizenAuthType = "registry-id-gov-ua"
	registryIdGovUaOfficerAuthType = "id-gov-ua-officer-redirector"
	hardKeyOsplmIniIndex           = "osplm.ini"
)

type BackupConfig struct {
	StorageLocation          string `form:"storage-location" binding:"required"`
	StorageType              string `form:"storage-type" binding:"required"`
	StorageCredentialsKey    string `form:"storage-credentials-key" binding:"required"`
	StorageCredentialsSecret string `form:"storage-credentials-secret" binding:"required"`
}

const (
	masterBranch = "master"
)

func (a *App) editGet(ctx *gin.Context) (router.Response, error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	mrExists, err := registry.ProjectHasOpenMR(ctx, a.ClusterRepo, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check project MR exists")
	}

	if mrExists {
		return router.MakeHTMLResponse(200, "registry/edit-mr-exists.html", gin.H{
			"page": "cluster",
		}), nil
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init cb service for user")
	}

	canUpdateCluster, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	}
	if !canUpdateCluster {
		return nil, errors.Wrap(err, "access denied")
	}

	cb, err := cbService.Get(a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	hasUpdate, branches, _, err := registry.HasUpdate(userCtx, a.Services.Gerrit, cb, registry.MRTargetClusterUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	hwINITemplateContent, err := registry.GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	values, err := getValuesFromGit(a.Config.CodebaseName, masterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load values")
	}
	// decode osplm.ini
	if values.DigitalSignature.Data.OsplmIni != "" {
		values.DigitalSignature.Data.OsplmIni = a.Vault.GetPropertyFromVault(values.DigitalSignature.Data.OsplmIni, hardKeyOsplmIniIndex)
	}

	valuesJs, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("unable to encode values to json, %w", err)
	}

	dnsManual, err := registry.GetManualURL(ctx, a.EDPComponent, a.DDMManualEDPComponent, a.RegistryDNSManualPath)
	if err != nil {
		return nil, fmt.Errorf("unable to get manual, %w", err)
	}
	usedKeys, err := a.getUsedSignKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get used in registries sign keys, %w", err)
	}
	registries, err := a.getRegistryNames()
	if err != nil {
		return nil, fmt.Errorf("unable to get registries, %w", err)
	}

	rspParams := gin.H{
		"dnsManual":      dnsManual,
		"page":           "cluster",
		"updateBranches": branches,
		"hasUpdate":      hasUpdate,
		"values":         string(valuesJs),
		"usedKeys":       usedKeys,
		"registries":     registries,
	}

	for _, f := range a.editDataLoaders() {
		if err := f(ctx, values, rspParams); err != nil {
			return nil, errors.Wrap(err, "unable to load edit data")
		}
	}

	templateArgs, err := json.Marshal(rspParams)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode template arguments")
	}

	return router.MakeHTMLResponse(200, "cluster/edit.html", gin.H{
		"page":                 "cluster",
		"templateArgs":         string(templateArgs),
		"hwINITemplateContent": hwINITemplateContent,
	}), nil
}

func (a *App) editDataLoaders() []func(context.Context, *Values, gin.H) error {
	return []func(context.Context, *Values, gin.H) error{
		a.loadAdminsConfig,
		a.loadCIDRConfig,
		a.loadBackupScheduleConfig,
		a.loadKeycloakDefaultHostname,
		a.loadDocumentation,
		a.loadGeneral,
	}
}

func (a *App) editPost(ctx *gin.Context) (router.Response, error) {
	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) loadAdminsConfig(_ context.Context, values *Values, rspParams gin.H) error {
	bts, err := json.Marshal(values.Admins)
	if err != nil {
		return errors.Wrap(err, "unable to json encode admins")
	}

	rspParams["admins"] = string(bts)
	return nil
}

func (a *App) loadCIDRConfig(_ context.Context, values *Values, rspParams gin.H) error {
	if values.Global.WhiteListIP.AdminRoutes != "" {
		cidrConfig, err := json.Marshal(map[string]interface{}{
			"admin": strings.Split(values.Global.WhiteListIP.AdminRoutes, " "),
		})
		if err != nil {
			return errors.Wrap(err, "unable to encode cidr to JSON")
		}
		rspParams["cidrConfig"] = string(cidrConfig)
	}

	return nil
}

func getValuesFromGit(projectName, branch string, gerritService gerrit.ServiceInterface) (*Values, error) {
	content, err := gerritService.GetFileFromBranch(projectName, branch, url.PathEscape(ValuesLocation))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values yaml")
	}

	var valuesStruct Values
	if err := yaml.Unmarshal([]byte(content), &valuesStruct); err != nil {
		return nil, errors.Wrap(err, "unable to decode values yaml")
	}

	return &valuesStruct, nil
}

func (a *App) getRegistryNames() ([]string, error) {
	cbs, err := a.Services.Codebase.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return nil, fmt.Errorf("unable to get codebases, %w", err)
	}

	var names []string
	for _, reg := range cbs {
		names = append(names, reg.Name)
	}
	return names, nil
}

func (a *App) getUsedSignKeys(ctx *gin.Context) (map[string][]string, error) {

	cbs, err := a.Services.Codebase.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return nil, fmt.Errorf("unable to get codebases, %w", err)
	}

	usedKeys := make(map[string][]string)
	for _, reg := range cbs {
		values, err := registry.GetValuesFromGit(strings.TrimLeft(*reg.Spec.GitUrlPath, "/"), registry.MasterBranch, a.Gerrit)
		if err != nil {
			return nil, fmt.Errorf("unable to decode values yaml, %w", err)
		}
		if values.Keycloak.CitizenAuthFlow.AuthType == registryIdGovUaCitizenAuthType {
			usedKeys[reg.Name] = append(usedKeys[reg.Name], values.Keycloak.CitizenAuthFlow.RegistryIdGovUa.KeyName)
		}
		if values.Keycloak.Realms.OfficerPortal.BrowserFlow == registryIdGovUaOfficerAuthType {
			usedKeys[reg.Name] = append(usedKeys[reg.Name], values.Keycloak.IdentityProviders.IDGovUA.KeyName)
		}
	}
	return usedKeys, nil
}
