package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/router"
	"ddm-admin-console/service/gerrit"
	"fmt"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
)

type ExtendedMergeRequests struct {
	gerrit.GerritMergeRequest
}

type valuesMrConfig struct {
	name          string
	authorName    string
	authorEmail   string
	commitMessage string
	targetBranch  string
	targetLabel   string
	values        string
}

func (v valuesMrConfig) TargetBranch() string {
	if v.targetBranch == "" {
		return "master"
	}

	return v.targetBranch
}

func (e ExtendedMergeRequests) RequestName() string {
	return e.Name
}

func (e ExtendedMergeRequests) StatusValue() string {
	if e.Labels[registry.MRLabelAction] == registry.MRLabelActionBranchMerge &&
		(e.Spec.SourceBranch == "" || e.Status.Value == "sourceBranch or changesConfigMap must be specified") {
		return "in progress"
	}

	if e.Status.Value == "" {
		return "-"
	}

	return e.Status.Value
}

func (e ExtendedMergeRequests) Action() string {
	if e.Labels[registry.MRLabelTarget] == MRTypeClusterAdmins {
		return "Оновлення адміністраторів платформи"
	}

	if e.Labels[registry.MRLabelTarget] == MRTypeClusterCIDR {
		return "Обмеження доступу"
	}

	if e.Labels[registry.MRLabelTarget] == MRTypeClusterKeycloakDNS {
		return "Редагування DNS Keycloak"
	}

	sourceBranch := e.Spec.SourceBranch
	if sourceBranch == "" {
		sourceBranch = e.Labels[registry.MRLabelSourceBranch]
	}

	if e.Labels[registry.MRLabelTarget] == registry.MRTargetClusterUpdate {
		return fmt.Sprintf("Оновлення платформи до %s", sourceBranch)
	}

	return "-"
}

func (a *App) ClusterGetMergeRequests(userCtx context.Context) ([]ExtendedMergeRequests, error) {
	mrs, err := a.Services.Gerrit.GetMergeRequestByProject(userCtx, a.CodebaseName)
	if err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit merge requests")
	}

	sort.Sort(gerrit.SortByCreationDesc(mrs))

	emrs := make([]ExtendedMergeRequests, 0, len(mrs))
	for _, mr := range mrs {
		emrs = append(emrs, ExtendedMergeRequests{GerritMergeRequest: mr})
	}

	return emrs, nil
}

func (a *App) createValuesMergeRequestCtx(ctx *gin.Context, target, commitMessage string, values *Values) error {
	valuesValue, err := yaml.Marshal(values)
	if err != nil {
		return errors.Wrap(err, "unable to encode new values")
	}

	if err := a.createValuesMergeRequest(router.ContextWithUserAccessToken(ctx), &valuesMrConfig{
		name:          fmt.Sprintf("mr-%s-%d", a.Config.CodebaseName, time.Now().Unix()),
		values:        string(valuesValue),
		targetLabel:   target,
		commitMessage: commitMessage,
		authorName:    ctx.GetString(router.UserNameSessionKey),
		authorEmail:   ctx.GetString(router.UserEmailSessionKey),
	}); err != nil {
		return errors.Wrap(err, "unable to create MR")
	}

	return nil
}

func (a *App) createValuesMergeRequest(ctx context.Context, cnf *valuesMrConfig) error {
	if err := a.Services.Gerrit.CreateMergeRequestWithContents(ctx, &gerrit.MergeRequest{
		ProjectName:   a.Config.CodebaseName,
		Name:          cnf.name,
		AuthorEmail:   cnf.authorEmail,
		AuthorName:    cnf.authorName,
		CommitMessage: cnf.commitMessage,
		TargetBranch:  cnf.TargetBranch(),
		Labels: map[string]string{
			registry.MRLabelTarget: cnf.targetLabel,
		},
	}, map[string]string{
		ValuesLocation: cnf.values,
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}
