package cluster

import "ddm-admin-console/app/registry"

type Values struct {
	Velero               Velero                    `yaml:"velero,omitempty" json:"velero"`
	Global               Global                    `yaml:"global" json:"global"`
	Admins               []Admin                   `yaml:"administrators" json:"administrators"`
	Keycloak             Keycloak                  `yaml:"keycloak,omitempty" json:"keycloak"`
	CdPipelineName       string                    `yaml:"cdPipelineName"`
	CdPipelineStageName  string                    `yaml:"cdPipelineStageName"`
	SourceCatalogVersion float32                   `yaml:"source_catalog_version"`
	DigitalSignature     registry.DigitalSignature `yaml:"digital-signature"`
}

type Velero struct {
	Backup BackupSchedule `yaml:"backup" json:"backup"`
}

type Global struct {
	WhiteListIP      WhiteListIP `yaml:"whiteListIP" json:"whiteListIP"`
	DeploymentMode   string      `yaml:"deploymentMode" json:"deploymentMode"`
	DemoRegistryName string      `json:"demoRegistryName" yaml:"demoRegistryName"`
	PlatformName     string      `json:"platformName" yaml:"platformName"`
	LogosPath        string      `json:"logosPath" yaml:"logosPath"`
	Language         string      `form:"language" json:"language"`
	Region           string      `json:"region" yaml:"region"`
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
