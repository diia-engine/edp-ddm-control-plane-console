package registry

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) listRegistry(ctx *gin.Context) (response *router.Response, retErr error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	allowedToCreate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "create", "")
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	cbs, err := a.Services.Codebase.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebases")
	}

	registries, err := a.Services.Codebase.CheckPermissions(cbs, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase permissions")
	}

	return router.MakeResponse(200, "registry/list.html", gin.H{
		"registries":      registries,
		"page":            "registry",
		"allowedToCreate": allowedToCreate,
		"timezone":        a.Config.Timezone,
	}), nil
}
