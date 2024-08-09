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
  <Typography variant="h3" class="h3">Підсистема управління геоданими</Typography>
  <template v-if="isEditAction">
    <div v-if="enabled" class="box">
      <img src="@/assets/img/status-active.png" />
      <Typography variant="bodyText" class="text">
        Підсистема управління геоданими розгорнута для цього реєстра.
      </Typography>
    </div>
    <div v-else class="box">
      <img src="@/assets/img/status-disabled.png" />
      <Typography variant="bodyText" class="text">
        Підсистема управління геоданими не розгорнута для цього реєстра.
      </Typography>
    </div>
  </template>
  <template v-else>
    <ToggleSwitch
      name="geoServerEnabled"
      label="Розгорнути підсистему управління геоданими"
      v-model="enabledGeoData"
    />
    <Banner
      description="Після створення реєстру змінити ці налаштування буде неможливо."
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
