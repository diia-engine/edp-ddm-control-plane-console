<script setup lang="ts">
import { toRefs } from 'vue';
import Typography from '@/components/common/Typography.vue';

interface RadioGroupProps {
  name: string;
  label: string;
  disabled?: boolean;
  modelValue?: any;
  classes?: string;
}
const $emit = defineEmits(['update:modelValue', 'change']);
const props = defineProps<RadioGroupProps>();
const { name, label, disabled, classes, modelValue } = toRefs(props);

const onChange = (event: Event) => {
  const checked = (event.target as HTMLInputElement).checked;
  $emit('update:modelValue', checked);
  $emit('change', checked);
};
</script>
<template>
  <div class="toggle-switch" :class="[classes, { disabled: disabled }]">
    <input
      class="switch-input"
      type="checkbox"
      :disabled="disabled"
      :id="name"
      v-model="modelValue"
      :name="name"
      @change="onChange"
    />
    <label :for="name">Toggle</label>
    <Typography variant="bodyText" class="span">{{ label }}</Typography>
  </div>
</template>

<style lang="scss" scoped>
.toggle-switch {
  display: flex;
  flex-direction: row;
}

.toggle-switch input[type='checkbox'] {
  height: 0;
  width: 0;
  visibility: hidden;
}

.toggle-switch label {
  cursor: pointer;
  text-indent: -9999px;
  width: 48px;
  height: 24px;
  background: grey;
  display: block;
  border-radius: 24px;
  position: relative;
}

.toggle-switch label:after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: $white-color;
  border-radius: 90px;
  transition: 0.3s;
}

.toggle-switch input:checked + label {
  background: $blue-main;
}

.toggle-switch input:checked + label:after {
  left: calc(100% - 2px);
  transform: translateX(-100%);
}

.toggle-switch label:active:after {
  width: 50px;
}

.toggle-switch .span {
  margin-left: 8px;
}

.disabled input:not(:checked) + label {
  background: $grey6;
  border: 1px solid $grey-border-color;
}
.disabled input:not(:checked) + label:after {
  top: 1px,
}

.disabled input:checked + label {
  background: $blue30;
}

.toggle-switch label:hover {
  background: $grey2;
}
.toggle-switch input:checked + label:hover {
  background: $blue125;
}
</style>
