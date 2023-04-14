<script setup lang="ts">
import { toRefs } from 'vue';

interface CidrModalProps {
  cidrPopupShow: any;
  cidrFormatError: any;
}

const props = defineProps<CidrModalProps>();
const { cidrPopupShow, cidrFormatError } = toRefs(props);

</script>
<script lang="ts">
export default {
  data() {
    return {
      disabled: false,
      value: ''
    };
  },
  methods: {
    submit() {
      this.disabled = true;
      this.$emit('createCidr');
    },
    hideCIDRForm() {
      this.$emit('hideCidrForm');
    },
    updateValue(event: any) {
      this.$emit('update:modelValue', (event.target as HTMLInputElement).value);
    }
  },
  watch: {
    cidrPopupShow() {
      this.disabled = false;
      this.value = '';
    }
  }
};
</script>

<template>
  <div class="popup-backdrop visible" v-cloak v-if="cidrPopupShow"></div>
  <div class="popup-window admin-window visible" v-cloak v-if="cidrPopupShow">
    <div class="popup-header">
      <p>Додати CIDR</p>
      <a href="#" @click.stop.prevent="hideCIDRForm" class="popup-close hide-popup">
        <img alt="close popup window" src="@/assets/img/close.png" />
      </a>
    </div>
    <form @submit.prevent="submit" id="cidr-form" method="post" action="">
      <div class="popup-body">
        <p class="popup-error" v-cloak v-if="cidrFormatError">Перевірте формат IP-адреси</p>
        <div class="rc-form-group">
          <label for="cidr-value">IP-адреси та маски</label>
          <input id="cidr-value" type="text" v-model="value" @input="updateValue" />
          <p>Допустимі символи "0-9", "/", ".". Приклад: 172.16.0.0/12.</p>
        </div>
      </div>
      <div class="popup-footer active">
        <a href="#" id="cidr-cancel" class="hide-popup" @click="hideCIDRForm">відмінити</a>
        <button value="submit" name="cidr-apply" type="submit"
                :disabled="disabled && !cidrFormatError">Підтвердити</button>
      </div>
    </form>
  </div>
</template>
