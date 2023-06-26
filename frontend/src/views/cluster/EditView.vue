<script setup lang="ts">
import { inject } from 'vue';
import { parseCronExpression } from 'cron-schedule';

interface PlatformUpdateTemplateVariables {
    updateBranches: any;
    errorsMap: any;
    hasUpdate: any;
    values: any;
    admins: any;
    backupSchedule: any;
    cidrConfig: any;
    keycloakHostname: any;
    dnsManual: string;
}
const variables = inject('TEMPLATE_VARIABLES') as PlatformUpdateTemplateVariables;
const updateBranches = variables?.updateBranches;
const errorsMap = variables?.errorsMap;
const hasUpdate = variables?.hasUpdate;
const values = variables?.values;
const adminsData = variables?.admins;
const backupSchedule = variables?.backupSchedule;
const cidrConfig = variables?.cidrConfig;
const keycloakHostname = variables?.keycloakHostname;
const dnsManual = variables?.dnsManual;

</script>
<script lang="ts">
import $ from 'jquery';
import Mustache from 'mustache';
import axios from 'axios';
import PlatformUpdateBlock from './components/PlatformUpdateBlock.vue';
import AdministratorsBlock from './components/AdministratorsBlock.vue';
import BackupBlock from './components/BackupBlock.vue';
import CidrBlock from './components/CidrBlock.vue';
import ClusterKeyBlock from './components/ClusterKeyBlock.vue';
import ClusterKeycloakBlock from './components/ClusterKeycloakBlock.vue';
import AdministratorModal from '@/components/AdministratorModal.vue';
import CidrModal from './components/ClusterCidrModal.vue';

