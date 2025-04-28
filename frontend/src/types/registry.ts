import type { LANGUAGES } from "@/constants/registry";
import type { StoredKey } from './cluster';

export interface ExternalReg {
  Enabled: boolean;
  External: boolean;
  KeyValue: string;
  Name: string;
  StatusRegistration: string;
}

export interface ClusterDigitalSignature {
  data: unknown,
  env: unknown,
  keys: Record<string, StoredKey>
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
  values: {
    global: Global & { notifications: any},
    trembita: any,
    externalSystems: any,
  };
  externalRegs: ExternalReg[];
  branches: any;
  mergeRequests: any;
  created: any;
  valuesJson: any;
  gerritURL: string;
  jenkinsURL: string;
  mrAvailable: string;
  createReleaseAvailable: boolean;
  registryVersion: string;
  allowedToCreate: boolean;
  registries: any;
  page: string;
  gerritBranches: string[];
  platformVersion: string;
  previousVersion: string;
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
  singleIdentityEnabled: boolean,
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
    keyName: string
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
    global: Global,
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
          keyName: string
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
    trembita: any,
    'digital-signature': {
      data: unknown,
      env: unknown,
      keys: Record<string, StoredKey>
    }
  }
  smtpConfig: string
  updateBranches: any[]
  gerritBranches: string[]
  registryTemplateName: string
  platformStatusType: PlatformStatusType
  isPlatformAdmin: boolean
  defaultRegistryValues: any
  clusterValues: {
    global: {
      language: keyof typeof LANGUAGES
    }
    "digital-signature": ClusterDigitalSignature
  },
  clusterDigitalSignature: ClusterDigitalSignature
}

interface Global {
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
    language: keyof typeof LANGUAGES
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
  name: string;
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
    datasource?: {
      maxPoolSize: number;
    };
    replicas?: number;
    enabled?: boolean;
  };
}

export enum PORTALS {
  citizen = 'citizen', 
  officer = 'officer', 
  admin   = 'admin',
}
