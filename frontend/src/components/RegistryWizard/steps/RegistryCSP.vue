<script setup lang="ts">
import CspModal from '../../CspModal.vue';
import { RegistryWizardTemplateVariables } from '@/types/registry';
import { ref, inject } from 'vue';
const templateVariables = inject('TEMPLATE_VARIABLES') as RegistryWizardTemplateVariables;
const connectSrc = ref(templateVariables?.registryValues?.global?.connectSourceList || []);

//        - hello
// - world
        // - *.example.com
        // - https://example.com/path
        // - https://example.com:8080
        // - data:
        // 'sha256-abc123'
        // - //hello.com

const cspChanged = ref(false);
const cspPopupShow = ref(false);

function deleteSrc(src: string) {
    connectSrc.value = connectSrc.value.filter((c) => c !== src);
    cspChanged.value = true;
}
function showCspForm(sources: string[]) {
    cspPopupShow.value = true;
    console.log('showCIDRForm', sources);
}
function onCspAdded(csp: string) {
      connectSrc.value.push(btoa(csp));
      cspChanged.value = true;
}
function decodeBase64(c: string) {
    try {
        return atob(c);
    } catch (e) {
        return c;
    }
}
</script>
<template>
  <h2>{{ $t('components.registryCsp.title') }}</h2>
  <p>{{ $t('components.registryCsp.desc') }}</p>
  <div class="rc-form-group">
      <label for="admins">{{ $t('components.registryCsp.connectSrc') }}</label>
      <input type="checkbox" style="display: none;" v-model="cspChanged" name="csp-changed" />

      <input type="hidden" id="registry-csp" name="registry-csp" :value="JSON.stringify(connectSrc)" />
      <div class="advanced-admins">
          <div v-cloak v-for="c in connectSrc" class="child-admin" v-bind:key="c">
              {{ decodeBase64(c) }}
              <a @click.stop.prevent="deleteSrc(c)" href="#">
                  <img src="@/assets/img/action-delete.png" />
              </a>
          </div>
          <button type="button" @click="showCspForm(connectSrc)">+</button>
      </div>
  </div>
  <CspModal
      v-model:show="cspPopupShow"
      :title="$t('components.registryCsp.modal.addConnectSrcTitle')"
     @csp-added="onCspAdded"
  />
</template>