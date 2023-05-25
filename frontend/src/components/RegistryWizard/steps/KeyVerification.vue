<script setup lang="ts">
import Typography from '@/components/common/Typography.vue';
</script>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  props: {
    registryAction: String,
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
        this.caCertRequired = false;
        this.caJSONRequired = false;
        let validationFailed = false;

        const keyCaCert = this.$refs.keyCaCert as any;
        if (keyCaCert.files.length === 0) {
          this.caCertRequired = true;
          validationFailed = true;
        }

        const keyCaJSON = this.$refs.keyCaJSON as any;
        if (keyCaJSON.files.length === 0) {
          this.caJSONRequired = true;
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
  },
  data() {
    return {
      caCertRequired: false,
      changed: false,
      caJSONRequired: false,
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
  <div class="form-group">
    <Typography variant="h3">Дані для перевірки підписів</Typography>
  </div>
  <Typography variant="bodyText" class="key-sign-page-description">Внесені сертифікати АЦСК для перевірки ключів
    системного підпису та КЕП користувачів будуть застосовані для налаштувань поточного реєстру.</Typography>

  <input type="checkbox" style="display: none;" v-model="changed" name="key-verification-changed" />
  <div class="rc-form-group" :class="{ 'error': caCertRequired }">
    <label for="ca-cert">Публічні сертифікати АЦСК (розширення .p7b)</label>
    <input @change="changed = true; caCertRequired = false;" ref="keyCaCert" type="file"
           name="ca-cert" id="ca-cert" accept=".p7b" />
    <span v-if="caCertRequired">Обов’язкове поле</span>
  </div>
  <div class="rc-form-group" :class="{ 'error': caJSONRequired }">
    <label for="ca-json">Перелік АЦСК (розширення .json)</label>
    <input @change="changed = true; caJSONRequired = false;" ref="keyCaJSON" type="file"
           name="ca-json" id="ca-json" accept=".json" />
    <span v-if="caJSONRequired">Обов’язкове поле</span>
  </div>
</template>