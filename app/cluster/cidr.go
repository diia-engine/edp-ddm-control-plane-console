package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (a *App) updateCIDRView(ctx *gin.Context) (router.Response, error) {
	if err := a.updateCIDR(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to update cidr")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) updateCIDR(ctx *gin.Context) error {
	userCtx := router.ContextWithUserAccessToken(ctx)

	cidrValue := ctx.PostForm("admin-cidr")

	var cidr []string
	if err := json.Unmarshal([]byte(cidrValue), &cidr); err != nil {
		return errors.Wrap(err, "unable to decode cidr")
	}

	if len(cidr) == 0 {
		return nil
	}

	vals, err := a.Services.Gerrit.GetFileContents(ctx, a.Config.CodebaseName, "master", registry.ValuesLocation)
	if err != nil {
		return errors.Wrap(err, "unable to get values contents")
	}

	var valuesDict map[string]interface{}
	if err := yaml.Unmarshal([]byte(vals), &valuesDict); err != nil {
		return errors.Wrap(err, "unable to decode values yaml")
	}

	globalInterface, ok := valuesDict["global"]
	if !ok {
		globalInterface = make(map[string]interface{})
	}
	globalDict := globalInterface.(map[string]interface{})

	whiteListInterface, ok := globalDict["whiteListIP"]
	if !ok {
		whiteListInterface = make(map[string]interface{})
	}
	whiteListDict := whiteListInterface.(map[string]interface{})
	whiteListDict["adminRoutes"] = strings.Join(cidr, " ")

	globalDict["whiteListIP"] = whiteListDict
	valuesDict["global"] = globalDict

	valuesValue, err := yaml.Marshal(valuesDict)
	if err != nil {
		return errors.Wrap(err, "unable to encode new values")
	}

	if err := a.createCIDRMergeRequest(userCtx, ctx, string(valuesValue)); err != nil {
		return errors.Wrap(err, "unable to create cidr merge request")
	}

	return nil
}

func (a *App) createCIDRMergeRequest(userCtx context.Context, ctx *gin.Context, values string) error {
	if err := a.createValuesMergeRequest(userCtx, &valuesMrConfig{
		name:          fmt.Sprintf("cidr-mr-%s-%d", a.Config.CodebaseName, time.Now().Unix()),
		values:        values,
		targetLabel:   MRTypeClusterCIDR,
		commitMessage: fmt.Sprintf("update cluster CIDR config"),
		authorName:    ctx.GetString(router.UserNameSessionKey),
		authorEmail:   ctx.GetString(router.UserEmailSessionKey),
	}); err != nil {
		return errors.Wrap(err, "unable to create MR")
	}

	return nil
}
