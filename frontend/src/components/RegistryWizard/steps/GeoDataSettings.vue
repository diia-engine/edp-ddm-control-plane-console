<script setup lang="ts">
import { toRefs, ref } from 'vue';
import Banner from '@/components/common/Banner.vue';
import Typography from '@/components/common/Typography.vue';
import ToggleSwitch from '@/components/common/ToggleSwitch.vue';

type GeoDataSettingsProps = {
  enabled: boolean;
  isEditAction: boolean;
};

const props = defineProps<GeoDataSettingsProps>();
const { isEditAction, enabled } = toRefs(props);
const enabledGeoData = ref(false);
</script>

<template>
  <Typography variant="h3" class="h3">{{ $t('components.geoDataSettings.title') }}</Typography>
  <template v-if="isEditAction">
    <div v-if="enabled" class="box">
      <img src="@/assets/img/status-active.png" />
      <Typography variant="bodyText" class="text">
        {{ $t('components.geoDataSettings.text.managementSubsystemNotDeployed') }}
      </Typography>
    </div>
    <div v-else class="box">
      <img src="@/assets/img/status-disabled.png" />
      <Typography variant="bodyText" class="text">
        {{ $t('components.geoDataSettings.text.managementSubsystemNotDeployed') }}
      </Typography>
    </div>
  </template>
  <template v-else>
    <ToggleSwitch
      name="geoServerEnabled"
      :label="$t('components.geoDataSettings.fields.geoServerEnabled.label')"
      v-model="enabledGeoData"
    />
    <Banner
      :description="$t('components.geoDataSettings.text.impossibleChangeAfterCreate')"
      class="banner"
    />
  </template>
</template>

<style scoped>
.h3 {
  margin-bottom: 24px;
}
.banner {
  margin-top: 8px;
}
.box {
  display: flex;
  align-items: center;
}
.text {
  margin-left: 8px;
}
</style>
