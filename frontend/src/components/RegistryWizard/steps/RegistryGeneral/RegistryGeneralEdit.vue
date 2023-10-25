<script setup lang="ts">
import { toRefs } from 'vue';
import Typography from '@/components/common/Typography.vue';
import type { RegistryWizardTemplateVariables } from '@/types/registry';
import LocalizationSettings from './components/LocalizationSettings.vue';

type RegistryGeneralEditProps = {
  templateVariables: Pick<RegistryWizardTemplateVariables, 'model' | 'registry' | 'registryValues'>;
};

const props = defineProps<RegistryGeneralEditProps>();
const { templateVariables } = toRefs(props);

function validator() {
  return new Promise((resolve) => {
    resolve(true);
  });
}

defineExpose({
  validator,
});
</script>

<template>
  <Typography variant="h3" class="h3">{{ $t('components.registryGeneralEdit.title') }}</Typography>
  <div class="rc-form-group">
    <Typography variant="subheading" class="subheading">{{ $t('components.registryGeneralEdit.text.name') }}</Typography>
    <Typography variant="bodyText">{{ templateVariables.model?.name }}</Typography>
  </div>

  <div class="rc-form-group">
    <label for="description">{{ $t('components.registryGeneralEdit.text.description') }}</label>
    <textarea
      rows="3"
      name="description"
      id="description"
      v-model="templateVariables.model.description"
      maxlength="250"
    ></textarea>
    <p>{{ $t('components.registryGeneralEdit.text.descriptionContainOfficialName') }}</p>
  </div>

  <div class="rc-form-group">
    <Typography variant="subheading" class="subheading">{{ $t('components.registryGeneralEdit.text.deploymentMode') }}</Typography>
    <Typography variant="bodyText">{{ templateVariables.registryValues?.global.deploymentMode }}</Typography>
  </div>

  <LocalizationSettings isEditAction :language="templateVariables.registryValues?.global?.language" />
</template>

<style scoped>
.h3 {
  margin-bottom: 24px;
}
.subheading {
  margin-bottom: 8px;
}
</style>
