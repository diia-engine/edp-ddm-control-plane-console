<script lang="ts">
import axios from 'axios';
import Modal from '@/components/common/Modal.vue';
import Typography from '@/components/common/Typography.vue';

export default {
  props: {
    publicApiPopupShow: { type: Boolean },
    publicApiName: { type: String },
    registry: { type: String },
  },
  components: { Modal, Typography },
    methods: {
      submit() {
        const formData = new FormData();

        formData.append("reg-name", this.publicApiName || '');
        axios.post(`/admin/registry/public-api-delete/${this.registry}`, formData).then(() => {
          window.location.assign(`/admin/registry/view/${this.registry}`);
        });
      },
      hideModalWindow() {
        this.$emit('hideModalWindow');
      },
    },
};
</script>

<template>
  <Modal
    :title="`Bидалити “${publicApiName}” з переліку?`"
    submitBtnText="Видалити"
    :show="publicApiPopupShow"
    @close="hideModalWindow"
    @submit="submit"
    redButton
  >
    <Typography variant="bodyText">Ви зможете надати доступ знову пізніше.</Typography>
  </Modal>
</template>
