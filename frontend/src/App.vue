<script setup lang="ts">
import { RouterView } from 'vue-router';
</script>

<script lang="ts">
import * as Yup from 'yup';

Yup.setLocale({
  mixed: {
    required: 'required',
  },
});

// Override the original URL validation
// workaround for Yup issue for urls that start with '//'
// https://github.com/jquense/yup/issues/672
const originalUrl = Yup.string().url;
Yup.addMethod(Yup.string, 'url', function (...args) {
  return originalUrl.call(this, ...args).test((value => !value?.startsWith('//')));
});
</script>

<template>
  <RouterView />
</template>
