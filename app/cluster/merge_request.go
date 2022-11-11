package cluster

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/service/gerrit"
	"fmt"
	"sort"

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

func (e ExtendedMergeRequests) Action() string {
	if e.Labels[registry.MRLabelTarget] == MRTypeClusterAdmins {
		return "Оновлення адміністраторів платформи"
	}

	if e.Labels[registry.MRLabelTarget] == MRTypeClusterCIDR {
		return "Обмеження доступу"
	}

	if e.Labels[registry.MRLabelTarget] == MRTypeClusterUpdate {
		return fmt.Sprintf("Оновлення платформи до %s", e.Spec.SourceBranch)
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
		registry.ValuesLocation: cnf.values,
	}); err != nil {
		return errors.Wrap(err, "unable to create MR with new values")
	}

	return nil
}
