package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
)

type BackupConfig struct {
	StorageLocation          string `form:"storage-location" binding:"required"`
	StorageType              string `form:"storage-type" binding:"required"`
	StorageCredentialsKey    string `form:"storage-credentials-key" binding:"required"`
	StorageCredentialsSecret string `form:"storage-credentials-secret" binding:"required"`
}

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

	valuesJs, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("unable to encode values to json, %w", err)
	}

	rspParams := gin.H{
		"updateBranches":       branches,
		"hasUpdate":            hasUpdate,
		"hwINITemplateContent": hwINITemplateContent,
		"values":               string(valuesJs),
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
		"page":         "cluster",
		"templateArgs": string(templateArgs),
	}), nil
}

func (a *App) editDataLoaders() []func(context.Context, *Values, gin.H) error {
	return []func(context.Context, *Values, gin.H) error{
		a.loadAdminsConfig,
		a.loadCIDRConfig,
		a.loadBackupScheduleConfig,
		a.loadKeycloakDefaultHostname,
	}
}

func (a *App) editPost(ctx *gin.Context) (router.Response, error) {
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
