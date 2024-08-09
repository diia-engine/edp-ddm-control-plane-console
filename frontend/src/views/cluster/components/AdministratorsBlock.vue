<script setup lang="ts">
import { toRefs } from 'vue';

interface AdministratorsProps {
    admins: any;
    adminsValue: any;
}

const props = defineProps<AdministratorsProps>();
const { admins, adminsValue } = toRefs(props);

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
        },
        deleteAdmin(adminEmail: string) {
            this.$emit('deleteAdmin', adminEmail);
        },
        showAdminForm() {
            this.$emit('showAdminForm');
        },
    }
};
</script>

<template>
    <h2>Адміністратори платформи</h2>
    <form id="platform-admin" class="registry-create-form wizard-form" method="post" action="/admin/cluster/admins"
        @submit="submit">
        <div class="rc-form-group">
            <label for="admins">Адміністратори</label>
            <input type="hidden" id="admins" name="admins" v-model="adminsValue" />
            <div class="advanced-admins">
                <div v-cloak v-for="adm in admins" :key="adm.email" class="child-admin">
                    {{ adm.email }}
                    <a @click.stop.prevent="deleteAdmin(adm.email)" :email="adm.email" href="#">
                        <img src="@/assets/img/action-delete.png" />
                    </a>
                </div>
                <button type="button" @click="showAdminForm">+</button>
            </div>
        </div>
        <div class="rc-form-group">
            <button onclick="window.localStorage.setItem('mr-scroll', 'true');" type="submit" name="submit" :disabled="disabled">Підтвердити</button>
        </div>
    </form>
</template>
