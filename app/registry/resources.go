package registry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"ddm-admin-console/router"
)

func (a *App) getValuesFromBranch(project, branch string) (map[string]any, error) {
	content, err := a.Gerrit.GetFileFromBranch(project, branch, url.PathEscape(ValuesLocation))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get project content")
	}

	var data map[string]any
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, errors.Wrap(err, "unable to decode yaml")
	}

	return data, nil
}

func (a *App) prepareRegistryResources(
	_ *gin.Context,
	newRegistryData *registry,
	existingData *Values,
	_ map[string]map[string]any,
	_ *[]string,
) (
	bool,
	error,
) {
	existingGlobalValues, ok := existingData.OriginalYaml[GlobalValuesIndex]
	if !ok {
		existingGlobalValues = make(map[string]any)
	}

	existingGlobalDict, ok := existingGlobalValues.(map[string]any)
	if !ok {
		return false, fmt.Errorf("failed to assume registry global type")
	}

	valuesChanged := false

	if newRegistryData.Resources != "" {
		var newResourcesConfig map[string]any

		if err := json.Unmarshal([]byte(newRegistryData.Resources), &newResourcesConfig); err != nil {
			return false, fmt.Errorf("failed to decode newResourcesConfig: %w", err)
		}

		existingResourcesConfig, ok := existingGlobalDict[ResourcesIndex].(map[string]any)
		if !ok {
			return false, fmt.Errorf("failed to assume resource config type")
		}

		for resourceName := range existingResourcesConfig {
			if _, exists := newResourcesConfig[resourceName]; !exists {
				newResourcesConfig[resourceName] = make(map[struct{}]struct{}) // should be {} in yaml when empty
			}
		}

		if !reflect.DeepEqual(newResourcesConfig, existingResourcesConfig) {
			valuesChanged = true
			existingGlobalDict[ResourcesIndex] = newResourcesConfig
			existingData.OriginalYaml[GlobalValuesIndex] = existingGlobalDict
		}
	}

	if newRegistryData.CrunchyPostgresMaxConnections != "" {
		maxCon, err := strconv.ParseInt(newRegistryData.CrunchyPostgresMaxConnections, 10, 32)
		if err != nil {
			return false, fmt.Errorf("unable to parse max connectrions, %w", err)
		}

		if existingData.Global.CrunchyPostgres.CrunchyPostgresPostgresql.CrunchyPostgresPostgresqlParameters.MaxConnections != int(maxCon) {
			existingData.Global.CrunchyPostgres.CrunchyPostgresPostgresql.CrunchyPostgresPostgresqlParameters.MaxConnections = int(maxCon)

			existingGlobalDict[CrunchyPostgresIndex] = existingData.Global.CrunchyPostgres
			existingData.OriginalYaml[GlobalValuesIndex] = existingGlobalDict
			valuesChanged = true
		}
	}

	if newRegistryData.CrunchyPostgresStorageSize != "" && existingData.Global.CrunchyPostgres.StorageSize != newRegistryData.CrunchyPostgresStorageSize {
		existingData.Global.CrunchyPostgres.StorageSize = newRegistryData.CrunchyPostgresStorageSize
		existingGlobalDict[CrunchyPostgresIndex] = existingData.Global.CrunchyPostgres
		existingGlobalDict[CrunchyPostgresIndex] = existingData.Global.CrunchyPostgres
		valuesChanged = true
	}

	return valuesChanged, nil
}

func (a *App) preloadTemplateValues(ctx *gin.Context) (rsp router.Response, retErr error) {
	template, branch := ctx.Query("template"), ctx.Query("branch")
	if template == "" || branch == "" {
		return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
	}

	data, err := a.getValuesFromBranch(template, branch)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get template content")
	}

	return router.MakeJSONResponse(http.StatusOK, data), nil
}
