package registry

import (
	"context"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (a *App) viewRegistry(ctx *gin.Context) (router.Response, error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Services.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	valuesJson, err := json.Marshal(values)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode values")
	}

	viewParams := gin.H{
		"page":       "registry",
		"timezone":   a.Config.Timezone,
		"values":     values,
		"valuesJson": string(valuesJson),
	}

	for _, f := range a.viewRegistryProcessFunctions() {
		if err := f(userCtx, registryName, values, viewParams); err != nil {
			return nil, errors.Wrap(err, "error during view registry function")
		}
	}

	return router.MakeHTMLResponse(200, "registry/view.html", viewParams), nil
}

func (a *App) viewRegistryProcessFunctions() []func(ctx context.Context, registryName string, values *Values, viewParams gin.H) error {
	return []func(ctx context.Context, registryName string, values *Values, viewParams gin.H) error{
		a.viewRegistryAllowedToEdit,
		a.viewRegistryGetRegistryAndBranches,
		a.viewRegistryGetEDPComponents,
		a.viewRegistryGetMergeRequests,
		a.viewRegistryExternalRegistration,
		a.viewDNSConfig,
		a.viewSMTPConfig,
		a.viewCIDRConfig,
		a.viewAdministratorsConfig,
		a.viewRegistryHasUpdates,
		a.viewUpdateTrembitaRegistries,
	}
}

func (a *App) viewUpdateTrembitaRegistries(userCtx context.Context, registryName string, values *Values, viewParams gin.H) error {
	mrs, ok := viewParams["mergeRequests"]
	if !ok {
		return nil
	}

	trembitaMrs := make(map[string]ExtendedMergeRequests)

	extendedMRs := mrs.([]ExtendedMergeRequests)
	for _, mr := range extendedMRs {
		name, ok := mr.Labels[MRLabelTrembitaRegsitryName]
		if ok {
			trembitaMrs[name] = mr
		}
	}

	for i, r := range values.Trembita.Registries {
		_, ok := trembitaMrs[i]
		if ok {
			r.UserID = "fake"
			values.Trembita.Registries[i] = r
		}
	}

	return nil
}

func (a *App) viewRegistryHasUpdates(userCtx context.Context, registryName string, _ *Values, viewParams gin.H) error {
	registry, ok := viewParams["registry"]
	if !ok {
		return nil
	}

	hasUpdate, _, err := HasUpdate(userCtx, a.Services.Gerrit, registry.(*codebase.Codebase), MRTargetRegistryVersionUpdate)
	if err != nil {
		return errors.Wrap(err, "unable to check for updates")
	}

	viewParams["hasUpdate"] = hasUpdate
	return nil
}

func (a *App) viewRegistryExternalRegistration(userCtx context.Context, registryName string, values *Values, viewParams gin.H) error {
	eRegs, mergeRequestsForER := make([]ExternalRegistration, 0), make(map[string]struct{})
	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit merge requests")
	}
	for _, mr := range mrs {
		if mr.Labels[MRLabelTarget] == "external-reg" && mr.Status.Value == gerrit.StatusNew {
			eRegs = append(eRegs, ExternalRegistration{Name: mr.Annotations[mrAnnotationRegName], Enabled: true,
				External: mr.Annotations[mrAnnotationRegType] == externalSystemTypeExternal, status: erStatusInactive})
			mergeRequestsForER[mr.Annotations[mrAnnotationRegName]] = struct{}{}
		} else if mr.Labels[MRLabelTarget] == "external-reg" && mr.Status.Value != gerrit.StatusMerged && mr.Status.Value != gerrit.StatusAbandoned {
			eRegs = append(eRegs, ExternalRegistration{Name: mr.Annotations[mrAnnotationRegName], Enabled: true,
				External: mr.Annotations[mrAnnotationRegType] == externalSystemTypeExternal, status: erStatusFailed})
			mergeRequestsForER[mr.Annotations[mrAnnotationRegName]] = struct{}{}
		}
	}

	//TODO: refactor to values struct
	_eRegs, err := decodeExternalRegsFromValues(values.OriginalYaml)
	if err != nil {
		return errors.Wrap(err, "unable to decode external regs")
	}

	for _, _er := range _eRegs {
		if _, ok := mergeRequestsForER[_er.Name]; !ok {
			eRegs = append(eRegs, _er)
		}
	}

	if err := a.loadKeysForExternalRegs(userCtx, registryName, eRegs); err != nil {
		return errors.Wrap(err, "unable load keys for ext regs")
	}

	viewParams["externalRegs"] = eRegs
	viewParams["values"] = values

	if err := a.loadCodebasesForExternalRegistrations(registryName, eRegs, viewParams); err != nil {
		return errors.Wrap(err, "unable to load codebases for external reg")
	}

	return nil
}

