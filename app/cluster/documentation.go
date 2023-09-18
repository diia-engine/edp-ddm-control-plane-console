package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"net/http"

	"gopkg.in/yaml.v3"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) updateDemoRegistryName(ctx *gin.Context) (router.Response, error) {
	demoRegistryNameValue := ctx.PostForm("demo-registry-name")

	vals, err := a.Services.Gerrit.GetFileContents(ctx, a.Config.CodebaseName, "master", registry.ValuesLocation)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values contents")
	}
	var valuesDict map[string]interface{}
	if err := yaml.Unmarshal([]byte(vals), &valuesDict); err != nil {
		return nil, errors.Wrap(err, "unable to decode values yaml")
	}

	globalInterface, ok := valuesDict["global"]
	if !ok {
		globalInterface = make(map[string]interface{})
	}
	globalDict := globalInterface.(map[string]interface{})

	globalDict["demoRegistryName"] = demoRegistryNameValue
	valuesDict["global"] = globalDict

	if err := a.createValuesMergeRequestCtx(ctx, MRTypeDemoRegistryName, "update cluster demoRegistryName config", valuesDict); err != nil {
		return nil, errors.Wrap(err, "unable to create demoRegistryName merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) loadDocumentation(_ context.Context, values *Values, rspParams gin.H) error {
	rspParams["demoRegistryName"] = values.Global.DemoRegistryName

	return nil
}
