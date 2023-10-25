<script setup lang="ts">
import { toRefs, computed, withDefaults } from 'vue';

import { numberRegexp } from '@/constants/form';
import FormField from '@/components/common/FormField.vue';
import MenuList from '@/components/common/MenuList.vue';
import Typography from '@/components/common/Typography.vue';
import TextField from '@/components/common/TextField.vue';
import IconButton from '@/components/common/IconButton.vue';
import type { PublicApiLimits } from '@/types/registry';
import i18n from '@/localization';

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
    title: i18n.global.t('domains.registry.rateLimitField.second'),
  },
  {
    value: 'minute',
    title: i18n.global.t('domains.registry.rateLimitField.minute'),
  },
  {
    value: 'hour',
    title: i18n.global.t('domains.registry.rateLimitField.hour'),
  },
  {
    value: 'day',
    title: i18n.global.t('domains.registry.rateLimitField.day'),
  },
  {
    value: 'month',
    title: i18n.global.t('domains.registry.rateLimitField.month'),
  },
  {
    value: 'year',
    title: i18n.global.t('domains.registry.rateLimitField.year'),
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
    <Typography variant="bodyText">{{ $t('domains.registry.rateLimitField.specifyLimitTimePeriod') }}:</Typography>
    <FormField
      v-for="(interval, i) in selectedIntervals"
      :key="i"
      :error="(errors?.[interval.value as keyof typeof errors] || '')"
      class="limit-field"
    >
      <div class="limit-block">
        <TextField
          :name="interval.value"
          :style="{ width: '120px' }"
          v-model="modelValue[interval.value as keyof PublicApiLimits]"
          :allowed-characters="numberRegexp"
        />
        <Typography variant="bodyText" class="limit-title">{{ interval.title }}</Typography>
        <IconButton @onClick="onIntervalRemove(interval)" class="limit-remove-icon">
          <img src="@/assets/svg/close.svg" alt="remove limit" />
        </IconButton>
      </div>
    </FormField>
    <MenuList :items="items" :onItemClick="onIntervalSelect" v-if="items.length" class="limit-select" />
  </FormField>
</template>

<style lang="scss" scoped>
.content-text {
  margin-bottom: 24px;
}

.limit-field {
  margin: 0;
}

.limit-remove-icon {
  margin-top: 8px;
}
.limit-block {
  display: flex;
  align-items: start;
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

