package registry

import (
	"context"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/vault"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationSMPTType            = "registry-parameters/smtp-type"
	AnnotationSMPTOpts            = "registry-parameters/smtp-opts"
	AnnotationTemplateName        = "registry-parameters/template-name"
	AnnotationCreatorUsername     = "registry-parameters/creator-username"
	AnnotationCreatorEmail        = "registry-parameters/creator-email"
	AnnotationValues              = "registry-parameters/values"
	AdministratorsValuesKey       = "administrators"
	ResourcesValuesKey            = "registry"
	VaultKeyCACert                = "caCertificate"
	VaultKeyCert                  = "certificate"
	VaultKeyPK                    = "key"
	externalSystemsKey            = "external-systems"
	externalSystemDefaultProtocol = "REST"
	externalSystemDeletableType   = "registry"
	trembitaRegistriesKey         = "registries"
	trembitaValuesKey             = "trembita"
	trembitaRegistriesValuesKet   = "registries"
)

func (a *App) createUpdateRegistryProcessors() []func(ctx *gin.Context, r *registry, values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
	return []func(*gin.Context, *registry, *Values,
		map[string]map[string]interface{}, *[]string) error{
		a.prepareDNSConfig,
		a.prepareCIDRConfig,
		a.prepareMailServerConfig,
		a.prepareAdminsConfig,
		a.prepareRegistryResources,
		a.prepareSupplierAuthConfig,
		a.prepareBackupSchedule,
		a.prepareKeycloakCustomHostname,
		a.prepareEDRCheck,
	}
}

func (a *App) validatePEMFile(ctx *gin.Context) (rsp router.Response, retErr error) {
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get form file")
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read file data")
	}

	if _, err := DecodePEM(data); err != nil {
		return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
	}

	return router.MakeStatusResponse(http.StatusOK), nil
}

func (a *App) registryNameAvailable(ctx *gin.Context) (rsp router.Response, retErr error) {
	name := ctx.Param("name")
	_, err := a.Codebase.Get(name)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return router.MakeStatusResponse(http.StatusNotFound), nil
		}

		return nil, errors.Wrap(err, "unable to check codebase existance")
	}

	return router.MakeStatusResponse(http.StatusOK), nil
}

func (a *App) createRegistryGet(ctx *gin.Context) (response router.Response, retErr error) {
	prjs, err := a.Services.Gerrit.GetProjects(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs)

	userCtx := router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	if err := a.checkCreateAccess(k8sService); err != nil {
		return nil, errors.Wrap(err, "error during create access check")
	}

	hwINITemplateContent, err := GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	dnsManual, err := a.getDNSManualURL(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get dns manual")
	}

	gerritBranches, err := formatGerritProjectBranches(prjs)
	if err != nil {
		return nil, errors.Wrap(err, "unable to format gerrit project branches")
	}

	keycloakHostname, err := LoadKeycloakDefaultHostname(ctx, a.KeycloakDefaultHostname, a.EDPComponent)
	if err != nil {
		return nil, fmt.Errorf("unable to load keycloak default hostname, %w", err)
	}

	keycloakHostnames, err := a.loadKeycloakHostnames()
	if err != nil {
		return nil, fmt.Errorf("unable to load keycloak hostnames, %w", err)
	}

	return router.MakeHTMLResponse(200, "registry/create.html", gin.H{
		"dnsManual":            dnsManual,
		"page":                 "registry",
		"gerritProjects":       prjs,
		"gerritBranches":       gerritBranches,
		"model":                registry{KeyDeviceType: KeyDeviceTypeFile},
		"hwINITemplateContent": hwINITemplateContent,
		"smtpConfig":           "{}",
		"cidrConfig":           "{}",
		"action":               "create",
		"registryData":         "{}",
		"keycloakHostname":     keycloakHostname,
		"keycloakHostnames":    keycloakHostnames,
	}), nil
}

func (a *App) getDNSManualURL(ctx context.Context) (string, error) {
	com, err := a.EDPComponent.Get(ctx, a.Config.DDMManualEDPComponent)
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			return "", nil
		}

		return "", errors.Wrap(err, "unable to get edp component")
	}

	u, err := url.Parse(com.Spec.Url)
	if err != nil {
		return "", errors.Wrap(err, "unable to parse url")
	}

	u.Path = path.Join(u.Path, a.Config.RegistryDNSManualPath)

	return u.String(), nil
}

