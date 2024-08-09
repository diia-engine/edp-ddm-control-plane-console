<script setup lang="ts">
import { toRefs } from 'vue';
import Typography from '@/components/common/Typography.vue';
import type { RegistryWizardTemplateVariables } from '@/types/registry';

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
  <Typography variant="h3" class="h3">Загальні</Typography>
  <div class="rc-form-group">
    <Typography variant="subheading" class="subheading">Назва</Typography>
    <Typography variant="bodyText">{{ templateVariables.model?.name }}</Typography>
  </div>

  <div class="rc-form-group">
    <label for="description">Опис</label>
    <textarea
      rows="3"
      name="description"
      id="description"
      v-model="templateVariables.model.description"
      maxlength="250"
    ></textarea>
    <p>Опис може містити офіційну назву реєстру чи його призначення.</p>
  </div>

  <div class="rc-form-group">
    <Typography variant="subheading" class="subheading">Режим розгортання</Typography>
    <Typography variant="bodyText">{{ templateVariables.registryValues?.global.deploymentMode }}</Typography>
  </div>

  <div class="rc-form-group">
    <Typography variant="subheading" class="subheading">Версія шаблону</Typography>
    <Typography variant="bodyText">{{ templateVariables?.registry.spec.branchToCopyInDefaultBranch }}</Typography>
  </div>
</template>

<style scoped>
.h3 {
  margin-bottom: 24px;
}
.subheading {
  margin-bottom: 8px;
}
</style>
