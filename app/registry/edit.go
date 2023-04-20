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
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	MasterBranch = "master"
)

func (a *App) editRegistryGet(ctx *gin.Context) (response router.Response, retErr error) {
	registryName := ctx.Param("name")

	mrExists, err := ProjectHasOpenMR(ctx, registryName, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check MRs exists")
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
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	if err := a.checkUpdateAccess(registryName, k8sService); err != nil {
		return nil, errors.New("access denied")
	}

	reg, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	model := registry{KeyDeviceType: KeyDeviceTypeFile, Name: reg.Name}
	if reg.Spec.Description != nil {
		model.Description = *reg.Spec.Description
	}

	hwINITemplateContent, err := GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	hasUpdate, branches, err := HasUpdate(userCtx, a.Services.Gerrit, reg, MRTargetRegistryVersionUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	dnsManual, err := a.getDNSManualURL(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get dns manual")
	}

	responseParams := gin.H{
		"dnsManual":            dnsManual,
		"registry":             reg,
		"page":                 "registry",
		"hwINITemplateContent": hwINITemplateContent,
		"updateBranches":       branches,
		"hasUpdate":            hasUpdate,
		"action":               "edit",
	}

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	if err := a.loadValuesEditConfig(ctx, values, responseParams, &model); err != nil {
		return nil, errors.Wrap(err, "unable to load edit values from config")
	}

	if err := a.viewDNSConfig(ctx, registryName, values, responseParams); err != nil {
		return nil, errors.Wrap(err, "unable to load dns config")
	}

	templateArgs, templateErr := json.Marshal(responseParams)
	if templateErr != nil {
		return nil, errors.Wrap(templateErr, "unable to encode template arguments")
	}

	responseParams["templateArgs"] = string(templateArgs)

	return router.MakeHTMLResponse(200, "registry/edit.html", responseParams), nil
}

func (a *App) loadValuesEditConfig(ctx context.Context, values *Values, rspParams gin.H, r *registry) error {
	if err := a.loadSMTPConfig(values.OriginalYaml, rspParams); err != nil {
		return errors.Wrap(err, "unable to load smtp config")
	}

	//TODO: refactor to values struct
	if err := a.loadAdminsConfig(values.OriginalYaml, r); err != nil {
		return errors.Wrap(err, "unable to load admins config")
	}

	//TODO: refactor to values struct
	if err := a.loadRegistryResourcesConfig(values.OriginalYaml, r); err != nil {
		return errors.Wrap(err, "unable to load resources config")
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
		return errors.Wrap(err, "unable to encode registry data")
	}
	rspParams["registryData"] = string(registryData)
	rspParams["registryValues"] = values

	return nil
}

func (a *App) loadIDGovUAClientID(values *Values) error {
	if values.Keycloak.IdentityProviders.IDGovUA.SecretKey == "" {
		return nil
	}

	modPath := ModifyVaultPath(values.Keycloak.IdentityProviders.IDGovUA.SecretKey)
	s, err := a.Vault.Read(modPath)
	if err != nil {
		return fmt.Errorf("unable to load id-gov-ua secret, err: %w", err)
	}

	data, ok := s.Data["data"]
	if !ok {
		return nil
	}

	dataDict, ok := data.(map[string]interface{})
	if !ok {
		return nil
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

func (a *App) loadRegistryResourcesConfig(values map[string]interface{}, r *registry) error {
	global, ok := values[GlobalValuesIndex]
	if !ok {
		return nil
	}
	globalDict, ok := global.(map[string]interface{})
	if !ok {
		return nil
	}

	resources, ok := globalDict[ResourcesValuesKey]
	if !ok {
		return nil
	}

	resJS, err := json.Marshal(resources)
	if err != nil {
		return errors.Wrap(err, "unable to encode resources config")
	}

	r.Resources = string(resJS)
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
		return errors.Wrap(err, "unable to marshal admins")
	}

	var admins []Admin
	if err := json.Unmarshal(adminsJs, &admins); err != nil {
		return errors.Wrap(err, "unable tro unmarshal admins")
	}

	//TODO: maybe load admin password
	for i := range admins {
		admins[i].TmpPassword = ""
		admins[i].PasswordVaultSecret = ""
		admins[i].PasswordVaultSecretKey = ""
	}

	adminsJs, err = json.Marshal(admins)
	if err != nil {
		return errors.Wrap(err, "unable to marshal admins")
	}

	r.Admins = string(adminsJs)

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
		return errors.Wrap(err, "unable to encode ot JSON smtp config")
	}

	rspParams["smtpConfig"] = string(mailConfig)
	return nil
}

func (a *App) checkUpdateAccess(codebaseName string, userK8sService k8s.ServiceInterface) error {
	allowedToUpdate, err := a.Services.Codebase.CheckIsAllowedToUpdate(codebaseName, userK8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check create access")
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
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.Param("name")

	allowed, err := cbService.CheckIsAllowedToUpdate(registryName, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access")
	}
	if !allowed {
		return nil, errors.New("access denied")
	}

	cb, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	r := registry{
		Name:                registryName,
		RegistryGitBranch:   cb.Spec.DefaultBranch,
		RegistryGitTemplate: cb.Spec.Repository.Url,
		Scenario:            ScenarioKeyNotRequired,
	}

	if err := ctx.ShouldBind(&r); err != nil {
		return nil, errors.Wrap(err, "unable to parse registry form")
	}

	if err := a.editRegistry(userCtx, ctx, &r, cb, cbService); err != nil {
		return nil, errors.Wrap(err, "unable to edit registry")
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
		return errors.Wrap(err, "unable to get values from git")
	}

	initialValuesHash, err := MapHash(values.OriginalYaml)
	if err != nil {
		return errors.Wrap(err, "unable to hash values")
	}

	vaultSecretData := make(map[string]map[string]interface{})
	mrActions := make([]string, 0)

	for _, proc := range a.createUpdateRegistryProcessors() {
		if err := proc(ginContext, r, values, vaultSecretData, &mrActions); err != nil {
			return errors.Wrap(err, "error during registry create")
		}
	}

	repoFiles := make(map[string]string)

	if _, err := PrepareRegistryKeys(keyManagement{
		r: r,
		vaultSecretPath: a.vaultRegistryPathKey(r.Name, fmt.Sprintf("%s-%s", KeyManagementVaultPath,
			time.Now().Format("20060201T150405Z"))),
	}, ginContext.Request, vaultSecretData, values.OriginalYaml, repoFiles); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	if err := CacheRepoFiles(a.TempFolder, r.Name, repoFiles, a.Cache); err != nil {
		return fmt.Errorf("unable to cache repo file, %w", err)
	}

	changedValuesHash, err := MapHash(values.OriginalYaml)
	if err != nil {
		return errors.Wrap(err, "unable to get values map hash")
	}

	if initialValuesHash != changedValuesHash || len(repoFiles) > 0 {
		if err := CreateEditMergeRequest(ginContext, r.Name, values.OriginalYaml, a.Gerrit, mrActions); err != nil {
			return errors.Wrap(err, "unable to create edit merge request")
		}
	}

	if len(vaultSecretData) > 0 {
		if err := CreateVaultSecrets(a.Vault, vaultSecretData, false); err != nil {
			return errors.Wrap(err, "unable to create vault secrets")
		}
	}

	if err := cbService.Update(ctx, cb); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	return nil
}

func MapHash(v map[string]interface{}) (string, error) {
	bts, err := json.Marshal(v)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode map")
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
		return errors.Wrap(err, "unable to encode values yaml")
	}

	mrExists, err := ProjectHasOpenMR(ctx, projectName, gerritService)
	if err != nil {
		return errors.Wrap(err, "unable to check project MR exists")
	}

	if mrExists {
		return MRExists("there is already open merge request(s) for this registry")
	}

	mrs, err := gerritService.GetMergeRequestByProject(ctx, projectName)
	if err != nil {
		return errors.Wrap(err, "unable to get MRs")
	}

	for _, mr := range mrs {
		if mr.Status.Value == gerrit.StatusNew {
			return MRExists("there is already open merge request(s) for this registry")
		}
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
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}

func ProjectHasOpenMR(ctx *gin.Context, projectName string, gerritService gerrit.ServiceInterface) (bool, error) {
	mrs, err := gerritService.GetMergeRequestByProject(ctx, projectName)
	if err != nil {
		return false, errors.Wrap(err, "unable to get MRs")
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
		return nil, errors.Wrap(err, "unable to unmarshal admins")
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
