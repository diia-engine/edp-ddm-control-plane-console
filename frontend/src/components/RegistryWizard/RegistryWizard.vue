<script setup lang="ts">
  import { inject } from 'vue';

  const templateVariables = inject('TEMPLATE_VARIABLES') as RegistryWizardTemplateVariables;
  const envVariables = inject('ENVIRONMENT_VARIABLES') as EnvVariables;
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
import RegistryTrembita from './steps/RegistryTrembita.vue';
import RegistryDigitalDocuments from './steps/RegistryDigitalDocuments.vue';
import KeyData from './steps/KeyData.vue';
import KeyVerification from "./steps/KeyVerification.vue";
import { RegistryGeneralCreate, RegistryGeneralEdit } from "./steps/RegistryGeneral";
import ParametersVirtualMachinesAWS from "./steps/ParametersVirtualMachinesAWS.vue";
import ParametersVirtualMachinesVSphere from "./steps/ParametersVirtualMachinesVSphere.vue";
import GeoDataSettings from "./steps/GeoDataSettings.vue";
import RegistryAdminAuth from "./steps/RegistryAdminAuth.vue";
import { type RegistryWizardTemplateVariables, PlatformStatusType, PORTALS} from '@/types/registry';
import type { EnvVariables } from '@/types/common';

