<script setup lang="ts">
import { inject, onMounted, ref } from 'vue';

import 'diff2html/bundles/css/diff2html.min.css';
import { Diff2HtmlUI, } from 'diff2html/lib/ui/js/diff2html-ui-slim';
import type { Diff2HtmlUIConfig, } from 'diff2html/lib/ui/js/diff2html-ui-slim';

interface ChangeTemplateVariables {
    changes: string;
    change: any;
    changeID: string;
}
const variables = inject('TEMPLATE_VARIABLES') as ChangeTemplateVariables;
const changes = variables?.changes;
const change = variables?.change;
const changeID = variables?.changeID;
const disabled = ref(false);

onMounted(() => {
    const targetElement = document.getElementById('changes') as HTMLElement;
    const configuration: Diff2HtmlUIConfig = {
        drawFileList: false,
        matching: 'lines',
        highlight: true,
        outputFormat: 'side-by-side',
        rawTemplates: { "tag-file-renamed": "" }
    };
    let originalDiffString = JSON.parse(changes);
    const diff2htmlUi = new Diff2HtmlUI(targetElement, originalDiffString, configuration);
    diff2htmlUi.draw();
    diff2htmlUi.highlightCode();
});

function handleClick(url: string) {
    disabled.value = true;
    window.location.href = url;
}

</script>

<template>
    <div id="changes"></div>
    <div class="change-actions" v-if="change.status === 'NEW'">
        <a href="#" class="change-abandon" :class="{ 'disabled': disabled }"
            @click="handleClick(`/admin/abandon-change/${changeID}`)">{{ $t('actions.reject') }}</a>
        <a href="#" class="change-submit" :class="{ 'disabled': disabled }"
            @click="handleClick(`/admin/submit-change/${changeID}`)">{{ $t('actions.confirm') }}</a>
    </div>
</template>

<style scoped>
.disabled {
    pointer-events: none;
}
</style>
