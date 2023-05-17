<script setup lang="ts">
import { inject } from 'vue';
interface UpdateTemplateVariables {
    registry: any;
    updateBranches: any;
}
const variables = inject('TEMPLATE_VARIABLES') as UpdateTemplateVariables;
const registry = variables?.registry;
const updateBranches = variables?.updateBranches;
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
        }
    }
};
</script>

<template>
    <div class="registry registry-create" id="registry-form">
        <div class="registry-header">
            <a href="/admin/registry/overview" class="registry-add">
                <img alt="add registry" src="@/assets/img/action-back.png" />
                <span>НАЗАД</span>
            </a>
        </div>

        <h1>Оновити реєстр "{{ registry.metadata.name }}"</h1>

        <form id="registry-update-form" class="registry-create-form" method="post" @submit="submit"
            :action="`/admin/registry/update/${registry.metadata.name}`">
            <div class="rc-form-group">
                <label for="branch">Оновити реєстр</label>
                <select id="branch" name="branch" required>
                    <option></option>
                    <option v-for="$val in updateBranches" :key="$val" :value="$val">{{ $val }}</option>
                </select>
            </div>
            <div class="rc-form-group">
                <button type="submit" name="submit" :disabled="disabled">Підтвердити</button>
            </div>
        </form>
    </div>
</template>
