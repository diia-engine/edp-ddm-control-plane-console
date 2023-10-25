<script setup lang="ts">
import IconButton from '@/components/common/IconButton.vue';
import Typography from '@/components/common/Typography.vue';
import { KEY_VARIANTS } from '@/constants/key';
import type { TableViewKey } from '@/types/cluster';
import { computed } from 'vue';

type KeysTableProps = {
  keys: TableViewKey[]
  title: string
}
defineProps<KeysTableProps>();

const emit = defineEmits(['onRemoveKeyClick', 'onEditKeyClick']);

const onOnRemoveKeyClick = (keyName: string) => {
  emit('onRemoveKeyClick', keyName);
};
const onEditKeyClick = (keyName: string) => {
  emit('onEditKeyClick', keyName);
};
const keyTypes = computed(() => KEY_VARIANTS());

</script>

<template>
  <template v-if="keys.length > 0">
    <Typography class="mt24" variant="h5">{{ title }}</Typography>
    <table class="rg-info-table rg-info-table-config mb24">
      <thead>
        <tr>
          <th>{{$t("components.keysManagement.table.techName")}}</th>
          <th class="key-type-column">{{$t("components.keysManagement.table.mediaType")}}</th>
          <th>{{$t("components.keysManagement.table.acsk")}}</th>
          <th>{{$t("components.keysManagement.table.allowedRegistries")}}</th>
          <th class="actions-header"></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="key in keys" v-bind:key="key.name">
          <td class="tech-name-cell cell">{{ key.name }}</td>
          <td class="cell">{{ keyTypes.find((el) => el.value === key.deviceType)?.title }}</td>
          <td class="cell">{{ key.issuer }}</td>
          <td class="cell">
            <div  v-for="reg in key.allowedRegistries" v-bind:key="reg">{{ reg }}</div>
          </td>
          <td class="actions-column">
            <IconButton @onClick="onEditKeyClick(key.name)" class="table-actions" :title="$t('actions.edit')">
              <img src="@/assets/img/action-edit.png" />
            </IconButton>
            <IconButton @onClick="onOnRemoveKeyClick(key.name)" class="table-actions" :title="$t('actions.remove')">
              <img src="@/assets/img/action-delete.png" />
            </IconButton>
          </td>
        </tr>
      </tbody>
    </table>
  </template> 
</template>

<style scoped lang="scss">
.tech-name-cell {
  word-wrap: break-word;
}
.key-type-column {
  width: 15%;
}
.table-actions {
  width: 24px;
  height: 24px;
}
.actions-header {
  width: 50px;
}
.actions-column {
  display: flex;
  gap: 12px;
  flex-wrap: nowrap;
}
.mt24 {
  margin-top: 24px;
}
.cell {
  vertical-align: top;
}
</style>