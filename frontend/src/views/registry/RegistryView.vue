<script setup lang="ts">
import { inject } from 'vue';
interface RegistryTemplateVariables {
    openMergeRequests: boolean;
    registry: any;
    allowedToEdit: any;
    hasUpdate: any;
    regisrtyAdministrationComponents: any;
    registryOperationalComponents: any;
    platformAdministrationComponents: any;
    platformOperationalComponents: any;
    externalRegAvailableRegistriesJSON: any;
    admins: any;
    citizenPortalHost: any;
    officerPortalHost: any;
    smtpType: any;
    officerCIDR: any;
    citizenCIDR: any;
    adminCIDR: any;
    values: any;
    externalRegs: any;
    branches: any;
    mergeRequests: any;
    created: any;
    valuesJson: any;
    gerritURL: string;
    jenkinsURL: string;
    mrAvailable: string;
}
const variables = inject('TEMPLATE_VARIABLES') as RegistryTemplateVariables;

const openMergeRequests = variables?.openMergeRequests;
const registry = variables?.registry;
const allowedToEdit = variables?.allowedToEdit;
const hasUpdate = variables?.hasUpdate;
const regisrtyAdministrationComponents = variables?.regisrtyAdministrationComponents;
const registryOperationalComponents = variables?.registryOperationalComponents;
const platformAdministrationComponents = variables?.platformAdministrationComponents;
const platformOperationalComponents = variables?.platformOperationalComponents;
const externalRegAvailableRegistriesJSON = variables?.externalRegAvailableRegistriesJSON;
const admins = variables?.admins;
const citizenPortalHost = variables?.citizenPortalHost;
const officerPortalHost = variables?.officerPortalHost;
const smtpType = variables?.smtpType;
const officerCIDR = variables?.officerCIDR;
const citizenCIDR = variables?.citizenCIDR;
const adminCIDR = variables?.adminCIDR;
const values = variables?.values;
const externalRegs = variables?.externalRegs;
const branches = variables?.branches;
const mergeRequests = variables?.mergeRequests;
const created = variables?.created;
const gerritURL = variables?.gerritURL;
const jenkinsURL = variables?.jenkinsURL;
const mrAvailable = variables?.mrAvailable;
</script>

<script lang="ts">
import $ from 'jquery';
import axios from 'axios';
import { getGerritURL, getImageUrl, getJenkinsURL } from '@/utils';
import MergeRequestsTable from '@/components/MergeRequestsTable.vue';
import { defineComponent } from 'vue';

