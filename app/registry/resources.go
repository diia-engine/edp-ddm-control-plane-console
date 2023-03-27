package registry

import (
	"ddm-admin-console/router"
	"encoding/json"
	"net/http"
	"net/url"

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

func (a *App) prepareRegistryResources(_ *gin.Context, r *registry, _values *Values,
	_ map[string]map[string]interface{}, mrActions *[]string) error {
	values := _values.OriginalYaml
	//TODO: refactor to new values

	if r.Resources != "" {
		var resources map[string]interface{}
		if err := json.Unmarshal([]byte(r.Resources), &resources); err != nil {
			return errors.Wrap(err, "unable to decode resources")
		}

		global, ok := values["global"]
		if !ok {
			global = make(map[string]interface{})
		}

		globalDict := global.(map[string]interface{})

		globalDict["registry"] = resources
		values["global"] = globalDict
	}

	return nil
}

func (a *App) preloadTemplateResources(ctx *gin.Context) (rsp router.Response, retErr error) {
	template, branch := ctx.Query("template"), ctx.Query("branch")
	if template == "" || branch == "" {
		return router.MakeStatusResponse(http.StatusUnprocessableEntity), nil
	}

	data, err := a.GetValuesFromBranch(template, branch)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get template content")
	}

	global, ok := data["global"]
	if !ok {
		return nil, errors.New("no global values in deploy templates")
	}
	globalDict, ok := global.(map[string]interface{})
	if !ok {
		return nil, errors.New("wrong global dict format")
	}

	resources, ok := globalDict["registry"]
	if !ok {
		return router.MakeJSONResponse(200, []string{}), nil
	}

	return router.MakeJSONResponse(http.StatusOK, resources), nil
}
