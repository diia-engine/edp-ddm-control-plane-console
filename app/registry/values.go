package registry

import "fmt"

type Values struct {
	Administrators  []Admin                   `yaml:"administrators" json:"administrators"`
	ExternalSystems map[string]ExternalSystem `yaml:"external-systems" json:"externalSystems"`
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
	UserID          string                  `yaml:"user-id" json:"userId"`
	Type            string                  `yaml:"type" json:"type"`
	ProtocolVersion string                  `yaml:"protocol-version" json:"protocolVersion"`
	URL             string                  `yaml:"url" json:"url"`
	Protocol        string                  `yaml:"protocol" json:"protocol"`
	Client          TrembitaRegistryClient  `yaml:"client" json:"client"`
	Service         TrembitaRegistryService `yaml:"service" json:"service"`
}

func (t TrembitaRegistry) StrType() string {
	return fmt.Sprintf("type-%s", t.Type)
}

func (e ExternalSystem) StrType() string {
	return fmt.Sprintf("type-%s", e.Type)
}

func (t TrembitaRegistry) Auth() string {
	if t.Service.Auth != nil {
		if t, ok := t.Service.Auth["type"]; ok {
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
	XRoadInstance string `yaml:"x-road-instance" json:"xRoadInstance"`
	MemberClass   string `yaml:"member-class" json:"memberClass"`
	MemberCode    string `yaml:"member-code" json:"memberCode"`
	SubsystemCode string `yaml:"subsystem-code" json:"subsystemCode"`
}

type TrembitaRegistryService struct {
	TrembitaRegistryClient
	Auth map[string]string `yaml:"auth" json:"auth"`
}