export default defineComponent({
    data() {
        return {
            forceMR: false,
            mergeRequest: {
                has: false,
                formShow: false,
            },
            backdropShow: false,
            accordion: 'general',
            mrView: false,
            externalKey: false,
            systemToShowKey: '',
            mrSrc: '',
            keyValue: '******',
            currentExternalKeyValue: '',
            removeExternalRegPopupShow: false,
            systemToDelete: '',
            accessGrantError: false,
            externalRegPopupShow: false,
            internalRegistryReg: true,
            externalSystemType: "internal-registry",
            mrError: false,
            systemToDeleteType: '',
            systemToDisable: '',
            systemToDisableType: '',
            trembitaClient: {} as any,
            externalSystem: {} as any,
            valuesData: {},
            registryName: '',
            activeTab: 'info',
            externalRegAvailableRegistriesNames: [],
            registrySelected: false,
        };
    },
    methods: {
        hasNewMergeRequests() {
            if (this.forceMR) {
                return true;
            }

            let statuses = $(".mr-status");
            for (let i = 0; i < statuses.length; i++) {
                if ($(statuses[i]).html().trim() === "NEW") {
                    return true;
                }
            }

            return false;
        },
        hideOpenMRForm(e: any) {
            e.preventDefault();
            this.mergeRequest.formShow = false;
            this.backdropShow = false;
            $("body").css("overflow", "scroll");
        },
        checkForOpenMRs(e: any) {
            if (this.mergeRequest.has) {
                e.preventDefault();
                this.showOpenMRForm();
            }
        },
        showOpenMRForm() {
            this.backdropShow = true;
            this.mergeRequest.formShow = true;
            this.accordion = 'merge-requests';
            //todo: load data

            $("body").css("overflow", "hidden");
        },
        disabledLink(e: any) {
            e.preventDefault();
            return false;
        },
        disableExternalReg(name: any, type: any, e: any) {
            e.preventDefault();

            if (this.hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }
            this.forceMR = true;

            this.systemToDisable = name;
            this.systemToDisableType = type;
            $('#disable-form-value').val(name);
            $("#disable-form-type").val(type);
            $("#disable-form").submit();

        },
        hideMrView(e: any) {
            $("body").css("overflow", "scroll");
            this.backdropShow = false;
            this.mrView = false;
            e.preventDefault();

            let mrFrame = this.$refs.mrIframe;
            if ((mrFrame as any).src !== (mrFrame as any).contentWindow.location.href) {
                document.location.reload();
            }
        },
        showExternalKeyValue(e: any) {
            if (this.keyValue === '******') {
                this.keyValue = this.currentExternalKeyValue;
            } else {
                this.keyValue = '******';
            }

            e.preventDefault();
        },
        hideExternalKey(e: any) {
            e.preventDefault();
            this.backdropShow = false;
            this.externalKey = false;
            this.keyValue = '******';
        },
        showMrError(e: any) {
            this.mrError = true;
            this.backdropShow = true;

            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);
        },
        removeExternalReg(name: string, _type: string, e: any) {
            e.preventDefault();

            if (this.hasNewMergeRequests()) {
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
        hideRemoveExternalReg(e: any) {
            e.preventDefault();
            this.backdropShow = false;
            this.removeExternalRegPopupShow = false;
            $("body").css("overflow", "scroll");

        },
        showExternalSystemForm(registry: string, e: any) {
            e.preventDefault();

            if (this.mergeRequest.has) {
                this.showOpenMRForm();
                return;
            }

            this.externalSystem = this.externalSystemDefaults();

            if (registry === '') {
                (this.externalSystem as any).registryNameEditable = true;
            }

            (this.externalSystem as any).registryName = registry;
            this.backdropShow = true;
            (this.externalSystem as any).formShow = true;

            this.mergeDeep((this.externalSystem as any).data, (this.valuesData as any).externalSystems[registry]);

            // eslint-disable-next-line no-prototype-builtins
            if ((this.externalSystem as any).data.auth.hasOwnProperty('secret')) {
                (this.externalSystem as any).secretInputTypes.secret = 'password';
            }

            // eslint-disable-next-line no-prototype-builtins
            if ((this.externalSystem as any).data.auth.hasOwnProperty('username') && this.externalSystem.data.auth.username !== '') {
                (this.externalSystem as any).secretInputTypes.username = 'password';
            }

            if (registry === 'diia') {
                (this.externalSystem as any).data.auth.type = 'AUTH_TOKEN+BEARER';
            }

            // eslint-disable-next-line no-prototype-builtins
            if (this.externalSystem.data.auth.hasOwnProperty('type') && this.externalSystem.data.auth.type === 'BASIC') {
                this.externalSystem.usernamePlaceholder = 'Завантаження...';

                axios.get(`/admin/registry/get-basic-username/${this.registryName}`,
                    { params: { "registry-name": this.externalSystem.registryName } })
                    .then((response) => {
                        this.externalSystem.data.auth.username = response.data;
                        this.externalSystem.usernamePlaceholder = '';
                    });

            }


            $("body").css("overflow", "hidden");
        },
        addExternalReg(e: any) {
            let names = $(".ereg-name");

            if (this.internalRegistryReg) {
                const selected = this.registrySelected;

                for (let i = 0; i < names.length; i++) {
                    if (($(names[i]) as any).html().trim() === selected) {
                        this.accessGrantError = `Доступ з таким ім'ям "${selected}" вже існує. Для вирішення конфлікту імен перестворіть доступ до зовнішньої системи з іншим ім'ям, а потім надайте доступ реєстру платформи: "${selected}` as unknown as boolean;
                        e.preventDefault();
                    }
                }
            }

            let inputName = ($("#ex-system") as any).val().trim() as never;

            if (this.externalRegAvailableRegistriesNames.includes(inputName)) {
                this.accessGrantError = `Доступ з таким ім'ям системи/або платформи "${inputName}" вже існує, оберіть інше ім'я` as unknown as boolean;
                e.preventDefault();
            }

            for (let i = 0; i < names.length; i++) {
                if ($(names[i]).html().trim() === inputName) {
                    this.accessGrantError = `Доступ з таким ім'ям системи/або платформи "${inputName}" вже існує, оберіть інше ім'я` as unknown as boolean;
                    e.preventDefault();
                }
            }
        },
        hideExternalReg(e: any) {
            e.preventDefault();
            $("body").css("overflow", "scroll");
            this.externalRegPopupShow = false;
            this.internalRegistryReg = true;
            this.backdropShow = false;
            this.accessGrantError = false;
        },
        showMrView(src: string) {
            this.mrView = true;
            this.backdropShow = true;
            $("body").css("overflow", "hidden");
            window.scrollTo(0, 0);
            this.mrSrc = src;
        },
        setInternalRegistryReg() {
            this.internalRegistryReg = true;
            this.externalSystemType = "internal-registry";
        },
        setExternalSystem() {
            this.internalRegistryReg = false;
            this.externalSystemType = "external-system";
        },
        hideMrError(e: any) {
            this.mrError = false;
            this.backdropShow = false;

            $("body").css("overflow", "scroll");
            e.preventDefault();
        },
        showExternalKey(name: string, keyValue: string, e: any) {
            e.preventDefault();
            this.backdropShow = true;
            this.externalKey = true;
            this.systemToShowKey = name;
            this.currentExternalKeyValue = keyValue;
        },
        showExternalReg(e: any) {
            if (this.mergeRequest.has) {
                this.showOpenMRForm();
                return;
            }

            if (this.hasNewMergeRequests()) {
                this.showMrError(e);
                return;
            }

            $("body").css("overflow", "hidden");
            e.preventDefault();
            window.scrollTo(0, 0);

            this.externalRegPopupShow = true;
            this.backdropShow = true;
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
                usernamePlaceholder: '',
                data: {
                    mock: false,
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
        trembitaClientDefaults() {
            return {
                registryName: '',
                registryNameExists: false,
                formShow: false,
                deleteFormShow: false,
                registryCreation: false,
                startValidation: false,
                tokenInputType: 'text',
                urlValidationFailed: false,
                data: {
                    mock: false,
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
                        serviceCode: '',
                        serviceVersion: '',
                    },
                    auth: {
                        type: 'NO_AUTH',
                        secret: '',
                    },
                },
            };

        },
        trembitaClientFormAction() {
            if (this.trembitaClient.registryCreation) {
                return `/admin/registry/trembita-client-create/${this.registryName}`;
            }
            return `/admin/registry/trembita-client/${this.registryName}`;
        },
        isSystemRegistry() {
            return this.trembitaClient.registryName === 'idp-exchange-service-registry' ||
                this.trembitaClient.registryName === 'dracs-registry' ||
                this.trembitaClient.registryName === 'edr-registry';
        },
        showDeleteTrembitaClientForm(registry: any, _type: any, e: any) {
            e.preventDefault();
            if (_type === 'platform') {
                return;
            }

            this.trembitaClient.registryName = registry;
            this.backdropShow = true;
            this.trembitaClient.deleteFormShow = true;
            $("body").css("overflow", "hidden");
        },

        deleteTrembitaClientLink() {
            return `/admin/registry/trembita-client-delete/${this.registryName}?trembita-client=${this.trembitaClient.registryName}`;
        },
        showTrembitaClientForm(registry: string, e: any) {
            e.preventDefault();

            if (this.mergeRequest.has) {
                this.showOpenMRForm();
                return;
            }

            this.trembitaClient = this.trembitaClientDefaults();

            if (registry === '') {
                this.trembitaClient.registryCreation = true;
            }


            (this.trembitaClient as any).registryName = registry;
            this.backdropShow = true;
            (this.trembitaClient as any).formShow = true;

            this.mergeDeep((this.trembitaClient as any).data, (this.valuesData as any).trembita.registries[registry]);
            // eslint-disable-next-line no-prototype-builtins
            if ((this.trembitaClient as any).data.auth.hasOwnProperty('secret')) {
                (this.trembitaClient as any).tokenInputType = 'password';
            }

            if (!this.trembitaClient.data.auth.type || (this.trembitaClient.data.type && this.trembitaClient.data.type !== 'registry')) {
                if (registry === 'idp-exchange-service-registry' || registry === 'dracs-registry') {
                    this.trembitaClient.data.auth.type = 'NO_AUTH';
                } else {
                    this.trembitaClient.data.auth.type = 'AUTH_TOKEN';
                }
            }


            $("body").css("overflow", "hidden");
        },
        showDeleteExternalSystemForm(registry: string, _type: string, e: any) {
            e.preventDefault();
            if (_type === 'platform') {
                return;
            }

            (this.externalSystem as any).registryName = registry;
            this.backdropShow = true;
            (this.externalSystem as any).deleteFormShow = true;
            $("body").css("overflow", "hidden");
        },
        isObject(item: any) {
            return (item && typeof item === 'object' && !Array.isArray(item));
        },
        mergeDeep(target: { [x: string]: any; }, ...sources: any[]): any {
            if (!sources.length) return target;
            const source = sources.shift();

            if (this.isObject(target) && this.isObject(source)) {
                for (const key in source) {
                    if (source[key] === null || source[key] === "") {
                        continue;
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
        selectTab(tabName: string) {
            this.activeTab = tabName;
        },
        isActiveTab(tabName: string) {
            return this.activeTab === tabName;
        },
        getIconName(url: string) {
            if (!url) {
                return "triangle-exclamation";
            }

            return "circle-check";
        },
        getType(type: string) {
            if (type === 'platform') {
                return "Системний";
            }
            return "Реєстровий";

        },
        getAuth(auth: any) {
            if (auth?.type) {
                return auth.type;
            }
            return '-';
        },
        getStatus(status: any) {
            switch (`status-${status}`) {
                case "status-active":
                case "status-SUCCESS":
                case "status-ok":
                    return "Активний";
                case "status-failed":
                case "status-failure":
                case "status-FAILURE":
                    return "Помилка";
                case "status-inactive":
                    return "В обробці";
            }
        },
        getExtStatus(status: string, enabled: boolean) {
            if (status === "") {
                return "status-active";
            }
            if (!enabled) {
                return "status-disabled";
            }
            return `status-${status}`;
        },
        inactive(status: string) {
            return status === "inactive" || status === "failed";
        },
        hideTrembitaClientForm(e: any) {
            e.preventDefault();
            this.backdropShow = false;
            this.trembitaClient.formShow = false;
            $("body").css("overflow", "scroll");
        },
        isURL(u: string) {
            return /^https?:\/\/(www\.)?[-a-zA-Z0-9@:%._\\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\\+.~#?&//=]*)$/.test(u);
        },
        setTrembitaClientForm(e: any) {
            this.trembitaClient.registryNameExists = false;
            this.trembitaClient.startValidation = true;
            this.trembitaClient.urlValidationFailed = false;

            // eslint-disable-next-line no-prototype-builtins
            if (this.trembitaClient.data.hasOwnProperty('url') && this.trembitaClient.data.url !== ''
                && !this.isURL(this.trembitaClient.data.url)) {
                this.trembitaClient.urlValidationFailed = true;
                e.preventDefault();
                return;
            }

            for (let i in this.trembitaClient.data) {
                if (typeof (this.trembitaClient.data[i]) == "string" && this.trembitaClient.data[i] === "") {
                    e.preventDefault();
                    return;
                }

            }

            for (let i in this.trembitaClient.data.client) {
                if (typeof (this.trembitaClient.data.client[i]) == "string" &&
                    this.trembitaClient.data.client[i] === "") {
                    e.preventDefault();
                    return;
                }
            }

            for (let i in this.trembitaClient.data.service) {
                if (typeof (this.trembitaClient.data.service[i]) == "string" &&
                    this.trembitaClient.data.service[i] === "" && i !== "serviceCode" && i !== "serviceVersion") {
                    e.preventDefault();
                    return;
                }
            }

            if (this.trembitaClient.data.auth.type === 'AUTH_TOKEN' &&
                this.trembitaClient.data.auth['secret'] === '') {
                e.preventDefault();
            }

            if (this.trembitaClient.registryCreation) {
                e.preventDefault();

                axios.get(`/admin/registry/trembita-client-check/${this.registryName}`,
                    { params: { "trembita-client": this.trembitaClient.registryName } })
                    .then(() => {
                        this.trembitaClient.registryNameExists = true;
                    })
                    .catch(function () {
                        $("#trembita-client-form").submit();
                    });
            }

        },
        trembitaFormSecretFocus() {
            if (this.trembitaClient.tokenInputType === 'password') {
                this.trembitaClient.data.auth.secret = '';
                this.trembitaClient.tokenInputType = 'text';
            }
        },
        changeTrembitaClientAuthType() {
            if (this.trembitaClient.data.auth.type === 'AUTH_TOKEN' &&
                // eslint-disable-next-line no-prototype-builtins
                !this.trembitaClient.data.auth.hasOwnProperty('secret')) {
                this.trembitaClient.data.auth['secret'] = '';
            }

            if (this.trembitaClient.data.auth.type === 'NO_AUTH' &&
                // eslint-disable-next-line no-prototype-builtins
                this.trembitaClient.data.auth.hasOwnProperty('secret')) {
                delete this.trembitaClient.data.auth['secret'];
            }
        },
        hideExternalSystemForm(e: any) {
            e.preventDefault();
            this.backdropShow = false;
            this.externalSystem.formShow = false;
            $("body").css("overflow", "scroll");
        },
        mockChanged(dataIndex: 'externalSystem' | 'trembitaClient') {
            let data = this[dataIndex].data;

            if (data.mock) {
                delete data['url'];
            } else {
                data['url'] = '';
            }
        },
        setExternalSystemForm(e: any) {
            this.externalSystem.registryNameExists = false;
            this.externalSystem.startValidation = true;
            this.externalSystem.urlValidationFailed = false;

            if (!this.externalSystem.registryName) {
                e.preventDefault();
                return;
            }

            // eslint-disable-next-line no-prototype-builtins
            if (this.externalSystem.data.url.hasOwnProperty('url') && this.externalSystem.data.url !== '' && !this.isURL(this.externalSystem.data.url)) {
                e.preventDefault();
                this.externalSystem.urlValidationFailed = true;
                return;
            }

            // eslint-disable-next-line no-prototype-builtins
            if (this.externalSystem.data.url.hasOwnProperty('url') && this.externalSystem.data.url === "") {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.data.auth.type !== 'NO_AUTH' &&
                // eslint-disable-next-line no-prototype-builtins
                (!this.externalSystem.data.auth.hasOwnProperty('secret') || this.externalSystem.data.auth['secret'] === '')) {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.data.auth.type === 'BASIC' && this.externalSystem.data.auth['username'] === '') {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.data.auth.type === 'AUTH_TOKEN+BEARER' &&
                // eslint-disable-next-line no-prototype-builtins
                (!this.externalSystem.data.auth.hasOwnProperty('auth-url') ||
                    this.externalSystem.data.auth['auth-url'] === '' ||
                    // eslint-disable-next-line no-prototype-builtins
                    !this.externalSystem.data.auth.hasOwnProperty('access-token-json-path') ||
                    this.externalSystem.data.auth['access-token-json-path'] === '')) {
                e.preventDefault();
                return;
            }

            if (this.externalSystem.registryNameEditable) {
                e.preventDefault();

                axios.get(`/admin/registry/external-system-check/${this.registryName}`,
                    { params: { "external-system": this.externalSystem.registryName } })
                    .then(() => {
                        this.externalSystem.registryNameExists = true;
                    })
                    .catch(function () {
                        $("#external-system-form").submit();
                    });
            }
        },
        externalSystemFormAction() {
            if (this.externalSystem.registryNameEditable) {
                return `/admin/registry/external-system-create/${this.registryName}`;
            }

            return `/admin/registry/external-system/${this.registryName}`;
        },
        externalSystemSecretFocus(name: string) {
            if (this.externalSystem.secretInputTypes[name] === 'password') {
                this.externalSystem.data.auth[name] = '';
                this.externalSystem.secretInputTypes[name] = 'text';
            }
        },
        hideDeleteForm() {
            this.backdropShow = false;
            this.externalSystem.deleteFormShow = false;
            this.trembitaClient.deleteFormShow = false;
            $("body").css("overflow", "scroll");

        },
        deleteExternalSystemLink() {
            return `/admin/registry/external-system-delete/${this.registryName}?external-system=${this.externalSystem.registryName}`;
        },
        changeExternalSystemAuthType() {
            this.externalSystem.startValidation = false;
        },
        mockAvailable() {
            return (this.valuesData as any)?.global?.deploymentMode === "development";
        },
    },
    mounted() {

        // eslint-disable-next-line no-prototype-builtins
        if (this.$refs.hasOwnProperty('valuesJson')) {
            this.valuesData = JSON.parse((this.$refs.valuesJson as any).value);
        }

        // eslint-disable-next-line no-prototype-builtins
        if (this.$refs.hasOwnProperty('registryName')) {
            this.registryName = (this.$refs.registryName as any).value;
        }

        // eslint-disable-next-line no-prototype-builtins
        if (this.$refs.hasOwnProperty('refOpenMergeRequests')) {
            this.mergeRequest.has = true;
        }

        // eslint-disable-next-line no-prototype-builtins
        if (this.$refs.hasOwnProperty('externalRegistries')) {
            this.externalRegAvailableRegistriesNames =
                JSON.parse((this.$refs.externalRegistries as any)?.value || '[]')?.map((reg: { metadata: { name: any; }; }) => reg?.metadata?.name) || [];
            if (this.externalRegAvailableRegistriesNames.length) {
                this.registrySelected = this.externalRegAvailableRegistriesNames[0];
            }
        }

        this.externalSystem = this.externalSystemDefaults();
        this.trembitaClient = this.trembitaClientDefaults();
    },
    components: { MergeRequestsTable },
});
</script>

<template>
    <div class="registry" id="registry-view">
        <input type="hidden" :value="variables.valuesJson" ref="valuesJson" />
        <template v-if="openMergeRequests">
            <input type="hidden" :value="openMergeRequests" ref="refOpenMergeRequests" />
            <div class="popup-window admin-window visible" v-cloak v-if="mergeRequest.formShow">
                <div class="popup-header">
                    <p>Дія недоступна</p>
                    <a href="#" @click="hideOpenMRForm" class="popup-close hide-popup">
                        <img alt="close popup window" src="@/assets/img/close.png" />
                    </a>
                </div>
                <div class="popup-body">
                    <p>Реєстр має не підтверджені запити на оновлення.</p>
                </div>
                <div class="popup-footer active">
                    <a href="#" id="admin-cancel" class="hide-popup" @click="hideOpenMRForm">Закрити</a>
                </div>
            </div>
        </template>
        <input type="hidden" :value="registry.metadata.name" ref="registryName" />
        <input type="hidden" :value="externalRegAvailableRegistriesJSON" ref="externalRegistries" />
        <div class="registry-header">
            <a href="/admin/registry/overview" class="registry-add">
                <img alt="add registry" src="@/assets/img/action-back.png" />
                <span>НАЗАД</span>
            </a>
        </div>
        <div class="registry-header registry-header-view">
            <h1>Реєстр {{ registry.metadata.name }}</h1>
            <template v-if="allowedToEdit">
                <div class="registry-view-actions">
                    <template v-if="hasUpdate">
                        <a :href="`/admin/registry/update/${registry.metadata.name}`" @click="checkForOpenMRs"
                            class="registry-add">
                            <i class="fa-solid fa-arrow-up"></i>
                            <span>Оновити</span>
                        </a>
                    </template>
                    <a :href="`/admin/registry/edit/${registry.metadata.name}`" @click="checkForOpenMRs"
                        class="registry-add">
                        <img alt="add registry" src="@/assets/img/action-edit.png" />
                        <span>Редагувати</span>
                    </a>
                </div>
            </template>
        </div>
        <div class="tabs">
            <div class="tab" @click="selectTab('info')" :class="{ active: isActiveTab('info') }">
                Інформація про реєстр
            </div>
            <div class="tab" @click="selectTab('links')" :class="{ active: isActiveTab('links') }">
                Швидкі посилання
            </div>
        </div>
        <div class="box" v-show="isActiveTab('info')">
            <div class="rg-info-block">
                <div class="rg-info-block-header" :class="{ 'border-bottom': accordion != 'general' }"
                    @click="accordion = 'general'">
                    <span>Загальна інформація</span>
                    <i class="fa-solid"
                        :class="{ 'fa-caret-up': accordion == 'general', 'fa-caret-down': accordion != 'general' }">
                    </i>
                </div>
                <div class="rg-info-block-body" v-show="accordion == 'general'">
                    <div class="rg-info-line-horizontal">
                        <span>Назва</span>
                        <span>{{ registry.metadata.name }}</span>
                    </div>
                    <div v-if="registry.spec.description" class="rg-info-line-horizontal">
                        <span>Опис</span>
                        <span>{{ registry.spec.description }}</span>
                    </div>
                    <div v-if="admins" class="rg-info-line-horizontal">
                        <span>Адміністратори</span>
                        <span class="cidr-values">
                            <div class="view-cidr" v-for="$adm in admins" :key="$adm.email">{{ $adm.email }}</div>
                        </span>
                    </div>
                    <div class="rg-info-line-horizontal">
                        <span>Час створення</span>
                        <span>{{ created }}</span>
                    </div>
                    <div class="rg-info-line-horizontal" v-if="citizenPortalHost">
                        <span>DNS ім’я для портала громадянина</span>
                        <span>{{ citizenPortalHost }}</span>
                    </div>

                    <div class="rg-info-line-horizontal" v-if="officerPortalHost">
                        <span>DNS ім’я для портала чиновника</span>
                        <span>{{ officerPortalHost }}</span>
                    </div>
                    <div class="rg-info-line-horizontal" v-if="smtpType">
                        <span>Поштовий сервер</span>
                        <span v-if="smtpType == 'external'">
                            Зовнішній поштовий сервер
                        </span>
                        <span v-else>
                            Платформенний поштовий сервер
                        </span>

                    </div>

                    <div class="rg-info-line-horizontal" v-if="officerCIDR">
                        <span>CIDR для портала чиновника</span>
                        <span class="cidr-values">
                            <div class="view-cidr" v-for="cidr in officerCIDR" :key="cidr">{{ cidr }}</div>
                        </span>
                    </div>
                    <div class="rg-info-line-horizontal" v-if="citizenCIDR">
                        <span>CIDR для портала громадянина</span>
                        <span class="cidr-values">
                            <div class="view-cidr" v-for="cidr in citizenCIDR" :key="cidr">{{ cidr }}</div>
                        </span>
                    </div>
                    <div class="rg-info-line-horizontal" v-if="adminCIDR">
                        <span>CIDR для адміністративних компонент</span>
                        <span class="cidr-values">
                            <div class="view-cidr" v-for="cidr in adminCIDR" :key="cidr">{{ cidr }}</div>
                        </span>
                    </div>
                </div>
            </div>
            <div class="rg-info-block">
                <div class="rg-info-block-header" :class="{ 'border-bottom': accordion != 'trembita-client' }"
                    @click="accordion = 'trembita-client'">
                    <span>налаштування взаємодії з реєстрами через Трембіту</span>
                    <i class="fa-solid"
                        :class="{ 'fa-caret-up': accordion == 'trembita-client', 'fa-caret-down': accordion != 'trembita-client' }"></i>
                </div>
                <div class="rg-info-block-body" v-show="accordion == 'trembita-client'">
                    <table class="rg-info-table rg-info-table-config">
                        <thead>
                            <tr>
                                <th>Статус</th>
                                <th>Назва</th>
                                <th>Рівень</th>
                                <th>Протокол інтеграції</th>
                                <th>Тип автентифікації</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="($al, $index) in values.trembita.registries" :key="$index">
                                <td>
                                    <i :class="`fa-solid fa-${getIconName($al.url)}`"></i>
                                </td>
                                <td>{{ $index }}</td>
                                <td>{{ getType($al.type) }}</td>
                                <td>{{ $al.protocol }}</td>
                                <td>{{ getAuth($al.auth) }}</td>
                                <td>
                                    <div class="trembita-actions">
                                        <a href="#" @click="showTrembitaClientForm(String($index), $event)">
                                            <i class="fa-solid fa-pen"></i>
                                        </a>
                                        <a :title="`${$al.type == 'platform' ? 'Система недоступна для видалення' : ''}`"
                                            href="#" @click="showDeleteTrembitaClientForm($index, $al.type, $event)">
                                            <i class="fa-solid fa-trash registry-trash"
                                                :class="{ inactive: $al.type == 'platform' }"></i>
                                        </a>
                                    </div>

                                </td>
                            </tr>
                        </tbody>
                    </table>
                    <div class="link-grant-access">
                        <a class="" href="#" @click="showTrembitaClientForm('', $event)">
                            <img alt="Додати реєстр" src="@/assets/img/plus.png" />
                            <span>Додати реєстр</span>
                        </a>
                    </div>


                    <!--trembita-client-form start-->

                    <div id="trembita-client-popup" class="popup-window admin-window visible scrollable-popup" v-cloak
                        v-if="trembitaClient.formShow">
                        <div class="popup-header">
                            <p v-if="!trembitaClient.registryCreation">Налаштувати взаємодію з реєстром через Трембіту</p>
                            <p v-if="trembitaClient.registryCreation">Додати взаємодію з реєстром через Трембіту</p>

                            <a href="#" @click="hideTrembitaClientForm" class="popup-close hide-popup">
                                <img alt="close popup window" src="@/assets/img/close.png" />
                            </a>
                        </div>
                        <div v-if="!trembitaClient.registryCreation" class="view-registry-warning">
                            Налаштування взаємодії доступне для версій реєстру {{ trembitaClient.data.type === 'platform' ? '1.9.3' : '1.9.4' }} і вище.
                        </div>
                        <div v-if="trembitaClient.registryCreation" class="view-registry-warning">
                            Додавання взаємодії доступне для версій реєстру 1.9.4 і вище.
                        </div>
                        <form @submit="setTrembitaClientForm" id="trembita-client-form" method="post"
                            :action="trembitaClientFormAction()">
                            <div class="popup-body">
                                <div class="popup-body-header">Налаштування ШБО Трембіти</div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.protocolVersion == '' && trembitaClient.startValidation }">
                                    <label>Версія протоколу</label>
                                    <input aria-label="" name="trembita-client-protocol-version" type="text"
                                        v-model="trembitaClient.data.protocolVersion" />
                                    <span
                                        v-if="trembitaClient.data.protocolVersion == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div v-if="mockAvailable()" class="rc-form-group">
                                    <label class="form-checkbox-container">
                                        Використати мок зовнішньої інтеграції (для версій реєстру 1.9.4 і вище)
                                        <input @change="mockChanged('trembitaClient')" v-model="trembitaClient.data.mock"
                                            type="checkbox">
                                        <span class="form-checkbox-checkmark"></span>
                                    </label>
                                </div>

                                <div v-if="!trembitaClient.data.mock" class="rc-form-group"
                                    :class="{ 'error': (trembitaClient.data.url == '' || trembitaClient.urlValidationFailed) && trembitaClient.startValidation }">
                                    <label>Адреса ШБО Трембіти</label>
                                    <input aria-label="" type="text" name="trembita-client-url"
                                        v-model="trembitaClient.data.url" />
                                    <p>URL, повинен починатись з http:// або https://</p>
                                    <span v-if="trembitaClient.data.url == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                    <span
                                        v-if="trembitaClient.urlValidationFailed && trembitaClient.startValidation">Невірний
                                        формат</span>
                                </div>

                                <div class="popup-body-header">Налаштування клієнта Трембіти</div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.userId == '' && trembitaClient.startValidation }">
                                    <label>Ідентифікатор клієнту</label>
                                    <input aria-label="" type="text" name="trembita-client-user-id"
                                        v-model="trembitaClient.data.userId" />
                                    <span v-if="trembitaClient.data.userId == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.client.xRoadInstance == '' && trembitaClient.startValidation }">
                                    <label>X-Road Instance</label>
                                    <input aria-label="" type="text" name="trembita-client-x-road-instance"
                                        v-model="trembitaClient.data.client.xRoadInstance" />
                                    <span
                                        v-if="trembitaClient.data.client.xRoadInstance == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.client.memberClass == '' && trembitaClient.startValidation }">
                                    <label>Member Class</label>
                                    <input aria-label="" type="text" name="trembita-client-member-class"
                                        v-model="trembitaClient.data.client.memberClass" />
                                    <span
                                        v-if="trembitaClient.data.client.memberClass == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.client.memberCode == '' && trembitaClient.startValidation }">
                                    <label>Member Code</label>
                                    <input aria-label="" type="text" name="trembita-client-member-code"
                                        v-model="trembitaClient.data.client.memberCode" />
                                    <span
                                        v-if="trembitaClient.data.client.memberCode == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.client.subsystemCode == '' && trembitaClient.startValidation }">
                                    <label>Subsystem Code</label>
                                    <input aria-label="" type="text" name="trembita-client-subsystem-code"
                                        v-model="trembitaClient.data.client.subsystemCode" />
                                    <span
                                        v-if="trembitaClient.data.client.subsystemCode == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="popup-body-header">Налаштування сервісу для інтеграції</div>

                                <div class="rc-form-group"
                                    :class="{ 'error': (trembitaClient.registryName == '' || trembitaClient.registryNameExists) && trembitaClient.startValidation }">

                                    <label>Службова назва реєстру</label>
                                    <input aria-label="" type="text" name="trembita-client-regitry-name"
                                        :readonly="!trembitaClient.registryCreation"
                                        v-model="trembitaClient.registryName" />
                                    <p v-if="trembitaClient.registryCreation">Назва не може бути змінена після додавання
                                        інтеграції!</p>
                                    <span v-if="trembitaClient.registryName == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                    <span v-if="trembitaClient.registryNameExists">
                                        Зовнішня система з таким ім'ям вже існує
                                    </span>
                                </div>


                                <div class="rc-form-group">
                                    <label>Протокол інтеграції</label>
                                    <input aria-label="" type="text" readonly name="trembita-client-protocol"
                                        v-model="trembitaClient.data.protocol" />
                                    <p>Наразі підтримується лише SOAP-протокол інтеграції.</p>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.service.xRoadInstance == '' && trembitaClient.startValidation }">
                                    <label>X-Road Instance</label>
                                    <input aria-label="" type="text" name="trembita-service-x-road-instance"
                                        v-model="trembitaClient.data.service.xRoadInstance" />
                                    <span
                                        v-if="trembitaClient.data.service.xRoadInstance == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.service.memberClass == '' && trembitaClient.startValidation }">
                                    <label>Member Class</label>
                                    <input aria-label="" type="text" name="trembita-service-member-class"
                                        v-model="trembitaClient.data.service.memberClass" />
                                    <span
                                        v-if="trembitaClient.data.service.memberClass == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.service.memberCode == '' && trembitaClient.startValidation }">
                                    <label>Member Code</label>
                                    <input aria-label="" type="text" name="trembita-service-member-code"
                                        v-model="trembitaClient.data.service.memberCode" />
                                    <span
                                        v-if="trembitaClient.data.service.memberCode == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': trembitaClient.data.service.subsystemCode == '' && trembitaClient.startValidation }">
                                    <label>Subsystem Code</label>
                                    <input aria-label="" type="text" name="trembita-service-subsystem-code"
                                        v-model="trembitaClient.data.service.subsystemCode" />
                                    <span
                                        v-if="trembitaClient.data.service.subsystemCode == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div v-if="!isSystemRegistry()" class="rc-form-group">
                                    <label>Service Code</label>
                                    <input aria-label="" type="text" name="trembita-service-service-code"
                                        v-model="trembitaClient.data.service.serviceCode" />
                                    <p>Необов'язковий параметр</p>
                                </div>

                                <div v-if="!isSystemRegistry()" class="rc-form-group">
                                    <label>Service Version</label>
                                    <input aria-label="" type="text" name="trembita-service-service-version"
                                        v-model="trembitaClient.data.service.serviceVersion" />
                                    <p>Необов'язковий параметр</p>
                                </div>


                                <div class="rc-form-group">
                                    <label>Вкажіть тип автентифікації</label>
                                    <select aria-label="" name="trembita-service-auth-type"
                                        v-model="trembitaClient.data.auth.type" @change="changeTrembitaClientAuthType">
                                        <option v-if="trembitaClient.registryName == 'idp-exchange-service-registry' ||
                                            trembitaClient.registryName == 'dracs-registry' ||
                                            trembitaClient.registryCreation || trembitaClient.data.type == 'registry'">
                                            NO_AUTH

                                        </option>

                                        <option v-if="trembitaClient.registryName == 'edr-registry' || trembitaClient.registryCreation
                                            || trembitaClient.data.type == 'registry'">AUTH_TOKEN</option>

                                    </select>
                                </div>

                                <div class="rc-form-group" v-if="trembitaClient.data.auth.type == 'AUTH_TOKEN'"
                                    :class="{ 'error': trembitaClient.data.auth.secret == '' && trembitaClient.startValidation }">
                                    <label>Вкажіть токен авторизації</label>
                                    <input aria-label="" name="trembita-service-auth-secret"
                                        v-model="trembitaClient.data.auth.secret" :type="trembitaClient.tokenInputType"
                                        @focus="trembitaFormSecretFocus" />
                                    <span v-if="trembitaClient.data.auth.secret == '' && trembitaClient.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="popup-footer active">
                                    <a href="#" id="admin-cancel" class="hide-popup"
                                        @click="hideTrembitaClientForm">відмінити</a>
                                    <button value="submit" class="submit-green" name="admin-apply" type="submit">
                                        <span v-if="!trembitaClient.registryCreation">Підтвердити</span>
                                        <span v-if="trembitaClient.registryCreation">Додати</span>
                                    </button>

                                </div>
                            </div>
                        </form>
                    </div>
                    <div class="popup-window admin-window visible" v-cloak v-if="trembitaClient.deleteFormShow">
                        <div class="popup-header">
                            <p>Видалити "{{ trembitaClient.registryName }}"?</p>
                            <a href="#" @click="hideDeleteForm" class="popup-close hide-popup">
                                <img alt="close popup window" src="@/assets/img/close.png" />
                            </a>
                        </div>
                        <div class="popup-body">
                            <p>Видалити усі налаштування інтеграціїї з реєстром?</p>
                        </div>
                        <div class="popup-footer active">
                            <a href="#" id="admin-cancel" class="hide-popup" @click="hideDeleteForm">Відмінити</a>
                            <a class="href-red" :href="deleteTrembitaClientLink()">Видалити</a>
                        </div>
                    </div>

                    <!--trembita-client-form end-->



                </div>
            </div>
            <div class="rg-info-block">
                <div class="rg-info-block-header" :class="{ 'border-bottom': accordion != 'external-systems' }"
                    @click="accordion = 'external-systems'">
                    <span>налаштування взаємодії з іншими системами</span>
                    <i class="fa-solid"
                        :class="{ 'fa-caret-up': accordion == 'external-systems', 'fa-caret-down': accordion != 'external-systems' }"></i>
                </div>
                <div class="rg-info-block-body" v-show="accordion == 'external-systems'">
                    <table class="rg-info-table rg-info-table-config">
                        <thead>
                            <tr>
                                <th>Статус</th>
                                <th>Назва</th>
                                <th>Рівень</th>
                                <th>Протокол інтеграції</th>
                                <th>Тип автентифікації</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="($al, $index) in values.externalSystems" :key="$index">
                                <td>
                                    <i :class="`fa-solid fa-${getIconName($al.url)}`"></i>
                                </td>
                                <td>{{ $index }}</td>
                                <td>{{ getType($al.type) }}</td>
                                <td>{{ $al.protocol }}</td>
                                <td>{{ getAuth($al.auth) }}</td>
                                <td>
                                    <div class="trembita-actions">
                                        <a href="#" @click="showExternalSystemForm(String($index), $event)">
                                            <i class="fa-solid fa-pen"></i>
                                        </a>
                                        <a :title="`${$al.type == 'platform' ? 'Система недоступна для видалення' : ''}`"
                                            href="#"
                                            @click="showDeleteExternalSystemForm(String($index), String($al.type), $event)">
                                            <i class="fa-solid fa-trash registry-trash"
                                                :class="{ inactive: $al.type == 'platform' }"></i>
                                        </a>
                                    </div>
                                </td>
                            </tr>

                        </tbody>
                    </table>
                    <div class="link-grant-access">
                        <a class="" href="#" @click="showExternalSystemForm('', $event)">
                            <img alt="Додати зовнішню систему" src="@/assets/img/plus.png" />
                            <span>Додати зовнішню систему</span>
                        </a>
                    </div>
                </div>

                <!-- registry-external-system start-->
                <div id="external-system-popup" class="popup-window admin-window visible scrollable-popup" v-cloak
                    v-if="externalSystem.formShow">
                    <div class="popup-header">
                        <p v-if="!externalSystem.registryNameEditable">Налаштувати зовнішню систему для взаємодії</p>
                        <p v-if="externalSystem.registryNameEditable">Додати зовнішню систему для взаємодії</p>
                        <a href="#" @click="hideExternalSystemForm" class="popup-close hide-popup">
                            <img alt="close popup window" src="@/assets/img/close.png" />
                        </a>
                    </div>
                    <p class="title-description" v-if="externalSystem.registryNameEditable">
                        Ви можете налаштувати інтеграцію з зовнішньою системою для подальшої взаємодії згідно регламенту
                        реєстру.
                        Мережеві політики доступу будуть створені автоматично.
                    </p>
                    <div class="view-registry-warning">
                        Налаштування взаємодії доступне для версій реєстру 1.9.3 і вище.
                    </div>
                    <form @submit="setExternalSystemForm" id="external-system-form" method="post"
                        :action="externalSystemFormAction()">
                        <div class="popup-body">
                            <div class="rc-form-group"
                                :class="{ 'error': (externalSystem.registryName == '' || externalSystem.registryNameExists) && externalSystem.startValidation }">
                                <label>Назва зовнішньої системи</label>
                                <input aria-label="" name="external-system-registry-name" type="text"
                                    v-model="externalSystem.registryName"
                                    :readonly="!externalSystem.registryNameEditable" />
                                <p v-if="externalSystem.registryNameEditable">Назва не може бути змінена після додавання
                                    інтеграції!</p>
                                <span v-if="externalSystem.registryName == '' && externalSystem.startValidation">
                                    Обов’язкове поле
                                </span>
                                <span v-if="externalSystem.registryNameExists">
                                    Зовнішня система з таким ім'ям вже існує
                                </span>
                            </div>

                            <div v-if="mockAvailable()" class="rc-form-group">
                                <label class="form-checkbox-container">
                                    Використати мок зовнішньої інтеграції (для версій реєстру 1.9.4 і вище)
                                    <input @change="mockChanged('externalSystem')" v-model="externalSystem.data.mock"
                                        type="checkbox">
                                    <span class="form-checkbox-checkmark"></span>
                                </label>
                            </div>

                            <div v-if="!externalSystem.data.mock" class="rc-form-group"
                                :class="{ 'error': (externalSystem.data.url == '' || externalSystem.urlValidationFailed) && externalSystem.startValidation }">
                                <label>Адреса зовнішньої системи</label>
                                <input aria-label="" type="text" name="external-system-url"
                                    v-model="externalSystem.data.url" />
                                <p>URL, повинен починатись з http:// або https://</p>
                                <span v-if="externalSystem.data.url == '' && externalSystem.startValidation">
                                    Обов’язкове поле
                                </span>
                                <span v-if="externalSystem.urlValidationFailed && externalSystem.startValidation">Невірний
                                    формат</span>

                            </div>

                            <div class="rc-form-group"
                                :class="{ 'error': externalSystem.data.protocol == '' && externalSystem.startValidation }">
                                <label>Протокол інтеграції</label>
                                <input aria-label="" type="text" name="external-system-protocol"
                                    v-model="externalSystem.data.protocol" readonly />
                                <span v-if="externalSystem.data.protocol == '' && externalSystem.startValidation">
                                    Обов’язкове поле
                                </span>
                                <p>Наразі підтримується лише REST-протокол інтеграції.</p>
                            </div>

                            <div class="rc-form-group">
                                <label>Вкажіть тип автентифікації</label>
                                <select aria-label="" name="external-system-auth-type"
                                    v-model="externalSystem.data.auth.type" @change="changeExternalSystemAuthType">
                                    <option v-if="externalSystem.registryName != 'diia'">NO_AUTH</option>
                                    <option v-if="externalSystem.registryName != 'diia'">AUTH_TOKEN</option>
                                    <option v-if="externalSystem.registryName != 'diia'">BEARER</option>
                                    <option v-if="externalSystem.registryName != 'diia'">BASIC</option>

                                    <option>AUTH_TOKEN+BEARER</option>
                                </select>
                                <p>Наразі підтримуються:
                                    <template v-if="externalSystem.registryName != 'diia'">NO_AUTH, AUTH_TOKEN, BEARER,
                                        BASIC,</template>
                                    AUTH_TOKEN+BEARER
                                </p>
                            </div>

                            <div class="auth-token-bearer" v-if="externalSystem.data.auth.type == 'AUTH_TOKEN+BEARER'">
                                <div class="rc-form-group"
                                    :class="{ 'error': externalSystem.data.auth['auth-url'] == '' && externalSystem.startValidation }">
                                    <label>Вкажіть енд-поінт автентифікації партнера</label>
                                    <input aria-label="" type="text" name="external-system-auth-url"
                                        v-model="externalSystem.data.auth['auth-url']" />
                                    <p>Вказується абсолютна адреса (https://example.ua/auth) або relative path відносно
                                        адреси,
                                        вказаної в полі Адреса зовнішньої системи (/auth)</p>
                                    <span
                                        v-if="externalSystem.data.auth['auth-url'] == '' && externalSystem.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': externalSystem.data.auth['access-token-json-path'] == '' && externalSystem.startValidation }">
                                    <label>Вкажіть json-path для отримання токена доступу</label>
                                    <input aria-label="" type="text" name="external-system-auth-access-token-json-path"
                                        v-model="externalSystem.data.auth['access-token-json-path']" />
                                    <span
                                        v-if="externalSystem.data.auth['access-token-json-path'] == '' && externalSystem.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                            </div>

                            <div class="rc-form-group"
                                v-if="externalSystem.data.auth.type == 'AUTH_TOKEN' || externalSystem.data.auth.type == 'BEARER' || externalSystem.data.auth.type == 'AUTH_TOKEN+BEARER'"
                                :class="{ 'error': externalSystem.data.auth.secret == '' && externalSystem.startValidation }">
                                <label>Вкажіть токен авторизації</label>
                                <input aria-label="" name="external-system-auth-secret"
                                    v-model="externalSystem.data.auth.secret" :type="externalSystem.secretInputTypes.secret"
                                    @focus="externalSystemSecretFocus('secret')" />
                                <span v-if="externalSystem.data.auth.secret == '' && externalSystem.startValidation">
                                    Обов’язкове поле
                                </span>
                            </div>

                            <div v-if="externalSystem.data.auth.type == 'BASIC'">
                                <div class="rc-form-group"
                                    :class="{ 'error': externalSystem.data.auth.username == '' && externalSystem.startValidation }">
                                    <label>Логін</label>
                                    <input aria-label="" name="external-system-auth-username"
                                        @focus="externalSystemSecretFocus('username')"
                                        v-model="externalSystem.data.auth.username"
                                        :type="externalSystem.secretInputTypes.username"
                                        :placeholder="externalSystem.usernamePlaceholder" />
                                    <span v-if="externalSystem.data.auth.username == '' && externalSystem.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>

                                <div class="rc-form-group"
                                    :class="{ 'error': externalSystem.data.auth.secret == '' && externalSystem.startValidation }">
                                    <label>Пароль</label>
                                    <input aria-label="" name="external-system-auth-secret"
                                        v-model="externalSystem.data.auth.secret"
                                        :type="externalSystem.secretInputTypes.secret"
                                        @focus="externalSystemSecretFocus('secret')" />
                                    <span v-if="externalSystem.data.auth.secret == '' && externalSystem.startValidation">
                                        Обов’язкове поле
                                    </span>
                                </div>
                            </div>
                        </div>
                        <div class="popup-footer active">
                            <a href="#" id="admin-cancel" class="hide-popup" @click="hideExternalSystemForm">відмінити</a>
                            <button class="submit-green" value="submit" name="admin-apply" type="submit">
                                <span v-if="!externalSystem.registryNameEditable">Підтвердити</span>
                                <span v-if="externalSystem.registryNameEditable">Додати</span>
                            </button>
                        </div>
                    </form>
                </div>


                <div class="popup-window admin-window visible" v-cloak v-if="externalSystem.deleteFormShow">
                    <div class="popup-header">
                        <p>Видалити "{{ externalSystem.registryName }}"?</p>
                        <a href="#" @click="hideDeleteForm" class="popup-close hide-popup">
                            <img alt="close popup window" src="@/assets/img/close.png" />
                        </a>
                    </div>
                    <div class="popup-body">
                        <p>Видалити усі налаштування інтеграції з системою?</p>
                    </div>
                    <div class="popup-footer active">
                        <a href="#" id="admin-cancel" class="hide-popup" @click="hideDeleteForm">Відмінити</a>
                        <a class="href-red" :href="deleteExternalSystemLink()">Видалити</a>
                    </div>
                </div>



                <!-- registry-external-system end-->




            </div>
            <div class="rg-info-block">
                <div class="rg-info-block-header" :class="{ 'border-bottom': accordion != 'external-access' }"
                    @click="accordion = 'external-access'">
                    <span>Доступ для реєстрів платформи та зовнішніх систем</span>
                    <i class="fa-solid"
                        :class="{ 'fa-caret-up': accordion == 'external-access', 'fa-caret-down': accordion != 'external-access' }"></i>
                </div>
                <div class="rg-info-block-body" v-show="accordion == 'external-access'">
                    <table class="rg-info-table rg-info-table-config" v-if="externalRegs && externalRegs.length">
                        <thead>
                            <tr>
                                <th>Статус</th>
                                <th>Назва</th>
                                <th>Тип системи</th>
                                <th></th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="($er, $index) in externalRegs" :key="$index">
                                <td>
                                    <img :alt="getStatus($er.StatusRegistration)"
                                        :src="getImageUrl(`status-${$er.StatusRegistration}`)"
                                        :title="getStatus($er.StatusRegistration)" />
                                </td>
                                <td class="ereg-name">
                                    {{ $er.Name }}
                                </td>
                                <td>
                                    <span v-if="$er.External"> Зовнішня система</span>

                                    <span v-else> Реєстр платформи</span>

                                </td>
                                <td>
                                    <div class="rg-external-system-actions"
                                        :class="{ inactive: getExtStatus($er.StatusRegistration, $er.Enabled) == 'status-inactive' }">

                                        <a v-if="$er.External"
                                            @click="
                                                getExtStatus($er.StatusRegistration, $er.Enabled) === 'status-active' ? showExternalKey(String($er.Name), String($er.KeyValue), $event) : disabledLink"
                                            href="#">
                                            <img title="Перевірити пароль" alt="key"
                                                :src="getImageUrl(`key-status-${$er.StatusRegistration}`)" />
                                        </a>


                                        <a :status="inactive($er.StatusRegistration)"
                                            @click="inactive($er.StatusRegistration) ? disabledLink : disableExternalReg($er.Name, $er.TypeStr, $event)"
                                            href="#">
                                            <img :title="$er.StatusRegistration == ' status-disabled' ? 'Розблокувати доступ' : 'Заблокувати доступ'"
                                                alt="key" :src="getImageUrl(`lock-status-${$er.StatusRegistration}`)" />
                                        </a>
                                        <a @click="inactive($er.StatusRegistration) ? disabledLink : removeExternalReg($er.Name, $er.TypeStr, $event)"
                                            href="#">
                                            <img title="Скасувати доступ" alt="key"
                                                :src="getImageUrl(`disable-status-${$er.StatusRegistration}`)" />
                                        </a>
                                    </div>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                    <div class="rg-info-block-no-content" v-else>
                        Немає реєстрів або систем, що мають доступ до цього реєстра.
                    </div>
                    <div class="link-grant-access">
                        <a class="" href="#" @click="showExternalReg">
                            <img alt="Надати доступ" src="@/assets/img/plus.png" />
                            <span>Надати доступ</span>
                        </a>
                    </div>
                </div>
            </div>

            <form id="disable-form" method="POST"
                :action="`/admin/registry/external-reg-disable/${registry.metadata.name}`">
                <input type="hidden" id="disable-form-value" :value="systemToDisable" name="reg-name" />
                <input type="hidden" id="disable-form-type" :value="systemToDisableType" name="external-system-type" />
            </form>

            <div class="rg-info-block" v-if="branches && branches.length">
                <div class="rg-info-block-header" :class="{ 'border-bottom': accordion != 'configuration' }"
                    @click="accordion = 'configuration'">
                    <span>Конфігурація</span>
                    <i class="fa-solid"
                        :class="{ 'fa-caret-up': accordion == 'configuration', 'fa-caret-down': accordion != 'configuration' }"></i>
                </div>
                <div class="rg-info-block-body" v-show="accordion == 'configuration'">
                    <table class="rg-info-table rg-info-table-config">
                        <thead>
                            <tr>
                                <th>Статус</th>
                                <th>Конфігурація</th>
                                <th>VCS</th>
                                <th>CI</th>
                                <th>Версія</th>
                                <th>Номер збірки</th>
                                <th>Остання вдала збірка</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="($br, $index) in branches" :key="$index">
                                <td>
                                    <img :title="getStatus($br.status.value)" :alt="getStatus($br.status.value)"
                                        :src="getImageUrl(`status-${$br.status.value}`)" />
                                </td>
                                <td>
                                    {{ $br.metadata.name }}
                                </td>
                                <td>
                                    <a :href="getGerritURL(gerritURL)" target="_blank">
                                        <img alt="vcs" src="@/assets/img/action-link.png" />
                                    </a>
                                </td>
                                <td>
                                    <a :href="getJenkinsURL(jenkinsURL, $br.spec.codebaseName, $br.spec.branchName)"
                                        target="_blank">
                                        <img alt="ci" src="@/assets/img/action-link.png" />
                                    </a>
                                </td>
                                <td>{{ $br.spec.version }}</td>
                                <td>{{ $br.status.build || '-' }}</td>
                                <td><span v-if="$br.status.lastSuccessfulBuild">{{ $br.status.lastSuccessfulBuild
                                }}</span><span v-else>-</span>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
            <div class="rg-info-block">
                <div class="rg-info-block-header" :class="{ 'border-bottom': accordion != 'merge-requests' }"
                    @click="accordion = 'merge-requests'">
                    <span>Запити на оновлення</span>
                    <i class="fa-solid"
                        :class="{ 'fa-caret-up': accordion == 'merge-requests', 'fa-caret-down': accordion != 'merge-requests' }">
                    </i>
                </div>
                <div class="rg-info-block-body mr-block-table" v-show="accordion == 'merge-requests'">
                    <template v-if="mergeRequests && mergeRequests.length">
                        <MergeRequestsTable :merge-requests="mergeRequests" @onViewClick="showMrView" :mr-available="mrAvailable"></MergeRequestsTable>
                    </template>
                    <div class="rg-info-block-no-content" v-else>
                        Запитів немає
                    </div>
                </div>
            </div>
        </div>


        <div class="box" v-show="isActiveTab('links')">
            <template v-if="regisrtyAdministrationComponents">
                <div class="rg-info-block">
                    <div class="rg-info-block-header">
                        <span>Адміністративна зона реєстру</span>
                    </div>
                    <div class="rg-info-block-body mr-block-table">
                        <div class="dashboard-panel registry-dashboard-panel">
                            <div class="list-item" v-for="$ec in regisrtyAdministrationComponents" :key="$ec.Url">
                                <img :src="`data:image/svg+xml;base64,${$ec.Icon}`" :alt="`${$ec.Type} logo`"
                                    class="item-image" />
                                <div class="item-content">
                                    <a target="_blank" :href="$ec.Url">
                                        {{ $ec.Title }}
                                        <img src="@/assets/img/action-link.png" :alt="`${$ec.Type} link`">
                                    </a>
                                    <div class="description">{{ $ec.Description }}</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
            <template v-if="registryOperationalComponents">
                <div class="rg-info-block">
                    <div class="rg-info-block-header">
                        <span>Операційна зона реєстру</span>
                    </div>
                    <div class="rg-info-block-body mr-block-table">
                        <div class="dashboard-panel registry-dashboard-panel">
                            <div class="list-item" v-for="$ec in registryOperationalComponents" :key="$ec.Url">
                                <img :src="`data:image/svg+xml;base64,${$ec.Icon}`" :alt="`${$ec.Type} logo`"
                                    class="item-image" />
                                <div class="item-content">
                                    <a target="_blank" :href="$ec.Url" :class="{ disabled: $ec.Visible == 'false' }">
                                        {{ $ec.Title }}
                                        <span v-if="$ec.Visible == 'false'">(вимкнено)</span>
                                        <img v-else src="@/assets/img/action-link.png" :alt="`${$ec.Type} link`">
                                    </a>
                                    <div class="description">{{ $ec.Description }}</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </template>

            <template v-if="platformAdministrationComponents">
                <div class="rg-info-block">
                    <div class="rg-info-block-header">
                        <span>Адміністративна зона платформи</span>
                    </div>
                    <div class="rg-info-block-body mr-block-table">
                        <div class="dashboard-panel registry-dashboard-panel">
                            <div class="list-item" v-for="$ec in platformAdministrationComponents" :key="$ec.Url">
                                <img :src="`data:image/svg+xml;base64,${$ec.Icon}`" :alt="`${$ec.Type} logo`"
                                    class="item-image" />
                                <div class="item-content">
                                    <a target="_blank" :href="$ec.Url">
                                        {{ $ec.Title }}
                                        <img src="@/assets/img/action-link.png" :alt="`${$ec.Type} link`">
                                    </a>
                                    <div class="description">{{ $ec.Description }}</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
            <template v-if="platformOperationalComponents">
                <div class="rg-info-block">
                    <div class="rg-info-block-header">
                        <span>Операційна зона платформи</span>
                    </div>
                    <div class="rg-info-block-body mr-block-table">
                        <div class="dashboard-panel registry-dashboard-panel">
                            <div class="list-item" v-for="$ec in platformOperationalComponents" :key="$ec.Url">
                                <img :src="`data:image/svg+xml;base64,${$ec.Icon}`" :alt="`${$ec.Type} logo`"
                                    class="item-image" />
                                <div class="item-content">
                                    <a target="_blank" :href="$ec.Url">
                                        {{ $ec.Title }}
                                        <img src="@/assets/img/action-link.png" :alt="`${$ec.Type} link`">
                                    </a>
                                    <div class="description">{{ $ec.Description }}</div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </template>
        </div>
        <div class="popup-backdrop visible" v-cloak v-if="backdropShow"></div>

        <div style="width:80%;left:10%;height:80%;" class="popup-window admin-window visible" v-cloak v-if="mrView">
            <div class="popup-header">
                <p>Запит на оновлення</p>
                <a href="#" @click="hideMrView" class="popup-close hide-popup">
                    <img alt="close popup window" src="@/assets/img/close.png" />
                </a>
            </div>
            <div class="popup-body mr-frame-body" style="border-bottom: none;">
                <iframe ref="mrIframe" id="mr-frame" :src="mrSrc" style="width:100%;"></iframe>
            </div>
        </div>

        <div class="popup-window admin-window visible" v-cloak v-if="externalKey">
            <div class="popup-header">
                <p>Перевірити пароль для "{{ systemToShowKey }}"</p>
                <a href="#" @click="hideExternalKey" class="popup-close hide-popup">
                    <img alt="close popup window" src="@/assets/img/close.png" />
                </a>
            </div>
            <form method="POST" :action="`/admin/registry/external-reg-remove/${registry.metadata.name}`">
                <div class="popup-body">
                    Не передавайте пароль стороннім особам.
                    <div class="er-ex-key">
                        <span id="key-value">{{ keyValue }}</span>
                        <a @click="showExternalKeyValue" href="#">
                            <img title="Показати" v-cloak v-show="keyValue == '******'" alt="display password"
                                src="@/assets/img/eye.png" />
                            <img title="Приховати" v-cloak v-show="keyValue != '******'" alt="display password"
                                style="height:20px;margin-top:2px;" src="@/assets/img/hide-eye.png" />
                        </a>
                    </div>
                </div>
                <div class="popup-footer active">
                    <a href="#" class="hide-popup" @click="hideExternalKey">закрити</a>
                </div>
            </form>
        </div>

        <div class="popup-window admin-window visible" v-cloak v-if="mrError">
            <div class="popup-header">
                <p>Помилка</p>
                <a href="#" @click="hideMrError" class="popup-close hide-popup">
                    <img alt="close popup window" src="@/assets/img/close.png" />
                </a>
            </div>
            <div class="popup-body" style="border-bottom: none;">
                Наразі у вас є відкриті запити на оновлення. Підтвердіть або відхиліть зміни щоб продовжити.
            </div>
        </div>

        <div class="popup-window admin-window visible" v-cloak v-if="removeExternalRegPopupShow">
            <div class="popup-header">
                <p>Видалити "{{ systemToDelete }} " з переліку ?</p>
                <a href="#" @click="hideRemoveExternalReg" class="popup-close hide-popup">
                    <img alt="close popup window" src="@/assets/img/close.png" />
                </a>
            </div>
            <form method="POST" :action="`/admin/registry/external-reg-remove/${registry.metadata.name}`">
                <div class="popup-body">
                    Ви зможете надати доступ знову пізніше.
                </div>
                <input type="hidden" :value="systemToDelete" name="reg-name" />
                <input type="hidden" id="delete-form-type" :value="systemToDeleteType" name="external-system-type" />
                <div class="popup-footer active">
                    <a href="#" class="hide-popup" @click="hideRemoveExternalReg">відмінити</a>
                    <button value="submit" @click="addExternalReg" type="submit">Видалити</button>
                </div>
            </form>
        </div>

        <div id="external-reg-popup" class="popup-window admin-window external-reg-window visible" v-cloak
            v-if="externalRegPopupShow">
            <div class="popup-header">
                <p>Надати доступ</p>
                <a href="#" @click="hideExternalReg" class="popup-close hide-popup">
                    <img alt="close popup window" src="@/assets/img/close.png" />
                </a>
            </div>
            <form @submit="addExternalReg" id="external-reg-form" method="post"
                :action="`/admin/registry/external-reg-add/${registry.metadata.name}`">
                <div class="popup-body">
                    <p class="popup-error" v-cloak v-if="accessGrantError">{{ accessGrantError }} </p>

                    <p>
                        Ви можете надати доступ до даних цього реєстру (master) іншим реєстрам на цій платформі або
                        зовнішнім системам (clients). Для цього в мастер-реєстрі буде створено окремого користувача
                        реєстра-клієнта, від імені якого здійснюватеметься доступ до мастер-реєстру.
                    </p>
                    <div class="er-radio">
                        <div @click="setInternalRegistryReg">
                            <span class="er-radio-button" :class="{ selected: internalRegistryReg }"></span>
                            <span class="er-radio-text">Внутрішній реєстр платформи</span>
                        </div>
                        <div @click="setExternalSystem">
                            <span class="er-radio-button" :class="{ selected: !internalRegistryReg }"></span>
                            <span class="er-radio-text">Зовнішня система</span>
                        </div>
                        <input type="hidden" :value="externalSystemType" name="external-system-type" />
                    </div>
                    <div v-if="internalRegistryReg" class="er-system-opts">
                        <label for="in-registry">Оберіть реєстр</label>
                        <select name="reg-name" id="in-registry" v-model="registrySelected">
                            <option v-for="(reg, index) in externalRegAvailableRegistriesNames" :key="index">
                                {{ reg }}
                            </option>
                        </select>

                        <span>Якщо реєстру не має в переліку – його потрібно створити заздалегідь.</span>
                    </div>
                    <div v-if="!internalRegistryReg" class="er-system-opts">
                        <label for="ex-system">Назва системи</label>
                        <input required name="reg-name" id="ex-system" placeholder="Введіть назву" maxlength="32"
                            pattern="^[a-z0-9][-a-z0-9]+?[a-z0-9]$" />
                        <span>Допустимі символи: “a-z”, 0-9, “-”. Назва не може перевищувати довжину у 32 символів.
                            Назва повинна починатись і закінчуватися символами латинського алфавіту або цифрами.</span>
                        <label>Пароль доступу</label>
                        <p>Пароль буде створено автоматично. Його можна буде перевірити після налагодження доступу до
                            мастер-реєстру.</p>
                    </div>

                </div>
                <div class="popup-footer active">
                    <a href="#" id="external-reg-cancel" class="hide-popup" @click="hideExternalReg">відмінити</a>
                    <button value="submit" name="external-reg-apply" @click="addExternalReg" type="submit">Надати</button>
                </div>
            </form>
        </div>
    </div>
</template>
