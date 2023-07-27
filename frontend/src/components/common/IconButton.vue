<script setup lang="ts">
const emit = defineEmits(['onClick']);

const onClick = () => {
  emit('onClick');
};
</script>

<template>
      <a href="#" @click.stop.prevent="onClick" class="icon-button">
        <slot></slot>
      </a>
</template>

<style lang="scss" scoped>
.icon-button {
  width: 40px;
  height: 40px;
  display: flex;
  justify-content: center;
  align-items: center;

  &:hover {
    background-color: #E6F3FA;
  }
}
</style>

