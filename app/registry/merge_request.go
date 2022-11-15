package registry

import (
	"ddm-admin-console/service/gerrit"
	"fmt"
)

type ExtendedMergeRequests struct {
	gerrit.GerritMergeRequest
}

func (e ExtendedMergeRequests) StatusValues() string {
	if e.Labels[MRLabelAction] == MRLabelActionBranchMerge && e.Spec.SourceBranch == "" {
		return "in progress"
	}

	return e.Status.Value
}

func (e ExtendedMergeRequests) RequestName() string {
	if e.Labels[MRLabelTarget] == mrTargetExternalReg {
		return e.Annotations[mrAnnotationRegName]
	}

	if e.Labels[MRLabelTarget] == MRTargetRegistryVersionUpdate {
		if e.Spec.SourceBranch != "" {
			return e.Spec.SourceBranch
		} else if e.Labels[MRLabelSourceBranch] != "" {
			return e.Labels[MRLabelSourceBranch]
		}
	}

	if e.Labels[MRLabelTarget] == mrTargetEditRegistry {
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

	return "-"
}
