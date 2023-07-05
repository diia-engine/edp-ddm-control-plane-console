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
