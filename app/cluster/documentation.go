package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) updateDemoRegistryName(ctx *gin.Context) (router.Response, error) {
	demoRegistryNameValue := ctx.PostForm("demo-registry-name")

	values, err := registry.GetValuesFromGit(a.Config.CodebaseName, registry.MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode values yaml")
	}

	values.OriginalYaml["demoRegistryName"] = demoRegistryNameValue

	if err := a.createValuesMergeRequestCtx(ctx, MRTypeDemoRegistryName, "update cluster demoRegistryName config", values.OriginalYaml); err != nil {
		return nil, errors.Wrap(err, "unable to create demoRegistryName merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) loadDocumentation(_ context.Context, values *Values, rspParams gin.H) error {
	rspParams["demoRegistryName"] = values.OriginalYaml["demoRegistryName"]

	return nil
}
