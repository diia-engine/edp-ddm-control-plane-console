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
    :title="$t('components.registryBackupDeleteModal.title')"
    :submitBtnText="$t('actions.remove')"
    :show="backupDeletePlacePopupShow"
    @close="hideBackupDeletePlaceModal"
    @submit="submit"
    redButton
  >
    <Typography variant="bodyText">{{ $t('components.registryBackupDeleteModal.text.valuesByDefault') }}</Typography>
  </Modal>
</template>
