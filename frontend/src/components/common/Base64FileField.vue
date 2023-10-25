<script setup lang="ts">
import { ref, type VNodeRef, type Ref } from 'vue';
import Typography from '@/components/common/Typography.vue';
import IconButton from '@/components/common/IconButton.vue';

interface Base64FileFieldProps {
  name: string,
  label: string,
  fileData: string,
  format: string,
  fileNameDescription?: string,
  error?: string,
}

const props = defineProps<Base64FileFieldProps>();
const fileInput: VNodeRef = ref(null);
const imageData: Ref<string | null> = ref(props.fileData);
const fileName: Ref<string | undefined> = ref(props.fileNameDescription);

const emit = defineEmits(['onSelect']);

const openFileDialog = () => {
  if (fileInput.value) {
    (fileInput.value as HTMLInputElement).click();
  }
};

const getDataFormat = () => {
  if (props.format === 'svg') {
    return 'data:image/svg+xml;base64';
  }
  return 'data:image/png;base64';
};

const selected = (event: Event) => {
  const inputElement = event.target as HTMLInputElement;
  const file = inputElement?.files?.[0];
  if (file) {
    const reader = new FileReader();
    reader.onloadend = () => {
      if (reader.result) {
        const decodedString = btoa(reader.result as string);
        const base64DataSplit = decodedString.split(',');
        const fileSplit = file.name.split('.');
        const format = fileSplit.length ? fileSplit[fileSplit.length - 1] : null;
        imageData.value = base64DataSplit[base64DataSplit.length - 1];
        emit('onSelect', imageData.value, format);
        if (!props.fileNameDescription) {
          fileName.value = file.name;
        }
      }
    };
    reader.readAsBinaryString(file);
  }
};

</script>

<template>
  <div class="file-wrapper">
    <label class="file-label">{{ label }}</label>
    <div class="file" :class="{ 'error': !!error }">
      <div class="file-image-wrapper" @click.stop.prevent="openFileDialog">
        <img class="file-image" :src="`${getDataFormat()},${imageData}`" />
      </div>
      <div class="file-description">
        <Typography variant="bodyText">
          {{ fileName }}
        </Typography>
      </div>
      <div class="file-actions">
        <IconButton class="file-action" @onClick="openFileDialog">
          <img src="@/assets/svg/update.svg" :title="$t('actions.replace')"/>
        </IconButton>
      </div>
      <input type="file" @change="selected" ref="fileInput" class="file-input" :accept="`.${format}`" />
    </div>
    <Typography variant="small" class="file-error-message" v-if="error">{{ error }}</Typography>
  </div>
</template>

<style lang="scss" scoped>
  input[type=file] {
    display: none;
  }

  .file-label {
    font-size: 16px;
    font-weight: bold;
    margin: 0 0 8px 0;
  }

  .file {
    display: flex;
    align-items: stretch;
    width: 480px;
    min-height: 32px;
    border-width: 1px;
    border-style: solid;
    border-color: #BFBFBF;
  }

  .error {
    border-color: $error-color;
  }

  .file-description {
    margin: 16px 8px;
    width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 4;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .file-image-wrapper {
    width: 128px;
    max-height: 128px;
    min-height: 56px;
    margin: 0;
    padding: 8px;
    display: flex;
    justify-content: center;
    align-items: center;
    flex-shrink: 0;
    background-color: #BFBFBF;
    box-sizing: border-box;
    cursor: pointer;
  }

  .file-image {
    max-width: 112px;
    max-height: 112px;
  }
  .file-actions {
    display: flex;
    margin: 8px 0;
  }

  .file-action {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 40px;
    height: 40px;
    margin: 0 8px;
    cursor: pointer;
  }

  .file-error-message {
    margin-top: 8px;
    color: $error-color;
  }

</style>
