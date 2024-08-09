package registry

import (
	"context"
	"crypto/sha1"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/k8s"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-version"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	MasterBranch      = "master"
	VersionQueryParam = "version"
)

func (a *App) editRegistryGet(ctx *gin.Context) (response router.Response, retErr error) {
	registryName := ctx.Param("name")

	mrExists, err := ProjectHasOpenMR(ctx, registryName, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to check MRs exists, %w", err)
	}

	if mrExists {
		return router.MakeHTMLResponse(200, "registry/edit-mr-exists.html", gin.H{
			"registryName": registryName,
			"page":         "registry",
		}), nil
	}

	userCtx := router.ContextWithUserAccessToken(ctx)
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, fmt.Errorf("unable to init service for user context, %w", err)
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, fmt.Errorf("unable to init service for user context, %w", err)
	}

	if err := a.checkUpdateAccess(registryName, k8sService); err != nil {
		return nil, errors.New("access denied")
	}

	reg, err := cbService.Get(registryName)
	if err != nil {
		return nil, fmt.Errorf("unable to get registry, %w", err)
	}

	model := registry{KeyDeviceType: KeyDeviceTypeFile, Name: reg.Name}
	if reg.Spec.Description != nil {
		model.Description = *reg.Spec.Description
	}

	hwINITemplateContent, err := GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get ini template data, %w", err)
	}

	hasUpdate, branches, registryVersion, err := HasUpdate(userCtx, a.Services.Gerrit, reg, MRTargetRegistryVersionUpdate)
	if err != nil {
		return nil, fmt.Errorf("unable to check for updates, %w", err)
	}

	if is, rsp := isVersionRedirect(ctx, "/admin/registry/edit", registryName, registryVersion); is {
		return rsp, nil
	}

	dnsManual, err := a.getDNSManualURL(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get dns manual, %w", err)
	}

	valuesFromDefaultBranch, err := a.GetValuesFromBranch(a.Config.RegistryTemplateName, registryVersion.Original())
	if err != nil {
		return nil, fmt.Errorf("unable to get template content, %w", err)
	}

	responseParams := gin.H{
		"dnsManual":             dnsManual,
		"registry":              reg,
		"page":                  "registry",
		"updateBranches":        branches,
		"hasUpdate":             hasUpdate,
		"action":                "edit",
		"registryVersion":       MajorVersion(registryVersion.Core().Original()),
		"platformStatusType":    a.Config.CloudProvider,
		"isPlatformAdmin":       ctx.GetBool(router.CanViewClusterManagementSessionKey),
		"defaultRegistryValues": valuesFromDefaultBranch,
	}

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to get values from git, %w", err)
	}

	if err := a.loadValuesEditConfig(ctx, values, responseParams, &model); err != nil {
		return nil, fmt.Errorf("unable to load edit values from config, %w", err)
	}

	if err := a.viewDNSConfig(ctx, reg, values, responseParams); err != nil {
		return nil, fmt.Errorf("unable to load dns config, %w", err)
	}

	templateArgs, templateErr := json.Marshal(responseParams)
	if templateErr != nil {
		return nil, fmt.Errorf("unable to encode template arguments, %w", templateErr)
	}

	responseParams["templateArgs"] = string(templateArgs)
	responseParams["hwINITemplateContent"] = hwINITemplateContent

	return router.MakeHTMLResponse(200, "registry/edit.html", responseParams), nil
}

func MajorVersion(fullVersion string) string {
	if fullVersion == "" {
		return ""
	}

	parts := strings.Split(fullVersion, ".")
	if len(parts) > 3 {
		parts = parts[:len(parts)-1]
	}

	return strings.Join(parts, ".")
}

