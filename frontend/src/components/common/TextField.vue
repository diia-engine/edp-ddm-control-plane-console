<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
import { getErrorMessage } from '@/utils';
import { watch, toRefs } from 'vue';

interface TextFieldProps {
  name?: string,
  label?: string,
  description?: string,
  error?: string,
  modelValue?: any,
  value?: HTMLInputElement['value']
  required?: boolean
  placeholder?: HTMLInputElement['placeholder'],
  rootClass?: string,
  allowedCharacters?: string,
}

const props = defineProps<TextFieldProps>();
const $emit = defineEmits(['update:modelValue']);
const { name, label, description, error, modelValue, required, placeholder, rootClass, allowedCharacters } = toRefs(props);

watch(modelValue, (value) => {
  const charactersRegexp = allowedCharacters?.value;
  if (charactersRegexp) {
    const match = value.match(RegExp(charactersRegexp, 'g'));
    const matchedValue = match?.length ? match.join('') : '';
    if (matchedValue !== value) {
      $emit('update:modelValue', matchedValue);
    }
  }
});

const onChange = (value: any, type: string) => {
  const val = type === 'number' ? +value : value;
  $emit('update:modelValue', val);
};
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
      @input="onChange(($event.target as any).value, $attrs.type as string)"
    />
    <Typography v-if="error" class="form-input-group-error-message" variant="small">{{ getErrorMessage(error) }}</Typography>
    <Typography v-if="description" class="form-input-group-error-description" variant="small">{{ description }}</Typography>
    <slot></slot>
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

  &::-webkit-outer-spin-button,
  &::-webkit-inner-spin-button {
    -webkit-appearance: none;
    margin: 0;
  }

  &[type=number] {
    -moz-appearance:textfield;
    appearance:textfield;
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

