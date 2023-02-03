package cluster

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
)

const (
	StorageLocation          = "backup-s3-like-storage-location"
	StorageType              = "backup-s3-like-storage-type"
	StorageCredentialsKey    = "backup-s3-like-storage-credentials-key"
	StorageCredentialsSecret = "backup-s3-like-storage-credentials-secret"
)

type BackupConfig struct {
	StorageLocation          string `form:"storage-location" binding:"required"`
	StorageType              string `form:"storage-type" binding:"required"`
	StorageCredentialsKey    string `form:"storage-credentials-key" binding:"required"`
	StorageCredentialsSecret string `form:"storage-credentials-secret" binding:"required"`
}

func (a *App) editGet(ctx *gin.Context) (router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

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

	var backupConfig BackupConfig
	secret, err := k8sService.GetSecret(a.Config.BackupSecretName)
	if err != nil && !k8sErrors.IsNotFound(err) {
		return nil, errors.Wrap(err, "unable to get backup config secret")
	}
	if err == nil {
		backupConfig = BackupConfig{
			StorageLocation:       string(secret.Data[StorageLocation]),
			StorageType:           string(secret.Data[StorageType]),
			StorageCredentialsKey: string(secret.Data[StorageCredentialsKey]),
		}
	}

	cb, err := cbService.Get(a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	hasUpdate, branches, err := registry.HasUpdate(userCtx, a.Services.Gerrit, cb, MRTypeClusterUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	hwINITemplateContent, err := registry.GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	values, err := a.getValuesDict(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load values")
	}

	rspParams := gin.H{
		"backupConf":           backupConfig,
		"page":                 "cluster",
		"updateBranches":       branches,
		"hasUpdate":            hasUpdate,
		"hwINITemplateContent": hwINITemplateContent,
	}

	for _, f := range a.editDataLoaders() {
		if err := f(values, rspParams); err != nil {
			return nil, errors.Wrap(err, "unable to load edit data")
		}
	}

	return router.MakeHTMLResponse(200, "cluster/edit.html", rspParams), nil
}

func (a *App) editDataLoaders() []func(*Values, gin.H) error {
	return []func(*Values, gin.H) error{
		a.loadAdminsConfig,
		a.loadCIDRConfig,
		a.loadBackupScheduleConfig,
	}
}

func (a *App) editPost(ctx *gin.Context) (router.Response, error) {
	//userCtx := a.router.ContextWithUserAccessToken(ctx)
	//
	//k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to init service for user context")
	//}
	//
	//canUpdateCluster, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", a.Config.CodebaseName)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	//}
	//if !canUpdateCluster {
	//	return nil, errors.Wrap(err, "access denied")
	//}
	//
	//jenkinsService, err := a.Services.Jenkins.ServiceForContext(userCtx)
	//if err != nil {
	//	return nil, errors.Wrap(err, "unable to init jenkins client")
	//}
	//
	//var backupConfig BackupConfig
	//if err := ctx.ShouldBind(&backupConfig); err != nil {
	//	validationErrors, ok := err.(validator.ValidationErrors)
	//	if !ok {
	//		return nil, errors.Wrap(err, "unable to parse registry form")
	//	}
	//
	//	return router.MakeHTMLResponse(200, "cluster/edit.html",
	//		gin.H{"page": "cluster", "errorsMap": validationErrors, "backupConf": backupConfig}), nil
	//}
	//
	//if err := k8sService.RecreateSecret(a.Config.BackupSecretName, map[string][]byte{
	//	StorageLocation:          []byte(backupConfig.StorageLocation),
	//	StorageType:              []byte(backupConfig.StorageType),
	//	StorageCredentialsKey:    []byte(backupConfig.StorageCredentialsKey),
	//	StorageCredentialsSecret: []byte(backupConfig.StorageCredentialsSecret),
	//}); err != nil {
	//	return nil, errors.Wrap(err, "unable to recreate backup secret")
	//}
	//
	//if err := jenkinsService.CreateJobBuildRun(ctx, fmt.Sprintf("cluster-update-%d", time.Now().Unix()),
	//	fmt.Sprintf("%s/job/MASTER-Build-%s/", a.Config.CodebaseName, a.Config.CodebaseName), nil); err != nil {
	//	return nil, errors.Wrap(err, "unable to trigger jenkins job build run")
	//}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) getValuesDict(ctx context.Context) (*Values, error) {
	vals, err := a.Gerrit.GetFileContents(ctx, a.Config.CodebaseName, "master", registry.ValuesLocation)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get admin values yaml")
	}

	var (
		valuesDict map[string]interface{}
		values     *Values
		valuesBts  = []byte(vals)
	)
	if err := yaml.Unmarshal(valuesBts, &valuesDict); err != nil {
		return nil, errors.Wrap(err, "unable to decode values")
	}

	if err := yaml.Unmarshal(valuesBts, &values); err != nil {
		return nil, errors.Wrap(err, "unable to decode values")
	}

	values.OriginalYaml = valuesDict

	return values, nil
}

func (a *App) loadAdminsConfig(values *Values, rspParams gin.H) error {
	bts, err := json.Marshal(values.Admins)
	if err != nil {
		return errors.Wrap(err, "unable to json encode admins")
	}

	rspParams["admins"] = string(bts)
	return nil
}

func (a *App) loadCIDRConfig(values *Values, rspParams gin.H) error {
	cidrConfig, err := json.Marshal(map[string]interface{}{
		"admin": strings.Split(values.Global.WhiteListIP.AdminRoutes, " "),
	})
	if err != nil {
		return errors.Wrap(err, "unable to encode cidr to JSON")
	}

	rspParams["cidrConfig"] = string(cidrConfig)

	return nil
}
