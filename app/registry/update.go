package registry

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
)

const (
	MRTargetRegistryVersionUpdate = "registry-version-update"
)

type updateRequest struct {
	Branch string `form:"branch" binding:"required"`
}

func (a *App) registryUpdateView(ctx *gin.Context) (*router.Response, error) {
	registryName := ctx.Param("name")

	userCtx := a.router.ContextWithUserAccessToken(ctx)

	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	reg, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	hasUpdate, branches, err := HasUpdate(userCtx, a.Services.Gerrit, reg, MRTargetRegistryVersionUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	return router.MakeResponse(200, "registry/update.html", gin.H{
		"updateBranches": branches,
		"hasUpdate":      hasUpdate,
		"registry":       reg,
		"page":           "registry",
	}), nil
}

func (a *App) registryUpdate(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	registryName := ctx.Param("name")

	cb, err := cbService.Get(registryName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry")
	}

	r := registry{
		Name:                registryName,
		RegistryGitBranch:   cb.Spec.DefaultBranch,
		RegistryGitTemplate: cb.Spec.Repository.Url,
		Scenario:            ScenarioKeyNotRequired,
	}

	var ur updateRequest
	if err := ctx.ShouldBind(&ur); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "registry/edit.html",
			gin.H{"page": "registry", "errorsMap": validationErrors, "registry": r, "model": r}), nil
	}

	if err := a.createMergeRequest(registryName, ur.Branch, userCtx, ctx); err != nil {
		return nil, errors.Wrap(err, "unable to create merge request")
	}

	if a.EnableBranchProvisioners {
		prov := branchProvisioner(ur.Branch)
		cb.Spec.JobProvisioning = &prov
		if err := a.Services.Codebase.Update(cb); err != nil {
			return nil, errors.Wrap(err, "unable to update codebase provisioner")
		}

		if err := a.Services.Jenkins.CreateJobBuildRun(fmt.Sprintf("ru-create-release-%d", time.Now().Unix()),
			fmt.Sprintf("%s/job/Create-release-%s/", r.Name, r.Name), map[string]string{
				"RELEASE_NAME": cb.Spec.DefaultBranch,
			}); err != nil {
			return nil, errors.Wrap(err, "unable to trigger jenkins job build run")
		}
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", r.Name)), nil
}

func (a *App) createMergeRequest(registryName, updateBranch string, userContext context.Context, ginContext *gin.Context) error {
	prj, err := a.Services.Gerrit.GetProject(userContext, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to get registry gerrit project")
	}

	if err := a.Services.Gerrit.CreateMergeRequest(userContext, &gerrit.MergeRequest{
		CommitMessage: fmt.Sprintf("Update registry to %s", updateBranch),
		SourceBranch:  updateBranch,
		ProjectName:   prj.Spec.Name,
		Name:          fmt.Sprintf("%s-update-%d", registryName, time.Now().Unix()),
		AuthorName:    ginContext.GetString(router.UserNameSessionKey),
		AuthorEmail:   ginContext.GetString(router.UserEmailSessionKey),
		Labels: map[string]string{
			MRLabelTarget: MRTargetRegistryVersionUpdate,
		},
		AdditionalArguments: []string{"-X", "ours"},
	}); err != nil {
		return errors.Wrap(err, "unable to create update merge request")
	}

	return nil
}
