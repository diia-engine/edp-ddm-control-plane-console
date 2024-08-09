<script setup lang="ts">
import { toRefs } from 'vue';

interface CidrProps {
    adminCIDRValue: any;
    adminCIDR: any;
}

const props = defineProps<CidrProps>();
const { adminCIDRValue, adminCIDR } = toRefs(props);

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
            this.$emit('registryFormSubmit');
        },
        deleteCIDR(c: any, adminCIDR: any, adminCIDRValue: any) {
            this.$emit('deleteCidr', c, adminCIDR, adminCIDRValue);
        },
        showCIDRForm(adminCIDR: any, adminCIDRValue: any) {
            this.$emit('showCidrForm', adminCIDR, adminCIDRValue);
        }
    }
};
</script>

<template>
    <h2>Перелік дозволених CIDR</h2>
    <form id="cluster-update-form" @submit="submit" class="registry-create-form wizard-form" method="post"
        action="/admin/cluster/cidr">
        <div class="rc-form-group">
            <label for="admins">CIDR для адміністративних компонент</label>
            <input type="hidden" id="admin-cidr" name="admin-cidr" v-model="adminCIDRValue.value" />
            <div class="advanced-admins">
                <div v-cloak v-for="c in adminCIDR" :key="c" class="child-admin">
                    {{ c }}
                    <a @click.stop.prevent="deleteCIDR(c, adminCIDR, adminCIDRValue)" href="#">
                        <img src="@/assets/img/action-delete.png" />
                    </a>
                </div>
                <button type="button" @click="showCIDRForm(adminCIDR, adminCIDRValue)">+</button>
            </div>
        </div>

        <div class="rc-form-group">
            <button type="submit" name="submit" :disabled="disabled" onclick="window.localStorage.setItem('mr-scroll', 'true');">Підтвердити</button>
        </div>
    </form>
</template>
