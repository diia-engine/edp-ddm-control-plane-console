<script setup lang="ts">
import { ref, computed } from 'vue';
import Typography from '@/components/common/Typography.vue';
import { DEFAULT_LANGUAGE, LANGUAGES } from '@/constants/registry';

type LocalizationSettingsProps = {
  language: keyof typeof LANGUAGES;
  isEditAction?: boolean;
};

const props = defineProps<LocalizationSettingsProps>();
const selectedLanguage = ref(props.language || DEFAULT_LANGUAGE);

const regionalFormat = computed(() => {
  return LANGUAGES[selectedLanguage.value].regionFormat;
});

</script>

<template>
  <Typography variant="h5" class="mt32" upperCase>{{ $t('components.registryGeneral.text.localizationTitle') }}</Typography>
  <Typography variant="bodyText" class="mt24">{{ $t('components.registryGeneral.text.localizationDescription') }}</Typography>
  <ul>
    <li><Typography variant="bodyText">{{ $t('components.registryGeneral.text.localizationOfficerPortal') }}</Typography></li>
    <li><Typography variant="bodyText">{{ $t('components.registryGeneral.text.localizationCitizenPortal') }}</Typography></li>
    <li><Typography variant="bodyText">{{ $t('components.registryGeneral.text.localizationAdminPortal') }}</Typography></li>
    <li><Typography variant="bodyText">{{ $t('components.registryGeneral.text.localizationReportsPortal') }}</Typography></li>
  </ul>
  <Typography variant="bodyText" v-if="isEditAction">{{ $t('components.registryGeneral.text.localizationWarning') }}</Typography>

  <div class="rc-form-group mt24">
    <label for="rec-auth-type">{{ $t('components.registryGeneral.fields.language.label') }}</label>
    <select name="language" v-model="selectedLanguage">
      <option v-for="(lang, index) in LANGUAGES" :key="index" :value="index">
        {{ lang.name }}
      </option>
    </select>
  </div>

  <div class="rc-form-group mt24">
    <Typography variant="subheading" class="mb8">{{ $t('components.registryGeneral.text.localizationRegionFormat') }}</Typography>
    <Typography variant="bodyText">{{regionalFormat}}</Typography>
  </div>
</template>

<style scoped>
.mt24 {
  margin-top: 24px;
}
.mt32 {
  margin-top: 32px;
}
.mb8 {
  margin-bottom: 8px;
}
ul {
  margin-bottom: 0;
}
</style>
