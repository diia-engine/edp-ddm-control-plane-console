<script lang="ts">
import { defineComponent } from 'vue';
import RegistryWizard from '@/components/RegistryWizard/RegistryWizard.vue';
import AdministratorModal from '@/components/AdministratorModal.vue';

import RegistryEditMixin from './mixins/RegistryEditMixin';

export default defineComponent({
  mixins: [RegistryEditMixin],
  components: { RegistryWizard, AdministratorModal}
});
</script>

<template>
    <div class="registry registry-create" id="registry-form">
        <div class="registry-header">
            <a href="/admin/registry/overview" class="registry-add">
                <img alt="add registry" src="@/assets/img/action-back.png" />
                <span>НАЗАД</span>
            </a>
        </div>
        <h1>Редагувати реєстр "{{templateVariables.registry.metadata.name}}"</h1>
        <div class="registry-description">Можлива зміна опису реєстру. Змінити назву неможливо.</div>

        <RegistryWizard ref="wizard" :form-submitted="registryFormSubmitted"/>

        <AdministratorModal
          :adminPopupShow="adminPopupShow"
          :editAdmin="editAdmin"
          :requiredError="requiredError"
          :emailFormatError="emailFormatError"
          :adminExistsError="adminExistsError"
          @create-admin="createAdmin"
          @hide-admin-form="hideAdminForm"
        />
    </div>
</template>
