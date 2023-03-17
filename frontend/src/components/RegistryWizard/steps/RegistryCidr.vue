<script setup lang="ts">
  import { inject } from 'vue';
  interface WizardTemplateVariables {
    cidrConfig: unknown;
  }
  const templateVariables = inject('TEMPLATE_VARIABLES') as WizardTemplateVariables;
</script>

<script lang="ts">
export default {
  data() {
    return {
      pageRoot: this.$parent?.$parent as any,
    };
  },
};
</script>

<template>
  <h2>Перелік дозволених CIDR</h2>
  <p>Ці налаштування є необов’язковими.</p>
  <div class="rc-form-group">
      <label for="admins">CIDR для портала чиновника</label>
      <input type="hidden" id="preload-cidr" ref="cidrEditConfig" :value="templateVariables.cidrConfig" />
      <input type="checkbox" style="display: none;" v-model="pageRoot.$data.cidrChanged" checked name="cidr-changed" />

      <input type="hidden" id="officer-cidr" name="officer-cidr" v-model="pageRoot.$data.officerCIDRValue.value" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in pageRoot.$data.officerCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click="pageRoot.deleteCIDR(c, pageRoot.$data.officerCIDR, pageRoot.$data.officerCIDRValue, $event)" :cidr="c" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="pageRoot.showCIDRForm(pageRoot.$data.officerCIDR, pageRoot.$data.officerCIDRValue)">+</button>
      </div>
  </div>

  <div class="rc-form-group">
      <label for="admins">CIDR для портала громадянина</label>
      <input type="hidden" id="citizen-cidr" name="citizen-cidr" v-model="pageRoot.$data.citizenCIDRValue.value" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in pageRoot.$data.citizenCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click="pageRoot.deleteCIDR(c, pageRoot.$data.citizenCIDR, pageRoot.$data.citizenCIDRValue, $event)" :cidr="c" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="pageRoot.$data.showCIDRForm(pageRoot.$data.citizenCIDR, pageRoot.$data.citizenCIDRValue)">+</button>
      </div>
  </div>

  <div class="rc-form-group">
      <label for="admins">CIDR для адміністративних компонент</label>
      <input type="hidden" id="admin-cidr" name="admin-cidr" v-model="pageRoot.$data.adminCIDRValue.value" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in pageRoot.$data.adminCIDR" class="child-admin" v-bind:key="c">
              {{ c }}
              <a @click="pageRoot.deleteCIDR(c, pageRoot.$data.adminCIDR, pageRoot.$data.adminCIDRValue, $event)" :cidr="c" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="pageRoot.showCIDRForm(pageRoot.$data.adminCIDR, pageRoot.$data.adminCIDRValue)">+</button>
      </div>
  </div>
</template>
