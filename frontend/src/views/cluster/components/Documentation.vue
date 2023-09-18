<script setup lang="ts">
import { onMounted, toRefs, ref } from 'vue';
import axios from 'axios';
import Typography from '@/components/common/Typography.vue';
import Banner from '@/components/common/Banner.vue';

interface DocumentationProps {
   demoRegistryName: string;
}
const props = defineProps<DocumentationProps>();
const { demoRegistryName } = toRefs(props);
const registry = ref(demoRegistryName.value || "");
const registries = ref([] as string[]);
const prevRegistryIsInvalid = ref(false);

onMounted(()=> {
    axios.get('/admin/registries')
        .then((response) => {
            registries.value = 
                response.data?.map((registry: { Codebase: { metadata: { name: any; }; }; }) => registry.Codebase.metadata.name);

            if (demoRegistryName.value && !registries.value.includes(demoRegistryName.value)) {
                registry.value = "";
                prevRegistryIsInvalid.value = true;
            }
        });
});

const registryChangeHandler = () => {
    prevRegistryIsInvalid.value = false;
};

</script>

<template>
    <h2>Документація</h2>
    <div class="documentation-description">
        <Typography variant="bodyText">На обраний демо-реєстр вестимуть посилання з документації.</Typography>
    </div>
    <form id="platform-admin" class="registry-create-form wizard-form" method="post" action="/admin/cluster/demo-registry-name">
        <div class="rc-form-group">
            <label for="demo-registry-name">Демо-реєстр</label>
            <select v-model="registry" id="demo-registry-name" name="demo-registry-name" :onChange="registryChangeHandler">
                <option value="">Не обрано</option>
                <option v-for="item in registries" :key="item" :value="item">
                    {{ item }}
                </option>
            </select>

            <Banner
                classes="banner"
                :description="`Попередній реєстр («${demoRegistryName}») не знайдено.`"
                v-if="prevRegistryIsInvalid"
            />
        </div>
        <div class="rc-form-group">
            <button type="submit" name="submit">Підтвердити</button>
        </div>
    </form>
</template>

<style lang="scss" scoped>
    .documentation-description {
      margin-bottom: 24px;
    }

    .banner {
        margin-top: 16px;
    }

</style>