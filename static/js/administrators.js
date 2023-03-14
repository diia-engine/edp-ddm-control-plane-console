const hostnameRegex = /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/;

let app = Vue.createApp({
    mounted() {
        if (this.$refs.hasOwnProperty('smtpServerTypeSelected')) {
            let selectedSMTP = this.$refs.smtpServerTypeSelected.value
            if (selectedSMTP === "") {
                selectedSMTP = "platform-mail-server"
            }
            this.smtpServerType = selectedSMTP;
        }

        if (this.$refs.hasOwnProperty('smtpEditConfig') && this.$refs.smtpEditConfig.value !== "") {
            let smtpConfig = JSON.parse(this.$refs.smtpEditConfig.value);
            if (smtpConfig['type'] === 'external') {
                this.smtpServerType = 'external-mail-server';
                this.externalSMTPOpts = smtpConfig;
                this.externalSMTPOpts['port'] = smtpConfig['port'].toString();
            } else {
                this.smtpServerType = 'platform-mail-server';
            }
        }

        if (this.$refs.hasOwnProperty('cidrEditConfig') && this.$refs.cidrEditConfig.value !== "") {
            let cidrConfig = JSON.parse(this.$refs.cidrEditConfig.value);

            if (cidrConfig.hasOwnProperty('citizen')) {
                this.citizenCIDR = cidrConfig.citizen;
                this.citizenCIDRValue.value = JSON.stringify(this.citizenCIDR);
            }

            if (cidrConfig.hasOwnProperty('officer')) {
                this.officerCIDR = cidrConfig.officer;
                this.officerCIDRValue.value = JSON.stringify(this.officerCIDR);
            }

            if (cidrConfig.hasOwnProperty('admin')) {
                this.adminCIDR = cidrConfig.admin;
                this.adminCIDRValue.value = JSON.stringify(this.adminCIDR);
            }
        }

        if (this.$refs.hasOwnProperty('resourcesEditConfig') && this.$refs.resourcesEditConfig.value !== "") {
            let resourcesConfig = JSON.parse(this.$refs.resourcesEditConfig.value);
            this.preloadRegistryResources(resourcesConfig);
        }

        if (this.$refs.hasOwnProperty('registryBranches') && this.$refs.registryBranches.value !== "") {
            this.wizard.tabs.template.projectBranches = JSON.parse(this.$refs.registryBranches.value);
        }

        if (this.$refs.hasOwnProperty('wizardAction')) {
            this.wizard.registryAction = this.$refs.wizardAction.value;

            if (this.$refs.wizardAction.value === "edit") {
                let registryData = JSON.parse(this.$refs.registryData.value);
                this.wizard.tabs.general.registryName = registryData.name;
                this.wizard.tabs.template.visible = false;
                this.wizard.tabs.confirmation.visible = false;
                this.adminsChanged = false;
                this.cidrChanged = false;
            }
        }

        if (this.$refs.hasOwnProperty('registryValues')) {
            this.registryValues = JSON.parse(this.$refs.registryValues.value);

            this.loadRegistryValues();
            this.dnsPreloadDataFromValues();
            this.wizardCronExpressionChange();
        }

        if (this.$refs.hasOwnProperty('registryUpdate')) {
            this.wizard.tabs.update.visible = true;
        }
    },
    data() {
        return {
            registryValues: null,
            registryFormSubmitted: false,
            cidrChanged: true,
            officerCIDRValue: { value: '' },
            officerCIDR: [],
            citizenCIDRValue: { value: '' },
            citizenCIDR: [],
            adminCIDRValue: { value: '' },
            adminCIDR: [],
            adminsValue: '',
            adminsChanged: true,
            currentCIDR: [],
            currentCIDRValue: '',
            cidrFormatError: false,
            adminPopupShow: false,
            cidrPopupShow: false,
            admins: [],
            editCIDR: '',
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
            wizard: {
                registryAction: 'create',
                activeTab: 'general',
                tabs: {
                    general: {
                        title: 'Загальні', validated: false, registryName: '', requiredError: false, existsError: false,
                        formatError: false, validator: this.wizardGeneralValidation,
                        visible: true,
                    },
                    administrators: {title: 'Адміністратори', validated: false, requiredError: false,
                        validator: this.wizardAdministratorsValidation, visible: true,},
                    template: {title: 'Шаблон реєстру', validated: false, registryTemplate: '', registryBranch: '',
                        branches: [], projectBranches: {}, templateRequiredError: false, branchRequiredError: false,
                        validator: this.wizardTemplateValidation, visible: true, },
                    mail: {title: 'Поштовий сервер', validated: false, beginValidation: false,
                        validator: this.wizardMailValidation, visible: true,},
                    key: {title: 'Дані про ключ', validated: false, deviceType: 'file', beginValidation: false,
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
                        allowedKeys: [{issuer: '', serial: '', removable: false}],
                        caCertRequired: false,
                        caJSONRequired: false,
                        key6Required: false,
                        validator: this.wizardKeyValidation, visible: true,
                        changed: false,
                    },
                    resources: {title: 'Ресурси реєстру', validated: false, beginValidation: false,
                        validator: /*this.wizardResourcesValidation*/ this.wizardEmptyValidation, visible: true,},
                    dns: {title: 'DNS', validated: false, data: {officer: '', citizen: '',},
                        beginValidation: false, formatError: {officer: false, citizen: false, },
                        requiredError: {officer: false, citizen: false, },
                        typeError: {officer: false, citizen: false,},
                        editVisible: {officer: false, citizen: false},
                        validator: this.wizardDNSValidation, visible: true,
                        preloadValues: {}},
                    cidr: {title: 'Обмеження доступу', validated: true, visible: true, validator: this.wizardEmptyValidation, },
                    supplierAuthentication: {
                        title: 'Автентифікація надавачів послуг', validated: false, validator: this.wizardSupAuthValidation,
                        beginValidation:false, visible: true,
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
                    backupSchedule: {
                        title: 'Резервне копіювання', validated: false, beginValidation:false, visible: true,
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
                    confirmation: {title: 'Підтвердження', validated: true, visible: true, validator: this.wizardEmptyValidation, }
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
                        key: 'keycloakDNS',
                        title: 'Keycloak DNS'
                    },
                ],
                keycloak: {
                    editHostname: '',
                    editCertPath: '',
                    submitInput: '',
                    hostname: '',
                    formShow: false,
                    hostnameError: '',
                    fileSelected: false,
                    pemError: '',
                },
            },
        }
    },
    methods: {
        editClusterKeycloakDNSHost(hostname, certificatePath, e) {
            e.preventDefault();
            this.clusterSettings.keycloak.editHostname = hostname;
            this.clusterSettings.keycloak.editCertificatePath = certificatePath;
            this.clusterSettings.keycloak.hostname = hostname;

            this.backdropShow = true;
            this.clusterSettings.keycloak.formShow = true;
            this.clusterSettings.keycloak.fileSelected = true;
        },
        clusterKeycloakDNSCustomHosts() {
            if (this.registryValues === null) {
                return [];
            }

            return this.registryValues.keycloak.customHosts;
        },
        submitKeycloakDNSForm(e){
            this.clusterSettings.keycloak.submitInput = JSON.stringify(this.registryValues.keycloak.customHosts);
        },
        clusterKeycloakDNSCertSelected(e) {
            e.preventDefault();
            this.clusterSettings.keycloak.fileSelected = true;
            this.clusterSettings.keycloak.pemError = '';
        },
        resetClusterKeycloakDNSForm(e) {
            e.preventDefault();
            this.clusterSettings.keycloak.fileSelected = false;
            this.clusterSettings.keycloak.pemError = '';
        },
        showClusterKeycloakDNSForm(e) {
            e.preventDefault();
            this.backdropShow = true;
            this.clusterSettings.keycloak.formShow = true;
            this.clusterSettings.keycloak.fileSelected = false;
            this.clusterSettings.keycloak.hostname = '';
        },
        hideClusterKeycloakDNSForm(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.clusterSettings.keycloak.formShow = false;
        },
        deleteClusterKeycloakDNS(hostname, e) {
            e.preventDefault();

            this.registryValues.keycloak.customHosts = this.registryValues.keycloak.customHosts.filter(
                i => i.host !== hostname);
        },
        addClusterKeycloakDNS(e) {
            e.preventDefault();
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

            for (let i=0;i<this.registryValues.keycloak.customHosts.length;i++) {
                if (this.registryValues.keycloak.customHosts[i].host === this.clusterSettings.keycloak.hostname &&
                    this.clusterSettings.keycloak.hostname !== this.clusterSettings.keycloak.editHostname) {
                    this.clusterSettings.keycloak.hostnameError = 'Така назва вже використовується';
                    return;
                }
            }

            if (this.clusterSettings.keycloak.fileSelected === false) {
                this.clusterSettings.keycloak.pemError = 'Поле обов’язкове для заповнення';
                return;
            }

            if (this.$refs.clusterKeycloakDNS.files.length > 0) {
                let formData = new FormData();
                formData.append("file", this.$refs.clusterKeycloakDNS.files[0]);
                formData.append("hostname", this.clusterSettings.keycloak.hostname);
                let $this = this;
                axios.post('/admin/cluster/upload-pem-dns', formData, {
                    headers: {
                        'Content-Type': 'multipart/form-data'
                    }
                }).then(function (rsp) {
                    if ($this.clusterSettings.keycloak.editHostname === '') {
                        $this.registryValues.keycloak.customHosts.push({
                            host: $this.clusterSettings.keycloak.hostname,
                            certificatePath: rsp.data,
                        });
                    } else {
                        for (let i=0;i<$this.registryValues.keycloak.customHosts.length;i++) {
                            if ($this.registryValues.keycloak.customHosts[i].host === $this.clusterSettings.keycloak.editHostname) {
                                $this.registryValues.keycloak.customHosts[i].host = $this.clusterSettings.keycloak.hostname;
                                $this.registryValues.keycloak.customHosts[i].certificatePath = rsp.data;
                            }
                        }
                    }

                    $this.backdropShow = false;
                    $this.clusterSettings.keycloak.formShow = false;
                    $this.clusterSettings.keycloak.editHostname = '';
                }).catch(function (error) {
                    $this.clusterSettings.keycloak.pemError = $this.localePEMError(error.response.data);
                });
            } else {
                for (let i=0;i<this.registryValues.keycloak.customHosts.length;i++) {
                    if (this.registryValues.keycloak.customHosts[i].host === this.clusterSettings.keycloak.editHostname) {
                        this.registryValues.keycloak.customHosts[i].host = this.clusterSettings.keycloak.hostname;
                    }
                }

                this.backdropShow = false;
                this.clusterSettings.keycloak.formShow = false;
                this.clusterSettings.keycloak.editHostname = '';
            }
        },
        localePEMError(message) {
            const messages = {
                "found in PEM file": "Перевірте формат файла",
                "certificate has expired or is not yet valid": "Сертифікат застарілий",
                "certificate is valid for": "Сертифікат не відповідає доменному імені",
            }

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
        wizardCronExpressionChange(e) {
            let bs = this.wizard.tabs.backupSchedule;
            if (bs.data.cronSchedule === '') {
                bs.nextLaunches = false;
                bs.wrongCronFormat = false;
                return;
            }

            try {
                const cron = cronSchedule.parseCronExpression(bs.data.cronSchedule)
                bs.nextDates = [];
                let dt = new Date();
                for (let i = 0; i < 3; i++) {
                    let next = cron.getNextDate(dt);
                    bs.nextDates.push(`${next.toLocaleDateString('uk')} ${next.toLocaleTimeString('uk')}`);
                    dt = next;
                }
                bs.nextLaunches = true;
                bs.wrongCronFormat = false;
            } catch (e) {
                bs.nextLaunches = false;
                bs.wrongCronFormat = true;
            }
        },
        wizardSupAuthFlowChange() {
            this.wizard.tabs.supplierAuthentication.validated = false;
            this.wizard.tabs.supplierAuthentication.beginValidation = false;

            let registryValues = this.registryValues;

            if (this.wizard.tabs.supplierAuthentication.data.authType === 'dso-officer-auth-flow') {
                if (registryValues && registryValues.signWidget.url !== '') {
                    this.wizard.tabs.supplierAuthentication.data.url = registryValues.signWidget.url;
                } else {
                    this.wizard.tabs.supplierAuthentication.data.url = this.wizard.tabs.supplierAuthentication.dsoDefaultURL;
                }

                if (registryValues && registryValues.keycloak.authFlows.officerAuthFlow.widgetHeight !== 0) {
                    this.wizard.tabs.supplierAuthentication.data.widgetHeight =
                        registryValues.keycloak.authFlows.officerAuthFlow.widgetHeight;
                }

            } else {
                if (registryValues && registryValues.keycloak.identityProviders.idGovUa.url !== '') {
                    this.wizard.tabs.supplierAuthentication.data.url = registryValues.keycloak.identityProviders.idGovUa.url;
                } else {
                    this.wizard.tabs.supplierAuthentication.data.url = '';
                }

                if (registryValues && registryValues.keycloak.identityProviders.idGovUa.clientId !== '') {
                    this.wizard.tabs.supplierAuthentication.data.clientId = registryValues.keycloak.identityProviders.idGovUa.clientId;
                    this.wizard.tabs.supplierAuthentication.data.secret = '*****';
                }
            }
        },
        loadRegistryValues() {
            try {
                if (this.registryValues.keycloak.realms.officerPortal.browserFlow !== '') {
                    this.wizard.tabs.supplierAuthentication.data.authType = this.registryValues.keycloak.realms.officerPortal.browserFlow;
                }

                if (this.wizard.tabs.supplierAuthentication.data.authType === 'dso-officer-auth-flow') {
                    this.wizard.tabs.supplierAuthentication.data.widgetHeight =
                        this.registryValues.keycloak.authFlows.officerAuthFlow.widgetHeight;
                    this.wizard.tabs.supplierAuthentication.data.url = this.registryValues.signWidget.url;
                } else {
                    this.wizard.tabs.supplierAuthentication.data.url =
                        this.registryValues.keycloak.identityProviders.idGovUa.url;
                    this.wizard.tabs.supplierAuthentication.data.clientId = this.registryValues.keycloak.identityProviders.idGovUa.clientId;
                    this.wizard.tabs.supplierAuthentication.data.secret = '*****';
                }
            } catch (e) {
                console.log(e);
            }

            try {
                this.wizard.tabs.backupSchedule.enabled = this.registryValues.global.registryBackup.enabled;
                this.wizard.tabs.backupSchedule.data.cronSchedule = this.registryValues.global.registryBackup.schedule;
                this.wizard.tabs.backupSchedule.data.days = this.registryValues.global.registryBackup.expiresInDays;
            } catch (e) {
                console.log(e);
            }

            if (this.registryValues.keycloak.customHosts === null) {
                this.registryValues.keycloak.customHosts = [];
            }
        },
        removeResourcesCatFromList(name)  {
            let searchIdx = this.registryResources.cats.indexOf(name);
            if (searchIdx !== -1) {
                this.registryResources.cats.splice(
                    searchIdx, 1);
            }
        },
        decodeResourcesEnvVars(inEnvVars) {
            let envVars = [];

            for (let j in inEnvVars) {
                envVars.push({
                    name: j,
                    value: inEnvVars[j],
                })
            }

            return envVars;
        },
        preloadRegistryResources(data) {
            //TODO: move to constant
            this.registryResources.cats = [
                'kong',
                'bpms',
                'digitalSignatureOps',
                'userTaskManagement',
                'userProcessManagement',
                'digitalDocumentService',
                'restApi',
                'kafkaApi',
                'soapApi',
            ];

            this.registryResources.addedCats = [];

            for (let i in data) {
                this.removeResourcesCatFromList(i);
                if (data[i].hasOwnProperty('container') &&
                    this.isObject(data[i].container) && data[i].container.hasOwnProperty('envVars')) {
                    data[i].container.envVars = this.decodeResourcesEnvVars(data[i].container.envVars);
                }

                let mergedData = this.mergeResource(data[i]);
                this.registryResources.addedCats.push({
                    name: i,
                    config: mergedData,
                });
            }
        },
        mergeResource(data) {
            let emptyResource = {
                istio: {
                    sidecar: {
                        enabled: false,
                            resources: {
                            requests: {
                                cpu: '',
                                memory: ''
                            },
                            limits: {
                                cpu: '',
                                memory: '',
                            },
                        },
                    },
                },
                container: {
                    resources: {
                        requests: {
                            cpu: '',
                                memory: ''
                        },
                        limits: {
                            cpu: '',
                                memory: '',
                        },
                    },
                    envVars: [{name: '', value: ''}],
                },

            };

            this.mergeDeep(emptyResource, data);
            return emptyResource;
        },
        isObject(item) {
            return (item && typeof item === 'object' && !Array.isArray(item));
        },
        mergeDeep(target, ...sources) {
            if (!sources.length) return target;
            const source = sources.shift();

            if (this.isObject(target) && this.isObject(source)) {
                for (const key in source) {
                    if (source[key] === null) {
                        continue
                    }

                    if (this.isObject(source[key])) {
                        if (!target[key]) Object.assign(target, { [key]: {} });
                        this.mergeDeep(target[key], source[key]);
                    } else {
                        Object.assign(target, { [key]: source[key] });
                    }
                }
            }

            return this.mergeDeep(target, ...sources);
        },
        wizardBackupScheduleChange(e) {
            console.log(e);
        },
        wizardDNSEditVisibleChange(name, event){
            console.log(name, event);
        },
        wizardEditSubmit(event) {
            let tab = this.wizard.tabs[this.wizard.activeTab];
            let $this = this;

            tab.validator(tab).then(function (){
                $this.registryFormSubmit(event);
                $this.$nextTick(() => {
                    $this.$refs.registryWizardForm.submit();
                });
            });
        },
        wizardNext() {
            let tabKeys = Object.keys(this.wizard.tabs);

            for (let i=0;i<tabKeys.length;i++) {
                if (tabKeys[i] === this.wizard.activeTab) {
                    let tab = this.wizard.tabs[tabKeys[i]];
                    let wizard = this.wizard;

                    tab.validator(tab).then(function (){
                        wizard.activeTab = tabKeys[i+1];
                    });

                    break;
                }
            }
        },
        dnsPreloadDataFromValues() {
            if (this.registryValues && this.registryValues.hasOwnProperty('portals')) {
                for (let p in this.registryValues.portals) {
                    if(this.registryValues.portals[p].hasOwnProperty('customDns')) {
                        this.wizard.tabs.dns.editVisible[p] =  this.registryValues.portals[p].customDns.enabled;
                        this.wizard.tabs.dns.data[p] =  this.registryValues.portals[p].customDns.host;
                        this.wizard.tabs.dns.preloadValues[p] = this.registryValues.portals[p].customDns.host;
                    }
                }
            }
        },
        wizardPrev(){
            let tabKeys = Object.keys(this.wizard.tabs);

            for (let i=0;i<tabKeys.length;i++) {
                if (tabKeys[i] === this.wizard.activeTab) {
                    let tab = this.wizard.tabs[tabKeys[i]];
                    let wizard = this.wizard;

                    tab.validator(tab).then(function (){
                        wizard.activeTab = tabKeys[i-1];
                    });

                    break;
                }
            }
        },
        selectWizardTab(tabName, e) {
            e.preventDefault();

            let tab = this.wizard.tabs[this.wizard.activeTab];
            let wizard = this.wizard;

            tab.validator(tab).then(function (){
                if (wizard.registryAction === "create") {
                    for (let k in wizard.tabs) {
                        if(!wizard.tabs[k].validated) {
                            return;
                        }

                        if (k === tabName) {
                            break
                        }
                    }
                }

                wizard.activeTab = tabName;
            });
        },
        selectClusterSettingsTab(tabName, e) {
            e.preventDefault();
            this.clusterSettings.activeTab = tabName;
        },
        wizardTabChanged(tabName) {
            this.wizard.tabs[tabName].changed = true;
        },
        wizardEmptyValidation(tab) {
            return new Promise((resolve) => {
                resolve();
            });
        },
        wizardGeneralValidation(tab) {
            return new Promise((resolve) => {
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
        wizardBackupScheduleValidation(tab) {
            return new Promise((resolve) => {
                let bs = this.wizard.tabs.backupSchedule;
                bs.data.cronSchedule = bs.data.cronSchedule.trim();

                if (!bs.enabled) {
                    resolve();
                    return;
                }

                tab.beginValidation = true;
                tab.validated = false;

                bs.wrongCronFormat = false;
                bs.wrongDaysFormat = false;

                if (bs.data.cronSchedule !== '') {
                    try {
                        cronSchedule.parseCronExpression(bs.data.cronSchedule)
                    } catch (e) {
                        bs.nextLaunches = false;
                        bs.wrongCronFormat = true;
                    }
                }

                const days = parseInt(bs.data.days);
                if (bs.data.days !== '' && (!/^[0-9]+$/.test(bs.data.days) || isNaN(days) || days <= 0)) {
                    bs.wrongDaysFormat = true;
                }

                if (bs.wrongDaysFormat || bs.wrongCronFormat|| bs.data.cronSchedule === '' || bs.data.days === '') {
                    return;
                }

                tab.validated = true;
                tab.beginValidation = false;
                resolve();
            });
        },
        wizardSupAuthValidation(tab){
            return new Promise((resolve) => {
                tab.beginValidation = true;
                tab.validated = false;
                tab.urlValidationFailed = false;
                tab.heightIsNotNumber = false;

                if (!/^(http(s)?:\/\/.)[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=,]*)$/.test(this.wizard.tabs.supplierAuthentication.data.url)) {
                    tab.urlValidationFailed = true;
                    return;
                }

                if (this.wizard.tabs.supplierAuthentication.data.authType === 'dso-officer-auth-flow') {
                    if (this.wizard.tabs.supplierAuthentication.data.widgetHeight === '') {
                        return;
                    }

                    if (!/^[0-9]+$/.test(this.wizard.tabs.supplierAuthentication.data.widgetHeight)) {
                        this.wizard.tabs.supplierAuthentication.heightIsNotNumber = true;
                        return;
                    }
                }

                if (this.wizard.tabs.supplierAuthentication.data.authType === 'id-gov-ua-officer-redirector' &&
                    (this.wizard.tabs.supplierAuthentication.data.clientId === '' ||
                        this.wizard.tabs.supplierAuthentication.data.secret === '')) {
                    return
                }

                tab.validated = true;
                tab.beginValidation = false;
                resolve();
            });
        },
        wizardDNSValidation(tab){
            return new Promise((resolve) => {
                tab.beginValidation = true;
                tab.validated = false;

                let validationFailed = false;
                let filesToCheck = [];

                for (let k in this.wizard.tabs.dns.data) {
                    this.wizard.tabs.dns.formatError[k] = false;
                    this.wizard.tabs.dns.requiredError[k] = false;
                    this.wizard.tabs.dns.typeError[k] = false;
                    let fileInput = this.$refs[`${k}SSL`];

                    if (this.wizard.tabs.dns.data[k] !== '') {
                        if (!/^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/.test(this.wizard.tabs.dns.data[k])) {
                            this.wizard.tabs.dns.formatError[k] = true;
                            validationFailed = true;
                        }

                        if (fileInput.files.length === 0 &&
                            (!this.wizard.tabs.dns.preloadValues.hasOwnProperty(k) || this.wizard.tabs.dns.preloadValues[k] !== this.wizard.tabs.dns.data[k])) {
                            this.wizard.tabs.dns.requiredError[k] = true;
                            validationFailed = true;
                        } else if (fileInput.files.length > 0) {
                            filesToCheck.push({name: k, file: fileInput.files[0]});
                        }
                    } else if (fileInput.files.length > 0) {
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
        wizardCheckPEMFiles(filesToCheck, resolve, tab) {
            if (filesToCheck.length === 0) {
                tab.validated = true;
                tab.beginValidation = false;
                resolve();
                return;
            }

            let f = filesToCheck.pop();
            let formData = new FormData();
            formData.append("file", f.file);
            let $this = this;

            axios.post('/admin/registry/check-pem', formData, {
                headers: {
                    'Content-Type': 'multipart/form-data'
                }
            }).then(function(rsp){
                $this.wizardCheckPEMFiles(filesToCheck, resolve, tab);
            }).catch(function(error){
                $this.wizard.tabs.dns.typeError[f.name] = true;
            });
        },
        wizardResourcesValidation(tab){
            return new Promise((resolve) => {
                tab.beginValidation = true;
                tab.validated = false;

                for (let i=0;i<this.registryResources.addedCats.length;i++) {
                    let cat = this.registryResources.addedCats[i];

                    if (!this.checkObjectFieldsEmpty(cat.config)) {
                        return;
                    }
                }

                tab.beginValidation = false;
                tab.validated = true;
                resolve();
            });
        },
        checkObjectFieldsEmpty(o){
            for (let i in o) {
                let t = typeof o[i];
                if (t === 'string' && o[i] === '') {
                    return false;
                }

                if (t === 'object') {
                    if(!this.checkObjectFieldsEmpty(o[i])) {
                        return false;
                    }
                }
            }

            return true;
        },
        wizardAdministratorsValidation(tab) {
            let admins = this.admins;

            return new Promise((resolve) => {
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
        wizardTemplateValidation(tab){
            return new Promise((resolve) => {
                tab.validated = false;
                tab.templateRequiredError = false;
                tab.branchRequiredError = false;

                if (tab.registryTemplate === '') {
                    tab.templateRequiredError = true;
                    return;
                }

                if (tab.registryBranch === '') {
                    tab.branchRequiredError = true;
                    return;
                }
                let $this = this;

                axios.get(
                    `/admin/registry/preload-resources`, { params: {
                            'template': tab.registryTemplate,
                            'branch': tab.registryBranch,
                        }
                    })
                    .then(function (response) {
                        $this.preloadRegistryResources(response.data);
                        tab.validated = true;
                        resolve();
                    })
                    .catch(function (error) {
                        tab.validated = true;
                        resolve();
                    });
            });
        },
        wizardMailValidation(tab){
            return new Promise((resolve) => {
                tab.validated = false;

                if (this.smtpServerType === 'platform-mail-server') {
                    tab.validated = true;
                    resolve();
                    return;
                }

                tab.beginValidation = true;

                for (let key in this.externalSMTPOpts) {
                    if (this.externalSMTPOpts[key] === '') {
                        return;
                    }
                }

                tab.beginValidation = false;
                tab.validated = true;
                resolve();
            });
        },
        clusterKeyFormSubmit(e) {
            if (!this.keyFormValidation(this.wizard.tabs.key, function () {

            })) {
                e.preventDefault();
            }
        },
        keyFormValidation(tab, resolve) {
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
            if (this.$refs.keyCaCert.files.length === 0) {
                this.wizard.tabs.key.caCertRequired = true;
                validationFailed = true;
            }

            if (this.$refs.keyCaJSON.files.length === 0) {
                this.wizard.tabs.key.caJSONRequired = true;
                validationFailed = true;
            }

            for (let i=0;i<this.wizard.tabs.key.allowedKeys.length;i++) {
                if (this.wizard.tabs.key.allowedKeys[i].issuer === '' ||
                    this.wizard.tabs.key.allowedKeys[i].serial === '') {
                    validationFailed = true;
                }
            }

            if (this.wizard.tabs.key.deviceType === 'hardware') {
                for (let key in this.wizard.tabs.key.hardwareData) {
                    if (this.wizard.tabs.key.hardwareData[key] === '') {
                        validationFailed = true;
                    }
                }
            } else {
                for (let key in this.wizard.tabs.key.fileData) {
                    if (this.wizard.tabs.key.hardwareData[key] === '') {
                        validationFailed = true;
                    }
                }

                if (this.$refs.key6.files.length === 0) {
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
        wizardKeyValidation(tab){
            return new Promise((resolve) => {
                this.keyFormValidation(tab, resolve);
            });
        },
        renderINITemplate() {
            let iniTemplate = document.getElementById('ini-template').innerHTML;
            this.wizard.tabs.key.hardwareData.iniConfig = Mustache.render(iniTemplate, {
                "CA_NAME": this.wizard.tabs.key.hardwareData.remoteCaName,
                "CA_HOST": this.wizard.tabs.key.hardwareData.remoteCaHost,
                "CA_PORT": this.wizard.tabs.key.hardwareData.remoteCaPort,
                "KEY_SN": this.wizard.tabs.key.hardwareData.remoteSerialNumber,
                "KEY_HOST": this.wizard.tabs.key.hardwareData.remoteKeyHost,
                "KEY_ADDRESS_MASK": this.wizard.tabs.key.hardwareData.remoteKeyMask,
            }).trim();
        },
        wizardKeyHardwareDataChanged(e) {
            this.renderINITemplate();
            this.wizard.tabs.key.changed = true;
        },
        wizardRemoveAllowedKey(item) {
            let searchIdx = this.wizard.tabs.key.allowedKeys.indexOf(item);
            if (searchIdx !== -1) {
                this.wizard.tabs.key.allowedKeys.splice(
                    searchIdx, 1);
            }
        },
        wizardAddAllowedKey() {
            this.wizard.tabs.key.allowedKeys.push({issuer: '', serial: '', removable: true});
        },
        addResourceCat(e) {
            e.preventDefault(e);
            if (this.registryResources.cat === '') {
                return;
            }

            this.registryResources.addedCats.unshift({
                name: this.registryResources.cat,
                config: {
                    istio: {
                        sidecar: {
                            enabled: false,
                            resources: {
                                requests: {
                                    cpu: '',
                                    memory: ''
                                },
                                limits: {
                                    cpu: '',
                                    memory: '',
                                },
                            },
                        },
                    },
                    container: {
                        resources: {
                            requests: {
                                cpu: '',
                                memory: ''
                            },
                            limits: {
                                cpu: '',
                                memory: '',
                            },
                        },
                        envVars: [{name: '', value: ''}],
                    },
                }
            });

            this.registryResources.cats.splice(
                this.registryResources.cats.indexOf(this.registryResources.cat), 1);
        },
        addEnvVar(envVars, event) {
            event.preventDefault();
            envVars.push({name: "", value: ""})
        },
        removeEnvVar(envVars, env) {
            envVars.splice(envVars.indexOf(env), 1);
        },
        removeResourceCat(cat, event) {
            event.preventDefault();

            this.registryResources.cats.push(cat.name);

            this.registryResources.addedCats.splice(
                this.registryResources.addedCats.indexOf(cat), 1);
        },
        encodeRegistryResources() {
            let prepare = {};

            this.registryResources.addedCats.forEach(function (el) {
                let cloneEL = JSON.parse(JSON.stringify(el));

                let envVars = {};
                cloneEL.config.container.envVars.forEach(function (el) {
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
        cleanEmptyProperties(obj) {
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
        registryFormSubmit(e) {
            if (this.registryFormSubmitted && e) {
                e.preventDefault();
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
        prepareDNSConfig(){
            for (let k in this.wizard.tabs.dns.data) {
                let fileInput = this.$refs[`${k}SSL`];

                if (this.wizard.tabs.dns.editVisible[k] &&
                    this.wizard.tabs.dns.data[k] === this.wizard.tabs.dns.preloadValues[k] &&
                    fileInput.files.length === 0) {
                    this.wizard.tabs.dns.data[k] = '';
                }
            }
        },
        loadAdmins(admins) {
            if (!this.adminsLoaded) {
                if (admins !== "") {
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
        showCIDRForm(cidr, value) {
            this.cidrPopupShow = true;
            $("body").css("overflow", "hidden");

            this.editCIDR = '';
            this.currentCIDR = cidr;
            this.currentCIDRValue = value;
            this.cidrFormatError = false;
        },
        hideCIDRForm() {
            this.cidrPopupShow = false;
            $("body").css("overflow", "scroll");
        },
        createCIDR(e) {
            e.preventDefault();
            let cidrVal = String(this.editCIDR).toLowerCase();
            if (cidrVal !== '0.0.0.0/0' && !cidrVal.
            match(/^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:\/[0-2]\d|\/3[0-2])?$/)) {
                this.cidrFormatError = true;
                return;
            }

            this.currentCIDR.push(this.editCIDR);
            this.currentCIDRValue.value = JSON.stringify(this.currentCIDR);
            this.hideCIDRForm();
            this.cidrChanged = true;
        },
        deleteCIDR(c, cidr, value, e) {
            e.preventDefault();

            for (let v in cidr) {
                if (cidr[v] === c) {
                    cidr.splice(v, 1);
                    break;
                }
            }

            value = JSON.stringify(cidr);
            this.cidrChanged = true;
        },
        hideAdminForm() {
            this.adminPopupShow = false;
            $("body").css("overflow", "scroll");
        },
        deleteAdmin(e) {
            e.preventDefault();
            let email = e.currentTarget.getAttribute('email');

            for (let v in this.admins) {
                if (this.admins[v].email === email) {
                    this.admins.splice(v, 1);
                    break;
                }
            }
            this.adminsValue = JSON.stringify(this.admins);
            this.adminsChanged = true;
        },
        createAdmin(e) {
            this.requiredError = false;
            this.emailFormatError = false;

            e.preventDefault();
            for (let v in this.editAdmin) {
                if (this.editAdmin[v] === "") {
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

            for (let i=0;i<this.admins.length;i++) {
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
        changeTemplateProject(){
            this.wizard.tabs.template.branches =
                this.wizard.tabs.template.projectBranches[this.wizard.tabs.template.registryTemplate];
        },
    }
})

app.config.compilerOptions.delimiters = ['[[', ']]'];
app.mount('#registry-form');

let editKeyChecked = function(){
    let checked = $("#edit-key").prop('checked'),
        keyBlock = $("#key-block");
    keyBlock.find("input").prop('disabled', !checked);

    if (checked) {
        keyBlock.show();
        $("#key-device-type").change();
    } else {
        keyBlock.hide();
    }
};

$(function () {
    let editKey = $("#edit-key");
    editKey.change(editKeyChecked);
    editKey.prop("checked", false);
    editKeyChecked();
});
