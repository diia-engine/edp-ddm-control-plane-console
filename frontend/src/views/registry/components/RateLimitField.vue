<script setup lang="ts">
import { toRefs, computed, withDefaults } from 'vue';

import { numberRegexp } from '@/constants/form';
import FormField from '@/components/common/FormField.vue';
import MenuList from '@/components/common/MenuList.vue';
import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
import IconButton from '@/components/common/IconButton.vue';
import type { PublicApiLimits } from '@/types/registry';

interface RateLimitFieldProps {
  label: string,
  name: string,
  errors?: {
    second?: string,
    minute?: string,
    hour?: string,
    day?: string,
    month?: string,
    year?: string,
    common?: string,
  },
  modelValue?: PublicApiLimits,
  required?: boolean,
}

const props = withDefaults(defineProps<RateLimitFieldProps>(), {
  modelValue: () => ({} as PublicApiLimits),
  required: false,
});
const emit = defineEmits(['update:modelValue']);
const { label, name, required, modelValue, errors } = toRefs(props);

const possibleIntervals = [
  {
    value: 'second',
    title: 'за секунду',
  },
  {
    value: 'minute',
    title: 'за хвилину',
  },
  {
    value: 'hour',
    title: 'за годину',
  },
  {
    value: 'day',
    title: 'за добу',
  },
  {
    value: 'month',
    title: 'за місяць',
  },
  {
    value: 'year',
    title: 'за рік',
  },
];
const selectedIntervals = computed(() => possibleIntervals.filter((item) => {
  const value: Record<string, unknown> = props.modelValue as Record<string, unknown> || {};
  return value[item.value] !== undefined;
}));
const items = computed(() => possibleIntervals.filter((item) => {
  const value: Record<string, unknown> = props.modelValue as Record<string, unknown> || {};
  return !Object.keys(props.modelValue || {}).find((key) => key === item.value && value[key] !== undefined);
}));

const onIntervalSelect = (item: { value: string }) => {  
  const value = modelValue?.value || {};
  emit('update:modelValue', {
    ...value,
    [item.value]: null,
  });
};

const onIntervalRemove = (item: { value: string }) => {  
  const value = modelValue?.value || {};
  emit('update:modelValue', {
    ...value,
    [item.value]: undefined,
  });
};
</script>

<template>
  <FormField :label="label" :name="name" :required="required" :error="errors?.common">
    <Typography variant="bodyText">Вкажіть ліміт мінімум для одного проміжку часу:</Typography>
    <div v-for="(interval, i) in selectedIntervals" :key="i" class="limit-block">
      <TextField
        :name="interval.value"
        :style="{ width: '120px' }"
        v-model="modelValue[interval.value as keyof PublicApiLimits]"
        :error="(errors?.[interval.value as keyof typeof errors] || '')"
        :allowed-characters="numberRegexp"
      />
      <Typography variant="bodyText" class="limit-title">{{ interval.title }}</Typography>
      <IconButton @onClick="onIntervalRemove(interval)">
        <img src="@/assets/svg/close.svg" alt="remove limit" />
      </IconButton>
    </div>
    <MenuList :items="items" :onItemClick="onIntervalSelect" v-if="items.length" class="limit-select" />
  </FormField>
</template>

<style lang="scss" scoped>
.content-text {
  margin-bottom: 24px;
}
.limit-block {
  display: flex;
  align-items: top;
}

.limit-title {
  flex-grow: 1;
  margin-left: 8px;
  margin-top: 14px;
}

:deep(.limit-select) {
  margin-top: 8px;
}
</style>

