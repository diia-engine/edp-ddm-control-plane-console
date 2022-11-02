package registry

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"

	"ddm-admin-console/service"
	"ddm-admin-console/service/codebase"
	"ddm-admin-console/service/gerrit"
)

type SortByVersion []string

func (a SortByVersion) Len() int           { return len(a) }
func (a SortByVersion) Less(i, j int) bool { return BranchVersion(a[i]).LessThan(BranchVersion(a[j])) }
func (a SortByVersion) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func BranchVersion(name string) *version.Version {
	v, err := version.NewVersion(name)
	if err != nil {
		v, _ := version.NewVersion("0")
		return v
	}
	return v
}

func HasUpdate(ctx context.Context, gerritService gerrit.ServiceInterface, cb *codebase.Codebase, mrTarget string) (bool, []string, error) {
	gerritProject, err := gerritService.GetProject(ctx, cb.Name)
	if service.IsErrNotFound(err) {
		return false, []string{}, nil
	}

	if err != nil {
		return false, nil, errors.Wrap(err, "unable to get gerrit project")
	}

	branches := UpdateBranches(gerritProject.Status.Branches)

	if len(branches) == 0 {
		return false, branches, nil
	}

	registryVersion := BranchVersion(cb.Spec.DefaultBranch)
	if cb.Spec.BranchToCopyInDefaultBranch != "" {
		registryVersion = BranchVersion(cb.Spec.BranchToCopyInDefaultBranch)
	}

	if registryVersion.Original() == "0" {
		registryVersion = LowestVersion(branches)
	}

	mrs, err := gerritService.GetMergeRequestByProject(ctx, gerritProject.Spec.Name)
	if err != nil {
		return false, branches, errors.Wrap(err, "unable to get merge requests")
	}

	branchesDict := make(map[string]string)
	for _, br := range branches {
		branchesDict[br] = br
	}

	for _, mr := range mrs {
		if mr.Labels[MRLabelTarget] != mrTarget {
			continue
		}

		if mr.Status.Value == gerrit.StatusNew {
			return false, branches, nil
		}

		if mr.Status.Value == gerrit.StatusMerged {
			mergedBranchVersion := BranchVersion(mr.Spec.SourceBranch)
			if registryVersion.LessThan(mergedBranchVersion) {
				registryVersion = mergedBranchVersion
			}

			delete(branchesDict, mr.Spec.SourceBranch)
		}
	}

	branches = []string{}
	for _, br := range branchesDict {
		if registryVersion.LessThan(BranchVersion(br)) {
			branches = append(branches, br)
		}
	}

	if len(branches) == 0 {
		return false, branches, nil
	}

	sort.Sort(SortByVersion(branches))
	return true, branches, nil
}

func LowestVersion(gerritProjectBranches []string) *version.Version {
	registryVersion := BranchVersion("0")
	if len(gerritProjectBranches) > 0 {
		registryVersion = BranchVersion(gerritProjectBranches[0])
		for i := 1; i < len(gerritProjectBranches); i++ {
			v := BranchVersion(gerritProjectBranches[i])
			if v.LessThan(registryVersion) {
				registryVersion = v
			}
		}
	}

	return registryVersion
}

func UpdateBranches(projectBranches []string) []string {
	var updateBranches []string
	for _, br := range projectBranches {
		if strings.Contains(br, "refs/heads") && !strings.Contains(br, "master") {
			updateBranches = append(updateBranches, strings.Replace(br, "refs/heads/", "", 1))
		}
	}

	return updateBranches
}

//func (a *App) filterUpdateBranchesByCluster(ctx context.Context, registryBranches []string) ([]string, error) {
//	clusterCodebase, err := a.Services.Codebase.Get(a.Config.ClusterCodebaseName)
//	if err != nil {
//		return nil, errors.Wrap(err, "unable to get cluster codebase")
//	}
//
//	hasUpdate, clusterBranches, err := HasUpdate(ctx, a.Services.Gerrit, clusterCodebase, MRTargetRegistryVersionUpdate)
//	if err != nil {
//		return nil, errors.Wrap(err, "unable to get cluster update branches")
//	}
//
//	if !hasUpdate {
//		return []string{}, nil
//	}
//
//	if len(clusterBranches) == 0 {
//		return []string{}, nil
//	}
//
//	clusterBranch := clusterBranches[0]
//
//	var filteredBranches []string
//	for _, rb := range registryBranches {
//		if BranchVersion(rb).LessThanOrEqual(BranchVersion(clusterBranch)) {
//			filteredBranches = append(filteredBranches, rb)
//		}
//	}
//
//	return filteredBranches, nil
//}
