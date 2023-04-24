package cluster

const (
	KeycloakValuesIndex = "keycloak"
)

type Values struct {
	Velero             Velero   `yaml:"velero" json:"velero"`
	Global             Global   `yaml:"global" json:"global"`
	Admins             []Admin  `yaml:"administrators" json:"administrators"`
	Keycloak           Keycloak `yaml:"keycloak" json:"keycloak"`
	OriginalYaml       map[string]interface{}
	IsRegistry194Lower bool `yaml:"-" json:"isRegistry194Lower"`
}

type Velero struct {
	Backup BackupSchedule `yaml:"backup" json:"backup"`
}

type Global struct {
	WhiteListIP WhiteListIP `yaml:"whiteListIP" json:"whiteListIP"`
}

type WhiteListIP struct {
	AdminRoutes string `json:"adminRoutes" yaml:"adminRoutes"`
}

type Keycloak struct {
	CustomHosts []CustomHost `json:"customHosts" yaml:"customHosts"`
}

type CustomHost struct {
	Host            string `json:"host" yaml:"host"`
	CertificatePath string `json:"certificatePath" yaml:"certificatePath"`
}
