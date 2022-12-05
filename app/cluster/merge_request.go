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

func (e ExtendedMergeRequests) RequestName() string {
	return e.Name
}

func (e ExtendedMergeRequests) StatusValue() string {
	if e.Labels[registry.MRLabelAction] == registry.MRLabelActionBranchMerge &&
		(e.Spec.SourceBranch == "" || e.Status.Value == "sourceBranch or changesConfigMap must be specified") {
		return "in progress"
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

	sourceBranch := e.Spec.SourceBranch
	if sourceBranch == "" {
		sourceBranch = e.Labels[registry.MRLabelSourceBranch]
	}

	if e.Labels[registry.MRLabelTarget] == MRTypeClusterUpdate {
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
