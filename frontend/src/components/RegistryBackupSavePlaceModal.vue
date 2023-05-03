<script lang="ts">
import Modal from '@/components/common/Modal.vue';
import TextField from '@/components/common/TextField.vue';
import * as yup from 'yup';
import { useForm } from 'vee-validate';

interface Data {
  backupBucket: string;
  endpoint: string;
  login: string;
  password: string;
}

yup.setLocale({
    mixed: {
        required: 'required',
    },
    string: {
        min: 'checkFormat',
        max: 'checkFormat',
        matches: 'checkFormat',
    },
});

const validationSchema = yup.object({
    backupBucket: yup.string().required().min(3).max(63).matches(/^[a-z0-9][a-z0-9.-]*$/i),
    endpoint: yup.string().required().matches(/^[a-z0-9.\-/:]+$/i),
    login: yup.string().required(),
    password: yup.string().required(),
});

export default {
  props: {
    backupPlacePopupShow: { type: Boolean },
    initialData: { type: Object },
  },
  components: { Modal, TextField },
  data() {
    return {
      handleSubmit: null as any,
      setValues: null as any,
      values: {} as Data,
      errors: {} as Data,
    };
  },
  mounted() {
    const { handleSubmit, setValues, values, errors } = useForm({
      validationSchema, initialValues: this.initialData
    });
    this.handleSubmit = handleSubmit;
    this.setValues = setValues;
    this.values = values as Data;
    this.errors = errors as unknown as Data;
  },
  updated() {
    this.setValues(this.initialData);
  },
  methods: {
    submit() {
      const validate = this.handleSubmit?.(async (data: Data) => {
        this.$emit('submitData', data);
      });

      validate(this.values);
    },
    hideBackupPlaceModal() {
      this.$emit('hideBackupPlaceModal');
    },
  },
};
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
