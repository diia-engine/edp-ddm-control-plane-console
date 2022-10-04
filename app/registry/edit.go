package registry

import (
	"context"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/k8s"
)

func (a *App) editRegistryGet(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.Param("name")

	if err := a.checkUpdateAccess(registryName, k8sService); err != nil {
		return nil, errors.New("access denied")
	}

	reg, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	model := registry{KeyDeviceType: KeyDeviceTypeFile}

	hwINITemplateContent, err := GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	hasUpdate, branches, err := HasUpdate(userCtx, a.Services.Gerrit, reg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	responseParams := gin.H{
		"dnsManual":            false,
		"registry":             reg,
		"page":                 "registry",
		"hwINITemplateContent": hwINITemplateContent,
		"updateBranches":       branches,
		"hasUpdate":            hasUpdate,
	}

	if err := a.loadValuesEditConfig(ctx, registryName, responseParams, &model); err != nil {
		return nil, errors.Wrap(err, "unable to load edit values from config")
	}

	return router.MakeResponse(200, "registry/edit.html", responseParams), nil
}

func (a *App) loadValuesEditConfig(ctx context.Context, registryName string, rspParams gin.H, r *registry) error {
	values, err := GetValuesFromGit(ctx, registryName, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to get values from git")
	}

	if err := a.loadSMTPConfig(values, rspParams); err != nil {
		return errors.Wrap(err, "unable to load smtp config")
	}

	if err := a.loadCIDRConfig(values, rspParams); err != nil {
		return errors.Wrap(err, "unable to load cidr config")
	}

	if err := a.loadAdminsConfig(values, r); err != nil {
		return errors.Wrap(err, "unable to load admins config")
	}

	rspParams["model"] = r

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

func (a *App) loadCIDRConfig(valuesDict map[string]interface{}, rspParams gin.H) error {
	rspParams["cidrConfig"] = "{}"

	//TODO: refactor
	global, ok := valuesDict["global"]
	if !ok {
		return nil
	}
	globalDict, ok := global.(map[string]interface{})
	if !ok {
		return nil
	}

	whiteListIP, ok := globalDict["whiteListIP"]
	if !ok {
		return nil
	}

	whiteListIPDict, ok := whiteListIP.(map[string]interface{})
	if !ok {
		return nil
	}

	if adminRoutes, ok := whiteListIPDict["adminRoutes"]; ok {
		whiteListIPDict["admin"] = strings.Split(adminRoutes.(string), " ")
	}

	if citizenPortal, ok := whiteListIPDict["citizenPortal"]; ok {
		whiteListIPDict["citizen"] = strings.Split(citizenPortal.(string), " ")
	}

	if officerPortal, ok := whiteListIPDict["officerPortal"]; ok {
		whiteListIPDict["officer"] = strings.Split(officerPortal.(string), " ")
	}

	cidrConfig, err := json.Marshal(whiteListIPDict)
	if err != nil {
		return errors.Wrap(err, "unable to encode cidr to JSON")
	}

	rspParams["cidrConfig"] = string(cidrConfig)
	return nil
}

func (a *App) loadSMTPConfig(values map[string]interface{}, rspParams gin.H) error {
	global, ok := values["global"]
	if !ok {
		rspParams["smtpConfig"] = "{}"
		return nil
	}

	globalDict := global.(map[string]interface{})
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

func (a *App) editRegistryPost(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
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
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/edit.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "registry": r, "model": r}), nil
	}

	if err := a.editRegistry(userCtx, ctx, &r, cb, cbService); err != nil {
		_, ok := errors.Cause(err).(MRExists)
		if ok {
			return router.MakeResponse(200, "registry/edit.html",
				gin.H{"page": "registry", "MRExists": true, "registry": r, "model": r}), nil
		}

		validationErrors, ok := errors.Cause(err).(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/edit.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "registry": r, "model": r}), nil
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

	values, err := GetValuesFromGit(ctx, r.Name, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to get values from git")
	}

	vaultSecretData := make(map[string]map[string]interface{})
	if err := a.prepareDNSConfig(ginContext, r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare dns config")
	}

	if err := CreateRegistryKeys(keyManagement{r: r, vaultSecretPath: a.keyManagementRegistryVaultPath(r.Name)}, ginContext.Request,
		vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	if err := a.prepareMailServerConfig(ginContext, r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare mail server config")
	}

	if err := a.prepareCIDRConfig(r, values); err != nil {
		return errors.Wrap(err, "unable to prepare cidr config")
	}

	if err := a.prepareAdminsConfig(r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare admin values config")
	}

	if len(values) > 0 {
		if err := CreateEditMergeRequest(ginContext, r.Name, values, a.Gerrit); err != nil {
			return errors.Wrap(err, "unable to create edit merge request")
		}
	}

	if len(vaultSecretData) > 0 {
		if err := CreateVaultSecrets(a.Vault, vaultSecretData); err != nil {
			return errors.Wrap(err, "unable to create vault secrets")
		}
	}

	if err := cbService.Update(cb); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	//if err := a.Services.Jenkins.CreateJobBuildRun(fmt.Sprintf("registry-update-%d", time.Now().Unix()),
	//	fmt.Sprintf("%s/job/MASTER-Build-%s/", r.Name, r.Name), nil); err != nil {
	//	return errors.Wrap(err, "unable to trigger jenkins job build run")
	//}

	return nil
}

func CreateEditMergeRequest(ctx *gin.Context, projectName string, editValues map[string]interface{},
	gerritService gerrit.ServiceInterface) error {
	values, err := GetValuesFromGit(ctx, projectName, gerritService)
	if err != nil {
		return errors.Wrap(err, "unable to get values from git")
	}

	for k, v := range editValues {
		values[k] = v
	}

	valuesYaml, err := yaml.Marshal(values)
	if err != nil {
		return errors.Wrap(err, "unable to encode values yaml")
	}

	mrs, err := gerritService.GetMergeRequestByProject(ctx, projectName)
	if err != nil {
		return errors.Wrap(err, "unable to get MRs")
	}

	for _, mr := range mrs {
		if mr.Status.Value == "NEW" {
			return MRExists("there is already open merge request(s) for this registry")
		}
	}

	if err := gerritService.CreateMergeRequestWithContents(ctx, &gerrit.MergeRequest{
		ProjectName:   projectName,
		Name:          fmt.Sprintf("reg-edit-mr-%s-%d", projectName, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("edit registry"),
		TargetBranch:  "master",
		Labels: map[string]string{
			MRLabelTarget: mrTargetEditRegistry,
		},
		Annotations: map[string]string{},
	}, map[string]string{
		ValuesLocation: string(valuesYaml),
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
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
