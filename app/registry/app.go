package registry

import (
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/keycloak"
	"ddm-admin-console/service/permissions"
	"ddm-admin-console/service/vault"
	"fmt"
	"strings"

	"github.com/patrickmn/go-cache"

	"github.com/pkg/errors"
)

type Config struct {
	GerritRegistryPrefix            string
	GerritRegistryHost              string
	HardwareINITemplatePath         string
	EnableBranchProvisioners        bool
	ClusterCodebaseName             string
	RegistryCodebaseLabels          string
	Timezone                        string
	UsersRealm                      string
	UsersNamespace                  string
	VaultRegistrySecretPathTemplate string
	VaultRegistrySMTPPwdSecretKey   string
	VaultKVEngineName               string
	VaultCitizenSSLPath             string
	VaultOfficerSSLPath             string
	TempFolder                      string
	RegistryDNSManualPath           string
	DDMManualEDPComponent           string
	RegistryVersionFilter           string
	KeycloakDefaultHostname         string
	WiremockAddr                    string
}

type Services struct {
	Codebase     codebase.ServiceInterface
	Gerrit       gerrit.ServiceInterface
	EDPComponent edpComponent.ServiceInterface
	K8S          k8s.ServiceInterface
	Jenkins      jenkins.ServiceInterface
	Keycloak     keycloak.ServiceInterface
	Vault        vault.ServiceInterface
	Cache        *cache.Cache //TODO: replace with interface
	Perms        permissions.ServiceInterface
}

type App struct {
	Config
	Services
	router         router.Interface
	codebaseLabels map[string]string
	admins         *Admins
	versionFilter  *VersionFilter
}

func Make(router router.Interface, services Services, cnf Config) (*App, error) {
	app := &App{
		Config:   cnf,
		Services: services,
		router:   router,
		admins:   MakeAdmins(services.Keycloak, cnf.UsersRealm, cnf.UsersNamespace),
	}

	if cnf.RegistryCodebaseLabels != "" {
		labels := strings.Split(cnf.RegistryCodebaseLabels, ",")
		if len(labels) > 0 {
			app.codebaseLabels = make(map[string]string)
			for _, l := range labels {
				labelParts := strings.Split(l, "=")
				if len(labelParts) == 2 {
					app.codebaseLabels[labelParts[0]] = labelParts[1]
				}
			}
		}
	}

	app.createRoutes()
	if err := app.registerCustomValidators(); err != nil {
		return nil, errors.Wrap(err, "unable to register validators")
	}

	vf, err := MakeVersionFilter(cnf.RegistryVersionFilter)
	if err != nil {
		return nil, fmt.Errorf("unable to init version filter, %w", err)
	}
	app.versionFilter = vf

	return app, nil
}
