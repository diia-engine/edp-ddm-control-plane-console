package registry

import (
	"encoding/json"
	"fmt"
)

const (
	DeploymentModeDevelopment = "development"
	GlobalValuesIndex         = "global"
	ResourcesIndex            = "registry"
	CrunchyPostgresIndex      = "crunchyPostgres"
	PortalsIndex              = "portals"
	WhiteListIPIndex          = "whiteListIP"
	NotificationsIndex        = "notifications"
)

type Values struct {
	Administrators   []Admin                   `yaml:"administrators" json:"administrators"`
	ExternalSystems  map[string]ExternalSystem `yaml:"external-systems" json:"externalSystems"`
	Global           Global                    `yaml:"global" json:"global"`
	Trembita         Trembita                  `yaml:"trembita" json:"trembita"`
	SignWidget       SignWidget                `yaml:"signWidget" json:"signWidget"`
	Keycloak         Keycloak                  `yaml:"keycloak" json:"keycloak"`
	Portals          Portals                   `yaml:"portals" json:"portals"`
	OriginalYaml     map[string]interface{}    `yaml:"-" json:"-"`
	DigitalDocuments DigitalDocuments          `yaml:"digitalDocuments" json:"digitalDocuments"`
	PublicApi        []PublicAPI               `yaml:"publicApi" json:"publicApi"`
	Griada           Griada                    `yaml:"griada" json:"griada"`
}

type Griada struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Ip      string `yaml:"ip" json:"ip"`
	Port    string `yaml:"port" json:"port"`
	Mask    string `yaml:"mask" json:"mask"`
}

type DigitalDocuments struct {
	MaxFileSize      string `yaml:"maxFileSize" json:"maxFileSize"`
	MaxTotalFileSize string `yaml:"maxTotalFileSize" json:"maxTotalFileSize"`
}

type Limits struct {
	Second int `json:"second,omitempty" yaml:"second,omitempty"`
	Minute int `json:"minute,omitempty" yaml:"minute,omitempty"`
	Hour   int `json:"hour,omitempty" yaml:"hour,omitempty"`
	Day    int `json:"day,omitempty" yaml:"day,omitempty"`
	Month  int `json:"month,omitempty" yaml:"month,omitempty"`
	Year   int `json:"year,omitempty" yaml:"year,omitempty"`
}

type PublicAPI struct {
	Name               string `yaml:"name" json:"name"`
	URL                string `yaml:"url" json:"url"`
	Limits             Limits `yaml:"limits,omitempty" json:"limits"`
	Enabled            bool   `yaml:"enabled" json:"enabled"`
	StatusRegistration string `yaml:"-"`
}

type CrunchyPostgres struct {
	CrunchyPostgresPostgresql CrunchyPostgresPostgresql `yaml:"postgresql" json:"postgresql"`
	StorageSize               string                    `yaml:"storageSize" json:"storageSize"`
	Backups                   interface{}               `yaml:"backups" json:"-"`
}

type CrunchyPostgresPostgresql struct {
	CrunchyPostgresPostgresqlParameters CrunchyPostgresPostgresqlParameters `yaml:"parameters" json:"parameters"`
}

type CrunchyPostgresPostgresqlParameters struct {
	MaxConnections int `yaml:"max_connections" json:"max_connections"`
}

type Portals struct {
	Citizen CitizenPortalSettings `yaml:"citizen" json:"citizen"`
	Officer OfficerPortalSettings `yaml:"officer" json:"officer"`
}
type OfficerPortalSettings struct {
	CustomDNS               CustomDNS  `yaml:"customDns" json:"customDns"`
	SignWidget              SignWidget `yaml:"signWidget" json:"signWidget"`
	IndividualAccessEnabled bool       `yaml:"individualAccessEnabled" json:"individualAccessEnabled"`
}
type CitizenPortalSettings struct {
	CustomDNS  CustomDNS  `yaml:"customDns" json:"customDns"`
	SignWidget SignWidget `yaml:"signWidget" json:"signWidget"`
}

type CustomDNS struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Host    string `yaml:"host" json:"host"`
}

type RegistryBackup struct {
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	Schedule      string `yaml:"schedule" json:"schedule"`
	ExpiresInDays int    `yaml:"expiresInDays" json:"expiresInDays"`
	OBC           OBC    `yaml:"obc" json:"obc"`
}

type OBC struct {
	CronExpression string `yaml:"cronExpression,omitempty" json:"cronExpression"`
	BackupBucket   string `yaml:"backupBucket,omitempty" json:"backupBucket"`
	Endpoint       string `yaml:"endpoint,omitempty" json:"endpoint"`
	Credentials    string `yaml:"credentials,omitempty" json:"credentials"`
}