func isVersionRedirect(ctx *gin.Context, basePath, registryName string, registryVersion *version.Version) (bool, router.Response) {
	qVersion := ctx.Query(VersionQueryParam)

	if qVersion == "" {
		return true, router.MakeRedirectResponse(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s/%s?%s=%s", basePath, registryName, VersionQueryParam,
				MajorVersion(registryVersion.Core().Original())))
	}

	if qVersion != MajorVersion(registryVersion.Core().Original()) {
		return true, router.MakeRedirectResponse(http.StatusTemporaryRedirect,
			fmt.Sprintf("%s/%s?%s=%s", basePath, registryName, VersionQueryParam,
				MajorVersion(registryVersion.Core().Original())))
	}

	return false, nil
}

func (a *App) loadValuesEditConfig(ctx context.Context, values *Values, rspParams gin.H, r *registry) error {
	//TODO: refactor to values struct
	if err := a.loadSMTPConfig(values.OriginalYaml, rspParams); err != nil {
		return fmt.Errorf("unable to load smtp config, %w", err)
	}

	//TODO: refactor to values struct
	if err := a.loadAdminsConfig(values.OriginalYaml, r); err != nil {
		return fmt.Errorf("unable to load admins config, %w", err)
	}

	if err := a.loadOBCConfig(values, r); err != nil {
		return fmt.Errorf("unable to load obc config, %w", err)
	}

	if err := a.loadIDGovUAClientID(values); err != nil {
		return fmt.Errorf("unable to load secret: %w", err)
	}

	keycloakHostname, err := LoadKeycloakDefaultHostname(ctx, a.KeycloakDefaultHostname, a.EDPComponent)
	if err != nil {
		return fmt.Errorf("unable to load keycloak default hostname, %w", err)
	}
	rspParams["keycloakHostname"] = keycloakHostname

	keycloakHostnames, err := a.loadKeycloakHostnames()
	if err != nil {
		return fmt.Errorf("unable to load keycloak hostnames, %w", err)
	}
	rspParams["keycloakHostnames"] = keycloakHostnames
	rspParams["keycloakCustomHost"] = values.Keycloak.CustomHost

	rspParams["model"] = r

	registryData, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("unable to encode registry data, %w", err)
	}
	rspParams["registryData"] = string(registryData)
	rspParams["registryValues"] = values

	return nil
}

func (a *App) loadIDGovUAClientID(values *Values) error {
	if values.Keycloak.IdentityProviders.IDGovUA.SecretKey == "" {
		return nil
	}

	dataDict, err := a.Vault.Read(values.Keycloak.IdentityProviders.IDGovUA.SecretKey)
	if err != nil {
		return fmt.Errorf("unable to load id-gov-ua secret, err: %w", err)
	}

	d, ok := dataDict[idGovUASecretClientID]
	if !ok {
		return nil
	}

	str, ok := d.(string)
	if !ok {
		return nil
	}

	values.Keycloak.IdentityProviders.IDGovUA.ClientID = str
	return nil
}

func (a *App) loadAdminsConfig(values map[string]interface{}, r *registry) error {
	adminsInterface, ok := values[AdministratorsValuesKey]
	if !ok {
		r.Admins = "[]"
		return nil
	}

	adminsJs, err := json.Marshal(adminsInterface)
	if err != nil {
		return fmt.Errorf("unable to marshal admins, %w", err)
	}

	var admins []Admin
	if err := json.Unmarshal(adminsJs, &admins); err != nil {
		return fmt.Errorf("unable tro unmarshal admins, %w", err)
	}

	//TODO: maybe load admin password
	for i := range admins {
		admins[i].TmpPassword = ""
		admins[i].PasswordVaultSecret = ""
		admins[i].PasswordVaultSecretKey = ""
	}

	adminsJs, err = json.Marshal(admins)
	if err != nil {
		return fmt.Errorf("unable to marshal admins, %w", err)
	}

	r.Admins = string(adminsJs)

	return nil
}

