<script setup lang="ts">
import { toRefs } from 'vue';

interface AdministratorModalProps {
    adminPopupShow: any;
    editAdmin: any;
    requiredError: any;
    emailFormatError: any;
    usernameFormatError: any;
    adminExistsError: any;
}

const props = defineProps<AdministratorModalProps>();
const { adminPopupShow, editAdmin, requiredError, emailFormatError, usernameFormatError, adminExistsError } = toRefs(props);

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
            return this.requiredError || this.emailFormatError || this.usernameFormatError || this.adminExistsError;
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
            <p>Адміністратор</p>
            <a href="#" @click.stop.prevent="hideAdminForm" class="popup-close hide-popup">
                <img alt="close popup window" src="@/assets/img/close.png" />
            </a>
        </div>
        <form id="admin-form" method="post" action="" @submit.prevent="submit">
            <div class="popup-body">
                <p class="popup-error" v-cloak v-if="requiredError">Всі поля обов'язкові для заповнення</p>
                <p class="popup-error" v-cloak v-if="emailFormatError">Невірний формат електронной пошти</p>
                <p class="popup-error" v-cloak v-if="usernameFormatError">Невірний формат ім'я користувача</p>
                <p class="popup-error" v-cloak v-if="adminExistsError">Адміністратор з таким email вже існує</p>

                <div class="rc-form-group">
                    <input aria-label="admin first name" type="text" placeholder="Ім'я" v-model="editAdmin.firstName" />
                </div>
                <div class="rc-form-group">
                    <input aria-label="admin last name" type="text" placeholder="Прізвище" v-model="editAdmin.lastName" />
                </div>
                <div class="rc-form-group">
                    <input aria-label="email" type="email" placeholder="Електронна пошта" v-model="editAdmin.email" />
                </div>
                <div class="rc-form-group">
                    <input aria-label="password" type="password" v-model="editAdmin.tmpPassword"
                        placeholder="Тимчасовий пароль" />
                </div>
            </div>
            <div class="popup-footer active">
                <a href="#" id="admin-cancel" class="hide-popup" @click.stop.prevent="hideAdminForm">відмінити</a>
                <button value="submit" name="admin-apply" type="submit"
                    :disabled="disabled && !hasError">Підтвердити</button>
            </div>
        </form>
    </div>
</template>
