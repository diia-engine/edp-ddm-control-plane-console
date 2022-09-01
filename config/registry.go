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
		VaultOfficerCACertKey:           cnf.VaultOfficerCACertKey,
		VaultOfficerCertKey:             cnf.VaultOfficerCertKey,
		VaultOfficerPKKey:               cnf.VaultOfficerPKKey,
		VaultCitizenCACertKey:           cnf.VaultCitizenCACertKey,
		VaultCitizenCertKey:             cnf.VaultCitizenCertKey,
		VaultCitizenPKKey:               cnf.VaultCitizenPKKey,
		VaultCitizenSSLPath:             cnf.VaultCitizenSSLPath,
		VaultOfficerSSLPath:             cnf.VaultOfficerSSLPath,
		TempFolder:                      cnf.TempFolder,
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
	}
}
