<script setup lang="ts">
import RegistryEditPublicApiModal from '@/components/RegistryEditPublicApiModal.vue';
import RegistryDeletePublicApiModal from '@/components/RegistryDeletePublicApiModal.vue';
import { getImageUrl } from '@/utils';
import { toRefs, ref } from 'vue';
import axios from 'axios';

interface PublicApi {
  name: string;
  url: string;
  enabled: boolean;
  StatusRegistration: string;
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

function getPublicApiBlockTitle(e: any, publicApi?: PublicApi) {
  if (inactive(publicApi?.StatusRegistration)) {
        return undefined;
    }
    if (publicApi?.enabled) {
        return 'Заблокувати доступ';
    }
    return 'Розблокувати доступ';
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
      <table class="rg-info-table rg-info-table-config" v-if="publicApi && publicApi.length">
          <thead>
              <tr>
                  <th>Статус</th>
                  <th>Назва</th>
                  <th>URL</th>
                  <th></th>
              </tr>
          </thead>
          <tbody>
              <tr v-for="(publicApiItem, $index) in publicApi" :key="$index">
                  <td>
                      <img :alt="getActualStatus(publicApiItem.StatusRegistration, publicApiItem.enabled)"
                          :src="getImageUrl(`status-${getActualStatus(publicApiItem.StatusRegistration, publicApiItem.enabled)}`)" />
                  </td>
                  <td>{{ publicApiItem.name }}</td>
                  <td>{{ publicApiItem.url }}</td>
                  <td>
                      <div class="rg-public-api-actions">
                          <a href="#" @click="showPublicApiEditReg($event, publicApiItem)">
                              <i class="fa-solid fa-pen"></i>
                          </a>
                          <a :class="inactive(publicApiItem.StatusRegistration) ? 'inactive' : ''"
                              @click="!inactive(publicApiItem.StatusRegistration) && disablePublicAccessReg(registry, publicApiItem.name, $event)"
                              href="#">
                              <img :title="getPublicApiBlockTitle(publicApiItem)"
                                  alt="key" :src="getImageUrl(`lock-status-${publicApiItem.StatusRegistration ? publicApiItem.StatusRegistration : (publicApiItem.enabled ? 'active' : 'disabled')}`)" />
                          </a>
                          <a href="#" @click="showDeletePublicAccessReg($event, publicApiItem)">
                              <i class="fa-solid fa-trash registry-trash"></i>
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
          <a class="" href="#" @click="showPublicApiEditReg($event)">
              <img alt="Надати доступ" src="@/assets/img/plus.png" />
              <span>Надати доступ</span>
          </a>
      </div>
  </div>

  <RegistryEditPublicApiModal :registry="registry" :publicApiPopupShow="publicApiPopupShow" :publicApi="publicApiValues" @hideModalWindow="hideModalWindow" />
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
}

.rg-external-system-actions a:last-of-type {
    margin: 0;
}

.rg-public-api-actions a.inactive {
    cursor: not-allowed;
}
</style>

