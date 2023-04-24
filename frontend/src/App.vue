<script setup lang="ts">
import { RouterView } from 'vue-router';
</script>

<script lang="ts">
import { setLocale } from 'yup';

setLocale({
  mixed: {
    required: 'required',
  },
});
</script>

<template>
  <RouterView />
</template>
