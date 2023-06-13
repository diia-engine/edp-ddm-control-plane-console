<script setup lang="ts">
import Modal from '@/components/common/Modal.vue';
import TextField from '@/components/common/TextField.vue';
import Typography from '@/components/common/Typography.vue';
import * as yup from 'yup';
import { useForm } from 'vee-validate';
import { onUpdated, toRefs } from 'vue';
import axios, { AxiosError } from 'axios';

interface Data {
  name: string;
  url: string;
}

interface RegistryEditPublicApiModalProps {
  publicApiPopupShow: boolean;
  publicApi: Data | null;
  registry: string;
}

const validationSchema = yup.object({
  name: yup.string().required().min(3).max(32).matches(/^[a-z0-9](?:[-]?[a-z0-9]){1,30}[a-z0-9]$/i),
  url: yup.string().required().matches(/^[A-Za-z0-9-._/]*$/i),
});

const props = defineProps<RegistryEditPublicApiModalProps>();
const { publicApiPopupShow, publicApi, registry } = toRefs(props);
const { handleSubmit, values, errors, setValues, setErrors } = useForm({
  validationSchema,
});

const emit = defineEmits(['hideModalWindow']);

function hideModalWindow() {
  emit('hideModalWindow');
}

onUpdated(()=> {
  if (publicApi?.value) {
    setValues(publicApi?.value);
  } else {
    setValues({ name: '', url: '' });
  }
  setErrors({});
});

const submit = handleSubmit(() => {
  const formData = new FormData();

  formData.append("reg-name", values.name);
  formData.append("reg-url", values.url);

  if (publicApi?.value?.name) {
    axios.post(`/admin/registry/public-api-edit/${registry.value}`, formData, {
      headers: {
          'Content-Type': 'multipart/form-data'
      }
    }).then(() => {
      window.location.assign(`/admin/registry/view/${registry.value}`);
    }).catch(({ response }: AxiosError<any>) => {
      setErrors(response?.data.errors);
    });
  } else {
    axios.post(`/admin/registry/public-api-add/${registry.value}`, formData, {
      headers: {
          'Content-Type': 'multipart/form-data'
      }
    }).then(() => {
      window.location.assign(`/admin/registry/view/${registry.value}`);
    }).catch(({ response }: AxiosError<any>) => {
      setErrors(response?.data.errors);
    });
  }
});

</script>

<template>
  <Modal :show="publicApiPopupShow" @close="hideModalWindow" @submit="submit" title="Надати публічний доступ" submitBtnText="Надати">
    <form id="backupPlace-form">
      <Typography variant="bodyText" class="content-text" v-if="!publicApi?.name">
        Ви можете надати публічний доступ до даних цього реєстру (master). Для цього в мастер-реєстрі буде створено окремого користувача реєстра-клієнта, від імені якого здійснюватеметься доступ до мастер-реєстру.
      </Typography>
      <TextField 
        v-if="!publicApi?.name"
        label="Службова назва запиту"
        name="name"
        description='Допустимі символи "a-z", цифри "0-9", "-". Назва не може перевищувати довжину у 32 символи. Назва повинна починатись і закінчуватись символами латинського алфавіту або цифрами та бути унікальною.'
        v-model="values.name"
        :error="errors?.name"
        required
      />
      <TextField 
        label="Точка інтеграції (шлях до публічного пошукового запиту)"
        name="url"
        description='Допустимі символи “A-Z”, "a-z", цифри "0-9", "-", крапка, "_", "/". Посилання повинно бути унікальним.'
        v-model="values.url"
        :error="errors?.url"
        required
      />
      <Typography variant="small">Наприклад: /search-laboratories-by-city.</Typography>
    </form>
  </Modal>
</template>

<style lang="scss" scoped>
.content-text {
  margin-bottom: 24px;
}
</style>

