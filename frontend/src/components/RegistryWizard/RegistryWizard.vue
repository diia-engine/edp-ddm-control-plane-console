<script setup lang="ts">
  import { inject } from 'vue';
  const templateVariables = inject('TEMPLATE_VARIABLES') as RegistryWizardTemplateVariables;
</script>

<script lang="ts">
import { defineComponent } from 'vue';

import RegistryBackupSchedule from './steps/RegistryBackupSchedule.vue';
import RegistryCidr from './steps/RegistryCidr.vue';
import RegistryDns from './steps/RegistryDns.vue';
import RegistryResources from './steps/RegistryResources.vue';
import RegistrySmtp from './steps/RegistrySmtp.vue';
import RegistrySupplierAuth from './steps/RegistrySupplierAuth.vue';
import RegistryRecipientAuth from './steps/RegistryRecipientAuth.vue';
import RegistryTemplate from './steps/RegistryTemplate.vue';
import RegistryTrembita from './steps/RegistryTrembita.vue';
import RegistryDigitalDocuments from './steps/RegistryDigitalDocuments.vue';
import KeyData from './steps/KeyData.vue';
import KeyVerification from "./steps/KeyVerification.vue";
import type { RegistryWizardTemplateVariables } from '@/types/registry';

export default defineComponent({
    props: {
      formSubmitted: Boolean,
    },
    data() {
        return {
            pageRoot: this.$parent as any, //TODO: remove this
            templatePreloadedData: {},
            parentFormSubmitted: false,
        };
    },
    components: {
      RegistrySmtp,
      RegistryResources,
      RegistryDns,
      RegistryCidr,
      RegistrySupplierAuth,
      RegistryBackupSchedule,
      RegistryRecipientAuth,
      KeyData,
      KeyVerification,
      RegistryTemplate,
      RegistryTrembita,
      RegistryDigitalDocuments
    },
    watch: {
      formSubmitted() {
        this.parentFormSubmitted = this.formSubmitted;
      },
    },
    methods: {
      onPreloadTemplateData(data: any) {
        this.templatePreloadedData = data;
      },
    },
});
</script>

