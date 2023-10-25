<script setup lang="ts">
import { toRefs } from 'vue';

interface ClusterKeycloakBlockProps {
    keycloakHostname: any;
    clusterKeycloakDNSCustomHosts: any;
    clusterSettings: any;
    backdropShow: any;
    dnsManual: string;
}

const props = defineProps<ClusterKeycloakBlockProps>();
const { keycloakHostname, clusterKeycloakDNSCustomHosts, clusterSettings, backdropShow, dnsManual } = toRefs(props);

</script>
<script lang="ts">
export default {
    data() {
        return {
            disabled: false,
            disableClusterKeycloakDNS: false,
        };
    },
    methods: {
        submit() {
            this.disabled = true;
            this.$emit('submitKeycloakDNSForm');
        },
        editClusterKeycloakDNSHost(host: string, certificatePath: string) {
            this.$emit('editClusterKeycloakDNSHost', host, certificatePath);
        },
        checkClusterDeleteKeycloakDNS(host: string) {
            this.$emit('checkClusterDeleteKeycloakDNS', host);
        },
        showClusterKeycloakDNSForm() {
            this.$emit('showClusterKeycloakDNSForm');
        },
        hideCheckClusterDeleteKeycloakDNS() {
            this.$emit('hideCheckClusterDeleteKeycloakDNS');
        },
        deleteClusterKeycloakDNS(host: string) {
            window.localStorage.setItem('mr-scroll', 'true');
            this.$emit('deleteClusterKeycloakDNS', host);
        },
        hideClusterCheckKeycloakDNS() {
            this.$emit('hideClusterCheckKeycloakDNS');
        },
        hideClusterKeycloakDNSForm() {
            this.$emit('hideClusterKeycloakDNSForm');
        },
        addClusterKeycloakDNS() {
            this.disableClusterKeycloakDNS = true;
            this.$emit('addClusterKeycloakDNS');
        },
        resetClusterKeycloakDNSForm() {
            this.$emit('resetClusterKeycloakDNSForm');
        },
        clusterKeycloakDNSCertSelected() {
            this.$emit('clusterKeycloakDNSCertSelected');
        },
    },
    watch: {
        clusterSettings: {
            handler(newClusterSettings) {
                if (newClusterSettings.keycloak.pemError || newClusterSettings.keycloak.hostnameError) {
                    this.disableClusterKeycloakDNS = false;
                }
                if (!newClusterSettings.keycloak.formShow) {
                    this.disableClusterKeycloakDNS = false;
                    this.clusterSettings.keycloak.pemError = '';
                }
            },
            deep: true
        },
    }
};
</script>

