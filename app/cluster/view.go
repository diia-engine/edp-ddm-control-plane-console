package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"encoding/json"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"ddm-admin-console/router"
)

func (a *App) view(ctx *gin.Context) (*router.Response, error) {
	userCtx := a.router.ContextWithUserAccessToken(ctx)
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

	branches, err := cbService.GetBranchesByCodebase(cb.Name)
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

	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(ctx, a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit merge requests")
	}

	adminsStr, err := a.displayAdmins(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get admins")
	}

	cidr, err := a.displayCIDR(userCtx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cidr")
	}

	return router.MakeResponse(200, "cluster/view.html", gin.H{
		"branches":         branches,
		"codebase":         cb,
		"jenkinsURL":       jenkinsComponent.Spec.Url,
		"gerritURL":        gerritComponent.Spec.Url,
		"page":             "cluster",
		"edpComponents":    namespacedEDPComponents,
		"canUpdateCluster": canUpdateCluster,
		"mergeRequests":    mrs,
		"admins":           adminsStr,
		"cidr":             cidr,
	}), nil
}

func (a *App) displayAdmins(ctx context.Context) (string, error) {
	js, err := a.getAdminsJSON(ctx)
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

func (a *App) displayCIDR(ctx context.Context) ([]string, error) {
	values, err := a.Gerrit.GetFileContents(ctx, a.Config.CodebaseName, "master", registry.ValuesLocation)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get admin values yaml")
	}

	//TODO: refactor
	var valuesDict map[string]interface{}
	if err := yaml.Unmarshal([]byte(values), &valuesDict); err != nil {
		return nil, errors.Wrap(err, "unable to decode values")
	}

	global, ok := valuesDict["global"]
	if !ok {
		return nil, nil
	}

	globalDict, ok := global.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	whiteListIP, ok := globalDict["whiteListIP"]
	if !ok {
		return nil, nil
	}

	whiteListIPDict, ok := whiteListIP.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	cidr, ok := whiteListIPDict["adminRoutes"]
	if !ok {
		return nil, nil
	}

	return strings.Split(cidr.(string), " "), nil
}
