<script lang="ts">
import Typography from '@/components/common/Typography.vue';

export default {
  components: { Typography },
  props: {
    show: { type: Boolean },
    redButton: { type: Boolean },
    title: { readonly: true, type: String },
    submitBtnText: { default: 'Підтвердити', type: String },
    hasCancelBtn: { default: true, type: Boolean },
  },
  updated() {
    if (this.show) {
      document.body.classList.add('body-modal-shown');
    } else {
      document.body.classList.remove('body-modal-shown');
    }
  },
  methods: {
    close() {
      this.$emit('close');
    },
    submit() {
      this.$emit('submit');
    },
  },
};
</script>

<template>
  <div>
    <div class="common-modal-backdrop" v-cloak v-if="show"></div>

    <div class="common-modal-window-wrapper" v-cloak v-if="show">
      <div class="common-modal-window" v-cloak v-if="show" @click.stop.prevent>
        <div class="common-modal-header">
            <div class="common-modal-title">
              <Typography variant="h3">{{ title }}</Typography>
            </div>
            <a href="#" @click.stop.prevent="close" class="common-modal-close common-modal-hide">
                <img alt="close modal window" class="common-modal-close-icon" src="@/assets/img/close.png" />
            </a>
        </div>

        <div class="popup-body">
          <slot></slot>
        </div>

        <div class="common-modal-footer">
          <button v-if="hasCancelBtn" class="common-modal-cancel" @click.stop.prevent="close">Відмінити</button>
          <button class="submit-button" :class="redButton && 'red-button'" @click.stop.prevent="submit">{{ submitBtnText }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>

.common-modal-backdrop {
  position: absolute;
  width: 100%;
  height: 100%;
  background: $black-color;
  opacity: 0.5;
  z-index: 1000;
  top: 0;
  left: 0;
}

.common-modal-close-icon {
  width: 14px;
  height: 14px;
  display: block;
  margin-right: 17px;
}

.common-modal-window {
  z-index: 1001;
  width: 496px;
  background: $white-color;
  opacity: 1;
  box-shadow: 0 6px 20px -5px $shadow-window-color;
  border-radius: 4px;
  padding: 8px;
  margin: auto;
}

.common-modal-window-wrapper {
  position: fixed;
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 1001;
  overflow-y: auto;
  left: 0; top: 0; right: 0; bottom: 0;
}

.common-modal-header {
  display: flex;
  justify-content: space-between;
  border-bottom: 1px solid $grey-border-color;
  align-items: baseline;
}

.common-modal-title {
  margin: 8px 0 15px 8px;
}

.common-modal-footer {
  display: flex;
  justify-content: flex-end;
  padding: 16px 8px 8px 8px;
}

.common-modal-cancel {
  font-family: 'Oswald', sans-serif;
  font-size: 18px;
  text-transform: uppercase;
  color: rgba(0, 0, 0, 0.5);
  background: none;
  border: none;
  padding: 8px 16px 8px 16px;
  cursor: pointer;
}

.common-modal-cancel:hover {
  border: none;
}

.submit-button {
  font-family: 'Oswald', sans-serif;
  font-size: 18px;
  text-transform: uppercase;
  border-radius: 4px;
  color: $white-color;
  border: none;
  padding: 8px 16px 8px 16px;
  cursor: pointer;
  background: $success-color;
}

.red-button {
  background: $error-color;
}

</style>
