package config

import (
	"ddm-admin-console/app/registry"
)

func (cnf *Settings) RegistryConfig() registry.Config {
	return registry.Config{
		UsersNamespace:                  cnf.UsersNamespace,
		UsersRealm:                      cnf.UsersRealm,
		RegistryCodebaseLabels:          cnf.RegistryCodebaseLabels,
		EnableBranchProvisioners:        cnf.EnableBranchProvisioners,
		ClusterCodebaseName:             cnf.ClusterCodebaseName,
		Timezone:                        cnf.Timezone,
		GerritRegistryHost:              cnf.RegistryRepoHost,
		GerritRegistryPrefix:            cnf.RegistryRepoPrefix,
		HardwareINITemplatePath:         cnf.RegistryHardwareKeyINITemplatePath,
		VaultRegistrySMTPPwdSecretKey:   cnf.VaultRegistrySMTPPwdSecretKey,
		VaultRegistrySecretPathTemplate: cnf.VaultRegistrySecretPathTemplate,
		VaultKVEngineName:               cnf.VaultKVEngineName,
		VaultCitizenSSLPath:             cnf.VaultCitizenSSLPath,
		VaultOfficerSSLPath:             cnf.VaultOfficerSSLPath,
		TempFolder:                      cnf.TempFolder,
		RegistryDNSManualPath:           cnf.RegistryDNSManualPath,
		DDMManualEDPComponent:           cnf.DDMManualEDPComponent,
	}
}

func (s *Services) RegistryServices() registry.Services {
	return registry.Services{
		Codebase:     s.Codebase,
		K8S:          s.K8S,
		Keycloak:     s.Keycloak,
		Jenkins:      s.Jenkins,
		Gerrit:       s.Gerrit,
		EDPComponent: s.EDPComponent,
		Vault:        s.Vault,
		Perms:        s.PermService,
	}
}
