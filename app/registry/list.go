package registry

import (
	"context"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) listRegistry(ctx *gin.Context) (response router.Response, retErr error) {
	userCtx := router.ContextWithUserAccessToken(ctx)
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

	if err := LoadRegistryVersions(userCtx, a.Gerrit, cbs); err != nil {
		return nil, errors.Wrap(err, "unable to load registry versions")
	}

	registries, err := a.Services.Perms.FilterCodebases(ctx, cbs, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase permissions")
	}

	prjs, err := a.Services.Gerrit.GetProjects(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}
	prjs = a.filterProjects(prjs, "cluster-mgmt")
	gerritBranches := formatGerritProjectBranches(prjs)

	responseParams := gin.H{
		"registries":      registries,
		"page":            "registry",
		"allowedToCreate": allowedToCreate,
		"timezone":        a.Config.Timezone,
		"gerritBranches":  gerritBranches,
		"platformVersion": a.Config.CurrentVersion,
		"previousVersion": a.Config.PreviousVersion,
	}

	templateArgs, templateErr := json.Marshal(responseParams)
	if templateErr != nil {
		return nil, errors.Wrap(templateErr, "unable to encode template arguments")
	}

	responseParams["templateArgs"] = string(templateArgs)

	return router.MakeHTMLResponse(200, "registry/list.html", responseParams), nil
}

func (a *App) getRegistries(ctx *gin.Context) (rsp router.Response, retErr error) {
	userCtx := router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init service for user context")
	}

	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase creation access")
	}

	cbs, err := a.Services.Codebase.GetAllByType(codebase.RegistryCodebaseType)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get codebases")
	}

	if err := LoadRegistryVersions(userCtx, a.Gerrit, cbs); err != nil {
		return nil, errors.Wrap(err, "unable to load registry versions")
	}

	registries, err := a.Services.Perms.FilterCodebases(ctx, cbs, k8sService)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check codebase permissions")
	}

	return router.MakeJSONResponse(http.StatusOK, registries), nil
}

func LoadRegistryVersions(ctx context.Context, gerritService gerrit.ServiceInterface, cbs []codebase.Codebase) error {
	mrs, err := gerritService.GetMergeRequests(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get merge requests")
	}

	registryMrs := make(map[string][]gerrit.GerritMergeRequest)

	for _, v := range mrs {
		registryMrs[v.Spec.ProjectName] = append(registryMrs[v.Spec.ProjectName], v)
	}

	for i, cb := range cbs {
		registryVersion := BranchVersion(cb.Spec.DefaultBranch)
		if cb.Spec.BranchToCopyInDefaultBranch != "" {
			registryVersion = BranchVersion(cb.Spec.BranchToCopyInDefaultBranch)
		}

		currentRegistryMrs := registryMrs[cb.Name]
		for _, mr := range currentRegistryMrs {
			if (mr.Labels[MRLabelTarget] == MRTargetRegistryVersionUpdate ||
				mr.Labels[MRLabelTarget] == MRTargetClusterUpdate) && mr.Status.Value == gerrit.StatusMerged {
				mergedBranchVersion := BranchVersion(mr.Spec.SourceBranch)
				if registryVersion.LessThan(mergedBranchVersion) {
					registryVersion = mergedBranchVersion
				}
			}
		}

		cbs[i].Version = registryVersion
		cbs[i].Spec.DefaultBranch = registryVersion.String()
	}

	return nil
}
