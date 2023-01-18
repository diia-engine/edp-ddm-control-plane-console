package registry

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/k8s"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AnnotationSMPTType        = "registry-parameters/smtp-type"
	AnnotationSMPTOpts        = "registry-parameters/smtp-opts"
	AnnotationTemplateName    = "registry-parameters/template-name"
	AnnotationCreatorUsername = "registry-parameters/creator-username"
	AnnotationCreatorEmail    = "registry-parameters/creator-email"
	AnnotationValues          = "registry-parameters/values"
	AdministratorsValuesKey   = "administrators"
	ResourcesValuesKey        = "registry"
	VaultKeyCACert            = "caCertificate"
	VaultKeyCert              = "certificate"
	VaultKeyPK                = "key"
)

func (a *App) validatePEMFile(ctx *gin.Context) (rsp router.Response, retErr error) {
	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get form file")
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read file data")
	}

	if _, _, _, err := decodePEM(data); err != nil {
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

	userCtx := a.router.ContextWithUserAccessToken(ctx)
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
	userCtx := a.router.ContextWithUserAccessToken(ctx)

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

	values, err := a.GetValuesFromBranch(r.RegistryGitTemplate, r.RegistryGitBranch)
	if err != nil {
		return errors.Wrap(err, "unable to load values from template")
	}

	vaultSecretData := make(map[string]map[string]interface{})
	if err := a.prepareDNSConfig(ginContext, r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare dns config")
	}

	if err := a.prepareCIDRConfig(ginContext, r, values); err != nil {
		return errors.Wrap(err, "unable to prepare cidr config")
	}

	if err := a.prepareMailServerConfig(ginContext, r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare mail server config")
	}

	if err := a.prepareAdminsConfig(r, vaultSecretData, values); err != nil {
		return errors.Wrap(err, "unable to prepare admins config")
	}

	if err := a.prepareRegistryResources(r, values); err != nil {
		return errors.Wrap(err, "unable to prepare registry resources config")
	}

	if _, err := prepareRegistryKeys(keyManagement{r: r}, ginContext.Request, k8sService); err != nil {
		return errors.Wrap(err, "unable to create registry keys")
	}

	if err := a.createVaultSecrets(vaultSecretData); err != nil {
		return errors.Wrap(err, "unable to create vault secrets")
	}

	if err := a.Services.Gerrit.CreateProject(ctx, r.Name); err != nil {
		return errors.Wrap(err, "unable to create gerrit project")
	}

	cb := a.prepareRegistryCodebase(r)
	valuesEncoded, err := a.encodeValues(r.Name, values)
	if err != nil {
		return errors.Wrap(err, "unable to encode values")
	}

	cb.Annotations = map[string]string{
		AnnotationTemplateName:    r.RegistryGitTemplate,
		AnnotationSMPTType:        r.MailServerType, //TODO: remove
		AnnotationSMPTOpts:        r.MailServerOpts, //TODO: remove
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

func (a *App) createVaultSecrets(secretData map[string]map[string]interface{}) error {
	for vaultPath, pathSecretData := range secretData {
		vaultPath = ModifyVaultPath(vaultPath)

		if _, err := a.Vault.Write(vaultPath, map[string]interface{}{
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

func (a *App) keyManagementRegistryVaultPath(registryName string) string {
	return a.vaultRegistryPath(registryName) + "/key-management"
}

func (a *App) prepareCIDRConfig(ginContext *gin.Context, r *registry, values map[string]interface{}) error {
	if ginContext.PostForm("action") == "edit" && ginContext.PostForm("cidr-changed") == "" {
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

	if r.CIDRCitizen != "" {
		var cidr []string
		if err := json.Unmarshal([]byte(r.CIDRCitizen), &cidr); err != nil {
			return errors.Wrap(err, "unable to decode cidr")
		}

		if len(cidr) > 0 {
			whiteListDict["citizenPortal"] = strings.Join(cidr, " ")
		} else if _, ok = whiteListDict["citizenPortal"]; ok {
			delete(whiteListDict, "citizenPortal")
		}
	}

	if r.CIDROfficer != "" {
		var cidr []string
		if err := json.Unmarshal([]byte(r.CIDROfficer), &cidr); err != nil {
			return errors.Wrap(err, "unable to decode cidr")
		}

		if len(cidr) > 0 {
			whiteListDict["officerPortal"] = strings.Join(cidr, " ")
		} else if _, ok = whiteListDict["officerPortal"]; ok {
			delete(whiteListDict, "officerPortal")
		}
	}

	if r.CIDRAdmin != "" {
		var cidr []string
		if err := json.Unmarshal([]byte(r.CIDRAdmin), &cidr); err != nil {
			return errors.Wrap(err, "unable to decode cidr")
		}

		if len(cidr) > 0 {
			whiteListDict["adminRoutes"] = strings.Join(cidr, " ")
		} else if _, ok = whiteListDict["adminRoutes"]; ok {
			delete(whiteListDict, "adminRoutes")
		}
	}

	globalDict["whiteListIP"] = whiteListDict
	values["global"] = globalDict

	return nil
}

func (a *App) prepareAdminsConfig(r *registry, secretData map[string]map[string]interface{},
	values map[string]interface{}) error {
	//TODO: don't recreate admin secrets for existing admin

	if r.Admins != "" && r.AdminsChanged == "on" {
		admins, err := validateAdmins(r.Admins)
		if err != nil {
			return errors.Wrap(err, "unable to validate admins")
		}

		adminsVaultPath := fmt.Sprintf("%s/administrators", a.vaultRegistryPath(r.Name))
		for i, adm := range admins {
			adminVaultPath := fmt.Sprintf("%s/%s", adminsVaultPath, adm.Email)
			secretData[adminVaultPath] = map[string]interface{}{
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

func (a *App) prepareDNSConfig(ginContext *gin.Context, r *registry, secretData map[string]map[string]interface{},
	values map[string]interface{}) error {
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

	if r.DNSNameOfficer != "" {
		if r.DNSNameOfficer == "-" {
			delete(portalsDict, "officer")
			delete(officerDict, "customDns")
		} else {
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

			caCert, cert, key, err := decodePEM(certData)
			if err != nil {
				return validator.ValidationErrors([]validator.FieldError{
					router.MakeFieldError("DNSNameOfficer", "pem-decode-error")})
			}

			secretPath := strings.ReplaceAll(a.Config.VaultOfficerSSLPath, "{registry}", r.Name)
			secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameOfficer)

			if _, ok := secretData[secretPath]; !ok {
				secretData[secretPath] = make(map[string]interface{})
			}

			secretData[secretPath][VaultKeyCACert] = caCert
			secretData[secretPath][VaultKeyCert] = cert
			secretData[secretPath][VaultKeyPK] = key
		}
	}

	if r.DNSNameCitizen != "" {
		if r.DNSNameCitizen == "-" {
			delete(portalsDict, "citizen")
			delete(citizenDict, "customDns")
		} else {
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

			caCert, cert, key, err := decodePEM(certData)
			if err != nil {
				return validator.ValidationErrors([]validator.FieldError{
					router.MakeFieldError("DNSNameCitizen", "pem-decode-error")})
			}

			secretPath := strings.ReplaceAll(a.Config.VaultCitizenSSLPath, "{registry}", r.Name)
			secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameCitizen)

			if _, ok := secretData[secretPath]; !ok {
				secretData[secretPath] = make(map[string]interface{})
			}

			secretData[secretPath][VaultKeyCACert] = caCert
			secretData[secretPath][VaultKeyCert] = cert
			secretData[secretPath][VaultKeyPK] = key
		}
	}

	//if r.DNSNameKeycloak != "" {
	//	certFile, _, err := ginContext.Request.FormFile("keycloak-ssl")
	//	if err != nil {
	//		return errors.Wrap(err, "unable to get citizen ssl certificate")
	//	}
	//
	//	certData, err := ioutil.ReadAll(certFile)
	//	if err != nil {
	//		return errors.Wrap(err, "unable to read citizen ssl data")
	//	}
	//
	//	caCert, cert, key, err := decodePEM(certData)
	//	if err != nil {
	//		return validator.ValidationErrors([]validator.FieldError{
	//			router.MakeFieldError("DNSNameKeycloak", "pem-decode-error")})
	//	}
	//
	//	secretPath := strings.ReplaceAll(a.Config.VaultCitizenSSLPath, "{registry}", r.Name)
	//	secretPath = strings.ReplaceAll(secretPath, "{host}", r.DNSNameKeycloak)
	//
	//	if _, ok := secretData[secretPath]; !ok {
	//		secretData[secretPath] = make(map[string]interface{})
	//	}
	//
	//	secretData[secretPath][VaultKeyCACert] = caCert
	//	secretData[secretPath][VaultKeyCert] = cert
	//	secretData[secretPath][VaultKeyPK] = key
	//
	//	kcInterface, ok := values["keycloak"]
	//	if !ok {
	//		kcInterface = make(map[string]interface{})
	//	}
	//	kcDict := kcInterface.(map[string]interface{})
	//	kcDict["customHost"] = r.DNSNameKeycloak
	//	values["keycloak"] = kcDict
	//}

	if len(citizenDict) > 0 {
		portalsDict["citizen"] = citizenDict
	}

	if len(officerDict) > 0 {
		portalsDict["officer"] = officerDict
	}

	//valuesPortalsInterface, ok := values["portals"]
	//if !ok {
	//	valuesPortalsInterface = map[string]interface{}{}
	//}
	//valuesPortalsDict := valuesPortalsInterface.(map[string]interface{})
	//
	//for k, v := range portalsDict {
	//	valuesPortalsDict[k] = v
	//}

	values["portals"] = portalsDict

	return nil
}

func decodePEM(buf []byte) (caCert string, cert string, privateKey string, retErr error) {
	var (
		block   *pem.Block
		caBlock bytes.Buffer
	)

	for {
		var tmp bytes.Buffer

		block, buf = pem.Decode(buf)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			x509Cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				panic(err)
			}

			if x509Cert.IsCA {
				if retErr = pem.Encode(&caBlock, block); retErr != nil {
					return
				}
			} else {
				if retErr = pem.Encode(&tmp, block); retErr != nil {
					return
				}
				cert = tmp.String()
			}
		} else {
			if retErr = pem.Encode(&tmp, block); retErr != nil {
				return
			}

			privateKey = tmp.String()
		}
	}

	caCert = caBlock.String()

	if privateKey == "" {
		retErr = errors.New("no key found in PEM file")
	} else if caCert == "" {
		retErr = errors.New("no CA certs found in PEM file")
	} else if cert == "" {
		retErr = errors.New("no cert found in PEM file")
	}

	return
}

func (a *App) prepareMailServerConfig(ginContext *gin.Context, r *registry, secretData map[string]map[string]interface{},
	values map[string]interface{}) error {

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

		if _, ok := secretData[a.vaultRegistryPath(r.Name)]; !ok {
			secretData[a.vaultRegistryPath(r.Name)] = make(map[string]interface{})
		}

		secretData[a.vaultRegistryPath(r.Name)][a.Config.VaultRegistrySMTPPwdSecretKey] = pwd
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

func validateRegistryKeys(rq *http.Request, r KeyManagement) (createKeys bool, key6Fl, caCertFl,
	caJSONFl multipart.File, err error) {

	var fieldErrors []validator.FieldError
	caCertFl, _, err = rq.FormFile("ca-cert")
	if err != nil {
		if !r.KeysRequired() {
			err = nil
			return
		}

		fieldErrors = append(fieldErrors, router.MakeFieldError("CACertificate", "required"))
	}

	caJSONFl, _, err = rq.FormFile("ca-json")
	if err != nil {
		fieldErrors = append(fieldErrors, router.MakeFieldError("CAsJSON", "required"))
	}

	if r.KeyDeviceType() == KeyDeviceTypeFile {
		key6Fl, _, err = rq.FormFile("key6")
		if err != nil {
			fieldErrors = append(fieldErrors, router.MakeFieldError("Key6", "required"))
		}
	}

	if len(fieldErrors) > 0 {
		err = validator.ValidationErrors(fieldErrors)
		return
	}

	createKeys = true
	return
}
