<script setup lang="ts">
import { toRefs, computed } from 'vue';
import Typography from '@/components/common/Typography.vue';

interface BannerProps {
  description: string,
  title?: string,
  color?: string,
  classes?: string,
}

const props = defineProps<BannerProps>();
const { title, description, color, classes } = toRefs(props);

const cssProps = computed(() => ({'background-color': color?.value ?? '#FFFBE6'}));
</script>
<template>
  <div class="banner" :style="cssProps" :class="classes">
      <Typography v-if="title" class="banner-title" variant="h5">
        {{ title }}
      </Typography>
      <Typography class="banner-body" variant="small">
        {{ description }}
      </Typography>
  </div>
</template>

<style lang="scss" scoped>
.banner {
  padding: 16px 24px;
}

.banner .banner-title {
  margin-bottom: 12px;
  text-transform: uppercase;
}

.banner-body {
  white-space: pre-line;
}
</style>
