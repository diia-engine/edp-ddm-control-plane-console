<script setup lang="ts">
  import { inject } from 'vue';
  interface WizardTemplateVariables {
    dnsManual: string;
    keycloakHostname: string;
    keycloakHostnames: Array<string>;
    keycloakCustomHost: string;
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
  <h2>Налаштування DNS</h2>

  <div class="rc-form-group show-errors-span dns-inputs">
    <label for="keycloak-hostname"
           class="header-label">Сервіс управління користувачами та ролями (Keycloak)</label>

    <div class="text-input-label label-ssl">Доменне імʼя для Keycloak</div>
    <select id="keycloak-hostname" name="keycloak-custom-hostname">
      <option>{{ templateVariables.keycloakHostname }}</option>
      <option v-for="hn in templateVariables.keycloakHostnames" :selected="hn === templateVariables.keycloakCustomHost" v-bind:key="hn">
        {{ hn }}
      </option>
    </select>
  </div>

  <div class="rc-form-group show-errors-span dns-inputs">
      <label for="officer-dns" class="header-label">кабінет посадової особи</label>
      <div class="toggle-switch">
          <input @change="pageRoot.wizardDNSEditVisibleChange('officer', $event)" v-model="pageRoot.$data.wizard.tabs.dns.editVisible.officer"
                  class="switch-input" type="checkbox"
                  id="officer-switch-input" name="officer-dns-enabled" />
          <label for="officer-switch-input">Toggle</label>
          <span style="color: #000000;">Використати власні значення</span>
      </div>

      <div v-show="pageRoot.$data.wizard.tabs.dns.editVisible.officer" class="text-input-label">Доменне імʼя для кабінету посадової особи</div>
      <input :class="{'error': pageRoot.$data.wizard.tabs.dns.formatError.officer}" v-show="pageRoot.$data.wizard.tabs.dns.editVisible.officer"
              pattern="^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$"
              type="text" name="officer-dns" id="officer-dns" v-model="pageRoot.$data.wizard.tabs.dns.data.officer"
              :placeholder="pageRoot.$data.wizard.tabs.dns.preloadValues.officer" />
      <span v-if="pageRoot.$data.wizard.tabs.dns.formatError.officer">Невірний формат</span>
      <p v-show="pageRoot.$data.wizard.tabs.dns.editVisible.officer">Назва не може перевищувати довжину у 63 символи. Допустимі символи “a-z”, “.”, “-”, “_”</p>
      <div v-show="pageRoot.$data.wizard.tabs.dns.editVisible.officer" for="" class="text-input-label label-ssl">SSL-сертифікат для кабінету чиновника (розширення .pem)</div>
      <input v-show="pageRoot.$data.wizard.tabs.dns.editVisible.officer" ref="officerSSL" id="officer-ssl" name="officer-ssl" type="file"
              :class="{'error': pageRoot.$data.wizard.tabs.dns.requiredError.officer || pageRoot.$data.wizard.tabs.dns.typeError.officer}"
          @change="pageRoot.$data.wizard.tabs.dns.requiredError.officer = false;"/>
      <span v-if="pageRoot.$data.wizard.tabs.dns.requiredError.officer">Обов’язкове поле</span>
      <span v-if="pageRoot.$data.wizard.tabs.dns.typeError.officer">Невірний тип файлу</span>
  </div>

  <div class="rc-form-group show-errors-span dns-inputs">
    <label for="citizen-dns" class="header-label">кабінет отримувача послуг</label>
    <div class="toggle-switch">
          <input @change="pageRoot.wizardDNSEditVisibleChange('citizen', $event)" v-model="pageRoot.$data.wizard.tabs.dns.editVisible.citizen"
                  class="switch-input" type="checkbox"
                  id="citizen-switch-input" name="citizen-dns-enabled" />
          <label for="citizen-switch-input">Toggle</label>
          <span style="color: #000000;">Використати власні значення</span>
      </div>
    <div v-show="pageRoot.$data.wizard.tabs.dns.editVisible.citizen" class="text-input-label">Доменне імʼя для кабінету отримувача послуг</div>
    <input :class="{'error': pageRoot.$data.wizard.tabs.dns.formatError.citizen }" v-show="pageRoot.$data.wizard.tabs.dns.editVisible.citizen"
              pattern="^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)+([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$"
              type="text" name="citizen-dns" id="citizen-dns" v-model="pageRoot.$data.wizard.tabs.dns.data.citizen"
              :placeholder="pageRoot.$data.wizard.tabs.dns.preloadValues.citizen" />
      <span v-if="pageRoot.$data.wizard.tabs.dns.formatError.citizen">Невірний формат</span>
      <p v-show="pageRoot.$data.wizard.tabs.dns.editVisible.citizen">Назва не може перевищувати довжину у 63 символи. Допустимі символи “a-z”, “.”, “-”, “_”</p>
      <div v-show="pageRoot.$data.wizard.tabs.dns.editVisible.citizen" for="" class="text-input-label label-ssl">SSL-сертифікат для кабінету громадянина (розширення .pem)</div>
      <input v-show="pageRoot.$data.wizard.tabs.dns.editVisible.citizen" ref="citizenSSL" id="citizen-ssl" name="citizen-ssl" type="file"
              :class="{'error': pageRoot.$data.wizard.tabs.dns.requiredError.citizen || pageRoot.$data.wizard.tabs.dns.typeError.citizen}"
          @change="pageRoot.$data.wizard.tabs.dns.requiredError.citizen = false;"/>
      <span v-if="pageRoot.$data.wizard.tabs.dns.requiredError.citizen">Обов’язкове поле</span>
      <span v-if="pageRoot.$data.wizard.tabs.dns.typeError.citizen">Невірний тип файлу</span>
  </div>

  <div class="yellow-banner">
      <div class="banner-title">
          Увага!
      </div>
      <div class="banner-body">
          Необхідно виконати додаткову зовнішню конфігурацію за межами OpenShift кластеру та реєстру.
          <template v-if="templateVariables.dnsManual">
            Інструкція доступна <a target="_blank" :href="templateVariables.dnsManual">за посиланням</a>.
          </template>
      </div>
  </div>
</template>
