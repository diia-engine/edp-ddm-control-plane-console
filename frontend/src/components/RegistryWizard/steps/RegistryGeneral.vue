<script setup lang="ts">
import { toRefs } from 'vue';
import * as Yup from 'yup';
import axios from 'axios';
import { useField, useForm } from 'vee-validate';
import TextField from '@/components/common/TextField.vue';
import Banner from '@/components/common/Banner.vue';
import Typography from '@/components/common/Typography.vue';
import { getErrorMessage } from '@/utils';

type RegistryGeneralProps = {
  gerritBranches: string[];
  registryTemplateName: string;
};

interface FormValues {
  registryName: string;
  registryDeploymentMode: string;
  registryGerritBranch: string;
}

const emit = defineEmits(['preloadTemplateData']);
const props = defineProps<RegistryGeneralProps>();
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

const { errors, validate, setErrors } = useForm<FormValues>({
  validationSchema,
  initialValues: {
    registryDeploymentMode: 'development',
    registryGerritBranch: gerritBranches.value.length === 1 ? gerritBranches.value[0] : '',
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

defineExpose({
  validator,
});
</script>

<template>
  <Typography variant="h3" class="h3">Загальні</Typography>
  <div class="rc-form-group">
    <TextField
      required
      label="Назва"
      name="name"
      v-model="registryName"
      :error="errors.registryName"
      description='Допустимі символи: "a-z", "-". Назва не може перевищувати довжину у 12 символів.'
    >
      <Banner
        classes="banner"
        description="Увага! Назва повинна бути унікальною і її неможливо буде змінити після створення реєстру!"
      />
    </TextField>
  </div>

  <div class="rc-form-group">
    <label for="description">Опис</label>
    <textarea
      rows="3"
      name="description"
      id="description"
      v-model="description"
      maxlength="250"
    ></textarea>
    <p>Опис може містити офіційну назву реєстру чи його призначення.</p>
  </div>
  <div
    class="rc-form-group"
    :class="{ error: !!errors.registryDeploymentMode }"
  >
    <label for="rec-auth-type"
      >Режим розгортання <b class="red-star">*</b></label
    >
    <select name="deployment-mode" v-model="registryDeploymentMode">
      <option selected value="development">development</option>
      <option value="production">production</option>
    </select>
    <span v-if="!!errors.registryDeploymentMode">
      {{ getErrorMessage(errors.registryDeploymentMode)}}
    </span>
    <Banner
      classes="banner"
      description="Після створення реєстру змінити режим розгортання реєстру буде неможливо."
    />
  </div>

  <div
    class="rc-form-group"
    :class="{ error: !!errors.registryGerritBranch }"
  >
    <label for="rec-auth-type">Версія шаблону <b class="red-star">*</b></label>
    <div v-if="gerritBranches.length === 1">
      {{ gerritBranches[0] }}
      <input type="hidden" name="registry-git-branch" v-model="registryGerritBranch" :value="gerritBranches[0]" />
    </div>
    <select v-else name="registry-git-branch" v-model="registryGerritBranch">
      <option value="" disabled selected>Оберіть версію шаблону</option>
      <option v-for="branch in gerritBranches" v-bind:key="branch">
        {{ branch }}
      </option>
    </select>
    <span v-if="!!errors.registryGerritBranch">
      {{ getErrorMessage(errors.registryGerritBranch) }}
    </span>
  </div>
</template>

<style scoped>
.h3 {
  margin-bottom: 24px;
}
.banner {
  margin-top: 16px;
}
</style>
