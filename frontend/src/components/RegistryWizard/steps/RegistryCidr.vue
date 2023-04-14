<script setup lang="ts">
  import CidrModal from '../../CidrModal.vue';
</script>

<script lang="ts">
import $ from 'jquery';
import {defineComponent, inject} from 'vue';

interface RegistryEditTemplateVariables {
  registryValues: any;
}

export default defineComponent({
  mounted() {
    const cidrConfig = this.templateVariables?.registryValues?.global?.whiteListIP;
    if (cidrConfig) {
      if (cidrConfig.adminRoutes !== "") {
        this.adminCIDR = cidrConfig.adminRoutes.split(" ");
      }
      if (cidrConfig.officerPortal !== "") {
        this.officerCIDR = cidrConfig.officerPortal.split(" ");
      }
      if (cidrConfig.citizenPortal !== "") {
        this.citizenCIDR = cidrConfig.citizenPortal.split(" ");
      }
    }
  },
  data() {
    const templateVariables = inject('TEMPLATE_VARIABLES') as RegistryEditTemplateVariables;
    return {
      templateVariables,
      cidrChanged: true,
      officerCIDRValue: { value: '' },
      officerCIDR: [],
      citizenCIDRValue: { value: '' },
      citizenCIDR: [],
      adminCIDRValue: { value: '' },
      adminCIDR: [],
      currentCIDR: [],
      currentCIDRValue: {},
      cidrPopupShow: false,
    };
  },
  methods: {
    showCIDRForm(cidr: never[], value: object) {
      this.cidrPopupShow = true;
      $("body").css("overflow", "hidden");
      this.currentCIDR = cidr;
      this.currentCIDRValue = value;
    },
    deleteCIDR(c: any, cidr: any, value: any) {
      for (let v in cidr) {
        if (cidr[v] === c) {
          cidr.splice(v, 1);
          break;
        }
      }
      value = JSON.stringify(cidr);
      this.cidrChanged = true;
    },
  },
  components: {CidrModal,}
});
</script>

<template>
  <h2>Перелік дозволених CIDR</h2>
  <p>Ці налаштування є необов’язковими.</p>
  <div class="rc-form-group">
      <label for="admins">CIDR для портала чиновника</label>
      <input type="checkbox" style="display: none;" v-model="cidrChanged" checked name="cidr-changed" />

      <input type="hidden" id="officer-cidr" name="officer-cidr" v-model="officerCIDRValue.value" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in officerCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click.stop.prevent="deleteCIDR(c, officerCIDR, officerCIDRValue)" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="showCIDRForm(officerCIDR, officerCIDRValue)">+</button>
      </div>
  </div>

  <div class="rc-form-group">
      <label for="admins">CIDR для портала громадянина</label>
      <input type="hidden" id="citizen-cidr" name="citizen-cidr" v-model="citizenCIDRValue.value" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in citizenCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click.stop.prevent="deleteCIDR(c, citizenCIDR, citizenCIDRValue)" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="showCIDRForm(citizenCIDR, citizenCIDRValue)">+</button>
      </div>
  </div>

  <div class="rc-form-group">
      <label for="admins">CIDR для адміністративних компонент</label>
      <input type="hidden" id="admin-cidr" name="admin-cidr" v-model="adminCIDRValue.value" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in adminCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click.stop.prevent="deleteCIDR(c, adminCIDR, adminCIDRValue)" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="showCIDRForm(adminCIDR, adminCIDRValue)">+</button>
      </div>
  </div>

  <cidr-modal
      v-model:cidr-popup-show="cidrPopupShow"
      v-model:currentCIDR="currentCIDR"
      v-model:currentCIDRValue="currentCIDRValue"
      v-model:cidr-changed="cidrChanged"
  />
</template>
