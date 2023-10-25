package registry

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
	"os"
)

const (
	ValuesLocation             = "deploy-templates/values.yaml"
	MRLabelTarget              = "console/target"
	MRLabelSubTarget           = "console/sub-target"
	MRLabelSourceBranch        = "console/source-branch"
	MRLabelTargetBranch        = "console/target-branch"
	MRLabelAction              = "console/action"
	MRLabelActionBranchMerge   = "branch-merge"
	mrAnnotationRegName        = "ext-reg/name"
	MRAnnotationActions        = "actions"
	mrAnnotationRegType        = "ext-reg/type"
	externalSystemTypeExternal = "external-system"
	erValuesIndex              = "nontrembita-external-registration"
	erStatusInactive           = "inactive"
	erStatusFailed             = "failed"
	erStatusActive             = "active"
	erStatusDisabled           = "disabled"
	mrTargetExternalReg        = "external-reg"
	mrTargetEditRegistry       = "edit-registry"
	mrTargetEditTrembita       = "trembita-registry-update"
	mrSubTargetCreation        = "creation"
	mrSubTargetDisable         = "disable"
	mrSubTargetEnable          = "enable"
	mrSubTargetDeletion        = "deletion"
	MRLabelApprove             = "console/approve"
	MRLabelApproveAuto         = "auto"
)

type MRExists string

func (m MRExists) Error() string {
	return string(m)
}

type ExternalRegistration struct {
	Name               string `yaml:"name"`
	Enabled            bool   `yaml:"enabled"`
	External           bool   `yaml:"external"`
	StatusRegistration string `yaml:"-"`
	KeyValue           string `yaml:"-"`
}

func (e ExternalRegistration) TypeStr() string {
	if e.External {
		return "external-system"
	}

	return "internal-registry"
}

func (e ExternalRegistration) Inactive() bool {
	return e.Status() == "status-inactive" || e.Status() == "status-failed"
}

func (e ExternalRegistration) Status() string {
	s := e.StatusRegistration
	if s == "" {
		s = erStatusActive
	}

	if !e.Enabled {
		s = erStatusDisabled
	}

	return fmt.Sprintf("status-%s", s)
}

