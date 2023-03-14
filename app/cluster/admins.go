package cluster

import (
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const (
	MRTypeClusterAdmins         = "cluster-admins"
	MRTypeClusterCIDR           = "cluster-cidr"
	MRTypeClusterKeycloakDNS    = "cluster-keycloak-dns"
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
	adminsValue := ctx.PostForm("admins")

	var admins []Admin
	if err := json.Unmarshal([]byte(adminsValue), &admins); err != nil {
		return errors.Wrap(err, "unable to decode admins from request")
	}

	if err := a.setAdminsVaultPassword(admins); err != nil {
		return errors.Wrap(err, "unable to create admins secrets")
	}

	values, err := registry.GetValuesFromGit(a.Config.CodebaseName, registry.MasterBranch, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to decode values yaml")
	}

	values.OriginalYaml[ValuesAdminsKey] = admins

	if err := a.createValuesMergeRequestCtx(ctx, MRTypeClusterAdmins, "update cluster admins",
		values.OriginalYaml); err != nil {
		return errors.Wrap(err, "unable to create admins merge request")
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
		//TODO: user registry.CreateVaultSecrets
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

func (a *App) getAdminsJSON(values *registry.Values) (string, error) {
	adminsInterface, ok := values.OriginalYaml[ValuesAdminsKey]
	if !ok {
		return "[]", nil
	}

	bts, err := json.Marshal(adminsInterface)
	if err != nil {
		return "", errors.Wrap(err, "unable to json encode admins")
	}

	return string(bts), nil
}
