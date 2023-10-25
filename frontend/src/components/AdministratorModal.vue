<script setup lang="ts">
import { toRefs } from 'vue';

interface AdministratorModalProps {
    adminPopupShow: any;
    editAdmin: any;
    requiredError: any;
    emailFormatError: any;
    passwordFormatError: boolean,
    usernameFormatError: any;
    adminExistsError: any;
}

const props = defineProps<AdministratorModalProps>();
const { adminPopupShow, editAdmin, requiredError, emailFormatError, passwordFormatError, usernameFormatError, adminExistsError } = toRefs(props);

</script>
<script lang="ts">
export default {
    data() {
        return {
            disabled: false,
        };
    },
    methods: {
        submit() {
            this.disabled = true;
            this.$emit('createAdmin');
        },
        hideAdminForm() {
            this.$emit('hideAdminForm');
        },
    },
    computed: {
        hasError() {
            return this.requiredError || this.emailFormatError || this.passwordFormatError || this.usernameFormatError || this.adminExistsError;
        }
    },
    watch: {
        adminPopupShow() {
            this.disabled = false;
        }
    }
};
</script>

<template>
    <div class="popup-backdrop visible" v-cloak v-if="adminPopupShow"></div>

    <div id="admin-popup admin-window" class="popup-window admin-window visible" v-cloak v-if="adminPopupShow">
        <div class="popup-header">
            <p>{{ $t('components.administratorModal.text.administrator') }}</p>
            <a href="#" @click.stop.prevent="hideAdminForm" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <form id="admin-form" method="post" action="" @submit.prevent="submit">
            <div class="popup-body">
                <p class="popup-error" v-cloak v-if="requiredError">{{ $t('components.administratorModal.errors.requiredAllFields') }}</p>
                <p class="popup-error" v-cloak v-if="emailFormatError">{{ $t('components.administratorModal.errors.invalidEmailFormat') }}</p>
                <p class="popup-error" v-cloak v-if="usernameFormatError">{{ $t('components.administratorModal.errors.invalidNameFormat') }}</p>
                <p class="popup-error" v-cloak v-if="adminExistsError">{{ $t('components.administratorModal.errors.adminEmailExists') }}</p>

                <div class="rc-form-group">
                    <input aria-label="admin first name" type="text" :placeholder="$t('components.administratorModal.fields.firstName.placeholder')" v-model="editAdmin.firstName" />
                </div>
                <div class="rc-form-group">
                    <input aria-label="admin last name" type="fields" :placeholder="$t('components.administratorModal.fields.surname.placeholder')" v-model="editAdmin.lastName" />
                </div>
                <div class="rc-form-group">
                    <input
                    aria-label="email"
                    type="email"
                    :placeholder="$t('components.administratorModal.fields.email.placeholder')" v-model="editAdmin.email"
                    />
                </div>
                <div class="rc-form-group"
                    :class="{ error: !!passwordFormatError }">
                  <input aria-label="password" type="password" v-model="editAdmin.tmpPassword"
                         :placeholder="$t('components.administratorModal.fields.temporaryPassword.placeholder')" />
                  <span v-if="!!passwordFormatError">
                         {{ $t("components.administratorModal.errors.invalidPasswordFormat") }}
                  </span>
                </div>
                <div>
                    {{ $t('components.administratorModal.fields.temporaryPassword.requirements.header') }}
                    <ul class="password-requirements">
                        <li>{{ $t('components.administratorModal.fields.temporaryPassword.requirements.point1') }}</li>
                        <li>{{ $t('components.administratorModal.fields.temporaryPassword.requirements.point2') }}</li>
                        <li>{{ $t('components.administratorModal.fields.temporaryPassword.requirements.point3') }}</li>
                        <li>{{ $t('components.administratorModal.fields.temporaryPassword.requirements.point4') }}</li>
                        <li v-html="$t('components.administratorModal.fields.temporaryPassword.requirements.point5')"></li>
                        <li>{{ $t('components.administratorModal.fields.temporaryPassword.requirements.point6') }}</li>
                        <li>{{ $t('components.administratorModal.fields.temporaryPassword.requirements.point7') }}</li>
                    </ul>
                </div>
            </div>
            <div class="popup-footer active">
                <a href="#" id="admin-cancel" class="hide-popup" @click.stop.prevent="hideAdminForm">
                    {{ $t('actions.cancel') }}
                </a>
                <button value="submit" name="admin-apply" type="submit"
                    :disabled="disabled && !hasError">{{ $t('actions.confirm') }}</button>
            </div>
        </form>
    </div>
</template>

<style>
.password-requirements {
  list-style-type: none;
  padding-left: 0.75em;
}
</style>