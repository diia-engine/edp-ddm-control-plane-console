package registry

import (
	"ddm-admin-console/service/gerrit"
	"fmt"
)

type ExtendedMergeRequests struct {
	gerrit.GerritMergeRequest
}

func (e ExtendedMergeRequests) StatusValue() string {
	if (e.Labels[MRLabelAction] == MRLabelActionBranchMerge && e.Spec.SourceBranch == "") || e.Status.Value == "" ||
		(e.Status.Value == "sourceBranch or changesConfigMap must be specified" && e.Spec.SourceBranch != "") {
		return "in progress"
	}

	return e.Status.Value
}

func (e ExtendedMergeRequests) RequestName() string {
	if e.Labels[MRLabelTarget] == mrTargetExternalReg {
		return e.Annotations[mrAnnotationRegName]
	}

	if e.Labels[MRLabelTarget] == MRTargetRegistryVersionUpdate {
		return "Оновлення версії реєстру"
	}

	if (e.Labels[MRLabelTarget] == mrTargetEditRegistry) || (e.Labels[MRLabelTarget] == mrTargetEditTrembita) {
		return "Редагування реєстру"
	}

	return e.Name
}

func (e ExtendedMergeRequests) Action() string {
	if e.Labels[MRLabelTarget] == mrTargetExternalReg {
		action, ok := e.Labels[MRLabelSubTarget]
		if ok {
			return fmt.Sprintf("mre-action-%s", action)
		}
	}

	if e.Labels[MRLabelTarget] == MRTargetRegistryVersionUpdate {
		sourceBranch := e.Spec.SourceBranch
		if sourceBranch == "" {
			sourceBranch = e.Labels[MRLabelSourceBranch]
		}

		return fmt.Sprintf("Оновлення реєстру до %s", sourceBranch)
	}

	return "-"
}
