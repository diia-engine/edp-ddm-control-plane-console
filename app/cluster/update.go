package cluster

import (
	"ddm-admin-console/app/registry"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
)

type updateRequest struct {
	Branch string `form:"branch" binding:"required"`
}

func (a *App) clusterUpdate(ctx *gin.Context) (*router.Response, error) {
	var ur updateRequest
	if err := ctx.ShouldBind(&ur); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			return nil, errors.Wrap(err, "unable to parse registry form")
		}

		return router.MakeResponse(200, "cluster/edit.html",
			gin.H{"page": "cluster", "errorsMap": validationErrors, "backupConf": BackupConfig{}}), nil
	}

	prj, err := a.Services.Gerrit.GetProject(ctx, a.Config.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get registry gerrit project")
	}

	if err := a.Services.Gerrit.CreateMergeRequest(ctx, &gerrit.MergeRequest{
		CommitMessage:       fmt.Sprintf("Update cluster to %s", ur.Branch),
		SourceBranch:        ur.Branch,
		ProjectName:         prj.Spec.Name,
		Name:                fmt.Sprintf("%s-update-%d", a.Config.CodebaseName, time.Now().Unix()),
		AuthorName:          ctx.GetString(router.UserNameSessionKey),
		AuthorEmail:         ctx.GetString(router.UserEmailSessionKey),
		AdditionalArguments: []string{"-X", "ours"},
		Labels: map[string]string{
			registry.MRLabelTarget: MRTypeClusterUpdate,
		},
	}); err != nil {
		return nil, errors.Wrap(err, "unable to create update merge request")
	}

	return router.MakeRedirectResponse(http.StatusFound, "/admin/cluster/management"), nil
}
