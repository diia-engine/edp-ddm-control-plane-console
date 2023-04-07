<script setup lang="ts">
import { toRefs } from 'vue';

interface ClusterKeyBlockProps {
    wizard: any;
}

const props = defineProps<ClusterKeyBlockProps>();
const { wizard } = toRefs(props);

</script>
<script lang="ts">
import KeyForm from '@/components/KeyForm.vue';
export default {
    components: {
        KeyForm,
    },
    data() {
        return {
            disabled: false,
            pageRoot: this.$parent as any,
        };
    },
    methods: {
        submit(event: any) {
            this.disabled = true;
            if (!this.pageRoot.keyFormValidation(this.pageRoot.$data.wizard.tabs.key, function () {

            })) {
                this.disabled = false;
                event.preventDefault();
            }
        },
        wizardTabChanged(key: string) {
            this.$emit('wizardTabChanged', key);
        },
        wizardKeyHardwareDataChanged() {
            this.$emit('wizardKeyHardwareDataChanged');
        },
        wizardAddAllowedKey() {
            this.$emit('wizardAddAllowedKey');
        },
        wizardRemoveAllowedKey(key: string) {
            this.$emit('wizardAddAllowedKey', key);
        }
    }
};
</script>

<template>
    <form @submit="submit" class="registry-create-form wizard-form" enctype="multipart/form-data" method="post"
        action="/admin/cluster/key">

        <key-form :wizard="wizard" @wizard-tab-changed="wizardTabChanged"
            @wizard-key-hardware-data-changed="wizardKeyHardwareDataChanged" @wizard-add-allowed-key="wizardAddAllowedKey"
            @wizard-remove-allowed-key="wizardRemoveAllowedKey" ref="keyFormRef" />

        <div class="rc-form-group">
            <button type="submit" name="submit" :disabled="disabled">Підтвердити</button>
        </div>
    </form>
</template>