func headsCount(refs []string) int {
	cnt := 0
	for _, ref := range refs {
		if strings.Contains(ref, "refs/heads") {
			cnt += 1
		}
	}

	return cnt
}

func (a *App) filterProjects(projects []gerrit.GerritProject) []gerrit.GerritProject {
	filteredProjects := make([]gerrit.GerritProject, 0, 4)
	for _, prj := range projects {
		if strings.Contains(prj.Spec.Name, a.Config.GerritRegistryPrefix) {
			if headsCount(prj.Status.Branches) > 1 {
				var branches []string
				for _, br := range prj.Status.Branches {
					if !strings.Contains(br, "master") {
						branches = append(branches, br)
					}
				}
				prj.Status.Branches = branches
			}

			filteredProjects = append(filteredProjects, prj)
		}
	}

	return filteredProjects
}

func formatGerritProjectBranches(projects []gerrit.GerritProject) (string, error) {
	res := make(map[string][]string)
	for _, p := range projects {
		var branches []string
		for _, b := range p.Status.Branches {
			idx := strings.Index(b, "heads/")
			if idx != -1 {
				branches = append(branches, b[idx+6:])
			}
		}
		res[p.Spec.Name] = branches
	}

	bts, err := json.Marshal(res)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode project branches")
	}

	return string(bts), nil
}

func GetINITemplateContent(path string) (string, error) {
	iniTemplate, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "unable to open ini template file")
	}

	data, err := ioutil.ReadAll(iniTemplate)
	if err != nil {
		return "", errors.Wrap(err, "unable to read ini template data")
	}

	return string(data), nil
}

func (a *App) createRegistryPost(ctx *gin.Context) (response router.Response, retErr error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get k8s service for user")
	}

	if err := a.checkCreateAccess(k8sService); err != nil {
		return nil, errors.Wrap(err, "error during create access check")
	}

	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	r := registry{Scenario: ScenarioKeyRequired}
	if err := ctx.ShouldBind(&r); err != nil {
		return nil, errors.Wrap(err, "unable to parse form")
	}

	if err := a.validateCreateRegistryGitTemplate(ctx, &r); err != nil {
		return nil, errors.Wrap(err, "unable to validate create registry git template")
	}

	if err := a.createRegistry(userCtx, ctx, &r, cbService, k8sService); err != nil {
		return nil, errors.Wrap(err, "unable to create registry")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/registry/overview"), nil
}

func (a *App) checkCreateAccess(userK8sService k8s.ServiceInterface) error {
	allowedToCreate, err := a.Services.Codebase.CheckIsAllowedToCreate(userK8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check create access")
	}
	if !allowedToCreate {
		return errors.New("access denied")
	}

	return nil
}

func (a *App) validateCreateRegistryGitTemplate(ctx context.Context, r *registry) error {
	prjs, err := a.Services.Gerrit.GetProjects(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs)

	for _, prj := range prjs {
		if prj.Spec.Name == r.RegistryGitTemplate {
			for _, br := range prj.Status.Branches {
				if br == fmt.Sprintf("refs/heads/%s", r.RegistryGitBranch) {
					return nil
				}
			}
		}
	}

	return errors.New("wrong registry template selected")
}

func (a *App) createRegistry(ctx context.Context, ginContext *gin.Context, r *registry,
	cbService codebase.ServiceInterface, k8sService k8s.ServiceInterface) error {
	_, err := cbService.Get(r.Name)
	if err == nil {
		return validator.ValidationErrors([]validator.FieldError{router.MakeFieldError("Name", "registry-exists")})
	}
	if !k8sErrors.IsNotFound(err) {
		return errors.Wrap(err, "unknown error")
	}

	values, err := GetValuesFromGit(r.RegistryGitTemplate, r.RegistryGitBranch, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to load values from template")
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
		return errors.Wrap(err, "unable to prepare registry keys")
	}

	if err := CacheRepoFiles(a.TempFolder, r.Name, repoFiles, a.Cache); err != nil {
		return fmt.Errorf("unable to cache repo file, %w", err)
	}

	if err := CreateVaultSecrets(a.Vault, vaultSecretData, false); err != nil {
		return errors.Wrap(err, "unable to create vault secrets")
	}

	if err := a.Services.Gerrit.CreateProject(ctx, r.Name); err != nil {
		return errors.Wrap(err, "unable to create gerrit project")
	}

	cb := a.prepareRegistryCodebase(r)
	valuesEncoded, err := a.encodeValues(r.Name, values.OriginalYaml)
	if err != nil {
		return errors.Wrap(err, "unable to encode values")
	}

	cb.Annotations = map[string]string{
		AnnotationTemplateName:    r.RegistryGitTemplate,
		AnnotationCreatorUsername: ginContext.GetString(router.UserNameSessionKey),
		AnnotationCreatorEmail:    ginContext.GetString(router.UserEmailSessionKey),
		AnnotationValues:          valuesEncoded,
	}

	if err := cbService.Create(cb); err != nil {
		return errors.Wrap(err, "unable to create codebase")
	}

	if err := cbService.CreateDefaultBranch(cb); err != nil {
		return errors.Wrap(err, "unable to create default branch")
	}

	return nil
}