func (a *App) loadKeysForExternalRegs(ctx context.Context, registryName string, eRegs []ExternalRegistration) error {
	for i, er := range eRegs {
		if er.External && er.Enabled {
			s, err := a.Services.K8S.GetSecretFromNamespace(ctx, fmt.Sprintf("keycloak-client-%s-secret", er.Name),
				registryName)
			if k8sErrors.IsNotFound(err) {
				eRegs[i].status = erStatusInactive
				continue
			} else if err != nil {
				return errors.Wrap(err, "unable to get er system key")
			}

			eRegs[i].KeyValue = string(s.Data["clientSecret"])
		}
	}

	return nil
}

func (a *App) loadCodebasesForExternalRegistrations(registryName string, eRegs []ExternalRegistration, viewParams gin.H) error {
	cbs, err := a.Services.Codebase.GetAllByType("registry")
	if err != nil {
		return errors.Wrap(err, "unable to get all registries")
	}

	var availableRegs []codebase.Codebase
	for _, cb := range cbs {
		skip := false
		for _, er := range eRegs {
			if er.Name == cb.Name && !er.External {
				skip = true
				break
			}
		}

		if !skip && cb.Name != registryName && cb.Status.Available && cb.DeletionTimestamp.IsZero() && cb.StrStatus() != "failed" {
			availableRegs = append(availableRegs, cb)
		}
	}
	viewParams["externalRegAvailableRegistries"] = availableRegs

	return nil
}

func convertExternalRegFromInterface(in interface{}) ([]ExternalRegistration, error) {
	js, err := json.Marshal(in)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode interface to json")
	}

	var res []ExternalRegistration
	if err := json.Unmarshal(js, &res); err != nil {
		return nil, errors.Wrap(err, "unable to decode json")
	}

	return res, nil
}

func (a *App) viewAdministratorsConfig(userCtx context.Context, registryName string, values *Values, viewParams gin.H) error {
	viewParams["admins"] = values.Administrators
	return nil
}

func (a *App) viewCIDRConfig(userCtx context.Context, registryName string, values *Values, viewParams gin.H) error {
	if values.Global.WhiteListIP.AdminRoutes != "" {
		viewParams["adminCIDR"] = strings.Split(values.Global.WhiteListIP.AdminRoutes, " ")
	}

	if values.Global.WhiteListIP.CitizenPortal != "" {
		viewParams["citizenCIDR"] = strings.Split(values.Global.WhiteListIP.CitizenPortal, " ")
	}

	if values.Global.WhiteListIP.OfficerPortal != "" {
		viewParams["officerCIDR"] = strings.Split(values.Global.WhiteListIP.OfficerPortal, " ")
	}

	return nil
}

func (a *App) viewSMTPConfig(userCtx context.Context, registryName string, values *Values, viewParams gin.H) error {
	if values.Global.Notifications.Email.Type != "" {
		viewParams["smtpType"] = values.Global.Notifications.Email.Type
	}

	return nil
}

func (a *App) viewDNSConfig(userCtx context.Context, registryName string, values *Values, viewParams gin.H) error {
	//TODO: refactor to values struct
	valuesDict := values.OriginalYaml

	portals, ok := valuesDict["portals"]
	if !ok {
		return nil
	}

	portalsDict := portals.(map[string]interface{})

	if _, ok := portalsDict["citizen"]; ok {
		citizenDict := portalsDict["citizen"].(map[string]interface{})
		if citizenCustomDNS, ok := citizenDict["customDns"]; ok {
			viewParams["citizenPortalHost"] = citizenCustomDNS.(map[string]interface{})["host"].(string)
		}
	}

	if _, ok := portalsDict["officer"]; ok {
		officerDict := portalsDict["officer"].(map[string]interface{})
		if officerCustomDNS, ok := officerDict["customDns"]; ok {
			viewParams["officerPortalHost"] = officerCustomDNS.(map[string]interface{})["host"].(string)
		}
	}

	return nil
}

func (a *App) viewRegistryGetMergeRequests(userCtx context.Context, registryName string, _ *Values, viewParams gin.H) error {
	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to list gerrit merge requests")
	}

	sort.Sort(gerrit.SortByCreationDesc(mrs))

	emrs := make([]ExtendedMergeRequests, 0, len(mrs))
	for _, mr := range mrs {
		if mr.Status.Value == gerrit.StatusNew {
			viewParams["openMergeRequests"] = true
		}
		emrs = append(emrs, ExtendedMergeRequests{GerritMergeRequest: mr})
	}

	viewParams["mergeRequests"] = emrs
	return nil
}

