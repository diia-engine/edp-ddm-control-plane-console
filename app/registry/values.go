package registry

type Values struct {
	Administrators  []Admin                   `yaml:"administrators" json:"administrators"`
	ExternalSystems map[string]ExternalSystem `yaml:"external-systems" json:"external-systems"`
	Global          Global                    `yaml:"global" json:"global"`
	Trembita        Trembita                  `yaml:"trembita" json:"trembita"`
	OriginalYaml    map[string]interface{}    `yaml:"-" json:"-"`
}

type Notifications struct {
	Email NotificationsEmail `yaml:"email" json:"email"`
}

type NotificationsEmail struct {
	Type string `yaml:"type" json:"type"`
}

type ExternalSystem struct {
	URL      string            `yaml:"url" json:"url"`
	Type     string            `yaml:"type" json:"type"`
	Protocol string            `yaml:"protocol" json:"protocol"`
	Auth     map[string]string `yaml:"auth" json:"auth"`
	//Auth     TrembitaServiceAuth `yaml:"auth" json:"auth"`
}

type Global struct {
	WhiteListIP   WhiteListIP   `json:"whiteListIP" yaml:"whiteListIP"`
	Notifications Notifications `json:"notifications" yaml:"notifications"`
}

type WhiteListIP struct {
	AdminRoutes   string `yaml:"adminRoutes" json:"adminRoutes"`
	CitizenPortal string `yaml:"citizenPortal" json:"citizenPortal"`
	OfficerPortal string `yaml:"officerPortal" json:"officerPortal"`
}

type Trembita struct {
	Registries map[string]TrembitaRegistry `yaml:"registries" json:"registries"`
}

type TrembitaRegistry struct {
	UserID          string                  `yaml:"user-id" json:"user-id"`
	Type            string                  `yaml:"type" json:"type"`
	ProtocolVersion string                  `yaml:"protocol-version" json:"protocol-version"`
	URL             string                  `yaml:"url" json:"url"`
	Protocol        string                  `yaml:"protocol" json:"protocol"`
	Client          TrembitaRegistryClient  `yaml:"client" json:"client"`
	Service         TrembitaRegistryService `yaml:"service" json:"service"`
}

// `yaml:"" json:""`

type TrembitaRegistryClient struct {
	XRoadInstance string `yaml:"x-road-instance" json:"x-road-instance"`
	MemberClass   string `yaml:"member-class" json:"member-class"`
	MemberCode    string `yaml:"member-code" json:"member-code"`
	SubsystemCode string `yaml:"subsystem-code" json:"subsystem-code"`
}

type TrembitaRegistryService struct {
	TrembitaRegistryClient
	Auth map[string]string `yaml:"auth" json:"auth"`
	//Auth TrembitaServiceAuth `yaml:"auth" json:"auth"`
}

type TrembitaServiceAuth struct {
	Type   string `yaml:"type" json:"type"`
	Secret string `yaml:"secret" json:"secret"`
}
