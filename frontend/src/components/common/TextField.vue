<script lang="ts">
import Typography from '@/components/common/Typography.vue';
import { getErrorMessage } from '@/utils';

export default {
  components: { Typography },
  props: {
    name: { readonly: true, type: String },
    label: { readonly: true, type: String },
    description: { type: String },
    value: { readonly: true, type: String },
    error: { type: String },
    type: { default: 'text', readonly: true, type: String },
    placeholder: { type: String },
  },
  methods: {
    getErrorMessage(key: string) {
      return getErrorMessage(key);
    }
  },
  computed: {
    inputVal: {
      get(): string {
        return this.value || '';
      },
      set(val: string): void {
        this.$emit('update', val);
      }
    }
  }
};

</script>

<template>
  <div class="form-input-group" :class="{ 'error': error }">
    <label :for="name">{{ label }}</label>
    <input :name="name" :aria-label="name" :type="type" :placeholder="placeholder" v-model="inputVal" v-on="$attrs" />
    <div v-if="error" class="form-input-group-error-message">
      <Typography variant="small">{{ getErrorMessage(error) }}</Typography>
    </div>
    <div class="form-input-group-error-description" v-if="description">
      <Typography variant="small">{{ description }}</Typography>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.form-input-group {
  margin: 0 0 24px 0;
  display: flex;
  flex-direction: column;
}

.form-input-group:last-of-type {
  margin: 0;
}

.form-input-group label {
  font-size: 16px;
  font-weight: bold;
  margin: 0 0 8px 0;
}

.form-input-group input {
  height: 40px;
  border: 1px solid $grey-border-color;
  background: $white-color;
  padding: 8px;

  &::placeholder {
    color: $black-color;
    opacity: 0.25;
  }
}

.form-input-group input:focus {
  outline: none;
}

.form-input-group p {
  margin: 8px 0 0 0;
  font-size: 12px;
}

.form-input-group textarea {
  background: $white-color;
  border: 1px solid $grey-border-color;
  border-radius: 2px;
}

.form-input-group textarea:focus {
  outline: none;
}

.form-input-group.error input {
  border: 1px solid $error-color;
}

.form-input-group.error select {
  border: 1px solid $error-color;
}

.form-input-group.error textarea {
  border: 1px solid $error-color;
}

.form-input-group input.error {
  border: 1px solid $error-color;
}

.form-input-group-error-message {
  color: $error-color;
}

.form-input-group-error-description {
  max-width: 464px;
}
</style>

