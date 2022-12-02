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
            trembitaClient: {
                registryName: '',
                formShow: false,
                startValidation: false,
                data: {
                    protocol: "SOAP",
                    service: {
                        auth: { type: 'NO_AUTH' },
                    },
                },
            },
        }
    }, // mrIframe
    methods: {
        setTrembitaClientForm(e) {
            this.trembitaClient.startValidation = true;
        },
        hideTrembitaClientForm(e) {
            e.preventDefault();
            this.backdropShow = false;
            this.trembitaClient.formShow = false;
            $("body").css("overflow", "scroll");
        },
        showTrembitaClientForm(registry, e) {
            e.preventDefault();

            this.trembitaClient.registryName = registry;
            this.backdropShow = true;
            this.trembitaClient.formShow = true;

            console.log(this.values.trembita.registries[registry]);
            this.mergeDeep(this.trembitaClient.data, this.values.trembita.registries[registry]);

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
