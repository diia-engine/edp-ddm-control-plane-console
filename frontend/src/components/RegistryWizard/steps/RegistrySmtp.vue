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
  <h2>{{ $t('components.registrySmtp.title') }}</h2>
  <div class="rc-form-group">
      <label for="smtp-server-type">{{ $t('components.registrySmtp.text.mailServer') }}</label>
      <input type="hidden" ref="smtpServerTypeSelected" :value="templateVariables.model?.MailServerType" />
      <input type="hidden" ref="smtpEditConfig" :value="templateVariables.smtpConfig" />
      <select v-model="pageRoot.$data.smtpServerType" id="smtp-server-type" name="smtp-server-type">
          <option :selected="templateVariables.model?.MailServerType === 'platform-mail-server'"
                  value="platform-mail-server">{{ $t('components.registrySmtp.text.platformMailServer') }}</option>
          <option :selected="templateVariables.model?.MailServerType === 'external-mail-server'"
                  value="external-mail-server">{{ $t('components.registrySmtp.text.externalMailServer') }}</option>
      </select>
  </div>

  <div v-cloak v-if="pageRoot.$data.smtpServerType == 'platform-mail-server'">
      <p>{{ $t('components.registrySmtp.text.settingConnectionPlatformMailServer') }}</p>
      <div class="rc-form-group">
          <label for="ms-opts">{{ $t('components.registrySmtp.text.registryAddress') }}</label>
          <input readonly type="text" id="ms-opts" name="mail-server-opts"
                  value="<registry_name>@<registry.platform-domain>" />
      </div>
  </div>

  <div v-cloak v-if="pageRoot.$data.smtpServerType == 'external-mail-server'">
      <p>{{ $t('components.registrySmtp.text.configuringConnectionExternalServer') }}</p>
      <input type="hidden" name="mail-server-opts" :value="pageRoot.$data.mailServerOpts" />
      <div class="rc-form-group" :class="{'error': pageRoot.$data.externalSMTPOpts.host == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-host">{{ $t('components.registrySmtp.text.host') }}</label>
          <input v-model="pageRoot.$data.externalSMTPOpts.host" id="smtp-host" />
          <span v-if="pageRoot.$data.externalSMTPOpts.host == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group" :class="{'error': pageRoot.$data.externalSMTPOpts.port == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-port">{{ $t('components.registrySmtp.text.port') }}</label>
          <input id="smtp-port" v-model="pageRoot.$data.externalSMTPOpts.port" />
          <span v-if="pageRoot.$data.externalSMTPOpts.port == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
            :class="{'error': pageRoot.$data.externalSMTPOpts.address == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-address">{{ $t('components.registrySmtp.text.mailAddress') }}</label>
          <input id="smtp-address" v-model="pageRoot.$data.externalSMTPOpts.address" />
          <span v-if="pageRoot.$data.externalSMTPOpts.address == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
      <div class="rc-form-group"
            :class="{'error': pageRoot.$data.externalSMTPOpts.password == '' && pageRoot.$data.wizard.tabs.mail.beginValidation}">
          <label for="smtp-password">{{ $t('components.registrySmtp.text.password') }}</label>
          <input placeholder="******" type="password" id="smtp-password" v-model="pageRoot.$data.externalSMTPOpts.password" />
          <span v-if="pageRoot.$data.externalSMTPOpts.password == '' && pageRoot.$data.wizard.tabs.mail.beginValidation">{{ $t('errors.requiredField') }}</span>
      </div>
  </div>
  <br />
</template>
