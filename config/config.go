package config

import (
	"ddm-admin-console/service/codebase"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/keycloak"
	"ddm-admin-console/service/openshift"
	"ddm-admin-console/service/vault"
)

type Settings struct {
	HTTPPort                              string `envconfig:"HTTP_PORT" default:"8080"`
	LogLevel                              string `envconfig:"LOG_LEVEL" default:"INFO"`
	LogEncoding                           string `envconfig:"LOG_ENCODING" default:"json"`
	Namespace                             string `envconfig:"NAMESPACE" default:"default"`
	SessionSecret                         string `envconfig:"SESSION_SECRET" default:"UdWaTEfULunPTkRC9sFLG26APz9W5gEC8x"`
	OCClientID                            string `envconfig:"OC_CLIENT_ID"`
	OCClientSecret                        string `envconfig:"OC_CLIENT_SECRET"`
	Host                                  string `envconfig:"HOST"`
	ClusterCodebaseName                   string `envconfig:"CLUSTER_CODEBASE_NAME"`
	ClusterRepo                           string `envconfig:"CLUSTER_REPO"`
	BackupSecretName                      string `envconfig:"BACKUP_SECRET_NAME" default:"backup-credential"`
	GinMode                               string `envconfig:"GIN_MODE"`
	Timezone                              string `envconfig:"TIMEZONE" default:"Europe/Kiev"`
	RegistryRepoPrefix                    string `envconfig:"REGISTRY_REPO_PREFIX" default:"registry-tenant-template-"`
	RegistryRepoHost                      string `envconfig:"REGISTRY_REPO_HOST"`
	RegistryHardwareKeyINITemplatePath    string `envconfig:"REGISTRY_HW_KEY_INI_TPL_PATH" default:"osplm.ini"`
	RootGerritName                        string `envconfig:"ROOT_GERRIT_NAME" default:"gerrit"`
	GroupGitRepo                          string `envconfig:"GROUP_GIT_REPO"`
	UsersNamespace                        string `envconfig:"USERS_NAMESPACE" default:"user-management"`
	UsersRealm                            string `envconfig:"USERS_REALM" default:"openshift"`
	EnableBranchProvisioners              bool   `envconfig:"ENABLE_BRANCH_PROVISIONERS"`
	RegistryCodebaseLabels                string `envconfig:"REGISTRY_CODEBASE_LABELS"`
	GerritAPIUrlTemplate                  string `envconfig:"GERRIT_API_URL_TPL" default:"http://{HOST}:8080/a/"`
	VaultNamespace                        string `envconfig:"VAULT_NAMESPACE" default:"user-management"`
	VaultSecretName                       string `envconfig:"VAULT_SECRET_NAME" default:"vault-root-token"`
	VaultSecretTokenKey                   string `envconfig:"VAULT_SECRET_TOKEN_KEY" default:"VAULT_ROOT_TOKEN"`
	VaultAPIAddr                          string `envconfig:"VAULT_API_ADDR" default:"http://hashicorp-vault.user-management:8200"`
	VaultRegistrySecretPathTemplate       string `envconfig:"V_REG_SEC_PATH_TPL" default:"{engine}/registry/{registry}"`
	VaultRegistrySMTPPwdSecretKey         string `envconfig:"V_REG_SMTP_SEC_KEY" default:"smtp-password"`
	VaultKVEngineName                     string `envconfig:"VAULT_KV_ENGINE_NAME" default:"registry-kv"`
	VaultClusterAdminsPathTemplate        string `envconfig:"V_CLS_ADM_PATH_TPL" default:"{engine}/cluster/{admin}"`
	VaultClusterKeyManagementPathTemplate string `envconfig:"V_CLS_KEYM_PATH_TPL" default:"{engine}/cluster/key-management"`
	VaultClusterAdminsPasswordKey         string `envconfig:"V_CLS_ADMIN_SEC_KEY" default:"password"`
	VaultOfficerCACertKey                 string `envconfig:"V_SSL_OFFICER_CA_CERT_KEY" default:"caCertificate"`
	VaultOfficerCertKey                   string `envconfig:"V_SSL_OFFICER_CA_CERT_KEY" default:"certificate"`
	VaultOfficerPKKey                     string `envconfig:"V_SSL_OFFICER_PK_KEY" default:"key"`
	VaultCitizenCACertKey                 string `envconfig:"V_SSL_CITIZEN_CA_CERT_KEY" default:"caCertificate"`
	VaultCitizenCertKey                   string `envconfig:"V_SSL_CITIZEN_CA_CERT_KEY" default:"certificate"`
	VaultCitizenPKKey                     string `envconfig:"V_SSL_CITIZEN_PK_KEY" default:"key"`
	VaultCitizenSSLPath                   string `encvonfig:"V_SSL_CITIZEN_PATH" default:"custom-dns-names/{registry}/citizen-portal/{host}"`
	VaultOfficerSSLPath                   string `encvonfig:"V_SSL_CITIZEN_PATH" default:"custom-dns-names/{registry}/officer-portal/{host}"`
	TempFolder                            string `envconfig:"TEMP_FOLDER" default:"/tmp"`
}

type Services struct {
	Codebase     codebase.ServiceInterface
	EDPComponent edpcomponent.ServiceInterface
	K8S          k8s.ServiceInterface
	OpenShift    openshift.ServiceInterface
	Gerrit       gerrit.ServiceInterface
	Jenkins      jenkins.ServiceInterface
	Keycloak     keycloak.ServiceInterface
	Vault        vault.ServiceInterface
}
