<script setup lang="ts">
  import { inject } from 'vue';
  const templateVariables = inject('TEMPLATE_VARIABLES') as any;
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
import KeyForm from '../KeyForm.vue';

export default defineComponent({
    data() {
        return {
            pageRoot: this.$parent as any,
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
      KeyForm,
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
                    <a :href="`/admin/registry/update/${templateVariables.registry.name}`">Оновити реєстр</a>
                </li>
            </ul>
        </div>
        <div class="wizard-body">
            <form id="create-form" @submit="pageRoot.$data.registryFormSubmit" class="registry-create-form wizard-form" method="post"
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
                    <h2>Налаштування шаблону реєстру</h2>
                    <input type="hidden" id="preload-registry-branches" ref="registryBranches"
                        :value="templateVariables.gerritBranches" />
                    <div class="rc-form-group"
                        :class="{'error': pageRoot.$data.wizard.tabs.template.templateRequiredError}">
                        <label for="registry-git-template">Шаблон реєстру</label>
                        <select name="registry-git-template" id="registry-git-template"
                                v-model="pageRoot.$data.wizard.tabs.template.registryTemplate" @change="pageRoot.changeTemplateProject">
                            <option></option>
                            <option v-for="project in templateVariables.gerritProjects" :selected="project.spec.name === templateVariables.model.RegistryGitTemplate" v-bind:key="project.spec.name">
                                {{project.spec.name}}
                            </option>
                        </select>
                        <span v-if="pageRoot.$data.wizard.tabs.template.templateRequiredError">Обов’язкове поле</span>
                    </div>
                    <div class="rc-form-group"
                        :class="{'error': pageRoot.$data.wizard.tabs.template.branchRequiredError}">
                        <label for="registry-git-branch">Гілка шаблону реєстру</label>
                        <select name="registry-git-branch" id="registry-git-branch" v-model="pageRoot.$data.wizard.tabs.template.registryBranch">
                            <option v-for="branch in pageRoot.$data.wizard.tabs.template.branches" v-bind:key="branch">{{ branch }}</option>
                        </select>
                        <span v-if="pageRoot.$data.wizard.tabs.template.branchRequiredError">Обов’язкове поле</span>
                    </div>
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'mail'">
                    <RegistrySmtp ref="smtpTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'key'">
                    <KeyForm
                      ref="keyTab"
                      :wizard="pageRoot.$data.wizard"
                      :model="templateVariables.model"
                      :action="templateVariables.action"
                      @wizard-tab-changed="pageRoot.wizardTabChanged"
                      @wizard-key-hardware-data-changed="pageRoot.wizardKeyHardwareDataChanged"
                      @wizard-add-allowed-key="pageRoot.wizardAddAllowedKey"
                      @wizard-remove-allowed-key="pageRoot.wizardRemoveAllowedKey"
                    />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'resources'">
                    <RegistryResources ref="resourcesTab" />
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
                  <RegistryRecipientAuth ref="recipientAuthTab"/>
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'backupSchedule'">
                  <RegistryBackupSchedule ref="backupScheduleTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'confirmation'">
                    <h2>Підтвердження</h2>
                    <p>Усе готово для створення реєстру. Ви можете перевірити внесені дані або натисніть "Створити реєстр".</p>
                </div>
                <div class="wizard-buttons" :class="{ 'no-prev': pageRoot.$data.wizard.activeTab == 'general' }">
                    <template v-if="templateVariables.action === 'create'">
                        <button class="wizard-prev" @click="pageRoot.wizardPrev" v-show="pageRoot.$data.wizard.activeTab != 'general'" type="button">Назад</button>
                        <button class="wizard-next" @click="pageRoot.wizardNext" v-show="pageRoot.$data.wizard.activeTab != 'confirmation'" type="button">Далі</button>
                        <button class="wizard-next" type="submit" name="submit" v-show="pageRoot.$data.wizard.activeTab == 'confirmation'">Створити реєстр</button>
                    </template>

                    <button v-if="templateVariables.action === 'edit'" v-show="pageRoot.$data.wizard.activeTab != 'update'" class="wizard-next" type="button"
                            @click="pageRoot.wizardEditSubmit">Підтвердити</button>
                </div>
            </form>
        </div>
    </div>
</template>
