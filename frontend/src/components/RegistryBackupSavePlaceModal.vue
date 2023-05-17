<script setup lang="ts">
import Modal from '@/components/common/Modal.vue';
import TextField from '@/components/common/TextField.vue';
import * as yup from 'yup';
import { useForm } from 'vee-validate';
import { onUpdated, toRefs } from 'vue';

interface Data {
  backupBucket: string;
  endpoint: string;
  login: string;
  password: string;
}

interface RegistryBackupSavePlaceModalProps {
  backupPlacePopupShow: boolean;
  initialData: Data;
}

const validationSchema = yup.object({
  backupBucket: yup.string().required().min(3).max(63).matches(/^[a-z0-9][a-z0-9.-]*$/i),
  endpoint: yup.string().required().matches(/^[a-z0-9.\-/:]+$/i),
  login: yup.string().required(),
  password: yup.string().required(),
});

const props = defineProps<RegistryBackupSavePlaceModalProps>();
const { backupPlacePopupShow, initialData } = toRefs(props);
const { handleSubmit, values, errors, setValues, setErrors } = useForm({
  validationSchema, initialValues: initialData,
});

onUpdated(()=> {
  setValues(initialData.value);
  setErrors({});
});

const emit = defineEmits(['hideBackupPlaceModal', 'submitData']);

function hideBackupPlaceModal() {
  emit('hideBackupPlaceModal');
}

const submit = handleSubmit((data: Data) => {
  emit('submitData', data);
});

</script>

<template>
  <Modal :show="backupPlacePopupShow" @close="hideBackupPlaceModal" @submit="submit" title="Задати власні значення для зберігання резервних копій реплікацій об’єктів S3">
    <form id="backupPlace-form">
      <TextField 
        label="Ім’я бакета"
        name="backupBucket"
        description="Довжина назви має бути від 3 до 63 символів. Допустимі символи “a-z”, “0-9”, “.”, “-”"
        :value="values.backupBucket"
        :error="errors?.backupBucket"
        @update="val => values.backupBucket = val" 
      />

      <TextField 
        label="Endpoint"
        name="endpoint"
        description="Наприклад: “https://endpoint.com”"
        :value="values.endpoint"
        :error="errors?.endpoint"
        @update="val => values.endpoint = val" 
      />

      <TextField 
        label="Логін"
        name="login"
        description="Надається постачальником послуги"
        :value="values.login"
        :error="errors?.login"
        @update="val => values.login = val" 
      />

      <TextField 
        label="Пароль"
        name="password"
        type="password"
        description="Надається постачальником послуги"
        :value="values.password"
        :error="errors?.password"
        @update="val => values.password = val" 
      />
    </form>
  </Modal>
</template>
