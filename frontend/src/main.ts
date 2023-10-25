import { createApp } from 'vue';
import { createVuetify } from 'vuetify';
import App from './App.vue';
import router from './router';
import i18n from './localization';

import './assets/main.css';
import '@mdi/font/css/materialdesignicons.css';

const vuetify = createVuetify({icons: {defaultSet: 'mdi'}});
const app = createApp(App).use(vuetify);
const templateEl = document.getElementById('template-args');
const envEl = document.getElementById('env-args');
const templateVariables = JSON.parse(templateEl?.getAttribute('data-args') || '{}');
const environmentVariables = JSON.parse(envEl?.getAttribute('data-args') || '{}');

i18n.global.locale  = environmentVariables.language || 'uk';

app.use(router);
app.use(i18n);
app.provide('TEMPLATE_VARIABLES', templateVariables);
app.provide('ENVIRONMENT_VARIABLES', environmentVariables);

app.mount('#app');