func (a *App) loadOBCConfig(values *Values, r *registry) error {
	if values.Global.RegistryBackup.OBC.Credentials == "" {
		return nil
	}

	dataDict, err := a.Vault.Read(values.Global.RegistryBackup.OBC.Credentials)
	if err != nil {
		return fmt.Errorf("unable to load obc credential, err: %w", err)
	}

	login, ok := dataDict[a.Config.BackupBucketAccessKeyID]
	if !ok {
		return nil
	}

	password, ok := dataDict[a.Config.BackupBucketSecretAccessKey]
	if !ok {
		return nil
	}

	loginStr, ok := login.(string)
	if !ok {
		return nil
	}

	passwordStr, ok := password.(string)
	if !ok {
		return nil
	}

	r.OBCLogin = loginStr
	r.OBCPassword = passwordStr
	return nil
}

func (a *App) loadSMTPConfig(values map[string]interface{}, rspParams gin.H) error {
	global, ok := values[GlobalValuesIndex]
	if !ok {
		rspParams["smtpConfig"] = "{}"
		return nil
	}

	globalDict := global.(map[string]interface{})
	if _, ok := globalDict["notifications"]; !ok {
		return nil
	}

	emailDict := globalDict["notifications"].(map[string]interface{})["email"].(map[string]interface{})
	mailConfig, err := json.Marshal(emailDict)
	if err != nil {
		return fmt.Errorf("unable to encode ot JSON smtp config, %w", err)
	}

	rspParams["smtpConfig"] = string(mailConfig)
	return nil
}

func (a *App) checkUpdateAccess(codebaseName string, userK8sService k8s.ServiceInterface) error {
	allowedToUpdate, err := a.Services.Codebase.CheckIsAllowedToUpdate(codebaseName, userK8sService)
	if err != nil {
		return fmt.Errorf("unable to check create access, %w", err)
	}
	if !allowedToUpdate {
		return errors.New("access denied")
	}

	return nil
}

func (a *App) editRegistryPost(ctx *gin.Context) (response router.Response, retErr error) {
	userCtx := router.ContextWithUserAccessToken(ctx)
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, fmt.Errorf("unable to init service for user context, %w", err)
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, fmt.Errorf("unable to init service for user context, %w", err)
	}

	registryName := ctx.Param("name")
	//TODO: move this to middleware
	allowed, err := cbService.CheckIsAllowedToUpdate(registryName, k8sService)
	if err != nil {
		return nil, fmt.Errorf("unable to check access, %w", err)
	}
	if !allowed {
		return nil, errors.New("access denied")
	}

	cb, err := cbService.Get(registryName)
	if err != nil {
		return nil, fmt.Errorf("unable to get registry, %w", err)
	}

	r := registry{
		Name:              registryName,
		RegistryGitBranch: cb.Spec.DefaultBranch,
		Scenario:          ScenarioKeyNotRequired,
	}

	if err := ctx.ShouldBind(&r); err != nil {
		return nil, fmt.Errorf("unable to parse registry form, %w", err)
	}

	if err := a.editRegistry(userCtx, ctx, &r, cb, cbService); err != nil {
		return nil, fmt.Errorf("unable to edit registry, %w", err)
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", r.Name)), nil
}

