package config

import (
	"ddm-admin-console/service/codebase"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/keycloak"
	"ddm-admin-console/service/openshift"
	"ddm-admin-console/service/permissions"
	"ddm-admin-console/service/vault"

	"github.com/patrickmn/go-cache"
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
	JenkinsAPIURL                         string `envconfig:"JENKINS_API_URL" default:"http://jenkins:8080"`
	JenkinsAdminSecretName                string `envconfig:"JENKINS_ADMIN_SECRET_NAME" default:"jenkins-admin-token"`
	VaultNamespace                        string `envconfig:"VAULT_NAMESPACE" default:"user-management"`
	VaultSecretName                       string `envconfig:"VAULT_SECRET_NAME" default:"vault-root-token"`
	VaultSecretTokenKey                   string `envconfig:"VAULT_SECRET_TOKEN_KEY" default:"VAULT_ROOT_TOKEN"`
	VaultAPIAddr                          string `envconfig:"VAULT_API_ADDR" default:"http://hashicorp-vault.user-management:8200"`
	VaultRegistrySecretPathTemplate       string `envconfig:"V_REG_SEC_PATH_TPL" default:"{engine}/registry/{registry}"`
	VaultRegistrySMTPPwdSecretKey         string `envconfig:"V_REG_SMTP_SEC_KEY" default:"smtp-password"`
	VaultKVEngineName                     string `envconfig:"VAULT_KV_ENGINE_NAME" default:"registry-kv"`
	VaultClusterAdminsPathTemplate        string `envconfig:"V_CLS_ADM_PATH_TPL" default:"{engine}/cluster/{admin}"`
	VaultClusterAdminsPasswordKey         string `envconfig:"V_CLS_ADMIN_SEC_KEY" default:"password"`
	VaultClusterPathTemplate              string `envconfig:"V_CLS_ADM_PATH_TPL" default:"{engine}/cluster"`
	VaultClusterKeyManagementPathTemplate string `envconfig:"V_CLS_KEYM_PATH_TPL" default:"{engine}/cluster/key-management"`
	VaultCitizenSSLPath                   string `encvonfig:"V_SSL_CITIZEN_PATH" default:"custom-dns-names/{registry}/citizen-portal/{host}"`
	VaultOfficerSSLPath                   string `encvonfig:"V_SSL_CITIZEN_PATH" default:"custom-dns-names/{registry}/officer-portal/{host}"`
	VaultKeycloakSSLPath                  string `encvonfig:"V_SSL_KEYCLOAK_PATH" default:"custom-dns-names/{registry}/officer-portal/{host}"`
	TempFolder                            string `envconfig:"TEMP_FOLDER" default:"/tmp"`
	RegistryDNSManualPath                 string `envconfig:"REGISTRY_DNS_MANUAL_PATH" default:"platform/1.6/tech/infrastructure/custom-dns.html"`
	DDMManualEDPComponent                 string `envconfig:"DDM_MANUAL_EDP_COMPONENT" default:"ddm-architecture"`
	OAuthUseExternalTokenURL              bool   `envconfig:"OAUTH_USE_EXTERNAL_TOKEN_URL"`
	OAuthInternalTokenHost                string `envconfig:"OAUTH_INTERNAL_TOKEN_HOST" default:"oauth-openshift.openshift-authentication.svc"`
	GitUsername                           string `envconfig:"GERRIT_GIT_USERNAME" default:"project-creator"`
	GitKeySecretName                      string `envconfig:"GERRIT_GIT_KEY_SECRET_NAME" default:"gerrit-project-creator"`
	GitHost                               string `envconfig:"GERRIT_GIT_HOSTNAME" default:"gerrit"`
	GitPort                               string `envconfig:"GERRIT_GIT_PORT" default:"31000"`
	KeycloakDefaultHostname               string `envconfig:"KEYCLOAK_DEFAULT_HOSTNAME"`
	Mock                                  string `envconfig:"MOCK"`
	RegistryVersionFilter                 string `envconfig:"REGISTRY_VERSION_FILTER"`
	WiremockAddr                          string `envconfig:"WIREMOCK_ADDR" default:"http://wiremock:9021/"`
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
	Cache        *cache.Cache //TODO: make interface
	PermService  permissions.ServiceInterface
}
