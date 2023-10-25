package cluster

import (
	"context"
	"ddm-admin-console/router"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) updateDemoRegistryName(ctx *gin.Context) (router.Response, error) {
	demoRegistryNameValue := ctx.PostForm("demo-registry-name")

	values, err := getValuesFromGit(a.Config.CodebaseName, masterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values contents")
	}

	values.Global.DemoRegistryName = demoRegistryNameValue

	if err := a.createValuesMergeRequestCtx(ctx, MRTypeDemoRegistryName, "update cluster demoRegistryName config", values); err != nil {
		return nil, errors.Wrap(err, "unable to create demoRegistryName merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) loadDocumentation(_ context.Context, values *Values, rspParams gin.H) error {
	rspParams["demoRegistryName"] = values.Global.DemoRegistryName

	return nil
}
