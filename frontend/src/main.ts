import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

import './assets/main.css';

const app = createApp(App);
const templateEl = document.getElementById('template-args');

app.use(router);
app.provide('TEMPLATE_VARIABLES', JSON.parse(templateEl?.getAttribute('data-args') || '{}'));

app.mount('#app');
