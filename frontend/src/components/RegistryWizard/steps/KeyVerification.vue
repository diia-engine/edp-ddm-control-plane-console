<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
import FileField from '@/components/common/FileField.vue';
</script>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  props: {
    registryAction: String,
    pageDescription: String,
  },
  methods: {
    validator() {
      return new Promise<void>((resolve, reject) => {
        if (this.registryAction === "edit" && !this.changed) {
          resolve();
          return true;
        }

        this.validated = false;
        this.beginValidation = true;
        this.caCertError = '';
        this.caJSONError = '';
        let validationFailed = false;

        if (!this.caCertSelected) {
          this.caCertError = 'Обов’язкове поле';
          validationFailed = true;
        }

        if (!this.caJSONSelected) {
          this.caJSONError = 'Обов’язкове поле';
          validationFailed = true;
        }

        if (validationFailed) {
          reject();
          return false;
        }

        this.beginValidation = false;
        this.validated = true;
        resolve();
        return true;
      });
    },
    onCACertFileSelected(){
      this.caCertSelected = true;
      this.caCertError = '';
      this.changed = true;
    },
    onCACertFileReset(){
      this.caCertSelected = false;
    },
    onCAJSONFileSelected(){
      this.caJSONSelected = true;
      this.caJSONError = '';
      this.changed = true;
    },
    onCAJSONFileReset(){
      this.caJSONSelected = false;
    },
  },
  data() {
    return {
      caCertError: '',
      caCertSelected: false,
      changed: false,
      caJSONError: '',
      caJSONSelected: false,
      validated: false,
      beginValidation: false,
    };
  }
});

</script>

<style scoped>
  .key-sign-page-description {
    margin-bottom: 32px;
  }
</style>

<template>
  <h2>Дані для перевірки підписів</h2>
  <Typography variant="bodyText" class="key-sign-page-description">{{ pageDescription }}</Typography>

  <input type="checkbox" style="display: none;" v-model="changed" name="key-verification-changed" />

  <FileField label="Публічні сертифікати АЦСК (CACertificate.p7b)" sub-label="Обрати файл" name="ca-cert" accept=".p7b"
             :error="caCertError" @selected="onCACertFileSelected" @reset="onCACertFileReset" id="ca-cert-upload" />

  <FileField label="Перелік АЦСК (розширення .json)" sub-label="Обрати файл" name="ca-json" accept=".json"
             :error="caJSONError" @selected="onCAJSONFileSelected" @reset="onCAJSONFileReset" id="ca-json-upload" />
</template>