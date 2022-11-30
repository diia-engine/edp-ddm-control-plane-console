package cluster

type Values struct {
	Velero       Velero  `yaml:"velero" json:"velero"`
	Global       Global  `yaml:"global" json:"global"`
	Admins       []Admin `yaml:"administrators" json:"administrators"`
	OriginalYaml map[string]interface{}
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
