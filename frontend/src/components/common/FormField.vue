<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
import { getErrorMessage } from '@/utils';

interface FormFieldProps {
  name?: string,
  label?: string,
  description?: string,
  required?: boolean,
  error?: string,
}

defineProps<FormFieldProps>();
</script>
<script lang="ts">
export default {
  inheritAttrs: false
};
</script>

<template>
  <div class="form-input-group" :class="{ 'error': error }">
    <label :for="name">
      {{ label }} <b v-if="required" class="red-star">*</b>
    </label>
    <slot></slot>
    <Typography v-if="error" class="form-input-group-error-message" variant="small">{{ getErrorMessage(error) }}</Typography>
    <Typography class="form-input-group-error-description" v-if="description" variant="small">{{ description }}</Typography>
  </div>
</template>

<style lang="scss" scoped>
.form-input-group {
  margin: 0 0 24px 0;
}

.form-input-group:last-of-type {
  margin: 0;
}

.form-input-group label {
  font-size: 16px;
  font-weight: bold;
  margin: 0 0 8px 0;
}

.form-input-group p {
  margin: 8px 0 0 0;
  font-size: 12px;
}

.form-input-group-error-message {
  color: $error-color;
}

.form-input-group-error-description {
  max-width: 464px;
}
</style>