<template>
    <h2>Keycloak DNS</h2>
    <br />
    <form @submit="submit" class="registry-create-form wizard-form cluster-keycloak" method="post"
        action="/admin/cluster/add-keycloak-dns">
        <p>{{ $t('domains.cluster.clusterKeycloak.text.setAdditionDnsForService') }}</p>

        <div class="keycloak-dns-manual">
            <a :href="dnsManual" target="_blank">{{ $t('domains.cluster.clusterKeycloak.text.instructionOutsideConfig') }}</a>
        </div>


        <div class="cluster-default-kc-dns-label">{{ $t('domains.cluster.clusterKeycloak.text.defaultDNS') }}</div>
        <div class="cluster-default-kc-dns-value">{{ keycloakHostname }}</div>

        <div v-for="h in clusterKeycloakDNSCustomHosts" :key="h.host">
            <div class="cluster-default-kc-dns-label">{{ $t('domains.cluster.clusterKeycloak.text.additionalDNS') }}</div>
            <div class="cluster-default-kc-dns-value">
                <span>{{ h.host }}</span>
                <div>
                    <a href="#" @click.stop.prevent="editClusterKeycloakDNSHost(h.host, h.certificatePath)">
                      <img :title="$t('actions.edit')" alt="pencil" src="@/assets/img/pencil.png" />
                    </a>
                    <a href="#" @click.stop.prevent="checkClusterDeleteKeycloakDNS(h.host)">
                      <img :title="$t('actions.remove')" class="img-trash" alt="trash" src="@/assets/img/trash.png" />
                    </a>
                </div>
            </div>
        </div>

        <div class="add-kc-dns-block">
            <a href="#" @click.stop.prevent="showClusterKeycloakDNSForm">
                <i class="fa-solid fa-plus"></i>
                <span>{{ $t('domains.cluster.clusterKeycloak.text.addDNS') }}</span>
            </a>
        </div>

        <input type="hidden" name="hostnames" v-model="clusterSettings.keycloak.submitInput" />
        <div class="rc-form-group">
            <button onclick="window.localStorage.setItem('mr-scroll', 'true');" type="submit" name="submit" :disabled="disabled">
                {{ $t('actions.confirm') }}
            </button>
        </div>
    </form>

    <div class="popup-backdrop visible" v-cloak v-if="backdropShow"></div>

    <div class="popup-window admin-window visible" v-cloak v-if="clusterSettings.keycloak.deleteHostname != ''">
        <div class="popup-header">
            <p>{{ $t('domains.cluster.clusterKeycloak.text.removeAdditionalDNS') }}</p>
            <a href="#" @click.stop.prevent="hideCheckClusterDeleteKeycloakDNS" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <div class="popup-body">
            <p>{{ $t('domains.cluster.clusterKeycloak.text.deleteHostname', { hostname: clusterSettings.keycloak.deleteHostname }) }}</p>
        </div>
        <div class="popup-footer active">
            <a href="#" class="hide-popup" @click.stop.prevent="hideCheckClusterDeleteKeycloakDNS">
                {{ $t('actions.cancel') }}
            </a>
            <button value="submit" name="cidr-apply"
                @click.stop.prevent="deleteClusterKeycloakDNS(clusterSettings.keycloak.deleteHostname)"
                type="submit">{{ $t('actions.confirm') }}</button>
        </div>
    </div>

    <div class="popup-window admin-window visible" v-cloak v-if="clusterSettings.keycloak.existHostname != ''">
        <div class="popup-header">
            <p>{{ $t('domains.cluster.clusterKeycloak.text.impossibleRemoveDNS') }}</p>
            <a href="#" @click.stop.prevent="hideClusterCheckKeycloakDNS" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <div class="popup-body">
            <p>{{ $t('domains.cluster.clusterKeycloak.text.domainIsUsedByRegistry', { existHostname: clusterSettings.keycloak.existHostname }) }}</p>
        </div>
        <div class="popup-footer active">
            <button class="submit-green" @click.stop.prevent="hideClusterCheckKeycloakDNS" name="admin-apply"
                type="submit">{{ $t('actions.gotIt') }}</button>
        </div>
    </div>

    <div class="popup-window admin-window visible" v-cloak v-if="clusterSettings.keycloak.formShow">
        <div class="popup-header">
            <p v-if="clusterSettings.keycloak.editHostname == ''">{{ $t('domains.cluster.clusterKeycloak.text.addDNS') }}</p>
            <p v-if="clusterSettings.keycloak.editHostname != ''">{{ $t('domains.cluster.clusterKeycloak.text.editDNS') }}</p>
            <a href="#" @click.stop.prevent="hideClusterKeycloakDNSForm" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <form @submit.prevent="addClusterKeycloakDNS" id="cluster-keycloak-dns-form" method="post" action="">
            <div class="popup-body">
                <div class="rc-form-group" :class="{ 'error': clusterSettings.keycloak.hostnameError != '' }">
                    <label for="cluster-keycloak-dns-value">{{ $t('domains.cluster.clusterKeycloak.text.domainNameKeycloak') }}</label>
                    <input id="cluster-keycloak-dns-value" maxlength="63" type="text"
                        v-model="clusterSettings.keycloak.hostname"
                        @blur="clusterSettings.keycloak.hostname = ($event.target as any).value.trim()"
                        :class="{ 'error': clusterSettings.keycloak.hostnameError != '' }"
                        :disabled="clusterSettings.keycloak.editDisabled" />
                        <p>{{ $t('domains.cluster.clusterKeycloak.text.clusterSettingsHostnameDescription') }}</p>
                    <span v-if="clusterSettings.keycloak.hostnameError != ''">{{ clusterSettings.keycloak.hostnameError
                    }}</span>
                </div>
                <div class="rc-form-group" :class="{ 'error': clusterSettings.keycloak.pemError != '' }">
                    <label>{{ $t('domains.cluster.clusterKeycloak.text.sslCertificateForKeycloak') }}</label>
                    <label v-show="!clusterSettings.keycloak.fileSelected" for="cluster-keycloak-dns-upload"
                        class="rc-form-upload-block">
                        <i class="fa-solid fa-plus"></i>
                        <span>{{ $t('domains.cluster.clusterKeycloak.text.uploadSSL') }}</span>
                    </label>
                    <div v-show="clusterSettings.keycloak.fileSelected" class="cluster-kc-dns-uploaded">
                        <div>
                            <i class="fa-solid fa-check"></i>
                            <span>{{ $t('domains.cluster.clusterKeycloak.text.fileUploaded') }}</span>
                        </div>
                        <a href="#" @click.stop.prevent="resetClusterKeycloakDNSForm"><i class="fa-solid fa-trash"></i></a>
                    </div>
                    <span v-if="clusterSettings.keycloak.pemError != ''">{{ clusterSettings.keycloak.pemError }}</span>
                    <input type="file" @change="clusterKeycloakDNSCertSelected" ref="clusterKeycloakDNS"
                        id="cluster-keycloak-dns-upload" style="display: none;" />
                </div>
            </div>
            <div class="popup-footer active">
                <a href="#" id="cidr-cancel" class="hide-popup"
                    @click.stop.prevent="hideClusterKeycloakDNSForm">{{ $t('actions.cancel') }}</a>
                <button value="submit" name="cidr-apply" onclick="window.localStorage.setItem('mr-scroll', 'true');" :disabled="disableClusterKeycloakDNS"
                    type="submit">{{ $t('actions.confirm') }}</button>
            </div>
        </form>
    </div>
</template>