func (a *App) editRegistry(ctx context.Context, ginContext *gin.Context, r *registry, cb *codebase.Codebase,
	cbService codebase.ServiceInterface) error {

	cb.Spec.Description = &r.Description
	if cb.Annotations == nil {
		cb.Annotations = make(map[string]string)
	}

	values, err := GetValuesFromGit(r.Name, MasterBranch, a.Gerrit)
	if err != nil {
		return fmt.Errorf("unable to get values from git, %w", err)
	}

	var (
		vaultSecretData = make(map[string]map[string]interface{})
		mrActions       = make([]string, 0)
		valuesChanged   = false
		repoFiles       = make(map[string]string)
	)

	for _, proc := range a.createUpdateRegistryProcessors() {
		procValuesChanged, err := proc(ginContext, r, values, vaultSecretData, &mrActions)
		if err != nil {
			return fmt.Errorf("error during registry create, %w", err)
		}
		if procValuesChanged {
			valuesChanged = true
		}
	}

	keysModified, err := PrepareRegistryKeys(keyManagement{
		r: r,
		vaultSecretPath: a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", KeyManagementVaultPath,
			time.Now().Format("20060201T150405Z"))),
	}, ginContext.Request, vaultSecretData, values.OriginalYaml, repoFiles)
	if err != nil {
		return fmt.Errorf("unable to create registry keys, %w", err)
	}

	if keysModified {
		if err := CacheRepoFiles(a.TempFolder, r.Name, repoFiles, a.Cache); err != nil {
			return fmt.Errorf("unable to cache repo file, %w", err)
		}
	}

	if valuesChanged || len(repoFiles) > 0 || keysModified {
		if err := CreateEditMergeRequest(ginContext, r.Name, values.OriginalYaml, a.Gerrit, mrActions); err != nil {
			return fmt.Errorf("unable to create edit merge request, %w", err)
		}
	}

	if len(vaultSecretData) > 0 {
		if err := CreateVaultSecrets(a.Vault, vaultSecretData, false); err != nil {
			return fmt.Errorf("unable to create vault secrets, %w", err)
		}
	}

	if err := cbService.Update(ctx, cb); err != nil {
		return fmt.Errorf("unable to update codebase, %w", err)
	}

	return nil
}

func MapHash(v map[string]interface{}) (string, error) {
	bts, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("unable to encode map, %w", err)
	}

	hasher := sha1.New()
	hasher.Write(bts)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil)), nil
}

type MRLabel struct {
	Key   string
	Value string
}

func CreateEditMergeRequest(ctx *gin.Context, projectName string, values map[string]interface{},
	gerritService gerrit.ServiceInterface, mrActions []string, labels ...MRLabel) error {

	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("unable to encode values yaml, %w", err)
	}

	mrExists, err := ProjectHasOpenMR(ctx, projectName, gerritService)
	if err != nil {
		return fmt.Errorf("unable to check project MR exists, %w", err)
	}

	if mrExists {
		return MRExists("there is already open merge request(s) for this registry")
	}

	_labels := map[string]string{
		MRLabelTarget: mrTargetEditRegistry,
	}

	for _, l := range labels {
		_labels[l.Key] = l.Value
	}

	mrActionsJsonBts, err := json.Marshal(mrActions)
	if err != nil {
		return fmt.Errorf("unable to encode mr actions, %w", err)
	}

	if err := gerritService.CreateMergeRequestWithContents(ctx, &gerrit.MergeRequest{
		ProjectName:   projectName,
		Name:          fmt.Sprintf("reg-edit-mr-%s-%d", projectName, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("edit registry"),
		TargetBranch:  "master",
		Labels:        _labels,
		Annotations: map[string]string{
			MRAnnotationActions: string(mrActionsJsonBts),
		},
	}, map[string]string{
		ValuesLocation: string(valuesYaml),
	}); err != nil {
		return fmt.Errorf("unable to create MR with new values, %w", err)
	}

	return nil
}

func ProjectHasOpenMR(ctx *gin.Context, projectName string, gerritService gerrit.ServiceInterface) (bool, error) {
	mrs, err := gerritService.GetMergeRequestByProject(ctx, projectName)
	if err != nil {
		return false, fmt.Errorf("unable to get MRs, %w", err)
	}

	for _, mr := range mrs {
		if mr.Status.Value == gerrit.StatusNew {
			return true, nil
		}
	}

	return false, nil
}

func validateAdmins(adminsLine string) ([]Admin, error) {
	var admins []Admin
	if err := json.Unmarshal([]byte(adminsLine), &admins); err != nil {
		return nil, fmt.Errorf("unable to unmarshal admins, %w", err)
	}

	validate := validator.New()
	for i, admin := range admins {
		errs := validate.Var(admin.Email, "required,email")
		if errs != nil {
			return nil,
				validator.ValidationErrors([]validator.FieldError{router.MakeFieldError("Admins", "required")})
		}
		admins[i].Username = admin.Email
	}

	return admins, nil
}