// eslint-disable-next-line no-useless-escape
const hostnameRegex = /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/;
export default {
    expose: ['keyFormValidation'],
    components: {
        PlatformUpdateBlock,
        AdministratorsBlock,
        BackupBlock,
        CidrBlock,
        ClusterKeyBlock,
        ClusterKeycloakBlock,
        AdministratorModal,
        CidrModal,
    },
    data() {
        return {
            registryValues: null,
            wizard: {
                registryAction: 'create',
                activeTab: 'general',
                tabs: {
                    general: {
                        title: 'Загальні', validated: false, registryName: '', requiredError: false, existsError: false,
                        formatError: false, validator: this.wizardGeneralValidation,
                        visible: true,
                    },
                    administrators: {
                        title: 'Адміністратори', validated: false, requiredError: false,
                        validator: this.wizardAdministratorsValidation, visible: true,
                    },
                    template: {
                        title: 'Шаблон реєстру', validated: false, registryTemplate: '', registryBranch: '',
                        branches: [], projectBranches: {}, templateRequiredError: false, branchRequiredError: false,
                        validator: this.wizardTemplateValidation, visible: true,
                    },
                    mail: {
                        title: 'Поштовий сервер', validated: false, beginValidation: false,
                        validator: this.wizardMailValidation, visible: true,
                    },
                    key: {
                        title: 'Дані про ключ', validated: false, deviceType: 'file', beginValidation: false,
                        hardwareData: {
                            remoteType: 'криптомод. ІІТ Гряда-301',
                            remoteKeyPWD: '',
                            remoteCaName: '',
                            remoteCaHost: '',
                            remoteCaPort: '',
                            remoteSerialNumber: '',
                            remoteKeyPort: '',
                            remoteKeyHost: '',
                            remoteKeyMask: '',
                            iniConfig: '',
                        },
                        fileData: {
                            signKeyIssuer: '',
                            signKeyPWD: '',
                        },
                        allowedKeys: [{ issuer: '', serial: '', removable: false }],
                        caCertRequired: false,
                        caJSONRequired: false,
                        key6Required: false,
                        validator: this.wizardKeyValidation, visible: true,
                        changed: false,
                    },
                    resources: {
                        title: 'Ресурси реєстру', validated: false, beginValidation: false,
                        validator: this.wizardEmptyValidation, visible: true,
                    },
                    dns: {
                        title: 'DNS', validated: false, data: { officer: '', citizen: '', },
                        beginValidation: false, formatError: { officer: false, citizen: false, },
                        requiredError: { officer: false, citizen: false, },
                        typeError: { officer: false, citizen: false, },
                        editVisible: { officer: false, citizen: false },
                        validator: this.wizardDNSValidation, visible: true,
                        preloadValues: {}
                    },
                    cidr: { title: 'Обмеження доступу', validated: true, visible: true, validator: this.wizardEmptyValidation, },
                    supplierAuthentication: {
                        title: 'Автентифікація надавачів послуг', validated: false, validator: this.wizardSupAuthValidation,
                        beginValidation: false, visible: true,
                        dsoDefaultURL: 'https://eu.iit.com.ua/sign-widget/v20200922/',
                        data: {
                            authType: 'dso-officer-auth-flow',
                            url: 'https://eu.iit.com.ua/sign-widget/v20200922/',
                            widgetHeight: '720',
                            clientId: '',
                            secret: '',
                        },
                        urlValidationFailed: false,
                        heightIsNotNumber: false,
                    },
                    recipientAuthentication: {
                        title: 'Автентифікація отримувачів послуг',
                        validated: true,
                        beginValidation: false,
                        validator: this.wizardEmptyValidation,
                        visible: true,
                        data: {
                            edrCheckEnabled: true
                        }
                    },
                    backupSchedule: {
                        title: 'Резервне копіювання', validated: false, beginValidation: false, visible: true,
                        validator: this.wizardBackupScheduleValidation, enabled: false,
                        nextLaunches: false,
                        wrongCronFormat: false,
                        wrongDaysFormat: false,
                        data: {
                            cronSchedule: '',
                            days: '',
                        },
                        nextDates: [],

                    },
                    confirmation: { title: 'Підтвердження', validated: true, visible: true, validator: this.wizardEmptyValidation, }
                },
            },
            clusterSettings: {
                activeTab: 'administrators',
                tabs: [
                    {
                        key: 'administrators',
                        title: 'Адміністратори'
                    },
                    {
                        key: 'backup',
                        title: 'Резервне копіювання'
                    },
                    {
                        key: 'allowedCIDR',
                        title: 'Дозволені CIDR'
                    },
                    {
                        key: 'dataAboutKey',
                        title: 'Дані про ключ'
                    },
                    {
                      key: 'dataAboutKeyVerification',
                      title: 'Дані для перевірки підписів'
                    },
                    {
                        key: 'keycloakDNS',
                        title: 'Keycloak DNS'
                    },
                ],
                keycloak: {
                    editDisabled: false,
                    deleteHostname: '',
                    editHostname: '',
                    editCertPath: '',
                    submitInput: '',
                    hostname: '',
                    formShow: false,
                    hostnameError: '',
                    fileSelected: false,
                    pemError: '',
                    existHostname: '',
                },
            },
            admins: [],
            adminsValue: '',
            adminsLoaded: false,
            adminsChanged: true,
            emailFormatError: false,
            adminExistsError: false,
            requiredError: false,
            adminPopupShow: false,
            usernameFormatError: false,
            editAdmin: {
                firstName: "",
                lastName: "",
                email: "",
                tmpPassword: ""
            },
            citizenCIDR: [],
            citizenCIDRValue: { value: '' },
            officerCIDR: [],
            officerCIDRValue: { value: '' },
            adminCIDR: [],
            adminCIDRValue: { value: '' },
            registryFormSubmitted: false,
            mailServerOpts: '',
            externalSMTPOpts: {
                host: '',
                port: '587',
                address: '',
                password: ''
            },
            registryResources: {
                encoded: '',
                cat: '',
                cats: [
                    'kong',
                    'bpms',
                    'digitalSignatureOps',
                    'userTaskManagement',
                    'userProcessManagement',
                    'digitalDocumentService',
                    'restApi',
                    'kafkaApi',
                    'soapApi',
                ],
                addedCats: [],
            },
            cidrChanged: true,
            cidrPopupShow: false,
            editCidr: "",
            currentCIDR: [],
            currentCIDRValue: '',
            cidrFormatError: false,
            backdropShow: false,
        };
    },
    methods: {
        selectClusterSettingsTab(tabName: string, e: any) {
            e.preventDefault();
            this.clusterSettings.activeTab = tabName;
        },
        loadRegistryValues() {
            try {
                if ((this.registryValues as any).keycloak.realms?.officerPortal.browserFlow !== '') {
                    this.wizard.tabs.supplierAuthentication.data.authType = (this.registryValues as any).keycloak.realms?.officerPortal.browserFlow;
                }

                if (this.wizard.tabs.supplierAuthentication.data.authType === 'dso-officer-auth-flow') {
                    this.wizard.tabs.supplierAuthentication.data.widgetHeight =
                        (this.registryValues as any).keycloak.authFlows.officerAuthFlow.widgetHeight;
                    this.wizard.tabs.supplierAuthentication.data.url = (this.registryValues as any).signWidget.url;
                } else {
                    this.wizard.tabs.supplierAuthentication.data.url =
                        (this.registryValues as any).keycloak.identityProviders?.idGovUa.url;
                    this.wizard.tabs.supplierAuthentication.data.clientId = (this.registryValues as any).keycloak.identityProviders?.idGovUa.clientId;
                    this.wizard.tabs.supplierAuthentication.data.secret = '*****';
                }

                this.wizard.tabs.recipientAuthentication.data.edrCheckEnabled = (this.registryValues as any).keycloak.citizenAuthFlow?.edrCheck;
            } catch (e) {
                console.log(e);
            }

            try {
                this.wizard.tabs.backupSchedule.enabled = (this.registryValues as any).global.registryBackup?.enabled;
                this.wizard.tabs.backupSchedule.data.cronSchedule = (this.registryValues as any).global.registryBackup?.schedule;
                this.wizard.tabs.backupSchedule.data.days = (this.registryValues as any).global.registryBackup?.expiresInDays;
            } catch (e) {
                console.log(e);
            }

            if ((this.registryValues as any).keycloak.customHosts === null) {
                (this.registryValues as any).keycloak.customHosts = [];
            }
        },
        dnsPreloadDataFromValues() {
            // eslint-disable-next-line no-prototype-builtins
            if (this.registryValues && (this.registryValues as any).hasOwnProperty('portals')) {
                for (let p in (this.registryValues as any).portals) {
                    // eslint-disable-next-line no-prototype-builtins
                    if ((this.registryValues as any).portals[p].hasOwnProperty('customDns')) {
                        (this.wizard as any).tabs.dns.editVisible[p] = (this.registryValues as any).portals[p].customDns.enabled;
                        (this.wizard as any).tabs.dns.data[p] = (this.registryValues as any).portals[p].customDns.host;
                        (this.wizard as any).tabs.dns.preloadValues[p] = (this.registryValues as any).portals[p].customDns.host;
                    }
                }
            }
        },
        wizardCronExpressionChange() {
            let bs = this.wizard.tabs.backupSchedule;
            if (bs.data.cronSchedule === '') {
                bs.nextLaunches = false;
                bs.wrongCronFormat = false;
                return;
            }

            try {
                const cron = parseCronExpression(bs.data.cronSchedule);
                bs.nextDates = [];
                let dt = new Date();
                for (let i = 0; i < 3; i++) {
                    let next = cron.getNextDate(dt);
                    bs.nextDates.push(`${next.toLocaleDateString('uk')} ${next.toLocaleTimeString('uk')}` as never);
                    dt = next;
                }
                bs.nextLaunches = true;
                bs.wrongCronFormat = false;
            } catch (e) {
                bs.nextLaunches = false;
                bs.wrongCronFormat = true;
            }
        },
        loadAdmins(admins: any) {
            if (!this.adminsLoaded) {
                if (admins !== "") {
                    this.admins = admins;
                    this.adminsValue = JSON.stringify(this.admins);
                    this.adminsLoaded = true;
                    this.adminsChanged = false;
                }
            }
        },
        deleteAdmin(adminEmail: string) {
            let email = adminEmail;

            for (let v in this.admins) {
                if ((this.admins as any)[v].email === email) {
                    this.admins.splice(+v, 1);
                    break;
                }
            }
            this.adminsValue = JSON.stringify(this.admins);
            this.adminsChanged = true;
        },
        showAdminForm() {
            this.emailFormatError = false;
            this.adminExistsError = false;
            this.requiredError = false;
            this.adminPopupShow = true;
            $("body").css("overflow", "hidden");
        },
        createAdmin() {
            this.requiredError = false;
            this.emailFormatError = false;


            for (let v in this.editAdmin) {
                if ((this.editAdmin as any)[v] === "") {
                    this.requiredError = true;
                    return;
                }
            }

            if (!String(this.editAdmin.email)
                .toLowerCase()
                .match(
                    /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
                )) {
                this.emailFormatError = true;
                return;
            }

            if (this.admins === null) {
                this.admins = [];
            }

            for (let i = 0; i < this.admins.length; i++) {
                if ((this.admins as any)[i].email.trim() === this.editAdmin.email.trim()) {
                    this.adminExistsError = true;
                    return;
                }
            }

            $("body").css("overflow", "scroll");
            this.adminPopupShow = false;

            (this.admins as any).push({
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
        hideAdminForm() {
            this.adminPopupShow = false;
            $("body").css("overflow", "scroll");
        },
        isObject(item: any) {
            return (item && typeof item === 'object' && !Array.isArray(item));
        },
        cleanEmptyProperties(obj: any) {
            if (this.isObject(obj)) {
                for (const key in obj) {
                    if (this.isObject(obj[key])) {
                        this.cleanEmptyProperties(obj[key]);

                        if (Object.keys(obj[key]).length === 0) {
                            delete obj[key];
                        }
                    } else if (obj[key] === '') {
                        delete obj[key];
                    }

                }
            }
        },
        encodeRegistryResources() {
            let prepare = {} as any;

            this.registryResources.addedCats.forEach(function (el) {
                let cloneEL = JSON.parse(JSON.stringify(el));

                let envVars = {} as any;
                cloneEL.config.container.envVars.forEach(function (el: any) {
                    envVars[el.name] = el.value;
                });
                cloneEL.config.container.envVars = envVars;

                prepare[cloneEL.name] = {
                    istio: cloneEL.config.istio,
                    container: cloneEL.config.container,
                };
            });

            this.cleanEmptyProperties(prepare);

            this.registryResources.encoded = JSON.stringify(prepare);
        },
        prepareDNSConfig() {
            for (let k in this.wizard.tabs.dns.data) {
                let fileInput = this.$refs[`${k}SSL`];

                if ((this.wizard as any).tabs.dns.editVisible[k] &&
                    (this.wizard as any).tabs.dns.data[k] === (this.wizard as any).tabs.dns.preloadValues[k] &&
                    (fileInput as any).files.length === 0) {
                    (this.wizard as any).tabs.dns.data[k] = '';
                }
            }
        },
        registryFormSubmit() {
            if (this.registryFormSubmitted) {
                return;
            }

            this.encodeRegistryResources();
            this.prepareDNSConfig();
            this.mailServerOpts = JSON.stringify(this.externalSMTPOpts);
            this.citizenCIDRValue.value = JSON.stringify(this.citizenCIDR);
            this.officerCIDRValue.value = JSON.stringify(this.officerCIDR);
            this.adminCIDRValue.value = JSON.stringify(this.adminCIDR);

            this.registryFormSubmitted = true;
        },
        deleteCIDR(c: any, cidr: any, value: any) {
            for (let v in cidr) {
                if (cidr[v] === c) {
                    cidr.splice(v, 1);
                    break;
                }
            }

            value = JSON.stringify(cidr);
            this.cidrChanged = true;
        },
        showCIDRForm(cidr: any, value: any) {
            this.cidrPopupShow = true;
            $("body").css("overflow", "hidden");

            this.editCidr = '';
            this.currentCIDR = cidr;
            this.currentCIDRValue = value;
            this.cidrFormatError = false;
        },
        hideCIDRForm() {
            this.cidrPopupShow = false;
            $("body").css("overflow", "scroll");
        },
        createCIDR() {
            let cidrVal = String(this.editCidr).toLowerCase();
            if (cidrVal !== '0.0.0.0/0' && !cidrVal.
                match(/^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:\/[0-2]\d|\/3[0-2])?$/)) {
                this.cidrFormatError = true;
                return;
            }

            this.currentCIDR.push(this.editCidr as never);
            (this.currentCIDRValue as any).value = JSON.stringify(this.currentCIDR);
            this.hideCIDRForm();
            this.cidrChanged = true;
        },
        keyFormValidation(tab: any, resolve: any) {
            if (this.wizard.registryAction === 'edit' && !this.wizard.tabs.key.changed) {
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
            const ref = (this.$refs as any).clusterKeyRef.$refs.keyFormRef.$refs;

            if (ref.keyCaCert.files.length === 0) {
                this.wizard.tabs.key.caCertRequired = true;
                validationFailed = true;
            }

            if (ref.keyCaJSON.files.length === 0) {
                this.wizard.tabs.key.caJSONRequired = true;
                validationFailed = true;
            }

            for (let i = 0; i < this.wizard.tabs.key.allowedKeys.length; i++) {
                if (this.wizard.tabs.key.allowedKeys[i].issuer === '' ||
                    this.wizard.tabs.key.allowedKeys[i].serial === '') {
                    validationFailed = true;
                }
            }

            if (this.wizard.tabs.key.deviceType === 'hardware') {
                for (let key in this.wizard.tabs.key.hardwareData) {
                    if ((this.wizard as any).tabs.key.hardwareData[key] === '') {
                        validationFailed = true;
                    }
                }
            } else {
                for (let key in this.wizard.tabs.key.fileData) {
                    if ((this.wizard as any).tabs.key.hardwareData[key] === '') {
                        validationFailed = true;
                    }
                }

                if (ref.key6.files.length === 0) {
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
        renderINITemplate() {
            let iniTemplate = document.getElementById('ini-template')?.innerHTML as string;
            this.wizard.tabs.key.hardwareData.iniConfig = Mustache.render(iniTemplate, {
                "CA_NAME": this.wizard.tabs.key.hardwareData.remoteCaName,
                "CA_HOST": this.wizard.tabs.key.hardwareData.remoteCaHost,
                "CA_PORT": this.wizard.tabs.key.hardwareData.remoteCaPort,
                "KEY_SN": this.wizard.tabs.key.hardwareData.remoteSerialNumber,
                "KEY_HOST": this.wizard.tabs.key.hardwareData.remoteKeyHost,
                "KEY_ADDRESS_MASK": this.wizard.tabs.key.hardwareData.remoteKeyMask,
            }).trim();
        },
        wizardTabChanged(tabName: string) {
            (this.wizard as any).tabs[tabName].changed = true;
        },
        wizardKeyHardwareDataChanged() {
            this.renderINITemplate();
            this.wizard.tabs.key.changed = true;
        },
        wizardAddAllowedKey() {
            this.wizard.tabs.key.allowedKeys.push({ issuer: '', serial: '', removable: true });
        },
        wizardRemoveAllowedKey(item: any) {
            let searchIdx = this.wizard.tabs.key.allowedKeys.indexOf(item);
            if (searchIdx !== -1) {
                this.wizard.tabs.key.allowedKeys.splice(
                    searchIdx, 1);
            }
        },
        submitKeycloakDNSForm() {
            this.clusterSettings.keycloak.submitInput = JSON.stringify((this.$refs.registryValues as any).keycloak.customHosts);
        },
        clusterKeycloakDNSCustomHosts() {
            if (this.registryValues === null) {
                return [];
            }

            return (this.$refs.registryValues as any).keycloak?.customHosts;
        },
        editClusterKeycloakDNSHost(hostname: string, certificatePath: string) {
            this.clusterSettings.keycloak.editHostname = hostname;
            (this.clusterSettings as any).keycloak.editCertificatePath = certificatePath;
            this.clusterSettings.keycloak.hostname = hostname;

            this.backdropShow = true;
            this.clusterSettings.keycloak.formShow = true;
            this.clusterSettings.keycloak.fileSelected = true;
            this.clusterSettings.keycloak.editDisabled = true;

            axios.get(`/admin/cluster/check-keycloak-hostname/${hostname}`)
                .then(() => {
                    this.clusterSettings.keycloak.editDisabled = false;
                });
        },
        checkClusterDeleteKeycloakDNS(hostname: string) {
            this.backdropShow = true;
            axios.get(`/admin/cluster/check-keycloak-hostname/${hostname}`)
                .then(() => {
                    this.clusterSettings.keycloak.deleteHostname = hostname;
                })
                .catch(() => {
                    this.clusterSettings.keycloak.existHostname = hostname;
                });
        },
        showClusterKeycloakDNSForm() {
            this.backdropShow = true;
            this.clusterSettings.keycloak.editDisabled = false;
            this.clusterSettings.keycloak.editHostname = '';
            this.clusterSettings.keycloak.formShow = true;
            this.clusterSettings.keycloak.fileSelected = false;
            this.clusterSettings.keycloak.hostname = '';
        },
        hideCheckClusterDeleteKeycloakDNS() {
            this.clusterSettings.keycloak.deleteHostname = '';
            this.backdropShow = false;
        },
        deleteClusterKeycloakDNS(hostname: string) {
            (this.registryValues as any).keycloak.customHosts = (this.registryValues as any).keycloak.customHosts.filter(
                (i: { host: string; }) => i.host !== hostname);
            this.clusterSettings.keycloak.deleteHostname = '';
            this.backdropShow = false;
        },
        hideClusterCheckKeycloakDNS() {
            this.backdropShow = false;
            this.clusterSettings.keycloak.existHostname = '';
        },
        hideClusterKeycloakDNSForm() {
            this.backdropShow = false;
            this.clusterSettings.keycloak.formShow = false;
        },
        addClusterKeycloakDNS() {
            this.clusterSettings.keycloak.hostnameError = '';
            this.clusterSettings.keycloak.pemError = '';

            if (this.clusterSettings.keycloak.hostname === '') {
                this.clusterSettings.keycloak.hostnameError = 'Поле обов’язкове для заповнення';
                return;
            }

            if (!hostnameRegex.test(this.clusterSettings.keycloak.hostname)) {
                this.clusterSettings.keycloak.hostnameError = 'Перевірте формат поля';
                return;
            }

            for (let i = 0; i < (this.registryValues as any).keycloak.customHosts.length; i++) {
                if ((this.registryValues as any).keycloak.customHosts[i].host === this.clusterSettings.keycloak.hostname &&
                    this.clusterSettings.keycloak.hostname !== this.clusterSettings.keycloak.editHostname) {
                    this.clusterSettings.keycloak.hostnameError = 'Така назва вже використовується';
                    return;
                }
            }

            if (this.clusterSettings.keycloak.fileSelected === false) {
                this.clusterSettings.keycloak.pemError = 'Поле обов’язкове для заповнення';
                return;
            }

            const ref = (this.$refs as any).clusterKeycloakRef.$refs.clusterKeycloakDNS;

            if (ref.files.length > 0) {
                let formData = new FormData();
                formData.append("file", ref.files[0]);
                formData.append("hostname", this.clusterSettings.keycloak.hostname);
                axios.post('/admin/cluster/upload-pem-dns', formData, {
                    headers: {
                        'Content-Type': 'multipart/form-data'
                    }
                }).then((rsp) => {
                    if (this.clusterSettings.keycloak.editHostname === '') {
                        (this.registryValues as any).keycloak.customHosts.push({
                            host: this.clusterSettings.keycloak.hostname,
                            certificatePath: rsp.data,
                        });

                    } else {
                        for (let i = 0; i < (this.registryValues as any).keycloak.customHosts.length; i++) {
                            if ((this.registryValues as any).keycloak.customHosts[i].host === this.clusterSettings.keycloak.editHostname) {
                                (this.registryValues as any).keycloak.customHosts[i].host = this.clusterSettings.keycloak.hostname;
                                (this.registryValues as any).keycloak.customHosts[i].certificatePath = rsp.data;
                            }
                        }
                    }

                    this.backdropShow = false;
                    this.clusterSettings.keycloak.formShow = false;
                    this.clusterSettings.keycloak.editHostname = '';
                }).catch((error) => {
                    this.clusterSettings.keycloak.pemError = this.localePEMError(error.response.data);
                });
            } else {
                for (let i = 0; i < (this.registryValues as any).keycloak.customHosts.length; i++) {
                    if ((this.registryValues as any).keycloak.customHosts[i].host === this.clusterSettings.keycloak.editHostname) {
                        (this.registryValues as any).keycloak.customHosts[i].host = this.clusterSettings.keycloak.hostname;
                    }
                }

                this.backdropShow = false;
                this.clusterSettings.keycloak.formShow = false;
                this.clusterSettings.keycloak.editHostname = '';
            }
        },
        localePEMError(message: string) {
            const messages: { [key: string]: string } = {
                "found in PEM file": "Перевірте формат файла",
                "certificate has expired or is not yet valid": "Сертифікат застарілий",
                "certificate is valid for": "Сертифікат не відповідає доменному імені",
            };

            // eslint-disable-next-line no-prototype-builtins
            if (messages.hasOwnProperty(message)) {
                return messages[message];
            }

            for (let m in messages) {
                if (message.indexOf(m) !== -1) {
                    return messages[m];
                }
            }

            return message;
        },
        resetClusterKeycloakDNSForm() {
            this.clusterSettings.keycloak.fileSelected = false;
            this.clusterSettings.keycloak.pemError = '';
        },
        clusterKeycloakDNSCertSelected() {
            this.clusterSettings.keycloak.fileSelected = true;
            this.clusterSettings.keycloak.pemError = '';
        },
    },
    mounted() {
        // eslint-disable-next-line no-prototype-builtins
        if (this.$refs.hasOwnProperty('registryValues')) {
            this.registryValues = JSON.parse((this.$refs.registryValues as any).value);

            this.loadRegistryValues();
            this.dnsPreloadDataFromValues();
            this.wizardCronExpressionChange();
        }

        // eslint-disable-next-line no-prototype-builtins
        if (this.$refs.hasOwnProperty('adminsDataRef')) {
            this.loadAdmins(JSON.parse((this.$refs.adminsDataRef as any).value));
        }

        // eslint-disable-next-line no-prototype-builtins
        if (this.$refs.hasOwnProperty('cidrEditConfig') && (this.$refs.cidrEditConfig as any).value !== "") {
            let cidrConfig = JSON.parse((this.$refs.cidrEditConfig as any).value);

            // eslint-disable-next-line no-prototype-builtins
            if (cidrConfig.hasOwnProperty('citizen')) {
                this.citizenCIDR = cidrConfig.citizen;
                this.citizenCIDRValue.value = JSON.stringify(this.citizenCIDR);
            }

            // eslint-disable-next-line no-prototype-builtins
            if (cidrConfig.hasOwnProperty('officer')) {
                this.officerCIDR = cidrConfig.officer;
                this.officerCIDRValue.value = JSON.stringify(this.officerCIDR);
            }

            // eslint-disable-next-line no-prototype-builtins
            if (cidrConfig.hasOwnProperty('admin')) {
                this.adminCIDR = cidrConfig.admin;
                this.adminCIDRValue.value = JSON.stringify(this.adminCIDR);
            }
        }
    },
    computed: {
        customHosts() {
            return (this.registryValues as any)?.keycloak?.customHosts || [];
        }
    }
};
</script>

<template>
    <div class="registry registry-create" id="registry-form">
        <input type="hidden" ref="registryValues" :value="values" />
        <input type="hidden" ref="adminsDataRef" :value="adminsData" />
        <input type="hidden" id="preload-cidr" ref="cidrEditConfig" :value="cidrConfig" />
        <div class="registry-header">
            <a href="/admin/cluster/management" class="registry-add">
                <img alt="add registry" src="@/assets/img/action-back.png" />
                <span>НАЗАД</span>
            </a>
        </div>
        <h1>Налаштування платформи</h1>
        <div class="reg-wizard">
            <div class="wizard-contents">
                <ul>
                    <template v-for="tab in clusterSettings.tabs" :key="tab.key">
                        <li :class="{ active: clusterSettings.activeTab == tab.key }">
                            <a @click="selectClusterSettingsTab(tab.key, $event)" href="#">{{ tab.title }}</a>
                        </li>
                    </template>
                    <li v-if="hasUpdate" :class="{ active: clusterSettings.activeTab == 'platformUpdate' }">
                        <a @click="selectClusterSettingsTab('platformUpdate', $event)" href="#">Оновлення платформи</a>
                    </li>
                </ul>
            </div>
            <div class="wizard-body">
                <div class="wizard-tab" v-show="clusterSettings.activeTab == 'administrators'">
                    <administrators-block :admins="admins" :adminsValue="adminsValue" @delete-admin="deleteAdmin"
                        @show-admin-form="showAdminForm" />
                </div>
                <div class="wizard-tab" v-show="clusterSettings.activeTab == 'backup'">
                    <backup-block :backupSchedule="backupSchedule" :errorsMap="errorsMap" />
                </div>
                <div class="wizard-tab" v-show="clusterSettings.activeTab == 'allowedCIDR'">
                    <cidr-block :adminCIDR="adminCIDR" :adminCIDRValue="adminCIDRValue" @delete-cidr="deleteCIDR"
                        @show-cidr-form="showCIDRForm" />
                </div>
                <div class="wizard-tab" v-show="clusterSettings.activeTab === 'dataAboutKey' || clusterSettings.activeTab === 'dataAboutKeyVerification'">
                    <cluster-key-block ref="clusterKeyRef" :active-tab="clusterSettings.activeTab" />
                </div>
                <div class="wizard-tab" v-show="clusterSettings.activeTab == 'keycloakDNS'">
                    <cluster-keycloak-block :keycloak-hostname="keycloakHostname"
                        :cluster-keycloak-d-n-s-custom-hosts="customHosts" :cluster-settings="clusterSettings"
                        :backdrop-show="backdropShow" :dns-manual="dnsManual" @submitKeycloakDNSForm="submitKeycloakDNSForm"
                        @editClusterKeycloakDNSHost="editClusterKeycloakDNSHost"
                        @checkClusterDeleteKeycloakDNS="checkClusterDeleteKeycloakDNS"
                        @showClusterKeycloakDNSForm="showClusterKeycloakDNSForm"
                        @hideCheckClusterDeleteKeycloakDNS="hideCheckClusterDeleteKeycloakDNS"
                        @deleteClusterKeycloakDNS="deleteClusterKeycloakDNS"
                        @hideClusterCheckKeycloakDNS="hideClusterCheckKeycloakDNS"
                        @hideClusterKeycloakDNSForm="hideClusterKeycloakDNSForm"
                        @addClusterKeycloakDNS="addClusterKeycloakDNS"
                        @resetClusterKeycloakDNSForm="resetClusterKeycloakDNSForm"
                        @clusterKeycloakDNSCertSelected="clusterKeycloakDNSCertSelected" ref="clusterKeycloakRef" />
                </div>
                <div class="wizard-tab" v-show="clusterSettings.activeTab == 'platformUpdate'">
                    <platform-update-block :updateBranches="updateBranches" :errorsMap="errorsMap" />
                </div>
            </div>
        </div>
        <administrator-modal :adminPopupShow="adminPopupShow" :editAdmin="editAdmin" :requiredError="requiredError"
            :emailFormatError="emailFormatError" :usernameFormatError="usernameFormatError"
            :adminExistsError="adminExistsError" @create-admin="createAdmin" @hide-admin-form="hideAdminForm" />

        <cidr-modal v-model="editCidr" :cidrPopupShow="cidrPopupShow" :cidrFormatError="cidrFormatError"
            @create-cidr="createCIDR" @hide-cidr-form="hideCIDRForm" />
    </div>
</template>
