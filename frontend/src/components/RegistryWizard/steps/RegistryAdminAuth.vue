<script lang="ts" setup>
import { ref } from 'vue';
import Typography from '@/components/common/Typography.vue';
import ToggleSwitch from '@/components/common/ToggleSwitch.vue';
import { PORTALS } from '@/types/registry';

interface RegistryAdminAuthProps {
  isEnabledPortal: boolean;
}

const props = defineProps<RegistryAdminAuthProps>();
const isEnabledPortal = ref(props.isEnabledPortal);
const portal = ref(props.isEnabledPortal ? '' : PORTALS.admin);

function handleEnabledPortalChange(enabled: boolean) {
  portal.value = enabled ? '' : PORTALS.admin;
}
</script>

<template>
  <Typography variant="h3" class="h3">{{ $t('components.registryAdminAuth.title') }}</Typography>
  <input type="hidden" name="excludePortals[]" :value="portal"/>
  <ToggleSwitch
    name="enabledAdminPortal"
    :label="$t('components.registryAdminAuth.fields.enabledAdminPortal.label')"
    v-model="isEnabledPortal"
    @change="handleEnabledPortalChange"
  />
</template>

<style scoped>
.h3 {
  margin-bottom: 24px;
}
</style>
