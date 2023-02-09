package codebase

import (
	"fmt"
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

func CheckCodebasePermission(name string, k8sService k8s.ServiceInterface) (canGet, canUpdate, canDelete bool, retErr error) {
	canGet, err := k8sService.CanI("v2.edp.epam.com", "codebases", "get", name)
	if err != nil {
		retErr = fmt.Errorf("unable to check access for codebase: %s, err: %w", name, err)
		return
	}
	if !canGet {
		return
	}

	canUpdate, err = k8sService.CanI("v2.edp.epam.com", "codebases", "update", name)
	if err != nil {
		retErr = fmt.Errorf("unable to check access for codebase: %s, err: %w", name, err)
		return
	}

	canDelete, err = k8sService.CanI("v2.edp.epam.com", "codebases", "delete", name)
	if err != nil {
		retErr = fmt.Errorf("unable to check access for codebase: %s, err: %w", name, err)
		return
	}

	return
}

func (s Service) CheckPermissions(initial []Codebase, k8sService k8s.ServiceInterface) ([]WithPermissions, error) {
	codebases := make([]WithPermissions, 0, len(initial))
	for i := range initial {
		canGet, canUpdate, canDelete, err := CheckCodebasePermission(initial[i].Name, k8sService)
		if err != nil {
			return nil, fmt.Errorf("unable to check codebase permissions: %w", err)
		}

		if !canGet {
			continue
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
