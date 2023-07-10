package registry

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"fmt"
	"net/http"

	goGerrit "github.com/andygrunwald/go-gerrit"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	publicAPISystemType       = "publicAPI-system"
	publicAPIValuesIndex      = "publicApi"
	publicAPIStatusInactive   = "inactive"
	publicAPIStatusFailed     = "failed"
	publicAPIStatusActive     = "active"
	publicAPIStatusDisabled   = "disabled"
	mrTargetPublicAPIReg      = "publicAPI-reg"
	MRLabelPublicApiTarget    = "console/target"
	MRLabelPublicApiSubTarget = "console/sub-target"
	MRLabelPublicApiName      = "publicAPI-reg-name"
)

func (a *App) editPublicAPIReg(ctx *gin.Context) (router.Response, error) {
	registryName := ctx.Param("name")
	systemName := ctx.PostForm("reg-name")
	regURL := ctx.PostForm("reg-url")
	regLimitsValue := ctx.PostForm("reg-limits")
	var regLimits Limits
	if systemName == "" {
		return nil, errors.New("reg-name is required")
	}

	if err := json.Unmarshal([]byte(regLimitsValue), &regLimits); err != nil {
		return nil, errors.Wrap(err, "unable to decode limits from request")
	}

	vals, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to get values from git, %w", err)
	}

	found := false
	for i, v := range vals.PublicApi {
		if v.Name == systemName {
			vals.PublicApi[i].URL = regURL
			vals.PublicApi[i].Limits = regLimits
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("reg-name not found")
	}

	vals.OriginalYaml[publicAPIValuesIndex] = vals.PublicApi

	if err := CreateEditMergeRequest(
		ctx,
		registryName,
		vals.OriginalYaml,
		a.Gerrit, []string{},
		MRLabel{Key: MRLabelPublicApiTarget, Value: mrTargetPublicAPIReg},
		MRLabel{Key: MRLabelPublicApiName, Value: ctx.PostForm("reg-name")},
		MRLabel{Key: MRLabelPublicApiSubTarget, Value: "edition"},
	); err != nil {
		return nil, fmt.Errorf("unable to create MR, %w", err)
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) addPublicAPIReg(ctx *gin.Context) (router.Response, error) {
	registryName := ctx.Param("name")
	var regLimits Limits
	regLimitsValue := ctx.PostForm("reg-limits")

	if err := json.Unmarshal([]byte(regLimitsValue), &regLimits); err != nil {
		return nil, errors.Wrap(err, "unable to decode limits from request")
	}

	publicAPI := PublicAPI{
		Name:    ctx.PostForm("reg-name"),
		URL:     ctx.PostForm("reg-url"),
		Limits:  regLimits,
		Enabled: true,
	}

	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to get values from git, %w", err)
	}

	for _, _er := range values.PublicApi {
		if publicAPI.Name == _er.Name && _er.URL == publicAPI.URL {
			return nil, errors.New("publicAPI reg system already exists")
		}
	}

	values.OriginalYaml[publicAPIValuesIndex] = append(values.PublicApi, PublicAPI{
		Name:    publicAPI.Name,
		Enabled: publicAPI.Enabled,
		URL:     publicAPI.URL,
		Limits:  publicAPI.Limits,
	})

	if err != nil {
		return nil, fmt.Errorf("unable to prepare registry values, %w", err)
	}

	if err := CreateEditMergeRequest(
		ctx,
		registryName,
		values.OriginalYaml,
		a.Gerrit, []string{},
		MRLabel{Key: MRLabelPublicApiTarget, Value: mrTargetPublicAPIReg},
		MRLabel{Key: MRLabelPublicApiName, Value: ctx.PostForm("reg-name")},
		MRLabel{Key: MRLabelPublicApiSubTarget, Value: "creation"},
	); err != nil {
		return nil, fmt.Errorf("unable to create MR, %w", err)
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) disablePublicAPIReg(ctx *gin.Context) (router.Response, error) {
	registryName := ctx.Param("name")
	systemName := ctx.PostForm("reg-name")
	if systemName == "" {
		return nil, errors.New("reg-name is required")
	}

	vals, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to get values from git, %w", err)
	}

	found := false
	var action string

	for i, v := range vals.PublicApi {
		if v.Name == systemName {
			vals.PublicApi[i].Enabled = !vals.PublicApi[i].Enabled
			found = true

			if vals.PublicApi[i].Enabled {
				action = "enable"
			} else {
				action = "disable"
			}
			break
		}
	}
	if !found {
		return nil, errors.New("reg-name not found")
	}

	vals.OriginalYaml[publicAPIValuesIndex] = vals.PublicApi

	if err := CreateEditMergeRequest(
		ctx,
		registryName,
		vals.OriginalYaml,
		a.Gerrit, []string{},
		MRLabel{Key: MRLabelPublicApiTarget, Value: mrTargetPublicAPIReg},
		MRLabel{Key: MRLabelPublicApiName, Value: ctx.PostForm("reg-name")},
		MRLabel{Key: MRLabelPublicApiSubTarget, Value: action},
	); err != nil {
		return nil, fmt.Errorf("unable to create MR, %w", err)
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) removePublicAPIReg(ctx *gin.Context) (router.Response, error) {
	registryName := ctx.Param("name")
	systemName := ctx.PostForm("reg-name")
	if systemName == "" {
		return nil, errors.New("reg-name is required")
	}

	vals, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, fmt.Errorf("unable to get values from git, %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("unable to decode publicAPI regs, %w", err)
	}

	found := false
	for i, v := range vals.PublicApi {
		if v.Name == systemName {
			vals.PublicApi = append(vals.PublicApi[:i], vals.PublicApi[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("reg-name not found")
	}

	vals.OriginalYaml[publicAPIValuesIndex] = vals.PublicApi
	if err != nil {
		return nil, fmt.Errorf("unable to encode new values yaml, %w", err)
	}

	if err := CreateEditMergeRequest(
		ctx,
		registryName,
		vals.OriginalYaml,
		a.Gerrit, []string{},
		MRLabel{Key: MRLabelPublicApiTarget, Value: mrTargetPublicAPIReg},
		MRLabel{Key: MRLabelPublicApiName, Value: ctx.PostForm("reg-name")},
		MRLabel{Key: MRLabelPublicApiSubTarget, Value: "deletion"},
	); err != nil {
		return nil, fmt.Errorf("unable to create MR, %w", err)
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) makeViewPublicAPIMR(mr gerrit.GerritMergeRequest, registryName string) (PublicAPI, error) {
	var publicAPI PublicAPI

	if mr.Status.ChangeID == "" {
		return publicAPI, nil
	}

	changeInfo, _, err := a.Gerrit.GoGerritClient().Changes.GetChangeDetail(mr.Status.ChangeID, &goGerrit.ChangeOptions{})

	if err != nil {
		return publicAPI, fmt.Errorf("unable to get gerrit change details, %w", err)
	}
	changesContent, err := a.getChangeContentData(changeInfo.ID, ValuesLocation, changeInfo.Project)

	if err != nil {
		return publicAPI, fmt.Errorf("unable to get gerrit change, %w", err)
	}

	var (
		url  string
		data Values
	)

	if err := yaml.Unmarshal([]byte(changesContent), &data); err != nil {
		return publicAPI, fmt.Errorf("unable to decode yaml, %w", err)
	}

	name := mr.Labels[MRLabelPublicApiName]

	vals, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return publicAPI, fmt.Errorf("unable to get values from git, %w", err)
	}

	for _, v := range vals.PublicApi {
		if v.Name == name {
			url = v.URL
		}
	}

	if len(data.PublicApi) > 0 {
		for _, pub := range data.PublicApi {
			if pub.Name == name {
				if pub.URL != "" {
					url = pub.URL
				}
			}
		}
	}

	if mr.Status.Value == gerrit.StatusNew {
		publicAPI = PublicAPI{Name: name, URL: url, Enabled: true, StatusRegistration: publicAPIStatusInactive}
	} else if mr.Status.Value != gerrit.StatusMerged && mr.Status.Value != gerrit.StatusAbandoned {
		publicAPI = PublicAPI{Name: name, URL: url, Enabled: true, StatusRegistration: publicAPIStatusFailed}
	}

	return publicAPI, nil
}
