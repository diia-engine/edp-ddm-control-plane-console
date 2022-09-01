package registry

import (
	"ddm-admin-console/service/gerrit"
	"fmt"
)

type ExtendedMergeRequests struct {
	gerrit.GerritMergeRequest
}

func (e ExtendedMergeRequests) RequestName() string {
	if e.Labels[MRLabelTarget] == mrTargetExternalReg {
		return e.Annotations[mrAnnotationRegName]
	}

	if e.Labels[MRLabelTarget] == mrTargetRegistryVersionUpdate {
		return e.Spec.SourceBranch
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
