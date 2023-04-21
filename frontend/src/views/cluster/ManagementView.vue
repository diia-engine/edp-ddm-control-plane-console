<script setup lang="ts">
import { inject } from 'vue';

interface ManagementTemplateVariables {
    canUpdateCluster: any;
    codebase: any;
    version: any;
    admins: any;
    cidr: any;
    branches: any;
    mergeRequests: any;
    edpComponents: any;
    jenkinsURL: any;
    gerritURL: any;
}
const variables = inject('TEMPLATE_VARIABLES') as ManagementTemplateVariables;
const canUpdateCluster = variables?.canUpdateCluster;
const codebase = variables?.codebase;
const version = variables?.version;
const admins = variables?.admins;
const cidr = variables?.cidr;
const branches = variables?.branches;
const mergeRequests = variables?.mergeRequests;
const edpComponents = variables?.edpComponents;
const jenkinsURL = variables?.jenkinsURL;
const gerritURL = variables?.gerritURL;
</script>
<script lang="ts">
import $ from 'jquery';
import { getFormattedDate, getGerritURL, getImageUrl, getJenkinsURL, getMergeRequestPlatformAction, getMergeRequestStatus, getStatus } from '@/utils';
import MergeRequestsTable from '@/components/MergeRequestsTable.vue';

export default {
    data() {
        return {
            backdropShow: false,
            mrView: false,
            mrSrc: '',
        };
    },
    methods: {
        showMrView(src: string) {
            this.mrView = true;
            this.backdropShow = true;
            $("body").css("overflow", "hidden");
            window.scrollTo(0, 0);
            this.mrSrc = src;
        },
        hideMrView() {
            $("body").css("overflow", "scroll");
            this.backdropShow = false;
            this.mrView = false;
            let mrFrame = this.$refs.mrIframe;
            if ((mrFrame as any).src !== (mrFrame as any).contentWindow.location.href) {
                document.location.reload();
            }
        },
    },
    components: { MergeRequestsTable }
};
</script>

<template>
    <div class="registry cluster" id="registry-view">
        <div class="registry-header registry-header-view cluster">
            <h1>Керування Платформою</h1>
            <a v-if="canUpdateCluster" href="/admin/cluster/edit" class="registry-add">
                <img alt="add registry" src="@/assets/img/action-edit.png" />
                <span>Редагувати</span>
            </a>
        </div>
        <div class="registry-description cluster">Керування обчислювальними ресурсами платформи, мережею платформи,
            налаштування компонентів платформи, доступу до них.</div>

        <div class="rg-info-block">
            <div class="rg-info-block-header">
                <span>Загальна інформація</span>
                <img src="@/assets/img/action-toggle.png" alt="toggle block" />
            </div>
            <div class="rg-info-block-body">
                <div class="rg-info-line-horizontal">
                    <span>Назва</span>
                    <span>{{ codebase.metadata.name }}</span>
                </div>
                <div class="rg-info-line-horizontal">
                    <span>Версія</span>
                    <span>{{ version }}</span>
                </div>
                <div class="rg-info-line-horizontal">
                    <span>Опис</span>
                    <span>{{ codebase.spec.description }}</span>
                </div>
                <div v-if="admins" class="rg-info-line-horizontal">
                    <span>Адміністратори</span>
                    <span>{{ admins }}</span>
                </div>
                <div v-if="cidr" class="rg-info-line-horizontal">
                    <span>CIDR для адміністративних компонент</span>
                    <span class="cidr-values">
                        <div v-for="(value, index) in cidr" :key="index" class="view-cidr">{{ value }}</div>
                    </span>
                </div>

                <div class="rg-info-line-horizontal">
                    <span>Час створення</span>
                    <span>{{ getFormattedDate(codebase.metadata.creationTimestamp) }}</span>
                </div>
            </div>
        </div>

        <div v-if="branches.length" class="rg-info-block">
            <div class="rg-info-block-header">
                <span>Конфігурація</span>
                <img src="@/assets/img/action-toggle.png" alt="toggle block" />
            </div>
            <div class="rg-info-block-body">
                <table class="rg-info-table rg-info-table-config">
                    <thead>
                        <tr>
                            <th>Статус</th>
                            <th>Конфігурація</th>
                            <th>VCS</th>
                            <th>CI</th>
                            <th>Версія</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="$br in branches" :key="$br.Name">
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
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <div v-if="mergeRequests.length" class="rg-info-block">
            <div class="rg-info-block-header">
                <span>Запити на оновлення</span>
                <img src="@/assets/img/action-toggle.png" alt="toggle block" />
            </div>
            <div class="rg-info-block-body mr-block-table">
                <MergeRequestsTable :merge-requests="mergeRequests" :in-platform="true" @onViewClick="showMrView"></MergeRequestsTable>
            </div>
        </div>
        <div v-if="edpComponents.length" class="dashboard-panel registry-dashboard-panel">
            <ul>
                <li v-for="$ec in edpComponents" :key="$ec.spec.type">
                    <img :src="`data:image/svg+xml;base64,${$ec.spec.icon}`" :alt="`${$ec.spec.type} logo`" />
                    <div class="dashboard-item-content">
                        <a :href="$ec.spec.url">
                            {{ $ec.spec.type }}
                            <img src="@/assets/img/action-link.png" :alt="`${$ec.spec.type} link`">
                        </a>
                    </div>
                </li>
            </ul>
        </div>
        <div class="popup-backdrop visible" v-cloak v-if="backdropShow"></div>

        <div style="width:80%;left:10%;height:80%;" class="popup-window admin-window visible" v-cloak v-if="mrView">
            <div class="popup-header">
                <p>Запит на оновлення</p>
                <a href="#" @click.stop.prevent="hideMrView" class="popup-close hide-popup">
                    <img alt="close popup window" src="@/assets/img/close.png" />
                </a>
            </div>
            <div class="popup-body mr-frame-body" style="border-bottom: none;">
                <iframe ref="mrIframe" id="mr-frame" :src="mrSrc" style="width:100%;"></iframe>
            </div>
        </div>
    </div>
</template>