func ModifyVaultPath(path string) string {
	if strings.Contains(path, "/data/") {
		return path
	}

	pathParts := strings.Split(path, "/")
	pathParts = append(pathParts[:1], append([]string{"data"}, pathParts[1:]...)...)
	return strings.Join(pathParts, "/")
}

func CreateVaultSecrets(v vault.ServiceInterface, secretData map[string]map[string]interface{}, append bool) error {
	for vPath, pathSecretData := range secretData {
		modPath := ModifyVaultPath(vPath)

		if append {
			sec, err := v.Read(modPath)
			if err != nil {
				return errors.Wrap(err, "unable to read secret")
			}

			if sec != nil {
				currentSecretData, ok := sec.Data["data"]
				if ok {
					currentSecretData, ok := currentSecretData.(map[string]interface{})
					if ok {
						for k, v := range pathSecretData {
							currentSecretData[k] = v
						}

						pathSecretData = currentSecretData
					}
				}
			}
		}

		if _, err := v.Write(modPath, map[string]interface{}{
			"data": pathSecretData,
		}); err != nil {
			return errors.Wrap(err, "unable to write to vault")
		}
	}

	return nil
}

func (a *App) encodeValues(registryName string, values map[string]interface{}) (string, error) {
	values["registryVaultPath"] = a.vaultRegistryPath(registryName)
	bts, err := json.Marshal(values)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode values to JSON")
	}

	return string(bts), nil
}

func (a *App) vaultRegistryPath(registryName string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(a.Config.VaultRegistrySecretPathTemplate, "{registry}", registryName),
		"{engine}", a.Config.VaultKVEngineName)
}

func (a *App) vaultRegistryPathKey(registryName, key string) string {
	return fmt.Sprintf("%s/%s", a.vaultRegistryPath(registryName), key)
}

func (a *App) keyManagementRegistryVaultPath(registryName string) string {
	return a.vaultRegistryPath(registryName) + "/key-management"
}

func (a *App) prepareCIDRConfig(ctx *gin.Context, r *registry, _values *Values,
	_ map[string]map[string]interface{}, mrActions *[]string) error {
	values := _values.OriginalYaml
	//TODO: refactor to new values

	//TODO: remove this check
	if ctx.PostForm("action") == "edit" && ctx.PostForm("cidr-changed") == "" {
		return nil
	}

	globalInterface, ok := values["global"]
	if !ok {
		globalInterface = make(map[string]interface{})
	}
	globalDict := globalInterface.(map[string]interface{})

	whiteListInterface, ok := globalDict["whiteListIP"]
	if !ok {
		whiteListInterface = make(map[string]interface{})
	}
	whiteListDict := whiteListInterface.(map[string]interface{})

	if err := handleCIDRCategory(r.CIDRCitizen, "citizenPortal", whiteListDict); err != nil {
		return errors.Wrap(err, "unable to handle cidr category")
	}

	if err := handleCIDRCategory(r.CIDROfficer, "officerPortal", whiteListDict); err != nil {
		return errors.Wrap(err, "unable to handle cidr category")
	}

	if err := handleCIDRCategory(r.CIDRAdmin, "adminRoutes", whiteListDict); err != nil {
		return errors.Wrap(err, "unable to handle cidr category")
	}

	globalDict["whiteListIP"] = whiteListDict
	values["global"] = globalDict

	return nil
}

func handleCIDRCategory(cidrCategory, dictIndex string, whiteListDict map[string]interface{}) error {
	if cidrCategory == "" {
		return nil
	}

	var cidr []string
	if err := json.Unmarshal([]byte(cidrCategory), &cidr); err != nil {
		return errors.Wrap(err, "unable to decode cidr")
	}

	if len(cidr) > 0 {
		whiteListDict[dictIndex] = strings.Join(cidr, " ")
	} else if _, ok := whiteListDict[dictIndex]; ok {
		delete(whiteListDict, dictIndex)
	}

	return nil
}

