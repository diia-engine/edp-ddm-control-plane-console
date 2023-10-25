<script setup lang="ts">
import { defineProps, toRefs } from 'vue';
import Typography from './Typography.vue';
import { getErrorMessage } from '@/utils';

interface SelectProps<T = unknown> {
  items: T[],
  label: string
  rootClass?: string
  description?: string,
  error?: string,
  required?: boolean
}

const props = defineProps<SelectProps>();

const { label, items, rootClass, description, error, required } = toRefs(props);

</script>
<script lang="ts">
export default {
  inheritAttrs: false
};
</script>
<template>
  <div :class="rootClass" v-show="!$attrs.hidden">
    <Typography variant="subheading" class="mb8">
      {{ label }} <b v-if="required" class="red-star">*</b>
    </Typography>
    <v-select v-bind:="$attrs" :class="['input', error && 'error']" density="compact" variant="outlined" :items="items" />
    <Typography v-if="error" class="form-input-group-error-message" variant="small">{{ getErrorMessage(error) }}</Typography>
    <Typography v-if="description" class="form-input-group-error-description" variant="small">{{ description }}</Typography>
  </div>
</template>

<style lang="scss" scoped>
  .error {
    & :deep(.v-field__outline__start) {
      border-color: $error-color !important;
    }
    & :deep(.v-field__outline__end) {
      border-color: $error-color !important;
    }
  }
  .input {
    & :deep(.v-field__input) {
      align-items: center;
    }
    & :deep(.v-input__details) {
      display: none;
    }
    & :deep(div.v-field__input > input) {
      opacity: 0;
    }
    & :deep(.v-field__outline__start) {
      border-radius: 0;
      opacity: 1;
      border-color: $grey3;
      &:hover {
        border-color: $black-color;
      }
      &:focus {
        border-color: $blue-main;
      }
    }
    & :deep(.v-field__outline__end) {
      border-radius: 0;
      opacity: 1;
      border-color: $grey3;
      &:hover {
        border-color: $black-color;
      }
      &:focus {
        border-color: $blue-main;
      }
    }
  }
  .mb8 {
    margin-bottom: 8px;
  }
  .form-input-group-error-message {
  color: $error-color;
  margin-top: 8px;
}

.form-input-group-error-description {
  max-width: 464px;
  margin-top: 8px;
}
</style>