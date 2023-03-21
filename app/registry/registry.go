package registry

import (
	"fmt"
)

const (
	ScenarioKeyRequired    = "key-required"
	ScenarioKeyNotRequired = "key-not-required"
	KeyDeviceTypeFile      = "file"
	KeyDeviceTypeHardware  = "hardware"
	SMTPTypePlatform       = "platform-mail-server"
	SMTPTypeExternal       = "external-mail-server"
)

type registry struct {
	Name                   string   `form:"name" binding:"required,min=3,max=12,registry-name" json:"name"`
	Description            string   `form:"description" valid:"max=250" json:"description"`
	Admins                 string   `form:"admins" json:"admins"`
	AdminsChanged          string   `form:"admins-changed"`
	SignKeyIssuer          string   `form:"sign-key-issuer" binding:"required_if=KeyDeviceType file Scenario key-required"`
	SignKeyPwd             string   `form:"sign-key-pwd" binding:"required_if=KeyDeviceType file Scenario key-required"`
	RegistryGitTemplate    string   `form:"registry-git-template" binding:"required"`
	RegistryGitBranch      string   `form:"registry-git-branch" binding:"required"`
	KeyDeviceType          string   `form:"key-device-type" binding:"oneof=file hardware"`
	RemoteType             string   `form:"remote-type" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPassword      string   `form:"remote-key-pwd" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAName           string   `form:"remote-ca-name" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAHost           string   `form:"remote-ca-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteCAPort           string   `form:"remote-ca-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteSerialNumber     string   `form:"remote-serial-number" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyPort          string   `form:"remote-key-port" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyHost          string   `form:"remote-key-host" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	RemoteKeyMask          string   `form:"remote-key-mask" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	Scenario               string   `binding:"oneof=key-required key-not-required"`
	INIConfig              string   `form:"remote-ini-config" binding:"required_if=KeyDeviceType hardware Scenario key-required"`
	AllowedKeysSerial      []string `form:"allowed-keys-serial[]" binding:"required_if=Scenario key-required"`
	AllowedKeysIssuer      []string `form:"allowed-keys-issuer[]" binding:"required_if=Scenario key-required"`
	MailServerType         string   `form:"smtp-server-type"`
	MailServerOpts         string   `form:"mail-server-opts"`
	DNSNameOfficer         string   `form:"officer-dns"`
	DNSNameOfficerEnabled  string   `form:"officer-dns-enabled"`
	DNSNameCitizen         string   `form:"citizen-dns"`
	DNSNameCitizenEnabled  string   `form:"citizen-dns-enabled"`
	DNSNameKeycloak        string   `form:"keycloak-dns"`
	CIDROfficer            string   `form:"officer-cidr"`
	CIDRCitizen            string   `form:"citizen-cidr"`
	CIDRAdmin              string   `form:"admin-cidr"`
	CIDRChanged            string   `form:"cidr-changed"`
	Resources              string   `form:"resources"`
	SupAuthBrowserFlow     string   `form:"sup-auth-browser-flow"`
	SupAuthURL             string   `form:"sup-auth-url"`
	SupAuthWidgetHeight    string   `form:"sup-auth-widget-height"`
	SupAuthClientID        string   `form:"sup-auth-client-id"`
	SupAuthClientSecret    string   `form:"sup-auth-client-secret"`
	BackupScheduleEnabled  string   `form:"backup-schedule-enabled"`
	CronSchedule           string   `form:"cron-schedule"`
	CronScheduleDays       string   `form:"cron-schedule-days"`
	KeycloakCustomHostname string   `form:"keycloak-custom-hostname"`
}

func (r *registry) KeysRequired() bool {
	return r.Scenario == ScenarioKeyRequired
}

type allowedKeysConfig struct {
	AllowedKeys []allowedKey `yaml:"allowed-keys"`
}

type allowedKey struct {
	Issuer string `yaml:"issuer"`
	Serial string `yaml:"serial"`
}

type keyManagement struct {
	r               *registry
	vaultSecretPath string
}

func (k keyManagement) VaultSecretPath() string {
	return k.vaultSecretPath
}

func (k keyManagement) KeyDeviceType() string {
	return k.r.KeyDeviceType
}

func (k keyManagement) AllowedKeysIssuer() []string {
	return k.r.AllowedKeysIssuer
}

func (k keyManagement) AllowedKeysSerial() []string {
	return k.r.AllowedKeysSerial
}

func (k keyManagement) SignKeyIssuer() string {
	return k.r.SignKeyIssuer
}

func (k keyManagement) SignKeyPwd() string {
	return k.r.SignKeyPwd
}

func (k keyManagement) RemoteType() string {
	return k.r.RemoteType
}

func (k keyManagement) RemoteSerialNumber() string {
	return k.r.RemoteSerialNumber
}

func (k keyManagement) RemoteKeyPort() string {
	return k.r.RemoteKeyPort
}

func (k keyManagement) RemoteKeyHost() string {
	return k.r.RemoteKeyHost
}

func (k keyManagement) RemoteKeyPassword() string {
	return k.r.RemoteKeyPassword
}

func (k keyManagement) INIConfig() string {
	return k.r.INIConfig
}

func (k keyManagement) KeysRequired() bool {
	return k.r.KeysRequired()
}

func (k keyManagement) FilesSecretName() string {
	return fmt.Sprintf("digital-signature-ops-%s-data", k.r.Name)
}

func (k keyManagement) EnvVarsSecretName() string {
	return fmt.Sprintf("digital-signature-ops-%s-env-vars", k.r.Name)
}