type Keycloak struct {
	CustomHost        string                           `yaml:"customHost,omitempty" json:"customHost"`
	Realms            KeycloakRealms                   `yaml:"realms" json:"realms"`
	AuthFlows         KeycloakAuthFlows                `yaml:"authFlows" json:"authFlows"`
	CitizenAuthFlow   KeycloakAuthFlowsCitizenAuthFlow `yaml:"citizenAuthFlow" json:"citizenAuthFlow"`
	IdentityProviders KeycloakIdentityProviders        `yaml:"identityProviders" json:"identityProviders"`
}

type KeycloakIdentityProviders struct {
	IDGovUA KeycloakIdentityProvidersIDGovUA `yaml:"idGovUa" json:"idGovUa"`
}

type KeycloakIdentityProvidersIDGovUA struct {
	URL       string `yaml:"url" json:"url"`
	SecretKey string `yaml:"secretKey" json:"secretKey"`
	ClientID  string `yaml:"-" json:"clientId"`
}

type KeycloakAuthFlows struct {
	OfficerAuthFlow KeycloakAuthFlowsOfficerAuthFlow `yaml:"officerAuthFlow" json:"officerAuthFlow"`
}

type KeycloakAuthFlowsOfficerAuthFlow struct {
	WidgetHeight int `yaml:"widgetHeight" json:"widgetHeight"`
}

type KeycloakWidgetAuthSettings struct {
	Url    string `yaml:"url" json:"url,omitempty"`
	Height int    `yaml:"height" json:"height,omitempty"`
}

type KeycloakRegistryIdGovUaSettings struct {
	Url          string `yaml:"url" json:"url,omitempty"`
	ClientSecret string `yaml:"clientSecret" json:"clientSecret,omitempty"`
	ClientId     string `yaml:"clientId" json:"clientId,omitempty"`
}

type KeycloakAuthFlowsCitizenAuthFlow struct {
	EDRCheck        bool                            `yaml:"edrCheck" json:"edrCheck"`
	AuthType        string                          `yaml:"authType" json:"authType"`
	Widget          KeycloakWidgetAuthSettings      `yaml:"widget" json:"widget"`
	RegistryIdGovUa KeycloakRegistryIdGovUaSettings `yaml:"registryIdGovUa" json:"registryIdGovUa"`
}

type KeycloakRealms struct {
	OfficerPortal KeycloakRealmsOfficerPortal `yaml:"officerPortal" json:"officerPortal"`
}

type KeycloakRealmsOfficerPortal struct {
	BrowserFlow      string `yaml:"browserFlow" json:"browserFlow"`
	SelfRegistration bool   `yaml:"selfRegistration" json:"selfRegistration"`
}

type SignWidget struct {
	URL                string `yaml:"url" json:"url"`
	Height             int    `yaml:"height" json:"height"`
	CopyFromAuthWidget bool   `yaml:"copyFromAuthWidget" json:"copyFromAuthWidget"`
}

type ExternalEmailSettings struct {
	Type      string `yaml:"type" json:"type"`
	Host      string `yaml:"host" json:"host"`
	Port      int64  `yaml:"port" json:"port"`
	Address   string `yaml:"address" json:"address"`
	VaultPath string `yaml:"vaultPath" json:"vaultPath"`
	VaultKey  string `yaml:"vaultKey" json:"vaultKey"`
}

type Notifications struct {
	Email ExternalEmailSettings `yaml:"email" json:"email"`
}

type ExternalSystem struct {
	URL      string            `yaml:"url,omitempty" json:"url"`
	Type     string            `yaml:"type" json:"type"`
	Protocol string            `yaml:"protocol" json:"protocol"`
	Auth     map[string]string `yaml:"auth,omitempty" json:"auth"`
	Mock     bool              `yaml:"mock" json:"mock"`
}

func (e ExternalSystem) StrAuth() string {
	if e.Auth != nil {
		if t, ok := e.Auth["type"]; ok {
			return t
		}
	}

	return "-"
}

func (e ExternalSystem) FaStatus() string {
	if e.URL == "" {
		return "triangle-exclamation"
	}

	return "circle-check"
}

type Global struct {
	WhiteListIP      WhiteListIP            `json:"whiteListIP" yaml:"whiteListIP"`
	Notifications    Notifications          `json:"notifications" yaml:"notifications"`
	RegistryBackup   RegistryBackup         `yaml:"registryBackup" json:"registryBackup"`
	DeploymentMode   string                 `yaml:"deploymentMode" json:"deploymentMode"`
	CrunchyPostgres  CrunchyPostgres        `yaml:"crunchyPostgres" json:"crunchyPostgres"`
	Registry         map[string]interface{} `yaml:"registry" json:"registry"`
	ComputeResources ComputeResources       `yaml:"computeResources" json:"computeResources"`
	ExcludePortals   []string               `yaml:"excludePortals" json:"excludePortals"`
	GeoServerEnabled bool                   `yaml:"geoServerEnabled" json:"geoServerEnabled"`
}

