package codebase

import (
	"time"

	"github.com/pkg/errors"

	"ddm-admin-console/service/k8s"
)

type WithPermissions struct {
	Codebase  *Codebase
	CanUpdate bool
	CanDelete bool
}

func (r WithPermissions) Available() bool {
	return r.Codebase.Available()
}

func (r WithPermissions) FormattedCreatedAtTimezone(timezone string) string {
	loc, _ := time.LoadLocation(timezone)
	return r.Codebase.CreationTimestamp.In(loc).Format(ViewTimeFormat)
}

func (s Service) CheckPermissions(initial []Codebase, k8sService k8s.ServiceInterface) ([]WithPermissions, error) {
	codebases := make([]WithPermissions, 0, len(initial))
	for i := range initial {
		canGet, err := k8sService.CanI("v2.edp.epam.com", "codebases", "get", initial[i].Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check access for codebase: %s", initial[i].Name)
		}
		if !canGet {
			continue
		}

		canUpdate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", initial[i].Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check access for codebase: %s", initial[i].Name)
		}

		canDelete, err := k8sService.CanI("v2.edp.epam.com", "codebases", "delete", initial[i].Name)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to check access for codebase: %s", initial[i].Name)
		}

		codebases = append(codebases, WithPermissions{Codebase: &initial[i], CanDelete: canDelete,
			CanUpdate: canUpdate})
	}

	return codebases, nil
}

func (s Service) CheckIsAllowedToCreate(k8sService k8s.ServiceInterface) (bool, error) {
	allowedToCreate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "create", "")
	if err != nil {
		return false, errors.Wrap(err, "unable to check codebase creation access")
	}

	return allowedToCreate, nil
}

func (s Service) CheckIsAllowedToUpdate(codebaseName string, k8sService k8s.ServiceInterface) (bool, error) {
	canUpdate, err := k8sService.CanI("v2.edp.epam.com", "codebases", "update", codebaseName)
	if err != nil {
		return false, errors.Wrapf(err, "unable to check access for codebase: %s", codebaseName)
	}

	return canUpdate, nil
}
