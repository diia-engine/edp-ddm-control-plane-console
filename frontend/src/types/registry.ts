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
  registryAdministrationComponents: any;
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
  registryVersion: string;
  allowedToCreate: boolean;
  registries: any;
  page: string;
  gerritBranches: string[];
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
  },
  customDns?: {
    enabled: boolean,
    host: string
  }
}
export interface OfficerPortalSettings extends PortalSettings {
  individualAccessEnabled: boolean,
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
    global: {
      registryBackup: {
        enabled: boolean;
        schedule: string;
        expiresInDays: string;
        obc: {
          cronExpression: string;
          backupBucket: string;
          endpoint: string;
        }
      }
      computeResources: ComputeResources,
      deploymentMode: string,
      registry: any,
      excludePortals: string[],
      geoServerEnabled: boolean,
    }
    keycloak: {
      authFlows: {
        officerAuthFlow: {
          widgetHeight: number
        }
      }
      citizenAuthFlow: CitizenAuthFlow
      customHost: string
      identityProviders: {
        idGovUa: {
          clientId: string
          url: string
          secretKey: string
        }
      }
      realms: {
        officerPortal: {
          browserFlow: string
          selfRegistration: boolean
        }
      }
    }
    portals: {
      citizen: PortalSettings
      officer: OfficerPortalSettings,
    }
    signWidget: {
      url: string
    }
    trembita: any
  }
  smtpConfig: string
  updateBranches: any[]
  gerritBranches: string[]
  registryTemplateName: string
  platformStatusType: PlatformStatusType
  isPlatformAdmin: boolean
  defaultRegistryValues: any
}

export interface PublicApiLimits {
  second?: string,
  minute?: string,
  hour?: string,
  day?: string,
  month?: string,
  year?: string,
}

export enum PlatformStatusType {
 AWS = 'AWS',
 VSphere = 'VSphere'
}

export type ComputeResources = {
  instanceCount: number;
  awsInstanceType: string;
  awsSpotInstance: boolean;
  awsSpotInstanceMaxPrice: number;
  awsInstanceVolumeType: string;
  instanceVolumeSize: number;
  vSphereInstanceCPUCount: number;
  vSphereInstanceCoresPerCPUCount: number;
  vSphereInstanceRAMSize: number;
}
export enum OfficerAuthType {
  widget = 'dso-officer-auth-flow',
  registryIdGovUa = 'id-gov-ua-officer-redirector',
}

export interface RegistryResource {
  name: REGISTRY_COMPONENTS;
  config: {
    istio: {
      sidecar: {
        enabled: boolean;
        resources: {
          requests: {
            cpu: string;
            memory: string;
          };
          limits: {
            cpu: string;
            memory: string;
          };
        };
      };
    };
    container: {
      resources: {
        requests: {
          cpu: string;
          memory: string;
        };
        limits: {
          cpu: string;
          memory: string;
        };
      };
      envVars: Array<{
        name: string;
        value: string;
      }>;
    };
    hpa?: {
      enabled: boolean;
      maxReplicas: number;
      minReplicas: number;
    };
    datasource?: {
      maxPoolSize: number;
    };
    replicas?: number;
    enabled?: boolean;
  };
}

export enum REGISTRY_COMPONENTS {
  bpms = 'bpms',
  digitalDocumentService = 'digitalDocumentService',
  digitalSignatureOps = 'digitalSignatureOps',
  geoServer = 'geoServer',
  kafkaApi = 'kafkaApi',
  kong = 'kong',
  redis = 'redis',
  restApi = 'restApi',
  sentinel = 'sentinel',
  soapApi = 'soapApi',
  userProcessManagement = 'userProcessManagement',
  userTaskManagement = 'userTaskManagement',
}

export enum PORTALS {
  citizen = 'citizen', 
  officer = 'officer', 
  admin   = 'admin',
}
