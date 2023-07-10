<script setup lang="ts">
interface FormFieldProps {
  items: Array<{ value: string, title: string }>,
  onItemClick: (item: { value: string, title: string }) => void;
}

defineProps<FormFieldProps>();
</script>

<template>
  <v-menu>
    <template v-slot:activator="{ isActive, props }">
      <v-btn
        v-bind="props"
        variant="text"
        :class="`button ${$attrs.class}`"
      >
        Додати ліміти
        <span v-if="isActive">
          <img src="@/assets/svg/toggle.svg" alt="active toggle" :style="{ transform: 'rotateX(180deg)' }"/>
        </span>
        <span v-if="!isActive">
          <img src="@/assets/svg/toggle.svg" alt="non-active-toggle" />
        </span>
      </v-btn>
    </template>
    <v-list>
      <v-list-item v-for="(item, i) in items" :key="i" @click="onItemClick(item)" class="drop-item">
        <v-list-item-title>{{ item.title }}</v-list-item-title>
      </v-list-item>
    </v-list>
  </v-menu>
</template>

<style lang="scss" scoped>
.root {
  margin: 0 0 24px 0;
  display: flex;
  flex-direction: column;
}

.button {
  display: inline-block;
  color: #00689B;
  padding: 4px 8px;
  text-transform: none;
  border-radius: 4px;

  &:hover {
    background-color: #E6F3FA;
  }

  &:focus {
    outline: none;
  }

  & :deep(.v-btn__overlay),
  & :deep(.v-ripple__container) {
    display: none;
  }
}

.drop-item {
  &:hover {
    background-color: #F2F2F2
  }

  & :deep(.v-list-item__overlay),
  & :deep(.v-ripple__container) {
    display: none;
  }
}
</style>

