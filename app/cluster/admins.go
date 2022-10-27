package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	MRTypeClusterAdmins = "cluster-admins"
	MRTypeClusterCIDR   = "cluster-cidr"
	MRTypeClusterUpdate = "cluster-update"
	ValuesAdminsKey     = "administrators"
)

type Admin struct {
	Username               string `json:"username" yaml:"username"`
	Email                  string `json:"email" yaml:"email"`
	FirstName              string `json:"firstName" yaml:"firstName"`
	LastName               string `json:"lastName" yaml:"lastName"`
	TmpPassword            string `json:"tmpPassword" yaml:"-"`
	PasswordVaultSecret    string `yaml:"passwordVaultSecret" json:"-"`
	PasswordVaultSecretKey string `yaml:"passwordVaultSecretKey" json:"-"`
}

func (a *App) updateAdminsView(ctx *gin.Context) (*router.Response, error) {
	if err := a.updateAdmins(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to update admins")
	}
	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) updateAdmins(ctx *gin.Context) error {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

	adminsValue := ctx.PostForm("admins")

	var admins []Admin
	if err := json.Unmarshal([]byte(adminsValue), &admins); err != nil {
		return errors.Wrap(err, "unable to decode admins from request")
	}

	if err := a.setAdminsVaultPassword(admins); err != nil {
		return errors.Wrap(err, "unable to create admins secrets")
	}

	valuesDict, err := registry.GetValuesFromGit(ctx, a.Config.CodebaseName, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to decode values yaml")
	}

	valuesDict[ValuesAdminsKey] = admins

	valuesValue, err := yaml.Marshal(valuesDict)
	if err != nil {
		return errors.Wrap(err, "unable to encode new values")
	}

	if err := a.createAdminsMergeRequest(userCtx, ctx, string(valuesValue)); err != nil {
		return errors.Wrap(err, "unable to create admins merge request")
	}

	return nil
}

func (a *App) createAdminsMergeRequest(userCtx context.Context, ctx *gin.Context, values string) error {
	if err := a.Services.Gerrit.CreateMergeRequestWithContents(userCtx, &gerrit.MergeRequest{
		ProjectName:   a.Config.CodebaseName,
		Name:          fmt.Sprintf("adm-mr-%s-%d", a.Config.CodebaseName, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("update cluster admins"),
		TargetBranch:  "master",
		Labels: map[string]string{
			registry.MRLabelTarget: MRTypeClusterAdmins,
		},
	}, map[string]string{
		registry.ValuesLocation: values,
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}

func (a *App) setAdminsVaultPassword(admins []Admin) error {
	for i, admin := range admins {
		vaultPath := strings.ReplaceAll(
			strings.ReplaceAll(a.Config.VaultClusterAdminsPathTemplate, "{admin}", admin.Email),
			"{engine}", a.Config.VaultKVEngineName)

		admins[i].PasswordVaultSecret = vaultPath
		vaultPath = registry.ModifyVaultPath(vaultPath)

		secretData := map[string]interface{}{
			a.Config.VaultClusterAdminsPasswordKey: admin.TmpPassword,
		}

		if _, err := a.Services.Vault.Write(
			vaultPath, map[string]interface{}{
				"data": secretData,
			}); err != nil {
			return errors.Wrap(err, "unable to write to vault")
		}

		admins[i].PasswordVaultSecretKey = a.VaultClusterAdminsPasswordKey
		admins[i].Username = admins[i].Email
	}

	return nil
}

func (a *App) getAdminsJSON(ctx context.Context) (string, error) {
	valuesDict, err := registry.GetValuesFromGit(ctx, a.Config.CodebaseName, a.Gerrit)
	if err != nil {
		return "", errors.Wrap(err, "unable to decode values")
	}

	adminsInterface, ok := valuesDict[ValuesAdminsKey]
	if !ok {
		return "[]", nil
	}

	bts, err := json.Marshal(adminsInterface)
	if err != nil {
		return "", errors.Wrap(err, "unable to json encode admins")
	}

	return string(bts), nil
}
