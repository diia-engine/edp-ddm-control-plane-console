<script setup lang="ts">
import { toRefs } from 'vue';

interface ButtonProps {
  faIconClass?: String,
  variant?: 'withIcon' | 'primary',
}

const props = defineProps<ButtonProps>();
const { faIconClass, variant = 'primary' } = toRefs(props);
const emit = defineEmits(['click']);
const onClick = () => {
  emit('click');
};
</script>
<script lang="ts">
export default {
  inheritAttrs: false
};
</script>

<template>
    <a v-bind="$attrs" v-if="variant === 'withIcon'" class="button-with-icon" href="#" @click.prevent="onClick">
      <i v-show="!!faIconClass" :class="faIconClass"></i>
      <div>
        <slot></slot>
      </div>
    </a>
    <button v-bind="$attrs" v-else-if="variant === 'primary'" class="primary-btn" @click="onClick">
      <i v-show="!!faIconClass" :class="faIconClass"></i>
      <slot></slot>
    </button>
</template>
<style scoped>
    .button-with-icon div {
      font-family: "Oswald", "MuseoSans", sans-serif;
      font-weight: 400;
      font-size: 18px;
      text-transform: uppercase;
      color: rgba(0, 0, 0, 0.5);
      margin-left: 13px;
      display: inline-block;
    }
    .primary-btn {
      color: white;
      display: flex;
      padding: 8px 16px;
      justify-content: center;
      align-items: center;
      font-family: Oswald;
      font-size: 18px;
      font-style: normal;
      font-weight: 400;
      line-height: 24px; /* 133.333% */
      text-transform: uppercase;
      border-radius: 4px;
      border: none;
      background: #6DA300;
    }
</style>