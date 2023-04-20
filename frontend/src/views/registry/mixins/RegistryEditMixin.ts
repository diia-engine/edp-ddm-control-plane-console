/* eslint-disable no-prototype-builtins */
import { defineComponent } from 'vue';
import axios from 'axios';
import $ from 'jquery';
import Mustache from 'mustache';
import { parseCronExpression } from 'cron-schedule';

export default defineComponent({
  expose: [
    'wizardCronExpressionChange',
    'wizardSupAuthFlowChange',
    'loadRegistryValues',
    'removeResourcesCatFromList',
    'decodeResourcesEnvVars',
    'preloadRegistryResources',
    'mergeResource',
    'isObject',
    'mergeDeep',
    'wizardBackupScheduleChange',
    'wizardDNSEditVisibleChange',
    'wizardEditSubmit',
    'wizardNext',
    'dnsPreloadDataFromValues',
    'wizardPrev',
    'selectWizardTab',
    'wizardTabChanged',
    'wizardEmptyValidation',
    'wizardGeneralValidation',
    'wizardBackupScheduleValidation',
    'wizardSupAuthValidation',
    'wizardDNSValidation',
    'wizardCheckPEMFiles',
    'checkObjectFieldsEmpty',
    'wizardAdministratorsValidation',
    'wizardTemplateValidation',
    'wizardMailValidation',
    'keyFormValidation',
    'wizardKeyValidation',
    'renderINITemplate',
    'wizardKeyHardwareDataChanged',
    'wizardRemoveAllowedKey',
    'wizardAddAllowedKey',
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

    if (childRefs.hasOwnProperty("resourcesEditConfig") && childRefs.resourcesEditConfig.value !== "") {
      const resourcesConfig = JSON.parse(childRefs.resourcesEditConfig.value);
      this.preloadRegistryResources(resourcesConfig);
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
      this.loadRegistryValues();
      this.dnsPreloadDataFromValues();
      this.wizardCronExpressionChange();
    }
    if (childRefs.hasOwnProperty("registryUpdate")) {
      this.wizard.tabs.update.visible = true;
    }
  },
  data() {
    return {
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
          key: { title: "Дані про ключ", validated: false, deviceType: "file", beginValidation: false, hardwareData: {
              remoteType: "криптомод. ІІТ Гряда-301",
              remoteKeyPWD: "",
              remoteCaName: "",
              remoteCaHost: "",
              remoteCaPort: "",
              remoteSerialNumber: "",
              remoteKeyPort: "",
              remoteKeyHost: "",
              remoteKeyMask: "",
              iniConfig: "",
            }, fileData: {
              signKeyIssuer: "",
              signKeyPWD: "",
            }, allowedKeys: [{ issuer: "", serial: "", removable: false }], caCertRequired: false, caJSONRequired: false, key6Required: false, validator: this.wizardKeyValidation, visible: true, changed: false,
          },
          resources: {
            title: "Ресурси реєстру", visible: true,
          },
          dns: { title: "DNS", validated: false, data: { officer: "", citizen: "", }, beginValidation: false, formatError: { officer: false, citizen: false, }, requiredError: { officer: false, citizen: false, }, typeError: { officer: false, citizen: false, }, editVisible: { officer: false, citizen: false }, validator: this.wizardDNSValidation, visible: true, preloadValues: {} },
          cidr: { title: "Обмеження доступу", validated: true, visible: true, },
          supplierAuthentication: {
            title: "Автентифікація надавачів послуг",
            validated: false,
            validator: this.wizardSupAuthValidation,
            beginValidation: false,
            visible: true,
            dsoDefaultURL: "https://eu.iit.com.ua/sign-widget/v20200922/",
            data: {
              authType: "dso-officer-auth-flow",
              url: "https://eu.iit.com.ua/sign-widget/v20200922/",
              widgetHeight: "720",
              clientId: "",
              secret: "",
            },
            urlValidationFailed: false,
            heightIsNotNumber: false,
            selfRegistrationEnabled: false,
          },
          recipientAuthentication: {
            title: 'Автентифікація отримувачів послуг',
            validated: true,
            beginValidation:false,
            visible: true,
            data: {
              edrCheckEnabled: true
            }
          },
          backupSchedule: {
            title: "Резервне копіювання",
            validated: false,
            beginValidation: false,
            visible: true,
            validator: this.wizardBackupScheduleValidation,
            enabled: false,
            nextLaunches: false,
            wrongCronFormat: false,
            wrongDaysFormat: false,
            data: {
              cronSchedule: "",
              days: "",
            },
            nextDates: [],
          },
          trembita: {
            title: "ШБО Трембіта",
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
      };
    },
    wizardCronExpressionChange(e: any) {
      const bs = this.wizard.tabs.backupSchedule;
      if (bs.data.cronSchedule === "") {
        bs.nextLaunches = false;
        bs.wrongCronFormat = false;
        return;
      }
      try {
        const cron = parseCronExpression(bs.data.cronSchedule);
        bs.nextDates = [];
        let dt = new Date();
        for (let i = 0; i < 3; i++) {
          const next = cron.getNextDate(dt);
          bs.nextDates.push(`${next.toLocaleDateString("uk")} ${next.toLocaleTimeString("uk")}`);
          dt = next;
        }
        bs.nextLaunches = true;
        bs.wrongCronFormat = false;
      }
      catch (e: any) {
        bs.nextLaunches = false;
        bs.wrongCronFormat = true;
      }
    },
    wizardSupAuthFlowChange() {
      this.wizard.tabs.supplierAuthentication.validated = false;
      this.wizard.tabs.supplierAuthentication.beginValidation = false;
      const registryValues = this.registryValues;
      if (this.wizard.tabs.supplierAuthentication.data.authType === "dso-officer-auth-flow") {
        if (registryValues && registryValues.signWidget.url !== "") {
          this.wizard.tabs.supplierAuthentication.data.url = registryValues.signWidget.url;
        }
        else {
          this.wizard.tabs.supplierAuthentication.data.url = this.wizard.tabs.supplierAuthentication.dsoDefaultURL;
        }
        if (registryValues && registryValues.keycloak.authFlows.officerAuthFlow.widgetHeight !== 0) {
          this.wizard.tabs.supplierAuthentication.data.widgetHeight =
            registryValues.keycloak.authFlows.officerAuthFlow.widgetHeight;
        }
      }
      else {
        if (registryValues && registryValues.keycloak.identityProviders.idGovUa.url !== "") {
          this.wizard.tabs.supplierAuthentication.data.url = registryValues.keycloak.identityProviders.idGovUa.url;
        }
        else {
          this.wizard.tabs.supplierAuthentication.data.url = "";
        }
        if (registryValues && registryValues.keycloak.identityProviders.idGovUa.clientId !== "") {
          this.wizard.tabs.supplierAuthentication.data.clientId = registryValues.keycloak.identityProviders.idGovUa.clientId;
          this.wizard.tabs.supplierAuthentication.data.secret = "*****";
        }
      }
    },
    loadRegistryValues() {
      try {
        if (this.registryValues.keycloak.realms.officerPortal.browserFlow !== "") {
          this.wizard.tabs.supplierAuthentication.data.authType = this.registryValues.keycloak.realms.officerPortal.browserFlow;
        }
        if (this.wizard.tabs.supplierAuthentication.data.authType === "dso-officer-auth-flow") {
          this.wizard.tabs.supplierAuthentication.data.widgetHeight =
            this.registryValues.keycloak.authFlows.officerAuthFlow.widgetHeight;
          this.wizard.tabs.supplierAuthentication.data.url = this.registryValues.signWidget.url;
        }
        else {
          this.wizard.tabs.supplierAuthentication.data.url =
            this.registryValues.keycloak.identityProviders.idGovUa.url;
          this.wizard.tabs.supplierAuthentication.data.clientId = this.registryValues.keycloak.identityProviders.idGovUa.clientId;
          this.wizard.tabs.supplierAuthentication.data.secret = "*****";
        }

        this.wizard.tabs.recipientAuthentication.data.edrCheckEnabled = this.registryValues.keycloak.citizenAuthFlow.edrCheck;
      } catch (e: any) {
        console.log(e);
      }
      try {
        this.wizard.tabs.backupSchedule.enabled = this.registryValues.global.registryBackup.enabled;
        this.wizard.tabs.backupSchedule.data.cronSchedule = this.registryValues.global.registryBackup.schedule;
        this.wizard.tabs.backupSchedule.data.days = this.registryValues.global.registryBackup.expiresInDays;
      }
      catch (e: any) {
        console.log(e);
      }
      if (this.registryValues.keycloak.customHosts === null) {
        this.registryValues.keycloak.customHosts = [];
      }
    },
    wizardBackupScheduleChange(e: any) {
      console.log(e);
      //TODO: remove
    },
    wizardDNSEditVisibleChange(name: string, event: any) {
      console.log(name, event);
      //TODO: remove
    },
    wizardEditSubmit(event: any) {
      const childRefs = this.getChildrenRefs();
      const tab = this.wizard.tabs[this.wizard.activeTab];

      this.callValidator(tab).then(() => {
        this.registryFormSubmit(event);
        this.$nextTick(() => {
          childRefs.registryWizardForm.submit();
        });
      });
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
    wizardTabChanged(tabName: string) {
      this.wizard.tabs[tabName].changed = true;
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
    wizardBackupScheduleValidation(tab: any) {
      return new Promise < void  > ((resolve) => {
        const bs = this.wizard.tabs.backupSchedule;
        bs.data.cronSchedule = bs.data.cronSchedule.trim();
        if (!bs.enabled) {
          resolve();
          return;
        }
        tab.beginValidation = true;
        tab.validated = false;
        bs.wrongCronFormat = false;
        bs.wrongDaysFormat = false;
        if (bs.data.cronSchedule !== "") {
          try {
            parseCronExpression(bs.data.cronSchedule);
          }
          catch (e: any) {
            bs.nextLaunches = false;
            bs.wrongCronFormat = true;
          }
        }
        const days = parseInt(bs.data.days);
        if (bs.data.days !== "" && (!/^[0-9]+$/.test(bs.data.days) || isNaN(days) || days <= 0)) {
          bs.wrongDaysFormat = true;
        }
        if (bs.wrongDaysFormat || bs.wrongCronFormat || bs.data.cronSchedule === "" || bs.data.days === "") {
          return;
        }
        tab.validated = true;
        tab.beginValidation = false;
        resolve();
      });
    },
    wizardSupAuthValidation(tab: any) {
      return new Promise < void  > ((resolve) => {
        tab.beginValidation = true;
        tab.validated = false;
        tab.urlValidationFailed = false;
        tab.heightIsNotNumber = false;
        if (!/^(http(s)?:\/\/.)[-a-zA-Z0-9@:%._+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_+.~#?&//=,]*)$/.test(this.wizard.tabs.supplierAuthentication.data.url)) {
          tab.urlValidationFailed = true;
          return;
        }
        if (this.wizard.tabs.supplierAuthentication.data.authType === "dso-officer-auth-flow") {
          if (this.wizard.tabs.supplierAuthentication.data.widgetHeight === "") {
            return;
          }
          if (!/^[0-9]+$/.test(this.wizard.tabs.supplierAuthentication.data.widgetHeight)) {
            this.wizard.tabs.supplierAuthentication.heightIsNotNumber = true;
            return;
          }
        }
        if (this.wizard.tabs.supplierAuthentication.data.authType === "id-gov-ua-officer-redirector" &&
          (this.wizard.tabs.supplierAuthentication.data.clientId === "" ||
            this.wizard.tabs.supplierAuthentication.data.secret === "")) {
          return;
        }
        tab.validated = true;
        tab.beginValidation = false;
        resolve();
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
    keyFormValidation(tab: any, resolve: () => void) {
      const childRefs = this.getChildrenRefs();
      if (this.wizard.registryAction === "edit" && !this.wizard.tabs.key.changed) {
        resolve();
        return true;
      }
      this.renderINITemplate();
      tab.validated = false;
      tab.beginValidation = true;
      this.wizard.tabs.key.caCertRequired = false;
      this.wizard.tabs.key.caJSONRequired = false;
      this.wizard.tabs.key.key6Required = false;
      let validationFailed = false;
      if (childRefs.keyCaCert.files.length === 0) {
        this.wizard.tabs.key.caCertRequired = true;
        validationFailed = true;
      }
      if (childRefs.keyCaJSON.files.length === 0) {
        this.wizard.tabs.key.caJSONRequired = true;
        validationFailed = true;
      }
      for (let i = 0; i < this.wizard.tabs.key.allowedKeys.length; i++) {
        if (this.wizard.tabs.key.allowedKeys[i].issuer === "" ||
          this.wizard.tabs.key.allowedKeys[i].serial === "") {
          validationFailed = true;
        }
      }
      if (this.wizard.tabs.key.deviceType === "hardware") {
        for (const key in this.wizard.tabs.key.hardwareData) {
          if (this.wizard.tabs.key.hardwareData[key] === "") {
            validationFailed = true;
          }
        }
      }
      else {
        for (const key in this.wizard.tabs.key.fileData) {
          if (this.wizard.tabs.key.hardwareData[key] === "") {
            validationFailed = true;
          }
        }
        if (childRefs.key6.files.length === 0) {
          this.wizard.tabs.key.key6Required = true;
          validationFailed = true;
        }
      }
      if (validationFailed) {
        return false;
      }
      tab.beginValidation = false;
      tab.validated = true;
      resolve();
      return true;
    },
    wizardKeyValidation(tab: any) {
      return new Promise < void  > ((resolve) => {
        this.keyFormValidation(tab, resolve);
      });
    },
    renderINITemplate() {
      const iniTemplate = document.getElementById("ini-template")?.innerHTML;
      this.wizard.tabs.key.hardwareData.iniConfig = Mustache.render(iniTemplate || '', {
        "CA_NAME": this.wizard.tabs.key.hardwareData.remoteCaName,
        "CA_HOST": this.wizard.tabs.key.hardwareData.remoteCaHost,
        "CA_PORT": this.wizard.tabs.key.hardwareData.remoteCaPort,
        "KEY_SN": this.wizard.tabs.key.hardwareData.remoteSerialNumber,
        "KEY_HOST": this.wizard.tabs.key.hardwareData.remoteKeyHost,
        "KEY_ADDRESS_MASK": this.wizard.tabs.key.hardwareData.remoteKeyMask,
      }).trim();
    },
    wizardKeyHardwareDataChanged(e: any) {
      this.renderINITemplate();
      this.wizard.tabs.key.changed = true;
    },
    wizardRemoveAllowedKey(item: unknown) {
      const searchIdx = this.wizard.tabs.key.allowedKeys.indexOf(item);
      if (searchIdx !== -1) {
        this.wizard.tabs.key.allowedKeys.splice(searchIdx, 1);
      }
    },
    wizardAddAllowedKey() {
      this.wizard.tabs.key.allowedKeys.push({ issuer: "", serial: "", removable: true });
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