func (a *App) prepareAdminsConfig(_ *gin.Context, r *registry, _values *Values,
	secrets map[string]map[string]interface{}, mrActions *[]string) error {
	values := _values.OriginalYaml
	//TODO: refactor to new values

	//TODO: don't recreate admin secrets for existing admin
	if r.Admins != "" && r.AdminsChanged == "on" {
		admins, err := validateAdmins(r.Admins)
		if err != nil {
			return errors.Wrap(err, "unable to validate admins")
		}

		adminsVaultPath := a.vaultRegistryPathKey(r.Name, "administrators")
		for i, adm := range admins {
			adminVaultPath := fmt.Sprintf("%s/%s", adminsVaultPath, adm.Email)
			secrets[adminVaultPath] = map[string]interface{}{
				"password": adm.TmpPassword,
			}

			admins[i].PasswordVaultSecret = adminVaultPath
			admins[i].PasswordVaultSecretKey = "password"
			admins[i].TmpPassword = ""
		}

		values[AdministratorsValuesKey] = admins
	}

	return nil
}

func (a *App) prepareDNSConfig(ginContext *gin.Context, r *registry, _values *Values,
	secretData map[string]map[string]interface{}, mrActions *[]string) error {
	//TODO: refactor to new values struct
	values := _values.OriginalYaml

	portals, ok := values["portals"]
	if !ok {
		portals = make(map[string]interface{})
	}
	portalsDict := portals.(map[string]interface{})

	citizen, ok := portalsDict["citizen"]
	if !ok {
		citizen = make(map[string]interface{})
	}
	citizenDict := citizen.(map[string]interface{})

	officer, ok := portalsDict["officer"]
	if !ok {
		officer = make(map[string]interface{})
	}
	officerDict := officer.(map[string]interface{})

	if r.DNSNameOfficerEnabled == "" {
		delete(portalsDict, "officer")
		delete(officerDict, "customDns")
	} else if r.DNSNameOfficerEnabled != "" && r.DNSNameOfficer != "" {
		customDNS := make(map[string]interface{})
		customDNS["enabled"] = true
		customDNS["host"] = r.DNSNameOfficer
		officerDict["customDns"] = customDNS

		certFile, _, err := ginContext.Request.FormFile("officer-ssl")
		if err != nil {
			return errors.Wrap(err, "unable to get officer ssl certificate")
		}

		certData, err := ioutil.ReadAll(certFile)
		if err != nil {
			return errors.Wrap(err, "unable to read officer ssl data")
		}

		pemInfo, err := DecodePEM(certData)
		if err != nil {
			return validator.ValidationErrors([]validator.FieldError{
				router.MakeFieldError("DNSNameOfficer", "pem-decode-error")})
		}

		secretPath := strings.ReplaceAll(a.Config.VaultOfficerSSLPath, "{registry}", r.Name)
		secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameOfficer)

		if _, ok := secretData[secretPath]; !ok {
			secretData[secretPath] = make(map[string]interface{})
		}

		secretData[secretPath][VaultKeyCACert] = pemInfo.CACert
		secretData[secretPath][VaultKeyCert] = pemInfo.Cert
		secretData[secretPath][VaultKeyPK] = pemInfo.PrivateKey
	}

	if r.DNSNameCitizenEnabled == "" {
		delete(portalsDict, "citizen")
		delete(citizenDict, "customDns")
	} else if r.DNSNameCitizenEnabled != "" && r.DNSNameCitizen != "" {
		customDNS := make(map[string]interface{})
		customDNS["enabled"] = true
		customDNS["host"] = r.DNSNameCitizen
		citizenDict["customDns"] = customDNS

		certFile, _, err := ginContext.Request.FormFile("citizen-ssl")
		if err != nil {
			return errors.Wrap(err, "unable to get citizen ssl certificate")
		}

		certData, err := ioutil.ReadAll(certFile)
		if err != nil {
			return errors.Wrap(err, "unable to read citizen ssl data")
		}

		pemInfo, err := DecodePEM(certData)
		if err != nil {
			return validator.ValidationErrors([]validator.FieldError{
				router.MakeFieldError("DNSNameCitizen", "pem-decode-error")})
		}

		secretPath := strings.ReplaceAll(a.Config.VaultCitizenSSLPath, "{registry}", r.Name)
		secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameCitizen)

		if _, ok := secretData[secretPath]; !ok {
			secretData[secretPath] = make(map[string]interface{})
		}

		secretData[secretPath][VaultKeyCACert] = pemInfo.CACert
		secretData[secretPath][VaultKeyCert] = pemInfo.Cert
		secretData[secretPath][VaultKeyPK] = pemInfo.PrivateKey
	}

	if len(citizenDict) > 0 {
		portalsDict["citizen"] = citizenDict
	}

	if len(officerDict) > 0 {
		portalsDict["officer"] = officerDict
	}

	values["portals"] = portalsDict

	return nil
}

