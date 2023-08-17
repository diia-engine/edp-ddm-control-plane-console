/* eslint-disable no-prototype-builtins */
import { defineComponent } from 'vue';
import axios from 'axios';
import $ from 'jquery';

export default defineComponent({
  expose: [
    'removeResourcesCatFromList',
    'decodeResourcesEnvVars',
    'wizardEditSubmit',
    'wizardNext',
    'dnsPreloadDataFromValues',
    'wizardPrev',
    'selectWizardTab',
    'wizardTabChanged',
    'wizardEmptyValidation',
    'wizardGeneralValidation',
    'wizardDNSValidation',
    'wizardCheckPEMFiles',
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
    'prepareDNSConfig',
    'loadAdmins',
    'showAdminForm',
    'hideAdminForm',
    'deleteAdmin',
    'createAdmin',
    'changeTemplateProject',
  ],
  inject: { templateVariables : { from: 'TEMPLATE_VARIABLES' } },
  mounted() {
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
        this.wizard.tabs.template.visible = false;
        this.wizard.tabs.confirmation.visible = false;
        this.adminsChanged = false;
        this.cidrChanged = false;
      }
    }

    if (this.templateVariables.registryValues) {
      this.registryValues = this.templateVariables.registryValues;
      this.dnsPreloadDataFromValues();
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
            title: "Загальні",
            validated: false,
            registryName: "",
            requiredError: false,
            existsError: false,
            formatError: false,
            validator: this.wizardGeneralValidation,
            visible: true,
          },
          administrators: { title: "Адміністратори", validated: false, requiredError: false,
            validator: this.wizardAdministratorsValidation, visible: true, },
          template: {
            title: "Шаблон реєстру",
            validatorRef: 'templateTab',
            visible: true
          },
          mail: { title: "Поштовий сервер", validated: false, beginValidation: false, validator: this.wizardMailValidation, visible: true, },
          key: {
            title: "Дані про ключ",
            visible: true,
            validatorRef: 'keyDataTab',
          },
          keyVerification: {
            title: 'Дані для перевірки підписів',
            visible: true,
            validatorRef: 'keyVerificationTab',
          },
          resources: {
            title: "Ресурси реєстру", visible: true, validatorRef: 'resourcesTab',
          },
          dns: { title: "DNS", validated: false, data: { officer: "", citizen: "", }, beginValidation: false, formatError: { officer: false, citizen: false, }, requiredError: { officer: false, citizen: false, }, typeError: { officer: false, citizen: false, }, editVisible: { officer: false, citizen: false }, validator: this.wizardDNSValidation, visible: true, preloadValues: {} },
          cidr: { title: "Обмеження доступу", validated: true, visible: true, },
          supplierAuthentication: {
            title: "Автентифікація надавачів послуг",
            validatorRef: 'supplierAuthTab',
            visible: true,
          },
          recipientAuthentication: {
            title: 'Автентифікація отримувачів послуг',
            validatorRef: 'recipientAuthTab',
            visible: true,
          },
          digitalDocuments: {
            title: "Цифрові документи",
            visible: true,
            validatorRef: 'digitalDocumentsTab',
          },
          backupSchedule: {
            title: "Резервне копіювання",
            validatorRef: 'backupScheduleTab',
            visible: true,
          },
          trembita: {
            title: "ШБО Трембіта",
            validatorRef: 'trembitaTab',
            visible: true,
          },
          confirmation: { title: "Підтвердження", validated: true, visible: true, }
        },
      },
    } as any;
  },
  methods: {
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
      };
    },
    wizardEditSubmit(event: any) {
      const childRefs = this.getChildrenRefs();
      const tab = this.wizard.tabs[this.wizard.activeTab];

      this.callValidator(tab).then(() => {
        this.disabled = true;
        this.registryFormSubmit(event);
        this.$nextTick(() => {
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
          });
          break;
        }
      }
    },
    dnsPreloadDataFromValues() {
      if (this.registryValues && this.registryValues.hasOwnProperty("portals")) {
        for (const p in this.registryValues.portals) {
          if (this.registryValues.portals[p].hasOwnProperty("customDns")) {
            this.wizard.tabs.dns.editVisible[p] = this.registryValues.portals[p].customDns.enabled;
            this.wizard.tabs.dns.data[p] = this.registryValues.portals[p].customDns.host;
            this.wizard.tabs.dns.preloadValues[p] = this.registryValues.portals[p].customDns.host;
          }
        }
      }
    },
    wizardPrev() {
      const tabKeys = Object.keys(this.wizard.tabs);
      for (let i = 0; i < tabKeys.length; i++) {
        if (tabKeys[i] === this.wizard.activeTab) {
          const tab = this.wizard.tabs[tabKeys[i]];
          const wizard = this.wizard;
          this.callValidator(tab).then(function () {
            wizard.activeTab = tabKeys[i - 1];
          });
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
    selectWizardTab(tabName: string, e: any) {
      e.preventDefault();
      const tab = this.wizard.tabs[this.wizard.activeTab];
      const wizard = this.wizard;
      this.callValidator(tab).then(function () {
        if (wizard.registryAction === "create") {
          for (const k in wizard.tabs) {
            if (!wizard.tabs[k].validated) {
              return;
            }
            if (k === tabName) {
              break;
            }
          }
        }
        wizard.activeTab = tabName;
      });
    },
    wizardEmptyValidation(tab: string) {
      return new Promise < void  > ((resolve) => {
        resolve();
      });
    },
    wizardGeneralValidation(tab: any) {
      return new Promise < void  > ((resolve) => {
        tab.requiredError = false;
        tab.formatError = false;
        tab.existsError = false;
        tab.validated = false;
        if (tab.registryName === "") {
          tab.requiredError = true;
          return;
        }
        if (tab.registryName.length < 3) {
          tab.formatError = true;
          return;
        }
        if (!/^[a-z0-9]([-a-z0-9]*[a-z0-9])?([a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/.test(tab.registryName)) {
          tab.formatError = true;
          return;
        }
        if (this.wizard.registryAction === "edit") {
          tab.validated = true;
          resolve();
          return;
        }
        axios.get(`/admin/registry/check/${tab.registryName}`)
          .then(function (response) {
            tab.existsError = true;
          })
          .catch(function (error) {
            tab.validated = true;
            resolve();
          });
      });
    },
    wizardDNSValidation(tab: any) {
      const childRefs = this.getChildrenRefs();
      return new Promise < void  > ((resolve) => {
        tab.beginValidation = true;
        tab.validated = false;
        let validationFailed = false;
        const filesToCheck = [];
        for (const k in this.wizard.tabs.dns.data) {
          this.wizard.tabs.dns.formatError[k] = false;
          this.wizard.tabs.dns.requiredError[k] = false;
          this.wizard.tabs.dns.typeError[k] = false;
          const fileInput = childRefs[`${k}SSL`];
          if (this.wizard.tabs.dns.data[k] !== "") {
            if (!/^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9-]*[A-Za-z0-9])$/.test(this.wizard.tabs.dns.data[k])) {
              this.wizard.tabs.dns.formatError[k] = true;
              validationFailed = true;
            }
            if (fileInput.files.length === 0 &&
              (!this.wizard.tabs.dns.preloadValues.hasOwnProperty(k) || this.wizard.tabs.dns.preloadValues[k] !== this.wizard.tabs.dns.data[k])) {
              this.wizard.tabs.dns.requiredError[k] = true;
              validationFailed = true;
            }
            else if (fileInput.files.length > 0) {
              filesToCheck.push({ name: k, file: fileInput.files[0] });
            }
          }
          else if (fileInput.files.length > 0) {
            this.wizard.tabs.dns.formatError[k] = true;
            validationFailed = true;
          }
        }
        if (validationFailed) {
          return;
        }
        if (filesToCheck.length > 0) {
          this.wizardCheckPEMFiles(filesToCheck, resolve, tab);
          return;
        }
        tab.validated = true;
        tab.beginValidation = false;
        resolve();
      });
    },
    wizardCheckPEMFiles(filesToCheck: Array<any>, resolve: () => void, tab: any) {
      if (filesToCheck.length === 0) {
        tab.validated = true;
        tab.beginValidation = false;
        resolve();
        return;
      }
      const f = filesToCheck.pop();
      const formData = new FormData();
      formData.append("file", f?.file || '');

      axios.post("/admin/registry/check-pem", formData, {
        headers: {
          "Content-Type": "multipart/form-data"
        }
      }).then((rsp) => {
        this.wizardCheckPEMFiles(filesToCheck, resolve, tab);
      }).catch((error) => {
        this.wizard.tabs.dns.typeError[f.name] = true;
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
      this.prepareDNSConfig();
      this.mailServerOpts = JSON.stringify(this.externalSMTPOpts);
      this.registryFormSubmitted = true;
    },
    prepareDNSConfig() {
      const childRefs = this.getChildrenRefs();
      for (const k in this.wizard.tabs.dns.data) {
        const fileInput = childRefs[`${k}SSL`];
        if (this.wizard.tabs.dns.editVisible[k] &&
          this.wizard.tabs.dns.data[k] === this.wizard.tabs.dns.preloadValues[k] &&
          fileInput.files.length === 0) {
          this.wizard.tabs.dns.data[k] = "";
        }
      }
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
