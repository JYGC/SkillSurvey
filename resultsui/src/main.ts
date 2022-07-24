// Declare routes
import { createRouter, createWebHistory } from 'vue-router';
import HelloWorld from '@/routes/HelloWorld.vue';
import MonthlyCountReport from '@/routes/MonthlyCountReport.vue';
import SkillAdd from '@/routes/SkillAdd.vue';
import SkillEdit from '@/routes/SkillEdit.vue';
import SkillsList from '@/routes/SkillsList.vue';
import SkillTypeAdd from '@/routes/SkillTypeAdd.vue';
import SkillTypeEdit from '@/routes/SkillTypeEdit.vue';
import SkillTypesList from '@/routes/SkillTypesList.vue';
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
        path: '/skill-add/',
        name: 'skill-add',
        component: SkillAdd
    },
    {
        path: '/skill-edit/:id',
        name: 'skill-edit',
        component: SkillEdit
    },
    {
        path: '/skills-list',
        name: 'skills-list',
        component: SkillsList
    },
    {
        path: '/skill-type-add',
        name: 'skill-type-add',
        component: SkillTypeAdd
    },
    {
        path: '/skill-type-edit/:id',
        name: 'skill-type-edit',
        component: SkillTypeEdit
    },
    {
        path: '/skill-types-list/',
        name: 'skill-types-list',
        component: SkillTypesList
    },
];
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
