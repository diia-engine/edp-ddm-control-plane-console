/* eslint-disable no-prototype-builtins */
import { defineComponent } from 'vue';
import $ from 'jquery';

export default defineComponent({
  expose: [
    'removeResourcesCatFromList',
    'decodeResourcesEnvVars',
    'wizardEditSubmit',
    'wizardNext',
    'wizardPrev',
    'selectWizardTab',
    'wizardTabChanged',
    'wizardEmptyValidation',
    'checkObjectFieldsEmpty',
    'wizardAdministratorsValidation',
    'wizardTemplateValidation',
    'wizardMailValidation',
    'addResourceCat',
    'addEnvVar',
    'removeEnvVar',
    'removeResourceCat',
    'encodeRegistryResources',
    'registryFormSubmit',
    'loadAdmins',
    'showAdminForm',
    'hideAdminForm',
    'deleteAdmin',
    'createAdmin',
    'changeTemplateProject',
  ],
  inject: { templateVariables : { from: 'TEMPLATE_VARIABLES' }, envVariables : { from: 'ENVIRONMENT_VARIABLES' }},
  mounted() {
    if(this.envVariables.region === "ua") {
      this.wizard.tabs.trembita =  {
        title: this.$t('domains.registry.editMixin.trembita'),
        validatorRef: 'trembitaTab',
        visible: true,
        disabled: true
      };
      this.wizard.tabs.keyVerification.validatorRef = "keyVerificationTab";
      this.wizard.tabs.key.validatorRef = "keyDataTab";
    }
    this.wizard.tabs.confirmation = {
      title: this.$t('domains.registry.editMixin.confirmation'),
      validated: true,
      visible: true,
      disabled: true
    };

    const childRefs = this.getChildrenRefs();
    if (childRefs.hasOwnProperty("smtpServerTypeSelected")) {
      // @ts-ignore
      let selectedSMTP = childRefs.smtpServerTypeSelected.value;
      if (selectedSMTP === "") {
        selectedSMTP = "platform-mail-server";
      }
      this.smtpServerType = selectedSMTP;
    }
    if (childRefs.hasOwnProperty("smtpEditConfig") && childRefs.smtpEditConfig.value !== "") {
      const smtpConfig = JSON.parse(childRefs.smtpEditConfig.value);
      if (smtpConfig["type"] === "external") {
        this.smtpServerType = "external-mail-server";
        this.externalSMTPOpts = smtpConfig;
        this.externalSMTPOpts["port"] = smtpConfig["port"].toString();
      }
      else {
        this.smtpServerType = "platform-mail-server";
      }
    }

    if (childRefs.hasOwnProperty("wizardAction")) {
      this.wizard.registryAction = childRefs.wizardAction.value;
      if (childRefs.wizardAction.value === "edit") {
        const registryData = JSON.parse(childRefs.registryData.value);
        this.wizard.tabs.general.registryName = registryData.name;
        this.wizard.tabs.confirmation.visible = false;
        this.adminsChanged = false;
        this.cidrChanged = false;
      }
    }

    if (this.templateVariables.registryValues) {
      this.registryValues = this.templateVariables.registryValues;
    }
    if (childRefs.hasOwnProperty("registryUpdate")) {
      this.wizard.tabs.update.visible = true;
    }

  },
  data() {
    return {
      disabled: false,
      registryValues: null,
      registryFormSubmitted: false,
      adminsValue: "",
      adminsChanged: true,
      adminPopupShow: false,
      admins: [],
      editAdmin: {
        firstName: "",
        lastName: "",
        email: "",
        tmpPassword: ""
      },
      requiredError: false,
      emailFormatError: false,
      passwordFormatError: false,
      adminsLoaded: false,
      adminsError: false,
      adminExistsError: false,
      smtpServerType: null,
      mailServerOpts: "",
      externalSMTPOpts: {
        host: "",
        port: "587",
        address: "",
        password: ""
      },
      wizard: {
        registryAction: "create",
        activeTab: "general",
        tabs: {
          general: {
            title: this.$t('domains.registry.editMixin.general'),
            validated: false,
            visible: true,
            validatorRef: 'generalTab',
            disabled: false,
          },
          administrators: {
            title: this.$t('domains.registry.editMixin.administrators'),
            validated: false,
             requiredError: false,
            validator: this.wizardAdministratorsValidation,
            visible: true,
            disabled: true
          },
          mail: {
            title: this.$t('domains.registry.editMixin.mailServer'),
            validated: false,
            beginValidation: false,
            validator: this.wizardMailValidation,
            visible: true,
            disabled: true
          },
          key: {
            title: this.$t('domains.registry.editMixin.keyInfo'),
            visible: true,
            disabled: true
          },
          keyVerification: {
            title: this.$t('domains.registry.editMixin.signatureVerificationInfo'),
            visible: true,
            disabled: true
          },
          parametersVirtualMachines: {
            title: this.$t('domains.registry.editMixin.parametersVirtualMachines'),
            visible: true,
            validatorRef: 'parametersVirtualMachinesTab',
            disabled: true
          },
          resources: {
            title: this.$t('domains.registry.editMixin.registryResources'),
            visible: true, 
            validatorRef: 'resourcesTab',
            disabled: true
          },
          dns: {
            title: "DNS",
            disabled: true,
            visible: true,
            validatorRef: 'dnsTab',
          },
          cidr: {
            title: this.$t('domains.registry.editMixin.accessRestrictions'),
            validated: true,
            visible: true,
            disabled: true
          },
          supplierAuthentication: {
            title: this.$t('domains.registry.editMixin.supplierAuthentication'),
            validatorRef: 'supplierAuthTab',
            visible: true,
            disabled: true
          },
          recipientAuthentication: {
            title: this.$t('domains.registry.editMixin.recipientAuthentication'),
            validatorRef: 'recipientAuthTab',
            visible: true,
            disabled: true
          },
          adminAuthentication: {
            title: this.$t('domains.registry.editMixin.officeAdministratorRegulations'),
            visible: true,
            disabled: true
          },
          geoDataSettings: {
            title: this.$t('domains.registry.editMixin.managementSubsystem'),
            visible: true,
            disabled: true
          },
          digitalDocuments: {
            title: this.$t('domains.registry.editMixin.digitalDocuments'),
            visible: true,
            validatorRef: 'digitalDocumentsTab',
            disabled: true
          },
          backupSchedule: {
            title: this.$t('domains.registry.editMixin.backup'),
            validatorRef: 'backupScheduleTab',
            visible: true,
            disabled: true
          }
        },
      },
    } as any;
  },
  methods: {
    goBack() {
      window.history.back();
    },
    getChildrenRefs() {
      const wizardRefs = this.$refs.wizard?.$refs || {};
      return {
        ...wizardRefs,
        ...(wizardRefs.smtpTab?.$refs || {}),
        ...(wizardRefs.keyTab?.$refs || {}),
        ...(wizardRefs.resourcesTab?.$refs || {}),
        ...(wizardRefs.dnsTab?.$refs || {}),
        ...(wizardRefs.cidrTab?.$refs || {}),
        ...(wizardRefs.supplierAuthTab?.$refs || {}),
        ...(wizardRefs.recipientAuthTab?.$refs || {}),
        ...(wizardRefs.backupScheduleTab?.$refs || {}),
        ...(wizardRefs.digitalDocumentsTab?.$refs || {}),
        ...(wizardRefs.generalTab?.$refs || {}),
        ...(wizardRefs.parametersVirtualMachinesTab?.$refs || {}),
      };
    },
    wizardEditSubmit(event: any) {
      const childRefs = this.getChildrenRefs();
      const tab = this.wizard.tabs[this.wizard.activeTab];

      this.callValidator(tab).then(() => {
        this.disabled = true;
        this.registryFormSubmit(event);
        this.$nextTick(() => {
          window.localStorage.setItem("mr-scroll", "true");
          childRefs.registryWizardForm.submit();
        });
      }).catch(() => {});
    },
    wizardNext() {
      const tabKeys = Object.keys(this.wizard.tabs);
      for (let i = 0; i < tabKeys.length; i++) {
        if (tabKeys[i] === this.wizard.activeTab) {
          const tab = this.wizard.tabs[tabKeys[i]];
          const wizard = this.wizard;
          this.callValidator(tab).then(function () {
            wizard.activeTab = tabKeys[i + 1];
            wizard.tabs[wizard.activeTab].disabled = false;
          });
          break;
        }
      }
    },
    wizardPrev() {
      const tabKeys = Object.keys(this.wizard.tabs);
      for (let i = 0; i < tabKeys.length; i++) {
        if (tabKeys[i] === this.wizard.activeTab) {
          this.wizard.activeTab = tabKeys[i - 1];
          this.wizard.tabs[tabKeys[i]].disabled = true;
          break;
        }
      }
    },
    callValidator(tab: any) {
      if ('validator' in tab) {
        return tab.validator(tab);
      }

      if ('validatorRef' in tab) {
        return this.$refs.wizard.$refs[tab.validatorRef].validator(tab);
      }

      return this.wizardEmptyValidation(tab);
    },
    selectWizardTab(tabName: string, action: string) {
      if (action === 'edit') {
        const tab = this.wizard.tabs[this.wizard.activeTab];
        this.callValidator(tab).then(() => {
          this.wizard.activeTab = tabName;
        });
        return;
      }
      let disabled = false;
      this.wizard.activeTab = tabName;
      for (const key in this.wizard.tabs) {
        if (key === tabName) {
          disabled = true;
          this.wizard.tabs[key].disabled = false;
        } else {
          this.wizard.tabs[key].disabled = disabled;
        }
      }
    },
    wizardEmptyValidation() {
      return new Promise < void  > ((resolve) => {
        resolve();
      });
    },
    wizardAdministratorsValidation(tab: any) {
      const admins = this.admins;
      return new Promise < void  > ((resolve) => {
        tab.requiredError = false;
        tab.validated = false;
        if (admins.length === 0) {
          tab.requiredError = true;
          return;
        }
        tab.validated = true;
        resolve();
      });
    },
    wizardMailValidation(tab: any) {
      return new Promise < void  > ((resolve) => {
        tab.validated = false;
        if (this.smtpServerType === "platform-mail-server") {
          tab.validated = true;
          resolve();
          return;
        }
        tab.beginValidation = true;
        for (const key in this.externalSMTPOpts) {
          if (this.externalSMTPOpts[key] === "") {
            return;
          }
        }
        tab.beginValidation = false;
        tab.validated = true;
        resolve();
      });
    },
    registryFormSubmit(e: any) {
      if (this.registryFormSubmitted && e) {
        e.preventDefault();
        return;
      }
      this.mailServerOpts = JSON.stringify(this.externalSMTPOpts);
      this.registryFormSubmitted = true;
    },
    loadAdmins(admins: string) {
      if (!this.adminsLoaded) {
        if (admins && admins !== "") {
          this.admins = JSON.parse(admins);
          this.adminsValue = JSON.stringify(this.admins);
          this.adminsLoaded = true;
          this.adminsChanged = false;
        }
      }
    },
    showAdminForm() {
      this.emailFormatError = false;
      this.passwordFormatError = false;
      this.adminExistsError = false;
      this.requiredError = false;
      this.adminPopupShow = true;
      $("body").css("overflow", "hidden");
    },
    hideAdminForm() {
      this.adminPopupShow = false;
      $("body").css("overflow", "scroll");
    },
    deleteAdmin(e: any) {
      e.preventDefault();
      const email = e.currentTarget.getAttribute("email");
      for (const v in this.admins) {
        if (this.admins[v].email === email) {
          this.admins.splice(v, 1);
          break;
        }
      }
      this.adminsValue = JSON.stringify(this.admins);
      this.adminsChanged = true;
    },
    createAdmin() {
      this.requiredError = false;
      this.emailFormatError = false;
      this.passwordFormatError = false;
      for (const v in this.editAdmin) {
        if (this.editAdmin[v] === "") {
          this.requiredError = true;
          return;
        }
      }
      if (!String(this.editAdmin.email)
        .toLowerCase()
        .match(/^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/)) {
        this.emailFormatError = true;
        return;
      }

      if (!String(this.editAdmin.tmpPassword)
        .match(/^(?=.*[0-9])(?=.*[a-z])(?=.*[A-Z])(?=.*[@#$%^&+=])(?=\S+$)[a-zA-Z0-9@#$%^&+=]{10,}$/)) {
        this.passwordFormatError = true;
        return;
      }

      if (this.admins === null) {
        this.admins = [];
      }
      for (let i = 0; i < this.admins.length; i++) {
        if (this.admins[i].email.trim() === this.editAdmin.email.trim()) {
          this.adminExistsError = true;
          return;
        }
      }
      $("body").css("overflow", "scroll");
      this.adminPopupShow = false;
      this.admins.push({
        email: this.editAdmin.email,
        firstName: this.editAdmin.firstName,
        lastName: this.editAdmin.lastName,
        tmpPassword: this.editAdmin.tmpPassword
      });
      this.editAdmin = {
        firstName: "",
        lastName: "",
        email: "",
        tmpPassword: ""
      };
      this.adminsValue = JSON.stringify(this.admins);
      this.adminsChanged = true;
      this.wizard.tabs.administrators.validated = false;
      this.wizard.tabs.administrators.requiredError = false;
    },
  },
});