<template>
    <div class="reg-wizard">
        <div class="wizard-contents">
            <ul>
                <template v-for="(tab, tabName) in pageRoot.$data.wizard.tabs">
                    <li :class="{ active: pageRoot.$data.wizard.activeTab == tabName }"
                        v-if="tab.visible" v-bind:key="tabName">
                        <a @click="pageRoot.selectWizardTab(tabName, $event)" href="#">{{ tab.title }}</a>
                    </li>
                </template>
                <li v-if="templateVariables.hasUpdate">
                    <a :href="`/admin/registry/update/${templateVariables.registry.metadata.name}`">Оновити реєстр</a>
                </li>
            </ul>
        </div>
        <div class="wizard-body">
            <form id="create-form" class="registry-create-form wizard-form" method="post"
                enctype="multipart/form-data" action="" ref="registryWizardForm">
                <input type="hidden" ref="wizardAction" name="action" :value="templateVariables.action" />
                <input type="hidden" ref="registryData" :value="templateVariables.registryData" />

                <div v-if="pageRoot.$data.error" class="rc-global-error">{{templateVariables.error}}</div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'general'">
                    <h2>Загальні налаштування</h2>
                    <div class="rc-title">
                        Увага! Назва повинна бути унікальною і її неможливо буде змінити після створення реєстру!
                    </div>
                    <div class="rc-form-group"
                        :class="{ 'error': pageRoot.$data.wizard.tabs.general.requiredError || pageRoot.$data.wizard.tabs.general.existsError || pageRoot.$data.wizard.tabs.general.formatError }">
                        <label for="name">Назва реєстру</label>
                        <input :disabled="templateVariables.action === 'edit'" type="text" id="name" name="name" maxlength="12"
                            v-model="pageRoot.$data.wizard.tabs.general.registryName"
                            pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?([a-z0-9]([-a-z0-9]*[a-z0-9])?)*$" />
                        <span v-if="pageRoot.$data.wizard.tabs.general.requiredError">Обов’язкове поле</span>
                        <span v-if="pageRoot.$data.wizard.tabs.general.existsError">Реєстр з такою назвою вже існує</span>
                        <span v-if="pageRoot.$data.wizard.tabs.general.formatError">Будь-ласка вкажіть назву у відповідному форматі</span>
                        <p>Допустимі символи: "a-z", "-". Назва не може перевищувати довжину у 12 символів.</p>
                    </div>

                    <div class="rc-form-group">
                        <label for="description">Опис</label>
                        <!-- eslint-disable -->
                        <textarea rows="3" name="description" id="description" maxlength="250">{{templateVariables.model?.description}}</textarea>
                        <!-- eslint-enable  -->
                        <p>Опис може містити офіційну назву реєстру чи його призначення.</p>
                    </div>
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'administrators'">
                    <h2>Адміністратори</h2>
                    <div :class="{ error: pageRoot.$data.wizard.tabs.administrators.requiredError }"
                        class="rc-form-group administrators">
                        <input type="hidden" id="admins" name="admins" v-model="pageRoot.$data.adminsValue" :admins="pageRoot.loadAdmins(templateVariables.model?.admins)" />
                        <input type="checkbox" style="display: none;" v-model="pageRoot.$data.adminsChanged" checked name="admins-changed" />

                        <div class="advanced-admins" ref="admins">
                            <div v-cloak v-for="adm in pageRoot.$data.admins" class="child-admin" v-bind:key="adm.email">
                                {{ adm.email }}
                                <a @click="pageRoot.deleteAdmin" :email="adm.email" href="#">
                                    <img src="@/assets/img/action-delete.png" />
                                </a>
                            </div>
                            <button type="button" @click="pageRoot.showAdminForm">+</button>
                        </div>
                        <span v-if="pageRoot.$data.wizard.tabs.administrators.requiredError">Обов’язкове поле</span>
                        <p>Допустимі символи: "0-9", "a-z", "_", "-", "@", ".", ",".</p>
                    </div>
                </div>
                <div v-if="templateVariables.action === 'create'" class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'template'">
                    <RegistryTemplate ref="templateTab" @preload-template-data="onPreloadTemplateData"
                                      :template-variables="templateVariables as any" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'mail'">
                    <RegistrySmtp ref="smtpTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'key'">
                    <KeyData
                      :registry-action="templateVariables.action"
                      page-description="Внесені ключі системного підпису та КЕП користувачів будуть застосовані для налаштувань поточного реєстру."
                      ref="keyDataTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'keyVerification'">
                  <KeyVerification
                      :registry-action="templateVariables.action"
                      page-description="Внесені сертифікати АЦСК для перевірки ключів системного підпису та КЕП користувачів будуть застосовані для налаштувань поточного реєстру."
                      ref="keyVerificationTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'resources'">
                    <RegistryResources ref="resourcesTab" :template-preloaded-data="templatePreloadedData"
                    :form-submitted="parentFormSubmitted" :template-variables="templateVariables" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'dns'">
                    <RegistryDns ref="dnsTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'cidr'">
                    <RegistryCidr ref="cidrTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'supplierAuthentication'">
                    <RegistrySupplierAuth ref="supplierAuthTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'recipientAuthentication'">
                    <RegistryRecipientAuth
                        :keycloak-settings="templateVariables.registryValues?.keycloak.citizenAuthFlow"
                        :citizen-portal-settings="templateVariables.registryValues?.citizenPortal"
                        ref="recipientAuthTab"
                    />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'digitalDocuments'">
                    <RegistryDigitalDocuments ref="digitalDocumentsTab"
                        :max-file-size-prop="templateVariables.registryValues?.digitalDocuments.maxFileSize"
                        :max-total-file-size-prop="templateVariables.registryValues?.digitalDocuments.maxTotalFileSize" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'backupSchedule'">
                  <RegistryBackupSchedule ref="backupScheduleTab" :template-variables="templateVariables" />
                </div>
              <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'trembita'">
                <RegistryTrembita ref="trembitaTab" />
              </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'confirmation'">
                    <h2>Підтвердження</h2>
                    <p>Усе готово для створення реєстру. Ви можете перевірити внесені дані або натисніть "Створити реєстр".</p>
                </div>
                <div class="wizard-buttons" :class="{ 'no-prev': pageRoot.$data.wizard.activeTab == 'general' }">
                    <template v-if="templateVariables.action === 'create'">
                        <button class="wizard-prev" @click="pageRoot.wizardPrev" v-show="pageRoot.$data.wizard.activeTab != 'general'" type="button">Назад</button>
                        <button class="wizard-next" @click="pageRoot.wizardNext" v-show="pageRoot.$data.wizard.activeTab != 'confirmation'" type="button">Далі</button>
                        <button class="wizard-next" @click="pageRoot.wizardEditSubmit" v-show="pageRoot.$data.wizard.activeTab == 'confirmation'">Створити реєстр</button>
                    </template>

                    <button v-if="templateVariables.action === 'edit'" v-show="pageRoot.$data.wizard.activeTab != 'update'" class="wizard-next" type="button"
                            @click="pageRoot.wizardEditSubmit" :disabled="pageRoot.$data.disabled">Підтвердити</button>
                </div>
            </form>
        </div>
    </div>
</template>
