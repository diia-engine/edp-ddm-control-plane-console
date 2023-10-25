package cluster

import (
	"ddm-admin-console/router"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func (a *App) updateCIDRView(ctx *gin.Context) (router.Response, error) {
	if err := a.updateCIDR(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to update cidr")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) updateCIDR(ctx *gin.Context) error {
	cidrValue := ctx.PostForm("admin-cidr")

	var cidr []string
	if err := json.Unmarshal([]byte(cidrValue), &cidr); err != nil {
		return errors.Wrap(err, "unable to decode cidr")
	}

	if len(cidr) == 0 {
		return nil
	}

	values, err := getValuesFromGit(a.Config.CodebaseName, masterBranch, a.Gerrit)
	if err != nil {
		return errors.Wrap(err, "unable to get values contents")
	}

	values.Global.WhiteListIP.AdminRoutes = strings.Join(cidr, " ")

	if err := a.createValuesMergeRequestCtx(ctx, MRTypeClusterCIDR, "update cluster CIDR config", values); err != nil {
		return errors.Wrap(err, "unable to create cidr merge request")
	}

	return nil
}
