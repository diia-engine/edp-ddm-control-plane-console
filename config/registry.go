package config

import (
	"fmt"
	"strconv"
	"strings"

	"ddm-admin-console/app/registry"
)

func (cnf *Settings) RegistryConfig() (registry.Config, error) {
	if cnf.PreviousPlatfromVersion == "" {
		prevVersion, err := calculatePrevVersion(cnf.PlatformVersion)
		if err != nil {
			return registry.Config{}, fmt.Errorf("failed to calculate previous version, %w", err)
		}

		cnf.PreviousPlatfromVersion = prevVersion
	}

	return registry.Config{
		UsersNamespace:                  cnf.UsersNamespace,
		UsersRealm:                      cnf.UsersRealm,
		RegistryCodebaseLabels:          cnf.RegistryCodebaseLabels,
		EnableBranchProvisioners:        cnf.EnableBranchProvisioners,
		ClusterCodebaseName:             cnf.ClusterCodebaseName,
		ClusterRepo:                     cnf.ClusterRepo,
		Timezone:                        cnf.Timezone,
		GerritRegistryHost:              cnf.RegistryRepoHost,
		HardwareINITemplatePath:         cnf.RegistryHardwareKeyINITemplatePath,
		VaultRegistrySMTPPwdSecretKey:   cnf.VaultRegistrySMTPPwdSecretKey,
		VaultRegistrySecretPathTemplate: cnf.VaultRegistrySecretPathTemplate,
		VaultKVEngineName:               cnf.VaultKVEngineName,
		VaultCitizenSSLPath:             cnf.VaultCitizenSSLPath,
		VaultOfficerSSLPath:             cnf.VaultOfficerSSLPath,
		TempFolder:                      cnf.TempFolder,
		RegistryDNSManualPath:           cnf.RegistryDNSManualPath,
		DDMManualEDPComponent:           cnf.DDMManualEDPComponent,
		RegistryVersionFilter:           cnf.RegistryVersionFilter,
		KeycloakDefaultHostname:         cnf.KeycloakDefaultHostname,
		WiremockAddr:                    cnf.WiremockAddr,
		BackupBucketAccessKeyID:         cnf.BackupBucketAccessKeyID,
		BackupBucketSecretAccessKey:     cnf.BackupBucketSecretAccessKey,
		RegistryTemplateName:            cnf.RegistryTemplateName,
		CloudProvider:                   cnf.CloudProvider,
		Region:                          cnf.Region,
		CurrentVersion:                  cnf.PlatformVersion,
		PreviousVersion:                 cnf.PreviousPlatfromVersion,
	}, nil
}

func calculatePrevVersion(currVersion string) (string, error) {
	versionBits := strings.Split(currVersion, ".")

	if len(versionBits) != 3 {
		return "", fmt.Errorf("version %v is not in the correct format", currVersion)
	}

	majorVersion, err := strconv.Atoi(versionBits[0])
	if err != nil {
		return "", fmt.Errorf("failed to convert major version to int, %w", err)
	}

	minorVersion, err := strconv.Atoi(versionBits[1])
	if err != nil {
		return "", fmt.Errorf("failed to convert minor version to int, %w", err)
	}

	patchVersion, err := strconv.Atoi(versionBits[2])
	if err != nil {
		return "", fmt.Errorf("failed to convert patch version to int, %w", err)
	}

	if patchVersion != 0 {
		return fmt.Sprintf("%v.%v.%v", majorVersion, minorVersion, patchVersion-1), nil
	}

	return "", fmt.Errorf("patch version is 0, behaviour currently undefined")
}

func (s *Services) RegistryServices() registry.Services {
	return registry.Services{
		Codebase:     s.Codebase,
		K8S:          s.K8S,
		Keycloak:     s.Keycloak,
		Jenkins:      s.Jenkins,
		Gerrit:       s.Gerrit,
		EDPComponent: s.EDPComponent,
		Vault:        s.Vault,
		Cache:        s.Cache,
		Perms:        s.PermService,
		OpenShift:    s.OpenShift,
	}
}
