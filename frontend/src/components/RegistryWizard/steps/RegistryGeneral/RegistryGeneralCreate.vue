<script setup lang="ts">
import { toRefs, watch } from 'vue';
import * as Yup from 'yup';
import axios from 'axios';
import { useField, useForm } from 'vee-validate';
import TextField from '@/components/common/TextField.vue';
import Banner from '@/components/common/Banner.vue';
import Typography from '@/components/common/Typography.vue';
import { getErrorMessage } from '@/utils';
import LocalizationSettings from './components/LocalizationSettings.vue';
import type { LANGUAGES } from '@/constants/registry';

type RegistryGeneralCreateProps = {
  gerritBranches: string[];
  registryTemplateName: string;
  language: keyof typeof LANGUAGES;
};

interface FormValues {
  registryName: string;
  registryDeploymentMode: string;
  registryGerritBranch: string;
}

const emit = defineEmits(['preloadTemplateData', 'onChooseGerritBranch']);
const props = defineProps<RegistryGeneralCreateProps>();
const { gerritBranches, registryTemplateName } = toRefs(props);

const validationSchema = Yup.object<FormValues>({
  registryName: Yup.string()
    .required()
    .max(12)
    .min(3)
    .matches(/^[a-z0-9]([-a-z0-9]*[a-z0-9])?([a-z0-9]([-a-z0-9]*[a-z0-9])?)*$/),
  registryDeploymentMode: Yup.string().required(),
  registryGerritBranch: Yup.string().required(),
});

const searchParams = new URLSearchParams(window.location.search);
const getLongBranchName = (short: string | undefined) => {
  if (!short) {
    return '';
  }
  return gerritBranches.value.find((long) => long.startsWith(short)) ?? '';
};

const { errors, validate, setErrors } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    registryDeploymentMode: 'development',
    registryGerritBranch: gerritBranches.value.length === 1 ? gerritBranches.value[0] : getLongBranchName(searchParams.get('version')?.toString()),
  },
});

const { value: registryName } = useField<string>('registryName');
const { value: description } = useField<string>('description');
const { value: registryDeploymentMode } = useField<string>('registryDeploymentMode');
const { value: registryGerritBranch } = useField<string>('registryGerritBranch');

function validator() {
  return new Promise((resolve, reject) => {
    validate().then(async (res) => {
      if (res.valid) {
        try {
          const { data } = await axios.get(
            `/admin/registry/check/${registryName.value}`
          );

          if (data.registryNameAvailable) {
            const { data } = await axios.get(`/admin/registry/preload-values`, {
              params: {
                template: registryTemplateName.value,
                branch: registryGerritBranch.value,
              },
            });
            emit('preloadTemplateData', data);
            return resolve(true);
          }
          setErrors({ registryName: 'registryNameAlreadyExists' });
          reject(false);
        } catch {
          reject(false);
        }
      }
    });
  });
}

watch(registryGerritBranch, (value) => {
  emit('onChooseGerritBranch', value);
});

defineExpose({
  validator,
});
</script>

<template>
  <Typography variant="h3" class="h3">{{ $t('components.registryGeneral.title') }}</Typography>
  <div
    class="rc-form-group"
    :class="{ error: !!errors.registryGerritBranch }"
  >
    <label for="rec-auth-type">{{ $t('components.registryGeneral.text.templateVersion') }} <b class="red-star">*</b></label>
    <div v-if="gerritBranches.length === 1">
      {{ gerritBranches[0] }}
      <input type="hidden" name="registry-git-branch" v-model="registryGerritBranch" :value="gerritBranches[0]" />
    </div>
    <select v-else name="registry-git-branch" v-model="registryGerritBranch">
      <option value="" disabled selected>{{ $t('components.registryGeneral.text.chooseTemplateVersion') }}</option>
      <option v-for="branch in gerritBranches" v-bind:key="branch">
        {{ branch }}
      </option>
    </select>
    <span v-if="!!errors.registryGerritBranch">
      {{ getErrorMessage(errors.registryGerritBranch) }}
    </span>
  </div>

  <div class="rc-form-group">
    <TextField
      required
      :label="$t('components.registryGeneral.fields.name.label')"
      name="name"
      v-model="registryName"
      :error="errors.registryName"
      :description="$t('components.registryGeneral.fields.name.description')"
    >
      <Banner
        classes="banner"
        :description="$t('components.registryGeneral.text.nameMustBeUnique')"
      />
    </TextField>
  </div>

  <div class="rc-form-group">
    <label for="description">{{ $t('components.registryGeneral.text.description') }}</label>
    <textarea
      rows="3"
      name="description"
      id="description"
      v-model="description"
      maxlength="250"
    ></textarea>
    <p>{{ $t('components.registryGeneral.text.descriptionContainOfficialName') }}</p>
  </div>
  <div
    class="rc-form-group"
    :class="{ error: !!errors.registryDeploymentMode }"
  >
    <label for="rec-auth-type">{{ $t('components.registryGeneral.text.deploymentMode') }} <b class="red-star">*</b></label>
    <select name="deployment-mode" v-model="registryDeploymentMode">
      <option selected value="development">development</option>
      <option value="production">production</option>
    </select>
    <span v-if="!!errors.registryDeploymentMode">
      {{ getErrorMessage(errors.registryDeploymentMode)}}
    </span>
    <Banner
      classes="banner"
      :description="$t('components.registryGeneral.text.impossibleChangeAfterDeploy')"
    />
  </div>
  <LocalizationSettings :language="language" />
</template>

<style scoped>
.h3 {
  margin-bottom: 24px;
}
.banner {
  margin-top: 16px;
}
</style>
