<script setup lang="ts">
import RegistryEditPublicApiModal from '@/components/RegistryEditPublicApiModal.vue';
import RegistryDeletePublicApiModal from '@/components/RegistryDeletePublicApiModal.vue';
import type { PublicApiLimits } from '@/types/registry';
import { getImageUrl } from '@/utils';
import { getExtStatus } from '@/utils/registry';
import { toRefs, ref } from 'vue';
import axios from 'axios';

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
    return "В обробці";
  }
  if (status === "failed") {
    return "Помилка";
  }
  if (enabled === false) {
      return "Заблокований";
  }

  return "Активний";
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
    return 'Розблокувати доступ';
  }
  return 'Заблокувати доступ';
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

function inactive(status?: string) {
  return status === "inactive" || status === "failed";
}

</script>

<template>
  <div class="rg-info-block-body">
    <table class="rg-info-table rg-info-table-config">
      <thead>
        <tr>
          <th>Статус</th>
          <th>Назва</th>
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
              <a href="#" @click.prevent="getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled) !== 'status-inactive' && showPublicApiEditReg($event, publicApiItem)" title="Редагувати">
                <i class="fa-solid fa-pen"></i>
              </a>
              <a @click.prevent="getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled) !== 'status-inactive' && disablePublicAccessReg(registry, publicApiItem.name, $event)" href="#">
                <img :title="getPublicApiBlockTitle(publicApiItem)"
                  alt="key" :src="getImageUrl(`lock-status-${publicApiItem.StatusRegistration ? publicApiItem.StatusRegistration : (publicApiItem.enabled ? 'active' : 'disabled')}`)" />
              </a>
              <a href="#" @click.prevent="getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled) !== 'status-inactive' && showDeletePublicAccessReg($event, publicApiItem)">
                <img title="Видалити" alt="key" :src="getImageUrl(`disable-${getExtStatus(publicApiItem.StatusRegistration, publicApiItem.enabled)}`)" />
              </a>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
    <div class="rg-info-block-no-content" v-if="!publicApi.length">
      Немає реєстрів або систем, що мають публічний доступ до даних цього реєтра.
    </div>
    <div class="link-grant-access">
      <a class="" href="#" @click="showPublicApiEditReg($event)">
        <img alt="Надати доступ" src="@/assets/img/plus.png" />
        <span>Надати доступ</span>
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

</style>