func (a *App) prepareMailServerConfig(_ *gin.Context, r *registry, _values *Values,
	secretData map[string]map[string]interface{}, mrActions *[]string) error {
	values := _values.OriginalYaml
	//TODO: refactor to new values

	notifications := make(map[string]interface{})

	if r.MailServerType == SMTPTypeExternal {
		var smptOptsDict map[string]string
		if err := json.Unmarshal([]byte(r.MailServerOpts), &smptOptsDict); err != nil {
			return errors.Wrap(err, "unable to decode mail server opts")
		}

		pwd, ok := smptOptsDict["password"]
		if !ok {
			return errors.New("no password in mail server opts")
		}

		vaultPath := a.vaultRegistryPathKey(r.Name, "smtp")

		if _, ok := secretData[vaultPath]; !ok {
			secretData[vaultPath] = make(map[string]interface{})
		}

		secretData[vaultPath][a.Config.VaultRegistrySMTPPwdSecretKey] = pwd
		//TODO: remove password from dict

		port, err := strconv.ParseInt(smptOptsDict["port"], 10, 32)
		if err != nil {
			return errors.Wrapf(err, "wrong smtp port value: %s", smptOptsDict["port"])
		}

		notifications["email"] = map[string]interface{}{
			"type":      "external",
			"host":      smptOptsDict["host"],
			"port":      port,
			"address":   smptOptsDict["address"],
			"password":  smptOptsDict["password"],
			"vaultPath": a.vaultRegistryPath(r.Name),
			"vaultKey":  a.Config.VaultRegistrySMTPPwdSecretKey,
		}
	} else {
		notifications["email"] = map[string]interface{}{
			"type": "internal",
		}
	}

	globalInterface, ok := values["global"]
	if !ok {
		globalInterface = make(map[string]interface{})
	}
	globalDict := globalInterface.(map[string]interface{})

	globalDict["notifications"] = notifications
	values["global"] = globalDict

	return nil
}

func (a *App) prepareRegistryCodebase(r *registry) *codebase.Codebase {
	jobProvisioning := "default"
	startVersion := "0.0.1"
	jenkinsSlave := "gitops"
	gitURL := codebase.RepoNotReady
	cb := codebase.Codebase{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1alpha1",
			Kind:       "Codebase",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: r.Name,
		},
		Spec: codebase.CodebaseSpec{
			Description:      &r.Description,
			Type:             "registry",
			BuildTool:        "gitops",
			Lang:             "other",
			DefaultBranch:    r.RegistryGitBranch,
			Strategy:         "import",
			DeploymentScript: "openshift-template",
			GitServer:        "gerrit",
			GitUrlPath:       &gitURL,
			CiTool:           "Jenkins",
			JobProvisioning:  &jobProvisioning,
			Versioning: codebase.Versioning{
				StartFrom: &startVersion,
				Type:      "edp",
			},
			Repository: &codebase.Repository{
				Url: gitURL,
			},
			JenkinsSlave: &jenkinsSlave,
		},
		Status: codebase.CodebaseStatus{
			Available:       false,
			LastTimeUpdated: time.Now(),
			Status:          "initialized",
			Action:          "codebase_registration",
			Value:           "inactive",
		},
	}

	if cb.Spec.DefaultBranch != "master" {
		cb.Spec.BranchToCopyInDefaultBranch = cb.Spec.DefaultBranch
		cb.Spec.DefaultBranch = "master"

		if a.EnableBranchProvisioners {
			jobProvisioning = branchProvisioner(cb.Spec.BranchToCopyInDefaultBranch)
			cb.Spec.JobProvisioning = &jobProvisioning
		}
	}

	if a.codebaseLabels != nil && len(a.codebaseLabels) > 0 {
		cb.SetLabels(a.codebaseLabels)
	}

	return &cb
}

func branchProvisioner(branch string) string {
	return "default-" + strings.Replace(
		strings.ToLower(branch), ".", "-", -1)
}
