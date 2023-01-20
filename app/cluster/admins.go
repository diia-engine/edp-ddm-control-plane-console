package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const (
	MRTypeClusterAdmins         = "cluster-admins"
	MRTypeClusterCIDR           = "cluster-cidr"
	MRTypeClusterUpdate         = "cluster-update"
	MRTypeClusterBackupSchedule = "cluster-backup-schedule"
	ValuesAdminsKey             = "administrators"
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

func (a *App) updateAdminsView(ctx *gin.Context) (router.Response, error) {
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
	if err := a.createValuesMergeRequest(userCtx, &valuesMrConfig{
		name:          fmt.Sprintf("adm-mr-%s-%d", a.Config.CodebaseName, time.Now().Unix()),
		values:        values,
		targetLabel:   MRTypeClusterAdmins,
		commitMessage: fmt.Sprintf("update cluster admins"),
		authorName:    ctx.GetString(router.UserNameSessionKey),
		authorEmail:   ctx.GetString(router.UserEmailSessionKey),
	}); err != nil {
		return errors.Wrap(err, "unable to create MR")
	}

	return nil
}

func (a *App) setAdminsVaultPassword(admins []Admin) error {
	for i, admin := range admins {
		//TODO: add separate folder for admins
		vaultPath := a.vaultPlatformPathKey(admin.Email)

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
