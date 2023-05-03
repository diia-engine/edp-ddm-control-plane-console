<script lang="ts">
import Modal from '@/components/common/Modal.vue';
import Typography from '@/components/common/Typography.vue';

export default {
  props: {
    backupDeletePlacePopupShow: { type: Boolean },
  },
  components: { Modal, Typography },
    methods: {
      submit() {
        this.$emit('submitData');
      },
      hideBackupDeletePlaceModal() {
        this.$emit('hideBackupDeletePlaceModal');
      },
    },
};
</script>

<template>
  <Modal
    title="Видалити власні значення для зберігання резервних копій реплікацій об’єктів S3?"
    :show="backupDeletePlacePopupShow"
    @close="hideBackupDeletePlaceModal"
    @submit="submit"
    redButton
  >
    <Typography variant="bodyText">Будуть використані значення за замовчуванням.</Typography>
  </Modal>
</template>