type ComputeResources struct {
	InstanceCount                   json.Number `yaml:"instanceCount,omitempty" json:"instanceCount,omitempty"`
	AwsInstanceType                 string      `yaml:"awsInstanceType,omitempty" json:"awsInstanceType,omitempty"`
	AwsSpotInstance                 *bool       `yaml:"awsSpotInstance,omitempty" json:"awsSpotInstance,omitempty"`
	AwsSpotInstanceMaxPrice         string      `yaml:"awsSpotInstanceMaxPrice,omitempty" json:"awsSpotInstanceMaxPrice,omitempty"`
	AwsInstanceVolumeType           string      `yaml:"awsInstanceVolumeType,omitempty" json:"awsInstanceVolumeType,omitempty"`
	InstanceVolumeSize              json.Number `yaml:"instanceVolumeSize,omitempty" json:"instanceVolumeSize,omitempty"`
	VSphereInstanceCPUCount         json.Number `yaml:"vSphereInstanceCPUCount,omitempty" json:"vSphereInstanceCPUCount,omitempty"`
	VSphereInstanceCoresPerCPUCount json.Number `yaml:"vSphereInstanceCoresPerCPUCount,omitempty" json:"vSphereInstanceCoresPerCPUCount,omitempty"`
	VSphereInstanceRAMSize          json.Number `yaml:"vSphereInstanceRAMSize,omitempty" json:"vSphereInstanceRAMSize,omitempty"`
}

type WhiteListIP struct {
	AdminRoutes   string `yaml:"adminRoutes,omitempty" json:"adminRoutes"`
	CitizenPortal string `yaml:"citizenPortal,omitempty" json:"citizenPortal"`
	OfficerPortal string `yaml:"officerPortal,omitempty" json:"officerPortal"`
}

type Trembita struct {
	Registries map[string]TrembitaRegistry `yaml:"registries" json:"registries"`
	IPList     []string                    `yaml:"ipList" json:"ipList"`
}

type TrembitaRegistry struct {
	UserID          string                  `yaml:"user-id,omitempty" json:"userId,omitempty"`
	Type            string                  `yaml:"type,omitempty" json:"type,omitempty"`
	ProtocolVersion string                  `yaml:"protocol-version,omitempty" json:"protocolVersion,omitempty"`
	URL             string                  `yaml:"url,omitempty" json:"url,omitempty"`
	Protocol        string                  `yaml:"protocol,omitempty" json:"protocol,omitempty"`
	Client          TrembitaRegistryClient  `yaml:"client,omitempty" json:"client,omitempty"`
	Service         TrembitaRegistryService `yaml:"service,omitempty" json:"service,omitempty"`
	Auth            map[string]string       `yaml:"auth,omitempty" json:"auth,omitempty"`
	Mock            bool                    `yaml:"mock" json:"mock"`
}

func (t TrembitaRegistry) StrType() string {
	return fmt.Sprintf("type-%s", t.Type)
}

func (e ExternalSystem) StrType() string {
	return fmt.Sprintf("type-%s", e.Type)
}

func (t TrembitaRegistry) StrAuth() string {
	if t.Auth != nil {
		if t, ok := t.Auth["type"]; ok {
			return t
		}
	}

	return "-"
}

func (t TrembitaRegistry) FaStatus() string {
	if t.UserID == "" {
		return "triangle-exclamation"
	}

	return "circle-check"
}

type TrembitaRegistryClient struct {
	XRoadInstance string `yaml:"x-road-instance,omitempty" json:"xRoadInstance,omitempty"`
	MemberClass   string `yaml:"member-class,omitempty" json:"memberClass,omitempty"`
	MemberCode    string `yaml:"member-code,omitempty" json:"memberCode,omitempty"`
	SubsystemCode string `yaml:"subsystem-code,omitempty" json:"subsystemCode,omitempty"`
}

type TrembitaRegistryService struct {
	XRoadInstance  string `yaml:"x-road-instance,omitempty" json:"xRoadInstance,omitempty"`
	MemberClass    string `yaml:"member-class,omitempty" json:"memberClass,omitempty"`
	MemberCode     string `yaml:"member-code,omitempty" json:"memberCode,omitempty"`
	SubsystemCode  string `yaml:"subsystem-code,omitempty" json:"subsystemCode,omitempty"`
	ServiceCode    string `yaml:"service-code,omitempty" json:"serviceCode,omitempty"`
	ServiceVersion string `yaml:"service-version,omitempty" json:"serviceVersion,omitempty"`
}

type ClusterValues struct {
	Keycloak ClusterKeycloak `yaml:"keycloak" json:"keycloak"`
}

type ClusterKeycloak struct {
	CustomHosts []CustomHost `json:"customHosts" yaml:"customHosts"`
}

type CustomHost struct {
	Host string `json:"host" yaml:"host"`
}
