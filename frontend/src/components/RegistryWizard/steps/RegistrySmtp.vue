<script setup lang="ts">
  import { inject } from 'vue';
  interface WizardTemplateVariables {
    model: {
      MailServerType: string;
    },
    smtpConfig: unknown;
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
  <h2>Вибір поштового серверу</h2>
  <div class="rc-form-group">
      <label for="smtp-server-type">Поштовий сервер</label>
      <input type="hidden" ref="smtpServerTypeSelected" :value="templateVariables.model?.MailServerType" />
      <input type="hidden" ref="smtpEditConfig" :value="templateVariables.smtpConfig" />
      <select v-model="pageRoot.$data.smtpServerType" id="smtp-server-type" name="smtp-server-type">
          <option :selected="templateVariables.model?.MailServerType === 'platform-mail-server'"
                  value="platform-mail-server">Платформенний поштовий сервер</option>
          <option :selected="templateVariables.model?.MailServerType === 'external-mail-server'"
                  value="external-mail-server">Зовнішній поштовий сервер</option>
      </select>
  </div>

  <div v-cloak v-if="pageRoot.$data.smtpServerType == 'platform-mail-server'">
      <p>Налаштування підключення до платформенного поштового серверу</p>
      <div class="rc-form-group">
          <label for="ms-opts">Поштова адреса реєстру</label>
          <input readonly type="text" id="ms-opts" name="mail-server-opts"
                  value="<registry_name>@<registry.platform-domain>" />
      </div>
  </div>

  <div v-cloak v-if="pageRoot.$data.smtpServerType == 'external-mail-server'">
      <p>Налаштування SMTP-підключення до зовнішнього поштового серверу</p>
      <input type="hidden" name="mail-server-opts" :value="pageRoot.$data.mailServerOpts" />
      <div class="rc-form-group" :class="{'error': pageRoot.$data.externalSMTPOpts.host == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-host">Хост</label>
          <input v-model="pageRoot.$data.externalSMTPOpts.host" id="smtp-host" />
          <span v-if="pageRoot.$data.externalSMTPOpts.host == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">Обов’язкове поле</span>
      </div>
      <div class="rc-form-group" :class="{'error': pageRoot.$data.externalSMTPOpts.port == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-port">Порт</label>
          <input id="smtp-port" v-model="pageRoot.$data.externalSMTPOpts.port" />
          <span v-if="pageRoot.$data.externalSMTPOpts.port == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">Обов’язкове поле</span>
      </div>
      <div class="rc-form-group"
            :class="{'error': pageRoot.$data.externalSMTPOpts.address == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-address">Поштова адреса</label>
          <input id="smtp-address" v-model="pageRoot.$data.externalSMTPOpts.address" />
          <span v-if="pageRoot.$data.externalSMTPOpts.address == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">Обов’язкове поле</span>
      </div>
      <div class="rc-form-group"
            :class="{'error': pageRoot.$data.externalSMTPOpts.password == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-password">Пароль</label>
          <input type="password" id="smtp-password" v-model="pageRoot.$data.externalSMTPOpts.password" />
          <span v-if="pageRoot.$data.externalSMTPOpts.password == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">Обов’язкове поле</span>
      </div>
  </div>
  <br />
</template>
