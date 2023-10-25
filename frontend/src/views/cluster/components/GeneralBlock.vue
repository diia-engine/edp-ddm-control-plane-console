<script setup lang="ts">
import { ref, computed } from 'vue';
import * as yup from 'yup';
import { useForm } from 'vee-validate';

import Banner from '@/components/common/Banner.vue';
import Typography from '@/components/common/Typography.vue';
import Base64FileField from '@/components/common/Base64FileField.vue';
import TextField from '@/components/common/TextField.vue';
import { DEFAULT_LANGUAGE, LANGUAGES } from '@/constants/cluster';
import axios, { AxiosError } from 'axios';
import i18n from '@/localization';

type LocalizationSettingsProps = {
  language: keyof typeof LANGUAGES;
  platformName: string;
  logoMain: string;
  logoFavicon: string;
};

const LOGO_MAIN_FORMAT = 'svg';
const LOGO_FAVICON_FORMAT = 'png';

const props = defineProps<LocalizationSettingsProps>();
const selectedLanguage = ref(props.language || DEFAULT_LANGUAGE);
const logoMain = ref(props.logoMain);
const logoFavicon = ref(props.logoFavicon);
const validationSchema = yup.object({
  name: yup.string().required(),
  logoMainFormat:
    yup.string().test('format-check', i18n.global.t('errors.checkFormat'), (value) => value === LOGO_MAIN_FORMAT),
  logoFaviconFormat:
    yup.string().test('format-check', i18n.global.t('errors.checkFormat'), (value) => value === LOGO_FAVICON_FORMAT),
});

const { useFieldModel, setErrors, errors, handleSubmit } = useForm({
  validationSchema,
  initialValues: {
    name: props.platformName,
    logoMainFormat: LOGO_MAIN_FORMAT,
    logoFaviconFormat: LOGO_FAVICON_FORMAT,
  },
});

const [
  name,
  logoMainFormat,
  logoFaviconFormat,
] = useFieldModel([
  'name',
  'logoMainFormat',
  'logoFaviconFormat',
]);

const regionalFormat = computed(() => {
  return LANGUAGES[selectedLanguage.value].regionFormat;
});

const onMainSelect = (data: string, format: string) => {
  logoMainFormat.value = format;
  logoMain.value = data;
};

const onFaviconSelect = (data: string, format: string) => {
  logoFaviconFormat.value = format;
  logoFavicon.value = data;
};

const submit = handleSubmit(() => {
  let formData = new FormData();

  formData.append("platform-name", name.value);
  formData.append("main", logoMain.value);
  formData.append("favicon", logoFavicon.value);
  formData.append("language", selectedLanguage.value);

  axios.post('/admin/cluster/general', formData, {
      headers: {
          'Content-Type': 'multipart/form-data'
      }
  }).then(() => {
    window.location.assign('/admin/cluster/management');
  }).catch(({ response }: AxiosError<any>) => {
    setErrors(response?.data.errors);
  });
});

</script>

<template>
  <form id="platform-general" @submit.prevent="submit">
    <Typography variant="h3">{{ $t('domains.cluster.general.title') }}</Typography>
    <Banner
      classes="mt24"
      :description="$t('domains.cluster.general.text.banner')"
    />
    <Typography variant="h5" class="mt24" upperCase>{{ $t('domains.cluster.general.text.platformName') }}</Typography>
    <Typography variant="bodyText" class="mt24 generalDescriptionBox">
      {{ $t('domains.cluster.general.text.platformDescription') }}
    </Typography>

    <div class="rc-form-group mt24">
      <TextField
        :label="$t('domains.cluster.general.fields.platformName.label')"
        name="platform-name"
        v-model="name"
        :error="errors.name || ''"
        required
      />
    </div>

    <div class="rc-form-group mt24">
      <Base64FileField
        name="main"
        :format="LOGO_MAIN_FORMAT"
        :fileData="props.logoMain"
        :label="$t('domains.cluster.general.fields.main.label')"
        :error="errors.logoMainFormat || ''"
        :fileNameDescription="$t('domains.cluster.general.fields.main.fileNameDescription')"
        @onSelect="onMainSelect"
      />
    </div>
    <div class="rc-form-group mt24">
      <Base64FileField
        name="favicon"
        :format="LOGO_FAVICON_FORMAT"
        :fileData="props.logoFavicon"
        :label="$t('domains.cluster.general.fields.favicon.label')"
        :error="errors.logoFaviconFormat || ''"
        :fileNameDescription="$t('domains.cluster.general.fields.favicon.fileNameDescription')"
        @onSelect="onFaviconSelect"
      />
    </div>

    <Typography variant="h5" class="mt32" upperCase>{{ $t('domains.cluster.general.text.localization') }}</Typography>
    <Typography variant="bodyText" class="mt24 generalDescriptionBox">
      {{ $t('domains.cluster.general.text.localizationDescription') }}
    </Typography>

    <div class="rc-form-group mt24">
      <label for="rec-auth-type">{{ $t('domains.cluster.general.fields.language.label') }}</label>
      <select name="language" v-model="selectedLanguage">
      <option v-for="(lang, index) in LANGUAGES" :key="index" :value="index">
          {{ lang.name }}
      </option>
      </select>
    </div>

    <div class="rc-form-group mt24">
      <Typography variant="subheading" class="mb8">{{ $t('domains.cluster.general.text.region') }}</Typography>
      <Typography variant="bodyText">{{ regionalFormat }}</Typography>
    </div>

    <div class="rc-form-group">
      <button type="submit" name="submit">{{ $t('actions.confirm') }}</button>
    </div>
  </form>
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
.generalDescriptionBox {
  max-width: 480px;
}

</style>
