export interface ExternalReg {
  Enabled: boolean;
  External: boolean;
  KeyValue: string;
  Name: string;
  StatusRegistration: string;
}
export interface RegistryTemplateVariables {
  openMergeRequests: boolean;
  registry: any;
  allowedToEdit: any;
  hasUpdate: any;
  regisrtyAdministrationComponents: any;
  registryOperationalComponents: any;
  platformAdministrationComponents: any;
  platformOperationalComponents: any;
  externalRegAvailableRegistriesJSON: any;
  publicApi: any;
  admins: any;
  citizenPortalHost: any;
  officerPortalHost: any;
  smtpType: any;
  officerCIDR: any;
  citizenCIDR: any;
  adminCIDR: any;
  values: any;
  externalRegs: ExternalReg[];
  branches: any;
  mergeRequests: any;
  created: any;
  valuesJson: any;
  gerritURL: string;
  jenkinsURL: string;
  mrAvailable: string;
}

export interface RegistryAdmin {
  email: string
  firstName: string
  lastName: string
  passwordVaultSecret: string
  passwordVaultSecretKey: string
  username: string
}

export interface PortalSettings {
  signWidget: {
    copyFromAuthWidget: boolean,
    url: string,
    height: number,
  }
}

export enum CitizenAuthType {
  widget = 'widget',
  registryIdGovUa = 'registry-id-gov-ua',
  platformIdGovUa = 'platform-id-gov-ua',
}

export interface CitizenAuthFlow {
  edrCheck: boolean
  authType: CitizenAuthType
  widget: {
    url: string
    height: number
  }
  registryIdGovUa: {
    url: string
    clientId: string
    clientSecret: string
  }
}

export interface RegistryWizardTemplateVariables {
  error?: any
  action: 'edit' | 'create'
  dnsManual: string
  hasUpdate: boolean
  hwINITemplateContent: string
  keycloakCustomHost: string
  keycloakHostname: string
  keycloakHostnames: string[]
  model: any
  page: string
  registry: any
  registryData: string
  registryValues: {
    administrators: RegistryAdmin[]
    digitalDocuments: {
      maxFileSize: string
      maxTotalFileSize: string
    }
    externalSystems: any
    global: any
    keycloak: {
      authFlows: any
      citizenAuthFlow: CitizenAuthFlow
      customHost: string
      identityProviders: any
      realms: Record<string, any>
    }
    portals: {
      citizen: PortalSettings
    }
    signWidget: {
      url: string
    }
    trembita: any
  }
  smtpConfig: string
  updateBranches: any[]
}
