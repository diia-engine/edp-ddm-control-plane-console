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
  <h2>Автентифікація надавачів послуг</h2>
  <p>Є можливість використовувати власний віджет автентифікації або налаштувати інтеграцію з id.gov.ua.</p>
  <div class="rc-form-group">
      <label for="sup-auth-type">Вкажіть тип автентифікації</label>
      <select name="sup-auth-browser-flow" id="sup-auth-type"
              v-model="pageRoot.$data.wizard.tabs.supplierAuthentication.data.authType" @change="pageRoot.wizardSupAuthFlowChange">
          <option value="dso-officer-auth-flow">Віджет</option>
          <option value="id-gov-ua-officer-redirector">id.gov.ua</option>
      </select>
  </div>
  <div class="rc-form-group"
        :class="{'error': pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && (pageRoot.$data.wizard.tabs.supplierAuthentication.data.url == '' || pageRoot.$data.wizard.tabs.supplierAuthentication.urlValidationFailed)}">
      <label for="sup-auth-url">Посилання <b class="red-star">*</b></label>
      <!--<input name="sup-auth-url" id="sup-auth-url" v-model="wizard.tabs.supplierAuthentication.data.url" />-->
      <input id="sup-auth-url" name="sup-auth-url" v-model="pageRoot.$data.wizard.tabs.supplierAuthentication.data.url">
      <span v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.data.url == ''">Обов’язкове поле</span>
      <span v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.urlValidationFailed">Перевірте формат поля</span>
      <p>URL, повинен починатись з http:// або https://</p>

  </div>
  <div v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.data.authType == 'dso-officer-auth-flow'" class="rc-form-group"
        :class="{'error': pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && (pageRoot.$data.wizard.tabs.supplierAuthentication.data.widgetHeight == '' || pageRoot.$data.wizard.tabs.supplierAuthentication.heightIsNotNumber)}">
      <label for="sup-auth-widget-height">Висота віджета, px <b class="red-star">*</b></label>
      <input id="sup-auth-widget-height" name="sup-auth-widget-height" v-model="pageRoot.$data.wizard.tabs.supplierAuthentication.data.widgetHeight">
      <span v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.data.widgetHeight == ''">Обов’язкове поле</span>
      <span v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.heightIsNotNumber">Перевірте формат поля</span>
  </div>

  <div v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.data.authType == 'id-gov-ua-officer-redirector'">
      <div class="rc-form-group"
            :class="{'error': pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.data.clientId == ''}">
          <label for="diia-client-id">Ідентифікатор клієнта (client_id) <b class="red-star">*</b></label>
          <input name="sup-auth-client-id" id="diia-client-id" v-model="pageRoot.$data.wizard.tabs.supplierAuthentication.data.clientId">
          <span v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.data.clientId == ''">Обов’язкове поле</span>
      </div>
      <div class="rc-form-group"
            :class="{'error': pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.data.secret == ''}">
          <label for="diia-client-secret">Клієнтський секрет (secret) <b class="red-star">*</b></label>
          <input type="password" name="sup-auth-client-secret" id="diia-client-secret" v-model="pageRoot.$data.wizard.tabs.supplierAuthentication.data.secret" />
          <span v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.beginValidation && pageRoot.$data.wizard.tabs.supplierAuthentication.data.secret == ''">Обов’язкове поле</span>
      </div>
  </div>

  <h2>Самостійна реєстрація користувачів</h2>
  <p>Передбачає наявність у реєстрі попередньо змодельованого бізнес-процесу самореєстрації.</p>
  <div class="toggle-switch backup-switch">
    <input v-model="pageRoot.$data.wizard.tabs.supplierAuthentication.selfRegistrationEnabled" class="switch-input"
           type="checkbox" id="self-registration-switch-input" name="self-registration-enabled" />
    <label for="self-registration-switch-input">Toggle</label>
    <span>Дозволити самостійну реєстрацію</span>
  </div>
  <div class="wizard-warning" v-if="pageRoot.$data.wizard.tabs.supplierAuthentication.selfRegistrationEnabled">
    При вимкненні можливості, користувачі, які почали процес самореєстрації, не зможуть виконати свої задачі, якщо вони змодельовані.
  </div>
</template>
