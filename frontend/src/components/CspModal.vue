<script setup lang="ts">
import { ref, watch, defineProps, defineEmits } from 'vue';
import TextField from './common/TextField.vue';

type CspModalProps = {
  show: boolean;
  title: string;
};

const props = defineProps<CspModalProps>();
const disabled = ref(false);
const editCsp = ref('');
const cspFormatError = ref(false);

const $emit = defineEmits(['cspAdded', 'update:show']);

function createCsp() {
    disabled.value = true;
    let cspVal = String(editCsp.value).toLowerCase();
    if (!isValidSource(cspVal)) {
        cspFormatError.value = true;
        return;
    }

    $emit('cspAdded', editCsp.value);
    hideCspForm();
}

function isValidSource(val: string) {
    return !!val && typeof val === 'string' && !val.includes('"');
}

function hideCspForm() {
    $emit('update:show', false);
    document.body.style.overflow = "scroll";
}

watch(() => props.show, () => {
    cspFormatError.value = false;
    disabled.value = false;
    editCsp.value = '';
});

</script>

<template>
    <div class="popup-backdrop visible" v-cloak v-if="props.show"></div>
    <div class="popup-window admin-window visible" v-cloak v-if="props.show">
        <div class="popup-header">
            <p>{{ props.title }}</p>
            <a href="#" @click.stop.prevent="hideCspForm" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <form @submit.prevent="createCsp()" id="csp-form" method="post" action="">
            <div class="popup-body">
                <TextField
                    v-model="editCsp"
                    :label="$t('components.registryCsp.modal.fields.source.label')"
                    :description="$t('components.registryCsp.modal.fields.source.description')"
                    required
                    :error="cspFormatError ? 'checkFormat' : undefined"
                />
            </div>
            <div class="popup-footer active">
                <a href="#" id="csp-cancel" class="hide-popup" @click="hideCspForm">{{ $t('actions.cancel') }}</a>
                <button class="submit-green" value="submit" name="csp-apply" type="submit"
                    :disabled="disabled && !cspFormatError">{{ $t('actions.confirm') }}</button>
            </div>
        </form>
    </div>
</template>
