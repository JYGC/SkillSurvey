import { createApp } from 'vue';
import App from './App.vue';
const app = createApp(App);

import { createRouter, createWebHistory } from 'vue-router';
import { routes } from './routes';
const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes
});
app.use(router);

import CarbonVue3 from '@carbon/vue';
app.use(CarbonVue3);

app.mount('#app');
