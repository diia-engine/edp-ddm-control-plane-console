<script lang="ts">
import { defineComponent } from 'vue';
import KeyData from '@/components/RegistryWizard/steps/KeyData.vue';
import KeyVerification from '@/components/RegistryWizard/steps/KeyVerification.vue';

export default defineComponent({
    props: {
      activeTab: String
    },
    components: {
      KeyData,
      KeyVerification,
    },
    data() {
        return {
            disabled: false,
        };
    },
    methods: {
        submit(event: any) {
            event.preventDefault();
            this.disabled = true;

            let validatorTab = this.$refs.keyDataTab as any;
            if (this.activeTab === 'dataAboutKeyVerification') {
              validatorTab = this.$refs.keyVerificationTab as any;
            }

            const clusterKeyDataForm = this.$refs.clusterKeyDataForm as any;
            validatorTab.validator().then(() => {
              this.$nextTick(() => {
                HTMLFormElement.prototype.submit.call(clusterKeyDataForm);
              });
            }).catch(() => {
              this.disabled = false;
            });
        },
    }
});
</script>

<style scoped>
  .rc-form-group button {
    margin-top: 20px;
  }
</style>

<template>
    <form @submit="submit" ref="clusterKeyDataForm" id="clusterKeyDataForm" class="registry-create-form wizard-form"
          enctype="multipart/form-data" method="post"
        action="/admin/cluster/key">

      <div v-if="activeTab === 'dataAboutKey'">
        <KeyData
            registry-action="create"
            ref="keyDataTab" />
      </div>

      <div v-if="activeTab === 'dataAboutKeyVerification'">
        <KeyVerification
            registry-action="create"
            ref="keyVerificationTab" />
      </div>

        <div class="rc-form-group">
            <button type="submit" name="submit" :disabled="disabled">Підтвердити</button>
        </div>
    </form>
</template>
