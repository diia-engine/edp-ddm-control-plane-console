<script setup lang="ts">
  import { inject } from 'vue';
  interface WizardTemplateVariables {
    model: {
      Resources: unknown;
    },
  }
  const templateVariables = inject('TEMPLATE_VARIABLES') as WizardTemplateVariables;
</script>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
  data() {
    return {
      pageRoot: this.$parent?.$parent as any,
    };
  },
});
</script>

<template>
  <h2>Ресурси реєстру</h2>

  <input type="hidden" name="resources" :value="pageRoot.$data.registryResources.encoded" />
  <input type="hidden" id="preload-resources" ref="resourcesEditConfig" :value=" templateVariables.model?.Resources" />

  <div class="registry-resources">
      <div class="rc-form-group res-cat-select">
          <select v-model="pageRoot.$data.registryResources.cat">
              <option v-for="cat in pageRoot.$data.registryResources.cats" v-bind:key="cat">{{ cat }}</option>
          </select>
          <button @click="pageRoot.addResourceCat">+</button>
      </div>

      <div class="cat-line" v-for="cat in pageRoot.$data.registryResources.addedCats" v-bind:key="cat.name">
          <h4>{{ cat.name }}
              <button type="button" @click="pageRoot.removeResourceCat(cat, $event)" class="remove-cat">-</button></h4>
          <div class="rc-form-group">
              <h5>Istio sidecar</h5>
              <div class="sidecar-enabled">
                  <input v-model="cat.config.istio.sidecar.enabled" type="checkbox" id="istio-sidecar-enabled">
                  <label for="istio-sidecar-enabled">Enabled</label>
              </div>
              <label>Requests</label>
              <input v-model="cat.config.istio.sidecar.resources.requests.cpu" type="text" placeholder="CPU"
                  :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.istio.sidecar.resources.requests.cpu == ''}"/>
              <input v-model="cat.config.istio.sidecar.resources.requests.memory" type="text" placeholder="Memory"
                      :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.istio.sidecar.resources.requests.memory == ''}"/>

              <label>Limits</label>
              <input v-model="cat.config.istio.sidecar.resources.limits.cpu" type="text" placeholder="CPU"
                      :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.istio.sidecar.resources.limits.cpu == ''}" />
              <input v-model="cat.config.istio.sidecar.resources.limits.memory" type="text" placeholder="Memory"
                      :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.istio.sidecar.resources.limits.memory == ''}">
          </div>

          <div class="rc-form-group">
              <h5>Container</h5>
              <label>Requests</label>
              <input v-model="cat.config.container.resources.requests.cpu" type="text" placeholder="CPU"
                      :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.container.resources.requests.cpu == ''}" />
              <input v-model="cat.config.container.resources.requests.memory" type="text" placeholder="Memory"
                      :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.container.resources.requests.memory == ''}" />

              <label>Limits</label>
              <input v-model="cat.config.container.resources.limits.cpu" type="text" placeholder="CPU"
                      :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.container.resources.limits.cpu == ''}" />
              <input v-model="cat.config.container.resources.limits.memory" type="text" placeholder="Memory"
                      :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && cat.config.container.resources.limits.memory == ''}">

              <label>Змінні оточення</label>
              <div class="env-vars">
                  <div class="env-var-line" v-for="env in cat.config.container.envVars" v-bind:key="env.value"
                        :class="{'error': pageRoot.$data.wizard.tabs.resources.beginValidation && (env.name == '' || env.value == '')}">
                      <input class="env-name" type="text" placeholder="Name" v-model="env.name" />
                      <input class="env-value" type="text" placeholder="Value" v-model="env.value" />
                      <button @click="pageRoot.removeEnvVar(cat.config.container.envVars, env)" class="remove-env-var">-</button>
                  </div>
                  <a class="env-add-lnk" @click="pageRoot.addEnvVar(cat.config.container.envVars, $event)" href="#">Додати змінну оточення</a>
              </div>
          </div>
      </div>
  </div>
</template>
