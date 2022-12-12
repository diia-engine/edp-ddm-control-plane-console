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

        if (this.$refs.hasOwnProperty('openMergeRequests')) {
            this.openMergeRequests.has = true;
        }
    },
    data() {
        return {
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
            externalSystem: {
                registryName: '',
                registryNameEditable: false,
                formShow: false,
                startValidation: false,
                tokenInputType: 'text',
                data: {
                    url: '',
                    protocol: "REST",
                    auth: {
                        type: 'NO_AUTH',
                        secret: '',
                        authUri: '',
                        accessTokenJSONPath: '',
                        username: '',
                    },
                },
            },
            trembitaClient: {
                registryName: '',
                formShow: false,
                startValidation: false,
                tokenInputType: 'text',
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
            },
        }
    }, // mrIframe
    methods: {
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
        setExternalSystemForm(e) {
            this.externalSystem.startValidation = true;

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
                (!this.externalSystem.data.auth.hasOwnProperty('authUri') ||
                    this.externalSystem.data.auth['authUri'] === '' ||
                    !this.externalSystem.data.auth.hasOwnProperty('accessTokenJSONPath') ||
                    this.externalSystem.data.auth['accessTokenJSONPath'] === '')) {
                e.preventDefault();
            }
        },
        setTrembitaClientForm(e) {
            this.trembitaClient.startValidation = true;
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
        externalSystemSecretFocus() {

        },
        showExternalSystemForm(registry, e) {
            e.preventDefault();
            this.externalSystem.registryName = registry;
            this.backdropShow = true;
            this.externalSystem.formShow = true;

            //todo: load data

            $("body").css("overflow", "hidden");
        },
        showTrembitaClientForm(registry, e) {
            e.preventDefault();

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
