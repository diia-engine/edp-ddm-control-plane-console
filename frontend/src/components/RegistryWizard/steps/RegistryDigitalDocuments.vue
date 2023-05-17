<script setup lang="ts">
import { toRefs, ref, computed } from 'vue';
import { object, string } from 'yup';
import { useForm } from 'vee-validate';
import Modal from '@/components/common/Modal.vue';
import TextField from '@/components/common/TextField.vue';

const MAX_FILE_SIZE = '100';

interface RegistryDigitalDocumentsProps {
  maxFileSizeProp: string;
  maxTotalFileSizeProp: string;
}

const validationRules = () => {
  return string()
    .required()
    .max(4)
    .matches(/^(?:[1-9][0-9]{0,3}(?:\.[1-9])?|\d{1}\.\d{1})?$/)
    .test({
      message: 'moreThanMaxValue',
      test: function (value) {
        return +MAX_FILE_SIZE >= +value;
      },
    });
};

const validationSchema = object({
  maxFileSize: validationRules()
    .test({
      message: 'moreThanMaxValue',
      test: function (value) {
        const parentMaxTotalFileSize = +(isNaN(+this.parent.maxTotalFileSize) ? MAX_FILE_SIZE : this.parent.maxTotalFileSize);
        return +value <= parentMaxTotalFileSize;
      },
    }),
  maxTotalFileSize: validationRules(),
});


const showMaxFileSizePopUp = ref(false);
const showMaxTotalFileSizePopUp = ref(false);
const props = defineProps<RegistryDigitalDocumentsProps>();
const { maxFileSizeProp, maxTotalFileSizeProp } = toRefs(props);


const { errors, useFieldModel, validate } = useForm({
  validationSchema,
  initialValues: {
    maxFileSize: maxFileSizeProp.value?.replace('MB', '') || MAX_FILE_SIZE,
    maxTotalFileSize: maxTotalFileSizeProp.value?.replace('MB', '') || MAX_FILE_SIZE
  }
});

const [maxFileSize, maxTotalFileSize] = useFieldModel(['maxFileSize', 'maxTotalFileSize']);

const preparedDigitalDocuments = computed(() => JSON.stringify({
  maxFileSize: maxFileSize.value + 'MB',
  maxTotalFileSize: maxTotalFileSize.value + 'MB',
}));

function handleShowMaxFileSizePopUp() {
  showMaxFileSizePopUp.value = !showMaxFileSizePopUp.value;
}

function handleShowMaxTotalFileSize() {
  showMaxTotalFileSizePopUp.value = !showMaxTotalFileSizePopUp.value;
}

function validator() {
  return new Promise((resolve) => {
    validate().then((res) => {
      if (res.valid) {
        resolve(true);
      }
    });
  });
}

defineExpose({
  validator
});
</script>

<template>
  <h2>Цифрові документи</h2>
  <p>Управління обмеженнями на завантаження файлів цифрових документів до реєстру користувачами та бізнес-процесами.</p>
  <div class="wizard-warning">Адміністративні обмеження діють в рамках обмежень на максимальний розмір запиту на
    завантаження в 100 МБ, встановленого на рівні Платформи.</div>
  <input type="hidden" name="digital-documents" :value="preparedDigitalDocuments" />
  <div class="rc-form-group">
    <TextField label="Максимальний розмір файлу для завантаження, MB *" name="maxFileSize" :value="maxFileSize"
      @update="val => maxFileSize = val" :error="errors.maxFileSize" />
    <p>Допустимі символи: “0-9”, “.”. Значення не може перевищувати довжину у 4 символи. <a href="#"
        @click.stop.prevent="handleShowMaxFileSizePopUp">Детальніше</a>.</p>
  </div>

  <div class="rc-form-group">
    <TextField label="Макс. сумарний розмір групи файлів для завантаження, MB *" name="maxTotalFileSize"
      :value="maxTotalFileSize" @update="val => maxTotalFileSize = val" :error="errors.maxTotalFileSize" />
    <p>Допустимі символи: “0-9”, “.”. Значення не може перевищувати довжину у 4 символи. <a href="#"
        @click.stop.prevent="handleShowMaxTotalFileSize">Детальніше</a>.</p>
  </div>

  <Modal title="Максимальний розмір файлу для завантаження" submitBtnText="Зрозуміло" :hasCancelBtn=false
    :show="showMaxFileSizePopUp" @close="handleShowMaxFileSizePopUp" @submit="handleShowMaxFileSizePopUp">
    <p>Значення, що вводяться, зчитуються в МВ, можуть бути десятковим дробом з крапкою в якості розділового знаку.</p>
    <p>Значення не може перевищувати обмеження, встановлене на рівні системи. Обмеження застосовується до файлів, які
      завантажуються користувачами та бізнес-процесами. Додаткові обмеження можуть бути встановлені на рівні окремих
      файлових полей при моделюванні UI-форм</p>
    <p>Значення максимального розміру файлу не може перевищувати максимальний сумарний розмір групи файлів для
      завантаження.</p>
    <p>Значення використовується для визначення параметру “File Maximum Size” в компоненті File конструктора UI-форм.</p>
  </Modal>

  <Modal title="Макс. сумарний розмір групи файлів для завантаження" submitBtnText="Зрозуміло" :hasCancelBtn=false
    :show="showMaxTotalFileSizePopUp" @close="handleShowMaxTotalFileSize" @submit="handleShowMaxTotalFileSize">
    <p>Значення, що вводяться, зчитуються в МВ, можуть бути десятковим дробом з крапкою в якості розділового знаку.</p>
    <p>Значення не може перевищувати обмеження, встановлене на рівні системи. Обмеження застосовується до групи файлів,
      які завантажуються користувачами через файлові поля UI-форми.</p>
    <p>Значення використовується для визначення параметру “Maximum total size” в компоненті File конструктора UI-форм.</p>
  </Modal>
</template>
