package registry

import (
	"ddm-admin-console/router"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
)

func (a *App) GetValuesFromBranch(project, branch string) (map[string]interface{}, error) {
	content, err := a.Gerrit.GetBranchContent(project, branch, url.PathEscape(ValuesLocation))
	if err != nil {
		return nil, errors.Wrap(err, "unable to get project content")
	}

	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &data); err != nil {
		return nil, errors.Wrap(err, "unable to decode yaml")
	}

	return data, nil
}

func (a *App) prepareRegistryResources(_ *gin.Context, r *registry, values *Values,
	_ map[string]map[string]interface{}, mrActions *[]string) error {

	if r.Resources != "" {
		var resources map[string]interface{}
		if err := json.Unmarshal([]byte(r.Resources), &resources); err != nil {
			return errors.Wrap(err, "unable to decode resources")
		}
		values.Global.Registry = resources
		values.OriginalYaml[GlobalValuesIndex] = values.Global
	}

	if r.CrunchyPostgresMaxConnections != "" {
		maxCon, err := strconv.ParseInt(r.CrunchyPostgresMaxConnections, 10, 32)
		if err != nil {
			return fmt.Errorf("unable to parse max connectrions, %w", err)
		}

		values.Global.CrunchyPostgres.CrunchyPostgresPostgresql.CrunchyPostgresPostgresqlParameters.MaxConnections = int(maxCon)
		values.OriginalYaml[GlobalValuesIndex] = values.Global
	}

	if r.CrunchyPostgresStorageSize != "" {
		values.Global.CrunchyPostgres.StorageSize = r.CrunchyPostgresStorageSize
		values.OriginalYaml[GlobalValuesIndex] = values.Global
	}

	return nil
}

func (a *App) preloadTemplateValues(ctx *gin.Context) (rsp router.Response, retErr error) {
	template, branch := ctx.Query("template"), ctx.Query("branch")
	if template == "" || branch == "" {
		return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
	}

	data, err := a.GetValuesFromBranch(template, branch)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get template content")
	}

	return router.MakeJSONResponse(http.StatusOK, data), nil
}
