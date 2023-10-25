<script setup lang="ts">
import { toRefs, ref } from 'vue';
import axios from 'axios';

import RegistryEditPublicApiModal from '@/components/RegistryEditPublicApiModal.vue';
import RegistryDeletePublicApiModal from '@/components/RegistryDeletePublicApiModal.vue';
import type { PublicApiLimits } from '@/types/registry';
import { getImageUrl } from '@/utils';
import { getExtStatus } from '@/utils/registry';
import i18n from '@/localization';

interface PublicApi {
  name: string;
  url: string;
  enabled: boolean;
  StatusRegistration: string;
  limits: PublicApiLimits;
}

interface RegistryBackupSavePlaceModalProps {
  publicApi: PublicApi[];
  registry: string;
  checkOpenedMR: (e: any) => boolean;
}

const props = defineProps<RegistryBackupSavePlaceModalProps>();
const { publicApi, registry, checkOpenedMR } = toRefs(props);
const publicApiPopupShow = ref(false);
const deletePublicApiPopupShow = ref(false);
const publicApiValues = ref(null as PublicApi | null);

function hideModalWindow() {
  publicApiPopupShow.value = false;
  deletePublicApiPopupShow.value = false; 
}

const getStatus = (status: string, enabled: boolean = true): string => {
  if (status === "inactive") {
    return i18n.global.t('domains.registry.publicApi.inProcessing');
  }
  if (status === "failed") {
    return i18n.global.t('domains.registry.publicApi.error');
  }
  if (enabled === false) {
      return i18n.global.t('domains.registry.publicApi.integrationIsNotConfigured');
  }

  return i18n.global.t('domains.registry.publicApi.integrationIsConfigured');
};

function showPublicApiEditReg(e: any, publicApi?: PublicApi) {
  e.preventDefault();
  if (checkOpenedMR.value(e)) {
    return;
  }
  publicApiValues.value = publicApi || null;
  publicApiPopupShow.value = true; 
}

function showDeletePublicAccessReg(e: any, publicApi?: PublicApi) {
  e.preventDefault();
  if (checkOpenedMR.value(e)) {
    return;
  }
  publicApiValues.value = publicApi || null;
  deletePublicApiPopupShow.value = true;
}

function getPublicApiBlockTitle(publicApi?: PublicApi) {
  if (!publicApi?.enabled) {
    return i18n.global.t('domains.registry.publicApi.enableAccess');
  }
  return i18n.global.t('domains.registry.publicApi.disableAccess');
}

function getActualStatus(status: string, enable: boolean) {
  if (status) {
    return status;
  }
  if (!enable) {
    return "disabled";
  }

  return "active";
}

function disablePublicAccessReg(registry: string, name: string, e: any) {
  e.preventDefault();
  if (checkOpenedMR.value(e)) {
    return;
  }

  const formData = new FormData();

  formData.append("reg-name", name);
  axios.post(`/admin/registry/public-api-disable/${registry}`, formData).then(() => {
      window.location.assign(`/admin/registry/view/${registry}`);
  });
}

</script>

<template>
  <div class="rg-info-block-body">
    <table class="rg-info-table rg-info-table-config">
      <thead>
        <tr>
          <th>{{ $t('domains.registry.publicApi.configured') }}</th>
          <th>{{ $t('domains.registry.publicApi.name') }}</th>
          <th>URL</th>
          <th></th>
        </tr>
      </thead>
      <tbody v-if="publicApi && publicApi.length">
        <tr v-for="(publicApiItem, $index) in publicApi" :key="$index">
          <td>
            <img :title="getStatus(publicApiItem.StatusRegistration, publicApiItem.enabled)" :alt="getActualStatus(publicApiItem.StatusRegistration, publicApiItem.enabled)"
              :src="getImageUrl(`status-${getActualStatus(publicApiItem.StatusRegistration, publicApiItem.enabled)}`)" />
          </td>
          <td>{{ publicApiItem.name }}</td>
          <td>{{ publicApiItem.url }}</td>
          <td>
            <div class="rg-public-api-actions" :class="{ inactive: getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled) == 'status-inactive' }">
              <a href="#" @click.prevent="getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled) !== 'status-inactive' && showPublicApiEditReg($event, publicApiItem)" :title="$t('actions.edit')">
                <i class="fa-solid fa-pen"></i>
              </a>
              <a @click.prevent="getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled) !== 'status-inactive' && disablePublicAccessReg(registry, publicApiItem.name, $event)" href="#">
                <img :title="getPublicApiBlockTitle(publicApiItem)"
                  alt="key" :src="getImageUrl(`lock-status-${publicApiItem.StatusRegistration ? publicApiItem.StatusRegistration : (publicApiItem.enabled ? 'active' : 'disabled')}`)" />
              </a>
              <a href="#" @click.prevent="getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled) !== 'status-inactive' && showDeletePublicAccessReg($event, publicApiItem)">
                <img :title="$t('actions.remove')" alt="key" :src="getImageUrl(`disable-${getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled)}`)" />
              </a>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div class="rg-info-block-no-content" v-if="!publicApi.length">
      {{ $t('domains.registry.publicApi.noRegistryOrSystem') }}
    </div>
    <div class="link-grant-access">
      <a class="" href="#" @click="showPublicApiEditReg($event)">
        <img :alt="$t('domains.registry.publicApi.grantAccess')" src="@/assets/img/plus.png" />
        <span>{{ $t('domains.registry.publicApi.grantAccess') }}</span>
      </a>
    </div>
  </div>

  <RegistryEditPublicApiModal 
    :registry="registry"
    :publicApiPopupShow="publicApiPopupShow"
    :publicApiList="publicApi"
    :publicApiValues="publicApiValues"
    @hideModalWindow="hideModalWindow"
  />
  <RegistryDeletePublicApiModal :registry="registry" :publicApiPopupShow="deletePublicApiPopupShow" :publicApiName="publicApiValues?.name" @hideModalWindow="hideModalWindow" />
</template>

<style lang="scss" scoped>
.rg-public-api-actions {
  display: flex;
  justify-content: flex-end;
}

.rg-public-api-actions a {
  margin-right: 36px;
  display: flex;
  align-items: center;
  text-decoration: none;
  background-color: transparent;
}

.rg-public-api-actions a:last-child {
  margin: 0;
}

.rg-public-api-actions.inactive a {
    cursor: not-allowed;
    color: #BFBFBF;
}

.link-grant-access a:hover {
  text-decoration: none;
}
.link-grant-access a {
  padding: 8px 10px 8px 10px;
  border-radius: 5px;
  display: flex;
  align-items: baseline;
  width: 170px;
  transition: 0.5s;
}
.link-grant-access a:hover {
  background: #E6F3FA;
}

</style>
