<script setup lang="ts">
import { getErrorMessage } from '@/utils';
</script>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  props: {
    label: String,
    subLabel: String,
    name: String,
    error: String,
    accept: String,
    id: String,
  },
  data() {
    return {
      fileSelected: false,
    };
  },
  methods: {
    reset() {
      this.fileSelected = false;
      (this.$refs.fileInput as HTMLFormElement).value = '';
      this.$emit('reset', this.$refs.fileInput);
    },
    selected: function (){
      this.fileSelected = true;
      this.$emit('selected', this.$refs.fileInput);
    },
  },
});
</script>

<style scoped>
  input[type=file] {
    display: none;
  }

  .uploaded {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
  }

  .uploaded div {
    display: flex;
    align-items: baseline;
  }

  .uploaded div > .fa-check {
    color: #6DA300;
  }

  .uploaded span {
    display: block;
    font-size: 16px;
    color: #000000;
    margin-left: 11px;
  }

  .uploaded a > .fa-trash {
    color: #D35D47;
  }
</style>

<template>
  <div class="rc-form-group" :class="[$attrs.class, error && 'error']">
    <label>{{ label }}</label>
    <label v-show="!fileSelected" :for="id"
           class="rc-form-upload-block">
      <i class="fa-solid fa-plus"></i>
      <span>{{ subLabel }}</span>
    </label>
    <div v-show="fileSelected" class="uploaded">
      <div>
        <i class="fa-solid fa-check"></i>
        <span>{{ $t('components.fileField.text.fileLoaded') }}</span>
      </div>
      <a href="#" @click.stop.prevent="reset"><i class="fa-solid fa-trash"></i></a>
    </div>
    <span v-if="error">{{ getErrorMessage(error) }}</span>
    <input :id="id" type="file" v-bind="$attrs" @change="selected" :name="name" ref="fileInput" :accept="accept" />
  </div>
</template>
