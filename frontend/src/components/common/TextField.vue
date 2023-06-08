<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
import { getErrorMessage } from '@/utils';

interface TextFieldProps {
  name?: string,
  label?: string,
  description?: string,
  error?: string,
  modelValue?: any,
  value?: HTMLInputElement['value']
  required?: boolean
  placeholder?: HTMLInputElement['placeholder'],
  rootClass?: string
}

defineProps<TextFieldProps>();
defineEmits(['update:modelValue']);
</script>
<script lang="ts">
export default {
  inheritAttrs: false
};
</script>

<template>
  <div class="form-input-group" :class="[ error ? 'error' : '', rootClass ? rootClass : '']">
    <label :for="name">
      {{ label }} <b v-if="required" class="red-star">*</b>
    </label>
    <input
      :name="name"
      :aria-label="name"
      :placeholder="placeholder"
      v-bind="$attrs"
      :value="modelValue ?? value"
      @input="$emit('update:modelValue', ($event.target as any).value)"
    />
    <Typography v-if="error" class="form-input-group-error-message" variant="small">{{ getErrorMessage(error) }}</Typography>
    <Typography class="form-input-group-error-description" v-if="description" variant="small">{{ description }}</Typography>
  </div>
</template>

<style lang="scss" scoped>
.form-input-group {
  margin: 0 0 24px 0;
  display: flex;
  flex-direction: column;
}

.form-input-group:last-of-type {
  margin: 0;
}

.form-input-group label {
  font-size: 16px;
  font-weight: bold;
  margin: 0 0 8px 0;
}

.form-input-group input {
  height: 40px;
  border: 1px solid $grey-border-color;
  background: $white-color;
  padding: 8px;

  &::placeholder {
    color: $black-color;
    opacity: 0.25;
  }
}

.form-input-group input:focus {
  outline: none;
}

.form-input-group p {
  margin: 8px 0 0 0;
  font-size: 12px;
}

.form-input-group textarea {
  background: $white-color;
  border: 1px solid $grey-border-color;
  border-radius: 2px;
}

.form-input-group textarea:focus {
  outline: none;
}

.form-input-group.error input {
  border: 1px solid $error-color;
}

.form-input-group.error select {
  border: 1px solid $error-color;
}

.form-input-group.error textarea {
  border: 1px solid $error-color;
}

.form-input-group input.error {
  border: 1px solid $error-color;
}

.form-input-group-error-message {
  color: $error-color;
}

.form-input-group-error-description {
  max-width: 464px;
}
</style>

