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

func (a *App) updateCIDRView(ctx *gin.Context) (*router.Response, error) {
	if err := a.updateCIDR(ctx); err != nil {
		return nil, errors.Wrap(err, "unable to update cidr")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}

func (a *App) updateCIDR(ctx *gin.Context) error {
	userCtx := a.router.ContextWithUserAccessToken(ctx)

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
	if err := a.Services.Gerrit.CreateMergeRequestWithContents(userCtx, &gerrit.MergeRequest{
		ProjectName:   a.Config.CodebaseName,
		Name:          fmt.Sprintf("cidr-mr-%s-%d", a.Config.CodebaseName, time.Now().Unix()),
		AuthorEmail:   ctx.GetString(router.UserEmailSessionKey),
		AuthorName:    ctx.GetString(router.UserNameSessionKey),
		CommitMessage: fmt.Sprintf("update cluster CIDR config"),
		TargetBranch:  "master",
		Labels: map[string]string{
			registry.MRLabelTarget: MRTypeClusterCIDR,
		},
	}, map[string]string{
		registry.ValuesLocation: values,
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}
