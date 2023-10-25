<script setup lang="ts">
import CidrModal from '../../CidrModal.vue';
</script>

<script lang="ts">

interface RegistryEditTemplateVariables {
  registryValues: any;
}

import { defineComponent, inject } from 'vue';
import $ from "jquery";
export default defineComponent({
  mounted() {
    const ipList = this.templateVariables?.registryValues?.trembita?.ipList;
    if (ipList) {
      this.currentCIDR = ipList;
      if (this.currentCIDR.length > 0) {
        this.enabled = true;
      }
    }
  },
  data() {
    const templateVariables = inject('TEMPLATE_VARIABLES') as RegistryEditTemplateVariables;

    return {
      templateVariables,
      enabled: false,
      value: '',
      requiredError: false,
      cidrPopupShow: false,
      currentCIDR: [] as Array<string>,
      cidrChanged: false,
      maxIPCount: 10,
    };
  },
  methods: {
    onEnabledChanged() {
      if (!this.enabled) {
        this.cidrChanged = true;
      }
    },
    deleteCIDR(cidr: string) {
      const spliceIndex = this.currentCIDR.findIndex((element) => element === cidr);
      if (spliceIndex === -1) {
        return;
      }

      this.currentCIDR.splice(spliceIndex, 1);
      this.cidrChanged = true;
    },
    addAllowed() {
      return this.currentCIDR.length < this.maxIPCount;
    },
    showCIDRForm() {
      if (!this.addAllowed()) {
        return;
      }
      this.cidrPopupShow = true;
      $("body").css("overflow", "hidden");
    },
    onCidrAdded(cidr: string) {
      if (!this.currentCIDR.includes(cidr)) {
        this.currentCIDR.push(cidr);
      }
      this.cidrChanged = true;
    },
    validator() {
      return new Promise<void>((resolve) => {
        this.requiredError = false;

        if (!this.enabled) {
          this.currentCIDR = [];
        }

        if (this.enabled && this.currentCIDR.length == 0) {
          this.requiredError = true;
          return;
        }

        resolve();
      });
    }
  },
  components: {CidrModal,}
});
</script>

<style scoped>
  .trembita-soap-ips {
    margin-top: 32px;
  }
  .add-cidr-disabled {
    background: #E1E3EB;
    cursor: not-allowed;
  }
</style>


<template>
  <h2>{{ $t('components.registryTrembita.title') }}</h2>
  <p>{{ $t('components.registryTrembita.text.registryTrembitaDescription') }}</p>
  <div class="toggle-switch">
    <input class="switch-input" type="checkbox" id="trembita-enable-input"
           v-model="enabled" @change="onEnabledChanged"/>
    <label for="trembita-enable-input">Toggle</label>
    <span>{{ $t('components.registryTrembita.text.enableAPIAccess') }}</span>
  </div>
  <div v-if="enabled" class="rc-form-group trembita-soap-ips" :class="{'error': requiredError}">
    <label for="admins">{{ $t('components.registryTrembita.text.addressesOfTrembita') }}</label>
    <input type="hidden" id="trembita-soap-ips" name="trembita-ip-list" :value="JSON.stringify(currentCIDR)" />
    <div class="advanced-admins">
      <div v-cloak v-for="c in currentCIDR" class="child-admin" v-bind:key="c">
        {{ c }}
        <a @click.stop.prevent="deleteCIDR(c)" href="#">
          <img src="@/assets/img/action-delete.png" />
        </a>
      </div>
      <button type="button" @click="showCIDRForm()" :class="{'add-cidr-disabled': !addAllowed()}">+</button>
    </div>
    <span v-if="requiredError">{{ $t('errors.fillingRequiredField') }}</span>
    <p>{{ $t('components.registryTrembita.text.validNumberValues') }} - {{ maxIPCount }}</p>
  </div>


  <cidr-modal
      v-model:cidr-popup-show="cidrPopupShow"
      :title="$t('components.registryTrembita.text.allowAccessFromAddress')"
      :sub-title="$t('components.registryTrembita.text.addressOfTrembita')"
      :mask-allowed="false"
      @cidrAdded="onCidrAdded"
  />
</template>
