package config

import "ddm-admin-console/service/vault"

func (cnf *Settings) VaultConfig() vault.Config {
	return vault.Config{
		SecretNamespace: cnf.VaultNamespace,
		SecretTokenKey:  cnf.VaultSecretTokenKey,
		SecretName:      cnf.VaultSecretName,
		APIAddr:         cnf.VaultAPIAddr,
		KVEngineName:    cnf.VaultKVEngineName,
	}
}
