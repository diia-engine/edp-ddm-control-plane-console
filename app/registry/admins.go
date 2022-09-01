package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/gosimple/slug"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"ddm-admin-console/service/keycloak"
)

type Admin struct {
	Username               string `json:"username" yaml:"username"`
	Email                  string `json:"email" yaml:"email"`
	FirstName              string `json:"firstName" yaml:"firstName"`
	LastName               string `json:"lastName" yaml:"lastName"`
	TmpPassword            string `json:"tmpPassword,omitempty" yaml:"tmpPassword,omitempty"`
	PasswordVaultSecret    string `yaml:"passwordVaultSecret" json:"passwordVaultSecret"`
	PasswordVaultSecretKey string `yaml:"passwordVaultSecretKey" json:"passwordVaultSecretKey"`
}

type Admins struct {
	keycloakService keycloak.ServiceInterface
	usersRealm      string
	usersNamespace  string
}

func MakeAdmins(keycloakService keycloak.ServiceInterface, usersRealm, usersNamespace string) *Admins {
	return &Admins{
		keycloakService: keycloakService,
		usersRealm:      usersRealm,
		usersNamespace:  usersNamespace,
	}
}

func (a *Admins) GetAdmins(ctx context.Context, registryName string) ([]Admin, error) {
	usrs, err := a.keycloakService.GetUsersByRealm(ctx, a.usersRealm)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get users by realm")
	}

	admins := []Admin{}
	for _, u := range usrs {
		for _, g := range u.Spec.Groups {
			if g == userGroupRoleNameFromRegistry(registryName) {
				admins = append(admins, Admin{
					Username:  u.Spec.Username,
					Email:     u.Spec.Email,
					FirstName: u.Spec.FirstName,
					LastName:  u.Spec.LastName,
				})
			}
		}
	}

	return admins, nil
}

//deprecated
func (a *Admins) SyncAdmins(ctx context.Context, registryName string, admins []Admin) error {
	usrs, err := a.keycloakService.GetUsersByRealm(ctx, a.usersRealm)
	if err != nil {
		return errors.Wrap(err, "unable to get users by realm")
	}

	adminsDict := make(map[string]Admin)
	for i, v := range admins {
		adminsDict[v.Email] = admins[i]
	}

	for i := range usrs {
		if _, ok := adminsDict[usrs[i].Spec.Email]; ok {
			// add to groups and roles
			// remove from dict
			if err := a.adminAddToGroupsAndRoles(ctx, registryName, &usrs[i]); err != nil {
				return errors.Wrap(err, "unable to add registry to realm user groups and roles")
			}

			delete(adminsDict, usrs[i].Spec.Email)
		} else {
			// remove from groups and roles
			if err := a.adminRemoveFromGroupsAndRoles(ctx, registryName, &usrs[i]); err != nil {
				return errors.Wrap(err, "unable to remote registry from realm user groups and roles")
			}
		}
	}

	for _, adm := range adminsDict {
		if err := a.adminCreate(ctx, registryName, &adm); err != nil {
			return errors.Wrap(err, "unable to create admin")
		}
	}

	return nil
}

func (a *Admins) adminCreate(ctx context.Context, registryName string, adm *Admin) error {
	if err := a.keycloakService.CreateUser(ctx, &keycloak.KeycloakRealmUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      userK8SNameFromUsername(adm.Username),
			Namespace: a.usersNamespace,
		},
		Spec: keycloak.KeycloakRealmUserSpec{
			FirstName:           adm.FirstName,
			LastName:            adm.LastName,
			Username:            adm.Username,
			Roles:               []string{userGroupRoleNameFromRegistry(registryName)},
			Groups:              []string{userGroupRoleNameFromRegistry(registryName)},
			Email:               adm.Email,
			Realm:               a.usersRealm,
			Password:            adm.TmpPassword,
			KeepResource:        true,
			Enabled:             true,
			EmailVerified:       true,
			RequiredUserActions: []string{"UPDATE_PASSWORD"},
		},
	}); err != nil {
		return errors.Wrap(err, "unable to create realm user from admin")
	}

	return nil
}

func userGroupRoleNameFromRegistry(registryName string) string {
	return fmt.Sprintf("cp-registry-admin-%s", registryName)
}

func userK8SNameFromUsername(username string) string {
	return fmt.Sprintf("admin-%s",
		strings.Replace(slug.Make(username), "_", "-", -1))
}

func (a *Admins) adminRemoveFromGroupsAndRoles(ctx context.Context, registryName string, u *keycloak.KeycloakRealmUser) error {
	groupRemoved, roleRemoved := false, false
	var newGroups, newRoles []string

	for _, g := range u.Spec.Groups {
		if g == userGroupRoleNameFromRegistry(registryName) {
			groupRemoved = true
			continue
		}

		newGroups = append(newGroups, g)
	}

	if groupRemoved {
		u.Spec.Groups = newGroups
	}

	for _, r := range u.Spec.Roles {
		if r == userGroupRoleNameFromRegistry(registryName) {
			roleRemoved = true
			continue
		}

		newRoles = append(newRoles, r)
	}

	if roleRemoved {
		u.Spec.Roles = newRoles
	}

	if groupRemoved || roleRemoved {
		if err := a.keycloakService.UpdateUser(ctx, u); err != nil {
			return errors.Wrap(err, "unable to update user")
		}
	}

	return nil
}

func (a *Admins) adminAddToGroupsAndRoles(ctx context.Context, registryName string, u *keycloak.KeycloakRealmUser) error {
	addToGroups, addToRoles := true, true
	for _, g := range u.Spec.Groups {
		if g == userGroupRoleNameFromRegistry(registryName) {
			addToGroups = false
			break
		}
	}

	if addToGroups {
		u.Spec.Groups = append(u.Spec.Groups, userGroupRoleNameFromRegistry(registryName))
	}

	for _, r := range u.Spec.Roles {
		if r == userGroupRoleNameFromRegistry(registryName) {
			addToRoles = false
			break
		}
	}

	if addToRoles {
		u.Spec.Roles = append(u.Spec.Roles, userGroupRoleNameFromRegistry(registryName))
	}

	if addToRoles || addToGroups {
		if err := a.keycloakService.UpdateUser(ctx, u); err != nil {
			return errors.Wrap(err, "unable to update user")
		}
	}

	return nil
}

func (a *Admins) formatViewAdmins(ctx context.Context, registryName string) (string, error) {
	usrs, err := a.keycloakService.GetUsersByRealm(ctx, a.usersRealm)
	if err != nil {
		return "", errors.Wrap(err, "unable to load admins")
	}

	var registryAdmins []string

	for _, u := range usrs {
		for _, role := range u.Spec.Roles {
			if role == userGroupRoleNameFromRegistry(registryName) {
				registryAdmins = append(registryAdmins, u.Spec.Email)
				break
			}
		}
	}

	return strings.Join(registryAdmins, ", "), nil
}