func (a *App) viewRegistryAllowedToEdit(userCtx context.Context, registryName string, _ *Values, viewParams gin.H) error {
	k8sService, err := a.Services.K8S.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	allowed, err := a.Services.Codebase.CheckIsAllowedToUpdate(registryName, k8sService)
	if err != nil {
		return errors.Wrap(err, "unable to check codebase creation access")
	}

	viewParams["allowedToEdit"] = allowed
	return nil
}

func (a *App) viewRegistryGetRegistryAndBranches(userCtx context.Context, registryName string, _ *Values, viewParams gin.H) error {
	cbService, err := a.Services.Codebase.ServiceForContext(userCtx)
	if err != nil {
		return errors.Wrap(err, "unable to init service for user context")
	}

	registry, err := cbService.Get(registryName)
	if err != nil {
		return errors.Wrapf(err, "unable to get registry by name: %s", registryName)
	}

	branches, err := cbService.GetBranchesByCodebase(registry.Name)
	if err != nil {
		return errors.Wrap(err, "unable to get registry branches")
	}

	if err := a.loadBranchesStatuses(userCtx, branches); err != nil {
		return errors.Wrap(err, "unable to load branch statuses")
	}

	registry.Branches = branches

	viewParams["registry"] = registry
	viewParams["branches"] = branches

	return nil
}

func (a *App) loadBranchesStatuses(ctx context.Context, branches []codebase.CodebaseBranch) error {
	for i, b := range branches {
		branchName := strings.ToUpper(b.Spec.BranchName)
		status, build, err := a.Jenkins.GetJobStatus(ctx, fmt.Sprintf("%s/view/%s/job/%s-Build-%s", b.Spec.CodebaseName,
			branchName, branchName, b.Spec.CodebaseName))
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				continue
			}

			return errors.Wrap(err, "unable to get branch build status")
		}
		buildString := strconv.FormatInt(build, 10)

		branches[i].Status.Value = status
		branches[i].Status.Build = &buildString
	}

	return nil
}

func (a *App) viewRegistryGetEDPComponents(userCtx context.Context, registryName string, _ *Values, viewParams gin.H) error {
	jenkinsComponent, err := a.Services.EDPComponent.Get(userCtx, "jenkins")
	if err != nil {
		return errors.Wrap(err, "unable to get jenkins edp component")
	}

	gerritComponent, err := a.Services.EDPComponent.Get(userCtx, "gerrit")
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit edp component")
	}

	categories, err := a.Services.EDPComponent.GetAllCategory(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to list namespaced edp components")
	}

	viewParams["jenkinsURL"] = jenkinsComponent.Spec.Url
	viewParams["gerritURL"] = gerritComponent.Spec.Url

	viewParams["platformOperationalComponents"] = categories[edpcomponent.PlatformOperationalZone]
	viewParams["platformAdministrationComponents"] = categories[edpcomponent.PlatformAdministrationZone]

	_, ok := categories[edpcomponent.RegistryOperationalZone]
	if ok {
		viewParams["registryOperationalComponents"] = categories[edpcomponent.RegistryOperationalZone]
		viewParams["regisrtyAdministrationComponents"] = categories[edpcomponent.RegistryAdministrationZone]
	} else {
		//TODO: remove this hotfix
		if err := a.loadRegistryEDPCats(userCtx, registryName, viewParams); err != nil {
			return fmt.Errorf("unable to load registry edp component")
		}
	}

	return nil
}

func (a *App) loadRegistryEDPCats(ctx context.Context, registryName string, viewParams gin.H) error {
	components, err := a.Services.EDPComponent.GetAllNamespace(ctx, registryName, true)
	if err != nil {
		return fmt.Errorf("unable to get edp components")
	}

	var (
		registryOperationalComponents    []edpcomponent.EDPComponentItem
		registryAdministrationComponents []edpcomponent.EDPComponentItem
	)

	comMap := map[string]*[]edpcomponent.EDPComponentItem{
		"gerrit":                                 &registryAdministrationComponents,
		"jenkins":                                &registryAdministrationComponents,
		"business-process-administration-portal": &registryAdministrationComponents,
		"citizen-portal":                         &registryOperationalComponents,
		"officer-portal":                         &registryOperationalComponents,
	}

	for _, cmp := range components {
		if list, ok := comMap[cmp.Name]; ok {
			item := edpcomponent.PrepareComponentItem(cmp)
			item.Title = cmp.Name
			*list = append(*list, item)
		}
	}

	viewParams["registryOperationalComponents"] = registryOperationalComponents
	viewParams["regisrtyAdministrationComponents"] = registryAdministrationComponents

	return nil
}
