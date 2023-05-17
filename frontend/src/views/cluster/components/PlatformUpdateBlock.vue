<script setup lang="ts">
import { getErrorMessage } from '@/utils';
import { toRefs } from 'vue';

interface PlatformUpdateProps {
    updateBranches: any;
    errorsMap: any;
}

const props = defineProps<PlatformUpdateProps>();
const { updateBranches, errorsMap } = toRefs(props);

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
    }
};
</script>

<template>
    <h2>Оновлення платформи</h2>
    <form method="post" action="/admin/cluster/upgrade" @submit="submit" class="registry-create-form">
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
