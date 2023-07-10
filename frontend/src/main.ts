import { createApp } from 'vue';
import { createVuetify } from 'vuetify';
import App from './App.vue';
import router from './router';

import './assets/main.css';

const vuetify = createVuetify();
const app = createApp(App).use(vuetify);
const templateEl = document.getElementById('template-args');

app.use(router);
app.provide('TEMPLATE_VARIABLES', JSON.parse(templateEl?.getAttribute('data-args') || '{}'));

app.mount('#app');
