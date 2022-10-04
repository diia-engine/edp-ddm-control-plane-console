package config

import "ddm-admin-console/app/cluster"

func (s *Services) ClusterServices() cluster.Services {
	return cluster.Services{
		Codebase:     s.Codebase,
		K8S:          s.K8S,
		EDPComponent: s.EDPComponent,
		Gerrit:       s.Gerrit,
		Jenkins:      s.Jenkins,
		Vault:        s.Vault,
	}
}

func (cnf *Settings) ClusterConfig() cluster.Config {
	return cluster.Config{
		RegistryRepoHost:                      cnf.RegistryRepoHost,
		BackupSecretName:                      cnf.BackupSecretName,
		ClusterRepo:                           cnf.ClusterRepo,
		CodebaseName:                          cnf.ClusterCodebaseName,
		VaultClusterAdminsPasswordKey:         cnf.VaultClusterAdminsPasswordKey,
		VaultClusterAdminsPathTemplate:        cnf.VaultClusterAdminsPathTemplate,
		VaultKVEngineName:                     cnf.VaultKVEngineName,
		HardwareINITemplatePath:               cnf.RegistryHardwareKeyINITemplatePath,
		VaultClusterKeyManagementPathTemplate: cnf.VaultClusterKeyManagementPathTemplate,
	}
}
