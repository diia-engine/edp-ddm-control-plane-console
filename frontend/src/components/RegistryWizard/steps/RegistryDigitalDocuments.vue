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
  <h2>{{ $t('components.registryDigitalDocuments.title') }}</h2>
  <p>{{ $t('components.registryDigitalDocuments.text.managementOfUploadingFiles') }}</p>
  <div class="wizard-warning">{{ $t('components.registryDigitalDocuments.text.administrativeRestrictionsMaximumRequest') }}</div>
  <input type="hidden" name="digital-documents" :value="preparedDigitalDocuments" />
  <div class="rc-form-group">
    <TextField :label="$t('components.registryDigitalDocuments.fields.maxFileSize.label')" name="maxFileSize"
      v-model="maxFileSize" required :error="errors.maxFileSize" />
    <p>{{ $t('components.registryDigitalDocuments.text.maximumSizeForDownloadValidation') }} <a href="#"
        @click.stop.prevent="handleShowMaxFileSizePopUp">{{ $t('components.registryDigitalDocuments.actions.moreDetails') }}</a>.</p>
  </div>

  <div class="rc-form-group">
    <TextField :label="$t('components.registryDigitalDocuments.fields.maxTotalFileSize.label')" name="maxTotalFileSize"
      v-model="maxTotalFileSize" required :error="errors.maxTotalFileSize" />
    <p>{{ $t('components.registryDigitalDocuments.text.maximumSizeForDownloadValidation') }} <a href="#"
        @click.stop.prevent="handleShowMaxTotalFileSize">{{ $t('components.registryDigitalDocuments.actions.moreDetails') }}</a>.</p>
  </div>

  <Modal :title="$t('components.registryDigitalDocuments.text.maximumSizeForDownload')" :submitBtnText="$t('actions.gotIt')" :hasCancelBtn=false
    :show="showMaxFileSizePopUp" @close="handleShowMaxFileSizePopUp" @submit="handleShowMaxFileSizePopUp">
    <p>{{ $t('components.registryDigitalDocuments.text.valueFormat') }}</p>
    <p>{{ $t('components.registryDigitalDocuments.text.valueLimit') }}</p>
    <p>{{ $t('components.registryDigitalDocuments.text.fileCannotExceedTotalSize') }}</p>
    <p>{{ $t('components.registryDigitalDocuments.text.valueMaximumSize') }}</p>
  </Modal>

  <Modal :title="$t('components.registryDigitalDocuments.text.maxTotalSize')" :submitBtnText="$t('actions.gotIt')" :hasCancelBtn=false
    :show="showMaxTotalFileSizePopUp" @close="handleShowMaxTotalFileSize" @submit="handleShowMaxTotalFileSize">
    <p>{{ $t('components.registryDigitalDocuments.text.valueFormat') }}</p>
    <p>{{ $t('components.registryDigitalDocuments.text.valueGroupLimit') }}</p>
    <p>{{ $t('components.registryDigitalDocuments.text.valueGroupMaximumSize') }}</p>
  </Modal>
</template>