func (a *App) addExternalReg(ctx *gin.Context) (router.Response, error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")
	er := ExternalRegistration{
		Name:     ctx.PostForm("reg-name"),
		External: ctx.PostForm("external-system-type") == externalSystemTypeExternal,
		Enabled:  true,
	}
	values, err := a.prepareRegistryValues(userCtx, registryName, &er)
	if err != nil {
		return nil, errors.Wrap(err, "unable to prepare registry values")
	}

	if err := a.createErMergeRequest(userCtx, ctx, registryName, er.Name, values, mrSubTargetCreation); err != nil {
		if _, ok := err.(MRExists); !ok {
			return nil, errors.Wrap(err, "unable to create MR")
		}
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func GetValuesFromGit(projectName, branch string, gerritService gerrit.ServiceInterface) (*Values, error) {
	content, err := gerritService.GetFileFromBranch(projectName, branch, url.PathEscape(ValuesLocation))
	if err != nil {
		return nil, fmt.Errorf("failed to get the values file from the branch: %w", err)
	}

	valuesBytes := []byte(content)
	valuesDict := make(map[string]any)

	if err := yaml.Unmarshal(valuesBytes, &valuesDict); err != nil {
		return nil, fmt.Errorf("failed to decode values yaml to a map: %w", err)
	}

	vals := new(Values)

	if err := yaml.Unmarshal(valuesBytes, vals); err != nil {
		return nil, fmt.Errorf("failed to decode values yaml to a struct: %w", err)
	}

	if os.Getenv("REGION") != "ua" {
		if externalSystems, ok := valuesDict["external-systems"].(map[string]interface{}); ok {
			delete(externalSystems, "diia")
			valuesDict["external-systems"] = externalSystems
		}
		delete(valuesDict, "diia")
	}

	vals.OriginalYaml = valuesDict

	if os.Getenv("REGION") != "ua" {
		delete(vals.ExternalSystems, "diia")
	}

	return vals, nil
}

func decodeExternalRegsFromValues(valuesDict map[string]interface{}) ([]ExternalRegistration, error) {
	eRegs := make([]ExternalRegistration, 0)
	var err error
	externalReg, ok := valuesDict[erValuesIndex]
	if ok {
		eRegs, err = convertExternalRegFromInterface(externalReg)
		if err != nil {
			return nil, errors.Wrap(err, "unable to convert external regs")
		}
	}

	return eRegs, nil
}

func (a *App) prepareRegistryValues(ctx context.Context, registryName string, er *ExternalRegistration) (string, error) {
	values, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return "", errors.Wrap(err, "unable to get values from git")
	}

	eRegs, err := decodeExternalRegsFromValues(values.OriginalYaml)
	if err != nil {
		return "", errors.Wrap(err, "unable to decode external regs")
	}

	for _, _er := range eRegs {
		if er.Name == _er.Name && _er.External == er.External {
			return "", errors.New("external reg system already exists")
		}
	}

	eRegs = append(eRegs, ExternalRegistration{
		Name:     er.Name,
		Enabled:  er.Enabled,
		External: er.External,
	})
	values.OriginalYaml[erValuesIndex] = eRegs

	newValues, err := yaml.Marshal(values.OriginalYaml)
	if err != nil {
		return "", errors.Wrap(err, "unable to encode new values yaml")
	}

	return string(newValues), nil
}

func (a *App) createErMergeRequest(userCtx context.Context, ctx *gin.Context, registryName, erName, values, action string) error {
	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(userCtx, registryName)
	if err != nil {
		return errors.Wrap(err, "unable to get MRs")
	}

	for _, mr := range mrs {
		if mr.Status.Value == gerrit.StatusNew {
			return MRExists("there is already open merge request(s) for this registry")
		}
	}

	if err := a.Services.Gerrit.CreateMergeRequestWithContents(userCtx, &gerrit.MergeRequest{
		ProjectName:   registryName,
		Name:          fmt.Sprintf("ers-mr-%s-%s-%d", registryName, erName, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("update registry external reg systems"),
		TargetBranch:  "master",
		Labels: map[string]string{
			MRLabelTarget:    mrTargetExternalReg,
			MRLabelSubTarget: action,
		},
		Annotations: map[string]string{
			mrAnnotationRegName: erName,
			mrAnnotationRegType: ctx.PostForm("external-system-type"),
		},
	}, map[string]string{
		ValuesLocation: values,
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}

func (a *App) disableExternalReg(ctx *gin.Context) (router.Response, error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")
	systemName := ctx.PostForm("reg-name")
	if systemName == "" {
		return nil, errors.New("reg-name is required")
	}

	vals, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	eRegs, err := decodeExternalRegsFromValues(vals.OriginalYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode external regs")
	}

	found := false
	mrSubTarget := mrSubTargetDisable
	for i, v := range eRegs {
		if v.Name == systemName {
			eRegs[i].Enabled = !eRegs[i].Enabled
			found = true
			if eRegs[i].Enabled {
				mrSubTarget = mrSubTargetEnable
			}
		}
	}
	if !found {
		return nil, errors.New("reg-name not found")
	}

	vals.OriginalYaml[erValuesIndex] = eRegs
	newValues, err := yaml.Marshal(vals.OriginalYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode new values yaml")
	}

	if err := a.createErMergeRequest(userCtx, ctx, registryName, systemName, string(newValues), mrSubTarget); err != nil {
		if _, ok := err.(MRExists); !ok {
			return nil, errors.Wrap(err, "unable to create MR")
		}
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}

func (a *App) removeExternalReg(ctx *gin.Context) (router.Response, error) {
	userCtx := router.ContextWithUserAccessToken(ctx)

	registryName := ctx.Param("name")
	systemName := ctx.PostForm("reg-name")
	if systemName == "" {
		return nil, errors.New("reg-name is required")
	}

	vals, err := GetValuesFromGit(registryName, MasterBranch, a.Gerrit)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get values from git")
	}

	eRegs, err := decodeExternalRegsFromValues(vals.OriginalYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode external regs")
	}

	found := false
	for i, v := range eRegs {
		if v.Name == systemName {
			eRegs = append(eRegs[:i], eRegs[i+1:]...)
			found = true
		}
	}
	if !found {
		return nil, errors.New("reg-name not found")
	}

	vals.OriginalYaml[erValuesIndex] = eRegs
	newValues, err := yaml.Marshal(vals.OriginalYaml)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode new values yaml")
	}

	if err := a.createErMergeRequest(userCtx, ctx, registryName, systemName, string(newValues), mrSubTargetDeletion); err != nil {
		if _, ok := err.(MRExists); !ok {
			return nil, errors.Wrap(err, "unable to create MR")
		}
	}

	return router.MakeRedirectResponse(http.StatusFound,
		fmt.Sprintf("/admin/registry/view/%s", registryName)), nil
}
