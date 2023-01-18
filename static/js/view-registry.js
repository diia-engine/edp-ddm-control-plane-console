let forceMR = false;

let hasNewMergeRequests = function () {
    if (forceMR) {
        return true;
    }

    let statuses = $(".mr-status");
    for (let i=0;i<statuses.length;i++) {
        if ($(statuses[i]).html().trim() === "NEW") {
            return true;
        }
    }

    return false;
}

/* ereg-name */

let app = Vue.createApp({
    mounted() {
        if (this.$refs.hasOwnProperty('valuesJson')) {
            this.values = JSON.parse(this.$refs.valuesJson.value);
        }

        if (this.$refs.hasOwnProperty('registryName')) {
            this.registryName = this.$refs.registryName.value
        }

        if (this.$refs.hasOwnProperty('openMergeRequests')) {
            this.openMergeRequests.has = true;
        }

        this.externalSystem = this.externalSystemDefaults();
        this.trembitaClient = this.trembitaClientDefaults();
    },
    data() {
        return {
            registryName: '',
            values: {},
            accordion: 'general',
            externalRegPopupShow: false,
            backdropShow: false,
            internalRegistryReg: true,
            externalSystemType: "internal-registry",
            removeExternalRegPopupShow: false,
            systemToDelete: '',
            systemToDisable: '',
            systemToDisableType: '',
            systemToDeleteType: '',
            mrError: false,
            externalKey: false,
            systemToShowKey: '',
            keyValue: '******',
            currentExternalKeyValue: '',
            accessGrantError: false,
            mrView: false,
            mrSrc: '',
            openMergeRequests: {
                has: false,
                formShow: false,
            },
            externalSystem: {},
            trembitaClient: {},
        }
    }, // mrIframe
    methods: {
        isURL(u){
            return /^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)$/.test(u);
        },
        trembitaClientDefaults() {
            return {
                registryName: '',
                formShow: false,
                startValidation: false,
                tokenInputType: 'text',
                urlValidationFailed: false,
                data: {
                    protocolVersion: '',
                    url: '',
                    userId: '',
                    protocol: "SOAP",
                    client: {
                        xRoadInstance: '',
                        memberClass: '',
                        memberCode: '',
                        subsystemCode: '',
                    },
                    service: {
                        xRoadInstance: '',
                        memberClass: '',
                        memberCode: '',
                        subsystemCode: '',
                        auth: {
                            type: 'NO_AUTH',
                        },
                    },
                },
            };
        },
        externalSystemDefaults() {
            return {
                registryName: '',
                registryNameExists: false,
                registryNameEditable: false,
                urlValidationFailed: false,
                formShow: false,
                deleteFormShow: false,
                startValidation: false,
                secretInputTypes: {
                    secret: 'text',
                    username: 'text',
                },
                tokenInputType: 'text',
                data: {
                    url: '',
                    protocol: "REST",
                    auth: {
                        type: 'NO_AUTH',
                        secret: '',
                        'auth-url': '',
                        'access-token-json-path': '',
                        username: '',
                    },
                },
            };
        },
        hideDeleteExternalSystemForm() {
            this.backdropShow = false;
            this.externalSystem.deleteFormShow = false;
            $("body").css("overflow", "scroll");
        },
        deleteExternalSystemLink() {
            return `/admin/registry/external-system-delete/${this.registryName}?external-system=${this.externalSystem.registryName}`
        },
        showDeleteExternalSystemForm(registry) {
            this.externalSystem.registryName = registry;
            this.backdropShow = true;
            this.externalSystem.deleteFormShow = true;
            $("body").css("overflow", "hidden");
        },
        checkForOpenMRs(e) {
            if (this.openMergeRequests.has) {
                e.preventDefault();
                this.showOpenMRForm();
            }
        },
        showOpenMRForm() {
            this.backdropShow = true;
            this.openMergeRequests.formShow = true;
            this.accordion = 'merge-requests';
            //todo: load data

            $("body").css("overflow", "hidden");
        },
        hideOpenMRForm(e) {
            e.preventDefault();
            this.openMergeRequests.formShow = false;
            this.backdropShow = false;
            $("body").css("overflow", "scroll");
        },
        externalSystemFormAction() {
            if (this.externalSystem.registryNameEditable) {
                return `/admin/registry/external-system-create/${this.registryName}`
            }

            return `/admin/registry/external-system/${this.registryName}`
        },
        setExternalSystemForm(e) {
            this.externalSystem.registryNameExists = false;
            this.externalSystem.startValidation = true;
            this.externalSystem.urlValidationFailed = false;

            if (this.externalSystem.data.url !== '' && !this.isURL(this.externalSystem.data.url)) {
                e.preventDefault();
                this.externalSystem.urlValidationFailed = true;
                return;
            }

            if (this.externalSystem.data.url === "") {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.data.auth.type !== 'NO_AUTH' &&
                (!this.externalSystem.data.auth.hasOwnProperty('secret') || this.externalSystem.data.auth['secret'] === '')) {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.data.auth.type === 'BASIC' && this.externalSystem.data.auth['username'] === '') {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.data.auth.type === 'AUTH_TOKEN+BEARER' &&
                (!this.externalSystem.data.auth.hasOwnProperty('auth-url') ||
                    this.externalSystem.data.auth['auth-url'] === '' ||
                    !this.externalSystem.data.auth.hasOwnProperty('access-token-json-path') ||
                    this.externalSystem.data.auth['access-token-json-path'] === '')) {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.registryNameEditable) {
                e.preventDefault();
                let $this = this;

                axios.get(`/admin/registry/external-system-check/${this.registryName}`,
                    {params: {"external-system": this.externalSystem.registryName}})
                    .then(function (response) {
                        $this.externalSystem.registryNameExists = true;
                    })
                    .catch(function (error) {
                        $("#external-system-form").submit();
                    });
            }
        },
        setTrembitaClientForm(e) {
            this.trembitaClient.startValidation = true;
            this.trembitaClient.urlValidationFailed = false;

            if (this.trembitaClient.data.url !== '' && !this.isURL(this.trembitaClient.data.url)) {
               this.trembitaClient.urlValidationFailed = true;
                e.preventDefault();
                return;
            }

            for (let i in this.trembitaClient.data) {
                if (typeof(this.trembitaClient.data[i]) == "string" && this.trembitaClient.data[i] === "") {
                    e.preventDefault();
                    return;
                }

            }

            for (let i in this.trembitaClient.data.client) {
                if (typeof(this.trembitaClient.data.client[i]) == "string" &&
                    this.trembitaClient.data.client[i] === "") {
                    e.preventDefault();
                    return;
                }
            }

            for (let i in this.trembitaClient.data.service) {
                if (typeof(this.trembitaClient.data.service[i]) == "string" &&
                    this.trembitaClient.data.service[i] === "") {
                    e.preventDefault();
                    return;
                }
            }

            if (this.trembitaClient.data.service.auth.type === 'AUTH_TOKEN' &&
                this.trembitaClient.data.service.auth['secret'] === '') {
                e.preventDefault();
            }
        },
        changeExternalSystemAuthType() {
            this.externalSystem.startValidation = false;
        },
        changeTrembitaClientAuthType() {
            if (this.trembitaClient.data.service.auth.type === 'AUTH_TOKEN' &&
                !this.trembitaClient.data.service.auth.hasOwnProperty('secret')) {
                this.trembitaClient.data.service.auth['secret'] = '';
            }

            if (this.trembitaClient.data.service.auth.type === 'NO_AUTH' &&
                this.trembitaClient.data.service.auth.hasOwnProperty('secret')) {
                delete this.trembitaClient.data.service.auth['secret']
            }
        },
        hideExternalSystemForm(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.externalSystem.formShow = false;
            $("body").css("overflow", "scroll");
        },
        hideTrembitaClientForm(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.trembitaClient.formShow = false;
            $("body").css("overflow", "scroll");
        },
        trembitaFormSecretFocus() {
            if (this.trembitaClient.tokenInputType === 'password') {
                this.trembitaClient.data.service.auth.secret = '';
                this.trembitaClient.tokenInputType = 'text';
            }
        },
        externalSystemSecretFocus(name) {
            if (this.externalSystem.secretInputTypes[name] === 'password') {
                this.externalSystem.data.auth[name] = '';
                this.externalSystem.secretInputTypes[name] = 'text';
            }
        },
        showExternalSystemForm(registry, e) {
            e.preventDefault();

            if (this.openMergeRequests.has) {
                this.showOpenMRForm();
                return;
            }

            this.externalSystem = this.externalSystemDefaults();

            if (registry === '') {
                this.externalSystem.registryNameEditable = true;
            }

            this.externalSystem.registryName = registry;
            this.backdropShow = true;
            this.externalSystem.formShow = true;

            this.mergeDeep(this.externalSystem.data, this.values.externalSystems[registry]);

            if (this.externalSystem.data.auth.hasOwnProperty('secret')) {
                this.externalSystem.secretInputTypes.secret = 'password';
            }

            if (this.externalSystem.data.auth.hasOwnProperty('username')) {
                this.externalSystem.secretInputTypes.username = 'password';
            }

            $("body").css("overflow", "hidden");
        },
        showTrembitaClientForm(registry, e) {
            e.preventDefault();

            if (this.openMergeRequests.has) {
                this.showOpenMRForm();
                return;
            }

            this.trembitaClient = this.trembitaClientDefaults();

            this.trembitaClient.registryName = registry;
            this.backdropShow = true;
            this.trembitaClient.formShow = true;

            this.mergeDeep(this.trembitaClient.data, this.values.trembita.registries[registry]);
            if (this.trembitaClient.data.service.auth.hasOwnProperty('secret')) {
                this.trembitaClient.tokenInputType = 'password';
            }

            $("body").css("overflow", "hidden");
        },
        isObject(item) {
            return (item && typeof item === 'object' && !Array.isArray(item));
        },
        mergeDeep(target, ...sources) {
            if (!sources.length) return target;
            const source = sources.shift();

            if (this.isObject(target) && this.isObject(source)) {
                for (const key in source) {
                    if (source[key] === null || source[key] === "") {
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
        hideMrView(e) {
            $("body").css("overflow", "scroll");
            this.backdropShow = false;
            this.mrView = false;
            e.preventDefault();

            let mrFrame = this.$refs.mrIframe;
            if (mrFrame.src !== mrFrame.contentWindow.location.href) {
                document.location.reload();
            }
        },
        showMrView(src, e) {
            this.mrView = true;
            this.backdropShow = true;
            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);
            this.mrSrc = src;
        },
        hideMrError(e) {
            this.mrError = false;
            this.backdropShow = false;

            $("body").css("overflow", "scroll");
            e.preventDefault();
        },
        showMrError(e) {
            this.mrError = true;
            this.backdropShow = true;

            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);
        },
        setInternalRegistryReg() {
            this.internalRegistryReg = true;
            this.externalSystemType = "internal-registry";
        },
        setExternalSystem() {
            this.internalRegistryReg = false;
            this.externalSystemType = "external-system";
        },
        showExternalReg(e) {
            if (hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }

            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);

            this.externalRegPopupShow = true;
            this.backdropShow = true;
        },
        hideExternalReg(e) {
            e.preventDefault();
            $("body").css("overflow", "scroll");
            this.externalRegPopupShow = false;
            this.internalRegistryReg = true;
            this.backdropShow = false;
            this.accessGrantError = false;
        },
        disabledLink(e) {
            e.preventDefault();
            return false;
        },
        disableExternalReg(name, _type, e) {
            e.preventDefault();

            if (hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }
            forceMR = true;

            this.systemToDisable = name;
            this.systemToDisableType = _type;
            $("#disable-form-value").val(name);
            $("#disable-form-type").val(_type);
            $("#disable-form").submit();

        },
        removeExternalReg(name, _type, e) {
            e.preventDefault();

            if (hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }

            this.systemToDelete = name;
            this.systemToDeleteType = _type;
            this.backdropShow = true;
            this.removeExternalRegPopupShow = true;
            window.scrollTo(0, 0);
            $("body").css("overflow", "hidden");
        },
        hideRemoveExternalReg(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.removeExternalRegPopupShow = false;
            $("body").css("overflow", "scroll");

        },
        showExternalKeyValue(e) {
            if (this.keyValue === '******') {
                this.keyValue = this.currentExternalKeyValue;
            } else {
                this.keyValue = '******'
            }

            e.preventDefault();
        },
        showExternalKey(name, keyValue, e) {
            e.preventDefault();
            this.backdropShow = true;
            this.externalKey = true;
            this.systemToShowKey = name;
            this.currentExternalKeyValue = keyValue;
        },
        hideExternalKey(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.externalKey = false;
            this.keyValue = '******'
        },
        addExternalReg(e) {
            let names = $(".ereg-name");
            let inputName = $("#ex-system").val().trim();
            for (let i=0;i<names.length;i++) {
                if ($(names[i]).html().trim() === inputName) {
                    this.accessGrantError = `Доступ з таким ім'ям системи/або платформи "${inputName}" вже існує, оберіть інше ім'я`;
                    e.preventDefault();
                }
            }
        },
    }
});

app.config.compilerOptions.delimiters = ['[[', ']]'];
app.mount('#registry-view');

$(document).ready(function() {
    //DataTable.datetime('DD.MM.YYYY h:mm');

    $("#mr-table").DataTable({
        ordering: true,
        paging: true,
        columnDefs: [
            { orderable: false, targets: 4 },
            {
                targets: 0,
                render: DataTable.render.datetime('DD.MM.YYYY h:mm'),
            },
        ],
        order: [[0, 'desc']],
        language: {
            "processing": "Зачекайте...",
            "lengthMenu": "Показати _MENU_ записів",
            "zeroRecords": "Записи відсутні.",
            "info": "Записи з _START_ по _END_ із _TOTAL_ записів",
            "infoEmpty": "Записи з 0 по 0 із 0 записів",
            "infoFiltered": "(відфільтровано з _MAX_ записів)",
            "search": "Пошук:",
            "paginate": {
                "first": "Перша",
                "previous": "Попередня",
                "next": "Наступна",
                "last": "Остання"
            },
            "aria": {
                "sortAscending": ": активувати для сортування стовпців за зростанням",
                "sortDescending": ": активувати для сортування стовпців за спаданням"
            }
        }
    });
});
