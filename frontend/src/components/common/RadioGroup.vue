<script setup lang="ts">
import { toRefs } from 'vue';
import Typography from '@/components/common/Typography.vue';

interface RadioGroupProps {
  name: string;
  items: {
    value: string;
    label: string;
  }[];
  modelValue?: any;
  classes?: string;
}
const $emit = defineEmits(['update:modelValue']);
const props = defineProps<RadioGroupProps>();
const { name, classes, modelValue } = toRefs(props);

const onChange = (event: Event) => {
  $emit('update:modelValue', (event.target as HTMLInputElement).value);
};
</script>
<template>
  <div :class="classes">
    <div class="radio" v-for="item in items" :key="item.value">
      <input
        type="radio"
        :name="name"
        :value="item.value"
        :id="item.value"
        :checked="modelValue === item.value"
        @input="onChange"
      />
      <label :for="item.value">
        <Typography variant="bodyText" class="labelText">
          {{ item.label }}
        </Typography>
      </label>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.radio {
  position: relative;
  display: flex;
  margin-top: 16px;
}

.radio:first-child {
  margin-top: 0;
}

.radio label {
  display: inline-flex;
  align-items: center;
  margin: 0;
}
.radio input[type='radio'] {
  display: none;
}

.labelText {
  display: inline;
}

.radio label:before {
  content: ' ';
  display: inline-block;
  position: relative;
  margin-right: 8px;
  width: 24px;
  height: 24px;
  border-radius: 15px;
  border: 2px solid $grey3;
  background-color: transparent;
}

.radio input[type='radio']:checked + label:before {
  border-color: $blue-main;
}

.radio input[type='radio']:checked + label:after {
  border-radius: 11px;
  width: 14px;
  height: 14px;
  position: absolute;
  top: 5px;
  left: 5px;
  content: ' ';
  display: block;
  background: $blue-main;
}

.radio input[type='radio']:checked:hover + label:after {
  background: $blue125;
}

.radio input[type='radio']:checked:hover + label:before {
  border-color: $blue125;
}
</style>
