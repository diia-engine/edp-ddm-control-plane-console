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
      officerCIDR: [],
      citizenCIDR: [],
      adminCIDR: [],
      currentCIDR: [] as Array<string>,
      cidrPopupShow: false,
    };
  },
  methods: {
    showCIDRForm(cidr: never[]) {
      this.cidrPopupShow = true;
      $("body").css("overflow", "hidden");
      this.currentCIDR = cidr;
    },
    onCidrAdded(cidr: string) {
      this.currentCIDR.push(cidr);
      this.cidrChanged = true;
    },
    deleteCIDR(c: any, cidr: any) {
      for (let v in cidr) {
        if (cidr[v] === c) {
          cidr.splice(v, 1);
          break;
        }
      }
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

      <input type="hidden" id="officer-cidr" name="officer-cidr" :value="JSON.stringify(officerCIDR)" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in officerCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click.stop.prevent="deleteCIDR(c, officerCIDR)" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="showCIDRForm(officerCIDR)">+</button>
      </div>
  </div>

  <div class="rc-form-group">
      <label for="admins">CIDR для портала громадянина</label>
      <input type="hidden" id="citizen-cidr" name="citizen-cidr" :value="JSON.stringify(citizenCIDR)" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in citizenCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click.stop.prevent="deleteCIDR(c, citizenCIDR)" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="showCIDRForm(citizenCIDR)">+</button>
      </div>
  </div>

  <div class="rc-form-group">
      <label for="admins">CIDR для адміністративних компонент</label>
      <input type="hidden" id="admin-cidr" name="admin-cidr" :value="JSON.stringify(adminCIDR)" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in adminCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click.stop.prevent="deleteCIDR(c, adminCIDR)" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="showCIDRForm(adminCIDR)">+</button>
      </div>
  </div>

  <cidr-modal
      v-model:cidr-popup-show="cidrPopupShow"
      title="Додати CIDR"
      sub-title="IP-адреси та маски"
     @cidrAdded="onCidrAdded"
  />
</template>
