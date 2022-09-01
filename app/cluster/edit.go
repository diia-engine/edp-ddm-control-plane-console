package cluster

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
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

func (a *App) editGet(ctx *gin.Context) (*router.Response, error) {
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

	hasUpdate, branches, err := registry.HasUpdate(userCtx, a.Services.Gerrit, cb)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	admins, err := a.getAdminsJSON(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode admins")
	}

	hwINITemplateContent, err := registry.GetINITemplateContent(a.Config.HardwareINITemplatePath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get ini template data")
	}

	return router.MakeResponse(200, "cluster/edit.html", gin.H{
		"backupConf":           backupConfig,
		"page":                 "cluster",
		"updateBranches":       branches,
		"hasUpdate":            hasUpdate,
		"admins":               admins,
		"hwINITemplateContent": hwINITemplateContent,
	}), nil
}

func (a *App) editPost(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	canUpdateCluster, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	}
	if !canUpdateCluster {
		return nil, errors.Wrap(err, "access denied")
	}

	jenkinsService, err := a.Services.Jenkins.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init jenkins client")
	}

	var backupConfig BackupConfig
	if err := ctx.ShouldBind(&backupConfig); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "cluster/edit.html",
			gin.H{"page": "cluster", "errorsMap": validationErrors, "backupConf": backupConfig}), nil
	}

	if err := k8sService.RecreateSecret(a.Config.BackupSecretName, map[string][]byte{
		StorageLocation:          []byte(backupConfig.StorageLocation),
		StorageType:              []byte(backupConfig.StorageType),
		StorageCredentialsKey:    []byte(backupConfig.StorageCredentialsKey),
		StorageCredentialsSecret: []byte(backupConfig.StorageCredentialsSecret),
	}); err != nil {
		return nil, errors.Wrap(err, "unable to recreate backup secret")
	}

	if err := jenkinsService.CreateJobBuildRun(fmt.Sprintf("cluster-update-%d", time.Now().Unix()),
		fmt.Sprintf("%s/job/MASTER-Build-%s/", a.Config.CodebaseName, a.Config.CodebaseName), nil); err != nil {
		return nil, errors.Wrap(err, "unable to trigger jenkins job build run")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}
