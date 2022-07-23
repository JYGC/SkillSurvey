// Declare routes
import { createRouter, createWebHistory } from 'vue-router';
import HelloWorld from '@/routes/HelloWorld.vue';
import MonthlyCountReport from '@/routes/MonthlyCountReport.vue';
import SkillAdd from '@/routes/SkillAdd.vue';
const routes = [
    {
        path: '/',
        name: 'home',
        component: HelloWorld
    },
    {
        path: '/monthly-count-report',
        name: 'monthly-count-report',
        component: MonthlyCountReport
    },
    {
        path: '/skill-add',
        name: 'skill-add',
        component: SkillAdd
    },
]
const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes
});

// Declare state management
// import { createStore } from 'vuex';
// const store = createStore({
//     state: {},
//     mutations: {},
//     actions: {}
// });

// Declare app
import { createApp } from 'vue';
import App from './App.vue';
const app = createApp(App);
app.use(router);
//app.use(store);
app.mount('#app');
