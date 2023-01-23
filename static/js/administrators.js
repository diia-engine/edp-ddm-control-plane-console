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
                    dns: {title: 'DNS', validated: false, data: {officer: '', citizen: '', /*keycloak: ''*/},
                        beginValidation: false, formatError: {officer: false, citizen: false, /*keycloak: false*/},
                        requiredError: {officer: false, citizen: false, /*keycloak: false*/},
                        typeError: {officer: false, citizen: false, /*keycloak: false*/},
                        editVisible: {officer: true, citizen: true},
                        validator: this.wizardDNSValidation, visible: true,
                        preloadValues: {}},
                    cidr: {title: 'Обмеження доступу', validated: true, visible: true, validator: this.wizardEmptyValidation, },
                    confirmation: {title: 'Підтвердження', validated: true, visible: true, validator: this.wizardEmptyValidation, }
                },
            },
        }
    },
    methods: {
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
        dnsUnsetPreloadedValue(name, e){
            this.wizard.tabs.dns.editVisible[name] = true;
            this.wizard.tabs.dns.data[name] = this.wizard.tabs.dns.preloadValues[name];
            e.preventDefault();
        },
        dnsSetPreloadedValue(name, value) {
            if (!this.wizard.tabs.dns.preloadValues.hasOwnProperty(name)) {
                this.wizard.tabs.dns.editVisible[name] = false;
                this.wizard.tabs.dns.preloadValues[name] = value;
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

                        if (fileInput.files.length === 0) {
                            this.wizard.tabs.dns.requiredError[k] = true;
                            validationFailed = true;
                        } else {
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
        wizardKeyValidation(tab){
            return new Promise((resolve) => {
                if (this.wizard.registryAction === 'edit' && !this.wizard.tabs.key.changed) {
                    resolve();
                    return;
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
                    return;
                }

                tab.beginValidation = false;
                tab.validated = true;

                resolve();
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
                if (this.wizard.tabs.dns.editVisible[k] && this.wizard.tabs.dns.data[k] === '' &&
                    this.wizard.tabs.dns.preloadValues.hasOwnProperty(k)) {
                    this.wizard.tabs.dns.preloadValues = '';
                    this.wizard.tabs.dns.editVisible[k] = false;
                    this.wizard.tabs.dns.data[k] = '-';
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
