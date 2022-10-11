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

            for (let i in resourcesConfig) {
                this.registryResources.cats.splice(
                    this.registryResources.cats.indexOf(i), 1);

                let envVars = [];

                for (let j in resourcesConfig[i].container.envVars) {
                    envVars.push({
                        name: j,
                        value: resourcesConfig[i].container.envVars[j],
                    })
                }

                resourcesConfig[i].container.envVars = envVars;

                this.registryResources.addedCats.push({
                    name: i,
                    config: resourcesConfig[i],
                });
            }
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
    },
    data() {
        return {
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
                    'digital-signature-ops',
                    'user-task-management',
                    'user-process-management',
                    'form-management-provider',
                    'digital-document-service',
                    'registry-rest-api',
                    'registry-kafka-api'
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
                        validator: this.wizardResourcesValidation, visible: true,},
                    dns: {title: 'DNS', validated: false, data: {officer: '', citizen: '', keycloak: ''},
                        beginValidation: false, formatError: {officer: false, citizen: false, keycloak: false},
                        requiredError: {officer: false, citizen: false, keycloak: false},
                        typeError: {officer: false, citizen: false, keycloak: false},
                        validator: this.wizardDNSValidation, visible: true, },
                    cidr: {title: 'Обмеження доступу', validated: true, visible: true, validator: this.wizardEmptyValidation, },
                    confirmation: {title: 'Підтвердження', validated: true, visible: true, validator: this.wizardEmptyValidation, }
                },
            },
        }
    },
    methods: {
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

                let filesToCheck = [];

                for (let k in this.wizard.tabs.dns.data) {
                    this.wizard.tabs.dns.formatError[k] = false;
                    this.wizard.tabs.dns.requiredError[k] = false;
                    this.wizard.tabs.dns.typeError[k] = false;

                    if (this.wizard.tabs.dns.data[k] !== '') {
                        if (!/^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/.test(this.wizard.tabs.dns.data[k])) {
                            this.wizard.tabs.dns.formatError[k] = true;
                            return;
                        }

                        let fileInput = this.$refs[`${k}SSL`];
                        if (fileInput.files.length === 0) {
                            this.wizard.tabs.dns.requiredError[k] = true;
                            return;
                        }

                        filesToCheck.push({name: k, file: fileInput.files[0]});
                    }
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

                tab.validated = true;
                resolve();
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

                if (this.$refs.keyCaCert.files.length === 0) {
                    this.wizard.tabs.key.caCertRequired = true;
                    return;
                }

                if (this.$refs.keyCaJSON.files.length === 0) {
                    this.wizard.tabs.key.caJSONRequired = true;
                    return;
                }

                for (let i=0;i<this.wizard.tabs.key.allowedKeys.length;i++) {
                    if (this.wizard.tabs.key.allowedKeys[i].issuer === '' ||
                        this.wizard.tabs.key.allowedKeys[i].serial === '') {
                        return;
                    }
                }

                if (this.wizard.tabs.key.deviceType === 'hardware') {
                    for (let key in this.wizard.tabs.key.hardwareData) {
                        if (this.wizard.tabs.key.hardwareData[key] === '') {
                            return;
                        }
                    }
                } else {
                    for (let key in this.wizard.tabs.key.fileData) {
                        if (this.wizard.tabs.key.hardwareData[key] === '') {
                            return;
                        }
                    }

                    if (this.$refs.key6.files.length === 0) {
                        this.wizard.tabs.key.key6Required = true;
                        return;
                    }
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
                let envVars = {};
                el.config.container.envVars.forEach(function (el) {
                    envVars[el.name] = el.value;
                });
                el.config.container.envVars = envVars;

                prepare[el.name] = {
                    istio: el.config.istio,
                    container: el.config.container,
                };
            });

            this.registryResources.encoded = JSON.stringify(prepare);
        },
        registryFormSubmit(e) {
            if (this.registryFormSubmitted) {
                e.preventDefault();
                return;
            }

            this.encodeRegistryResources();
            this.mailServerOpts = JSON.stringify(this.externalSMTPOpts);
            this.registryFormSubmitted = true;
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
            if (!String(this.editCIDR).toLowerCase().match(/^([01]?\d\d?|2[0-4]\d|25[0-5])(?:\.(?:[01]?\d\d?|2[0-4]\d|25[0-5])){3}(?:\/[0-2]\d|\/3[0-2])?$/)) {
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