export default defineComponent({
    props: {
      formSubmitted: Boolean,
    },
    data() {
        return {
            pageRoot: this.$parent as any, //TODO: remove this
            templatePreloadedData: {},
            parentFormSubmitted: false,
            gerritBranch: new URLSearchParams(window.location.search).get('version')?.toString() ?? '',
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
      RegistryTrembita,
      RegistryDigitalDocuments,
      RegistryGeneralCreate,
      RegistryGeneralEdit,
      ParametersVirtualMachinesAWS,
      ParametersVirtualMachinesVSphere,
      GeoDataSettings,
      RegistryAdminAuth,
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
      onChooseGerritBranch(branch: string) {
        // set short version of branch to url
        if (branch.split('.').length >= 4) {
          const searchParams = new URLSearchParams(window.location.search);
          const branchShortName = branch.split('.').slice(0, 3).join('.');
          searchParams.set("version", branchShortName);
          window.location.search = searchParams.toString();
        }
      },
    },
});
</script>

<template>
    <div class="reg-wizard">
        <div class="wizard-contents">
            <ul>
                <template v-for="(tab, tabName) in pageRoot.$data.wizard.tabs">
                    <li :class="{
                          active: pageRoot.$data.wizard.activeTab == tabName,
                          disabled: tab.disabled && templateVariables.action === 'create'
                        }"
                        v-if="tab.visible" v-bind:key="tabName">
                        <a @click.stop.prevent="pageRoot.selectWizardTab(tabName, templateVariables.action)" href="#">{{ tab.title }}</a>
                    </li>
                </template>
                <li v-if="templateVariables.hasUpdate">
                    <a :href="`/admin/registry/update/${templateVariables.registry.metadata.name}`">{{ $t('components.registryWizard.actions.updateRegister') }}</a>
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
                  <RegistryGeneralCreate
                    ref="generalTab"
                    v-if="templateVariables.action === 'create'"
                    @preload-template-data="onPreloadTemplateData"
                    @on-choose-gerrit-branch="onChooseGerritBranch"
                    :gerritBranches="templateVariables.gerritBranches"
                    :registryTemplateName="templateVariables.registryTemplateName"
                    :language="templateVariables.clusterValues?.global?.language"
                  />
                  <RegistryGeneralEdit
                    ref="generalTab"
                    v-if="templateVariables.action === 'edit'"
                    :templateVariables="templateVariables"
                  />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'parametersVirtualMachines'">
                  <ParametersVirtualMachinesAWS
                    v-if="templateVariables.platformStatusType === PlatformStatusType.AWS"
                    ref="parametersVirtualMachinesTab"
                    :compute-resources="templateVariables.registryValues?.global.computeResources"
                    :is-platform-admin="templateVariables.isPlatformAdmin"
                    :is-edit-action="templateVariables.action === 'edit'"
                  />
                  <ParametersVirtualMachinesVSphere
                    v-if="templateVariables.platformStatusType === PlatformStatusType.VSphere"
                    ref="parametersVirtualMachinesTab"
                    :compute-resources="templateVariables.registryValues?.global.computeResources"
                    :is-platform-admin="templateVariables.isPlatformAdmin"
                    :is-edit-action="templateVariables.action === 'edit'"
                  />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'administrators'">
                    <h2>{{ $t('components.registryWizard.text.administrators') }}</h2>
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
                        <span v-if="pageRoot.$data.wizard.tabs.administrators.requiredError">{{ $t('errors.requiredField') }}</span>
                        <p>{{ $t('components.registryWizard.text.advancedAdminsDescription') }}</p>
                    </div>
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'mail'">
                    <RegistrySmtp ref="smtpTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'key'">
                    <KeyData
                      :registry-action="templateVariables.action"
                      :page-description="$t('components.registryWizard.text.enteredSystemSignatureKeys')"
                      :region="envVariables.region"
                      ref="keyDataTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'keyVerification'">
                  <KeyVerification
                      :registry-action="templateVariables.action"
                      :page-description="$t('components.registryWizard.text.enteredCertificatesVerification')"
                      :region="envVariables.region"
                      ref="keyVerificationTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'resources'">
                    <RegistryResources
                      ref="resourcesTab"
                      :template-preloaded-data="templatePreloadedData"
                      :form-submitted="parentFormSubmitted"
                      :template-variables="templateVariables"
                      :is-edit-action="templateVariables.action === 'edit'"
                    />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'dns'">
                    <RegistryDns
                      ref="dnsTab"
                      :dns-manual="templateVariables.dnsManual"
                      :keycloak-hostname="templateVariables.keycloakHostname"
                      :keycloak-hostnames="templateVariables.keycloakHostnames"
                      :keycloak-custom-host="templateVariables.keycloakCustomHost"
                      :portals="templateVariables.registryValues?.portals"
                    />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'cidr'">
                    <RegistryCidr ref="cidrTab" />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'supplierAuthentication'">
                    <RegistrySupplierAuth
                        :keycloak-settings="templateVariables.registryValues?.keycloak"
                        :sign-widget-settings="templateVariables.registryValues?.signWidget"
                        :officer-portal-settings="templateVariables.registryValues?.portals?.officer"
                        :is-enabled-portal="!templateVariables.registryValues?.global.excludePortals?.includes(PORTALS.officer)"
                        :digital-signature-keys="templateVariables?.clusterDigitalSignature?.keys || templateVariables?.clusterValues?.['digital-signature']?.keys"
                        :region="envVariables.region"
                        :registryName="templateVariables.registry?.metadata?.name || ''"
                        ref="supplierAuthTab"
                    />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'recipientAuthentication'">
                    <RegistryRecipientAuth
                        :keycloak-settings="templateVariables.registryValues?.keycloak.citizenAuthFlow"
                        :citizen-portal-settings="templateVariables.registryValues?.portals?.citizen"
                        :is-enabled-portal="!templateVariables.registryValues?.global.excludePortals?.includes(PORTALS.citizen)"
                        :region="envVariables.region"
                        ref="recipientAuthTab"
                        :digital-signature-keys="templateVariables?.clusterDigitalSignature?.keys  || templateVariables?.clusterValues?.['digital-signature']?.keys"
                       :registryName="templateVariables.registry?.metadata?.name || ''"
                    />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'adminAuthentication'">
                    <RegistryAdminAuth
                      :is-enabled-portal="!templateVariables.registryValues?.global.excludePortals?.includes(PORTALS.admin)"
                    />
                </div>
                <div class="wizard-tab" v-show="pageRoot.$data.wizard.activeTab == 'geoDataSettings'">
                    <GeoDataSettings
                      :enabled="templateVariables.registryValues?.global?.geoServerEnabled"
                      :is-edit-action="templateVariables.action === 'edit'"
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
                    <h2>{{ $t('components.registryWizard.text.confirmation') }}</h2>
                    <p>{{ $t('components.registryWizard.text.everythingReadyCreateRegistry') }}</p>
                </div>
                <div class="wizard-buttons" :class="{ 'no-prev': pageRoot.$data.wizard.activeTab == 'general' }">
                    <template v-if="templateVariables.action === 'create'">
                        <button class="wizard-prev" @click="pageRoot.wizardPrev" v-show="pageRoot.$data.wizard.activeTab != 'general'" type="button">{{ $t('actions.back') }}</button>
                        <button class="wizard-next" @click="pageRoot.wizardNext" v-show="pageRoot.$data.wizard.activeTab != 'confirmation'" type="button">{{ $t('actions.next') }}</button>
                        <button class="wizard-next" @click="pageRoot.wizardEditSubmit" v-show="pageRoot.$data.wizard.activeTab == 'confirmation'">{{ $t('components.registryWizard.actions.createRegister') }}</button>
                    </template>

                    <button v-if="templateVariables.action === 'edit'" v-show="pageRoot.$data.wizard.activeTab != 'update'" class="wizard-next" type="button"
                            @click="pageRoot.wizardEditSubmit" :disabled="pageRoot.$data.disabled">{{ $t('actions.confirm') }}</button>
                </div>
            </form>
        </div>
    </div>
</template>
