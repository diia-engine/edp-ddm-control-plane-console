<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
  data() {
    return {
      pageRoot: this.$parent?.$parent as any,
    };
  },
});
</script>

<template>
  <h2>Автентифікація отримувачів послуг</h2>
  <p>Перевірка даних з КЕП користувачів в ЄДР відбувається за умови налаштованої інтеграції поточного реєстру з ЄДР через
    ШБО Трембіта.</p>
  <div class="toggle-switch">
    <input class="switch-input" type="checkbox" id="edr-check-input" name="edr-check-enabled"
           v-model="pageRoot.$data.wizard.tabs.recipientAuthentication.data.edrCheckEnabled"/>
    <label for="edr-check-input">Toggle</label>
    <span>Перевіряти наявність активного запису в ЄДР для бізнес-користувачів</span>
  </div>
</template>
