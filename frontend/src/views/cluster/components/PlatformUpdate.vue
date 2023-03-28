<script setup lang="ts">
import { inject } from 'vue';
interface PlatformUpdateTemplateVariables {
    updateBranches: any;
    errorsMap: any;
}
const variables = inject('TEMPLATE_VARIABLES') as PlatformUpdateTemplateVariables;
const updateBranches = variables?.updateBranches;
const errorsMap = variables?.errorsMap;
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
        getErrorMessage(error: string) {
            switch (error) {
                case 'required':
                    return "Не може бути порожнім";
                default:
                    return "";
            }
        }
    }
};
</script>

<template>
    <h2>Оновлення платформи</h2>
    <form class="registry-create-form" method="post" action="/admin/cluster/upgrade" @submit="submit">
        <div class="rc-form-group" :class="{ error: errorsMap?.branch }">
            <label for="branch">Оновити платформу</label>
            <select id="branch" name="branch" required>
                <option></option>
                <option v-for="$val in updateBranches" :key="$val" :value="$val">{{ $val }}</option>
            </select>
            <span v-for="$val in errorsMap?.branch" :key="$val.tag">
                {{ getErrorMessage($val.tag) }}
            </span>
        </div>
        <div class="rc-form-group">
            <button type="submit" name="submit" :disabled="disabled">Підтвердити</button>
        </div>
    </form>
</template>
