package cluster

import (
	"ddm-admin-console/app/registry"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
)

func (a *App) view(ctx *gin.Context) (router.Response, error) {
	userCtx := router.ContextWithUserAccessToken(ctx)
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s service for user")
	}

	canUpdateCluster, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check access to cluster codebase")
	}

	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init codebase service for user")
	}

	cb, err := cbService.Get(a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster codebase")
	}

	branches, err := cbService.GetBranchesByCodebase(ctx, cb.Name)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry branches")
	}
	cb.Branches = branches

	jenkinsComponent, err := a.Services.EDPComponent.Get(userCtx, "jenkins")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.Services.EDPComponent.Get(userCtx, "gerrit")
	if err != nil {
		return nil, errors.Wrap(err, "unable to get gerrit edp component")
	}

	namespacedEDPComponents, err := a.Services.EDPComponent.GetAllNamespace(userCtx, cb.Name, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list namespaced edp components")
	}

	clusterValues, err := registry.GetValuesFromGit(a.ClusterRepo, registry.MasterBranch, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to get cluster values, %w", err)
	}

	categories, err := a.Services.EDPComponent.GetAllCategoryPlatform(userCtx)
	if err != nil {
		return nil, fmt.Errorf("unable to get edp components categories")
	}

	adminsStr, err := a.displayAdmins(clusterValues)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get admins")
	}

	cidr, err := a.displayCIDR(clusterValues)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cidr")
	}

	emrs, err := a.ClusterGetMergeRequests(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load gerrit merge requests")
	}

	clusterProject, err := a.Gerrit.GetProject(userCtx, a.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster gerrit project")
	}

	hasUpdate, _, _, err := registry.HasUpdate(userCtx, a.Services.Gerrit, cb, registry.MRTargetClusterUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check for updates")
	}

	rspParams := gin.H{
		"hasUpdate":        hasUpdate,
		"branches":         branches,
		"codebase":         cb,
		"jenkinsURL":       jenkinsComponent.Spec.Url,
		"gerritURL":        gerritComponent.Spec.Url,
		"edpComponents":    namespacedEDPComponents,
		"canUpdateCluster": canUpdateCluster,
		"admins":           adminsStr,
		"cidr":             cidr,
		"version":          a.getClusterVersion(clusterProject.Status.Branches, emrs),
		"mergeRequests":    emrs,
	}

	rspParams["platformOperationalComponents"] = categories[edpcomponent.PlatformOperationalZone]
	rspParams["platformAdministrationComponents"] = categories[edpcomponent.PlatformAdministrationZone]
	rspParams["platformCentralComponents"] = categories[edpcomponent.PlatformCentralZone]

	templateArgs, err := json.Marshal(rspParams)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode template arguments")
	}

	return router.MakeHTMLResponse(200, "cluster/view.html", gin.H{
		"page":         "cluster",
		"templateArgs": string(templateArgs),
	}), nil
}

func (a *App) getClusterVersion(gerritProjectBranches []string, mrs []ExtendedMergeRequests) string {
	registryVersion := registry.LowestVersion(registry.UpdateBranches(gerritProjectBranches))

	for _, mr := range mrs {
		if mr.Labels[registry.MRLabelTarget] != MRTargetClusterUpdate {
			continue
		}

		if mr.Status.Value == gerrit.StatusMerged {
			mergedBranchVersion := registry.BranchVersion(mr.Spec.SourceBranch)
			if registryVersion.LessThan(mergedBranchVersion) {
				registryVersion = mergedBranchVersion
			}
		}
	}

	return registryVersion.String()
}

func (a *App) displayAdmins(values *registry.Values) (string, error) {
	js, err := a.getAdminsJSON(values)
	if err != nil {
		return "", errors.Wrap(err, "unable to get admins json")
	}

	var admins []Admin
	if err := json.Unmarshal([]byte(js), &admins); err != nil {
		return "", errors.Wrap(err, "unable to decode admins")
	}

	var adminsStr []string
	for _, adm := range admins {
		adminsStr = append(adminsStr, adm.Email)
	}

	return strings.Join(adminsStr, ", "), nil
}

func (a *App) displayCIDR(values *registry.Values) ([]string, error) {
	//TODO: refactor
	global, ok := values.OriginalYaml["global"]
	if !ok {
		return []string{}, nil
	}

	globalDict, ok := global.(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	whiteListIP, ok := globalDict["whiteListIP"]
	if !ok {
		return []string{}, nil
	}

	whiteListIPDict, ok := whiteListIP.(map[string]interface{})
	if !ok {
		return []string{}, nil
	}

	cidr, ok := whiteListIPDict["adminRoutes"]
	if !ok {
		return []string{}, nil
	}

	cidrStr := cidr.(string)
	if cidrStr == "" {
		return []string{}, nil
	}

	return strings.Split(cidrStr, " "), nil
}
