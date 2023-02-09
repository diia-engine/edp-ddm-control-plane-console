package registry

import (
	"ddm-admin-console/router"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) deleteRegistry(ctx *gin.Context) (response router.Response, retErr error) {
	userCtx := router.ContextWithUserAccessToken(ctx)
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.PostForm("registry-name")
	err = cbService.Delete(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	a.Perms.DeleteRegistry(registryName)

	return router.MakeRedirectResponse(http.StatusFound, "/admin/registry/overview"), nil
}
