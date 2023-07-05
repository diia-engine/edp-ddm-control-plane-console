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
  publicApiValues: Data | null;
  publicApiList: Data[];
  registry: string;
}

const props = defineProps<RegistryEditPublicApiModalProps>();
const { publicApiPopupShow, publicApiValues, publicApiList, registry } = toRefs(props);

const validationSchema = yup.object({
  name: yup.string().required().min(3).max(32).matches(/^[a-z0-9]([a-z0-9-]){1,30}[a-z0-9]$/).test({
    message: 'isUnique',
    test: function (value) {
      if (publicApiValues.value?.name) {
        return true;
      }
      return publicApiList.value?.findIndex(({ name }) => name === value) === -1;
    },
  }),
  url: yup.string().required().matches(/^[A-Za-z0-9-/]*$/i).test({
    message: 'isUnique',
    test: function (value) {
      return publicApiList.value?.findIndex(({ url }) => url === value) === -1;
    },
  }),
});

const { handleSubmit, values, errors, setValues, setErrors } = useForm({
  validationSchema,
});

const emit = defineEmits(['hideModalWindow']);

function hideModalWindow() {
  emit('hideModalWindow');
}

onUpdated(()=> {
  if (publicApiValues?.value) {
    setValues(publicApiValues?.value);
  } else {
    setValues({ name: '', url: '' });
  }
  setErrors({});
});

const submit = handleSubmit(() => {
  const formData = new FormData();

  formData.append("reg-name", values.name);
  formData.append("reg-url", values.url);

  if (publicApiValues.value?.name) {
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
  <Modal 
    :show="publicApiPopupShow"
    :title="publicApiValues?.name ? `Редагувати “${publicApiValues?.name}”` : `Надати публічний доступ`"
    :submitBtnText="publicApiValues?.name ? 'Підтвердити' : 'Надати'"
    @close="hideModalWindow" @submit="submit"
  >
    <form id="backupPlace-form">
      <Typography variant="bodyText" class="content-text" v-if="!publicApiValues?.name">
        Ви можете надати публічний доступ до даних цього реєстру (master). Для цього в мастер-реєстрі буде створено окремого користувача реєстра-клієнта, від імені якого здійснюватиметься доступ до мастер-реєстру.
      </Typography>
      <TextField 
        v-if="!publicApiValues?.name"
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
        description='Допустимі символи "A-Z", "a-z", цифри "0-9", "-", "/". Посилання повинно бути унікальним.'
        v-model="values.url"
        :error="errors?.url"
        required
      />
      <Typography variant="small">Наприклад: /search-laboratories-by-city</Typography>
    </form>
  </Modal>
</template>

<style lang="scss" scoped>
.content-text {
  margin-bottom: 24px;
}
</style>

