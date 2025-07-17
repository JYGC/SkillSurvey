// Declare app
import { createApp } from 'vue';
import App from './App.vue';
const app = createApp(App);

// Declare routes
import { createRouter, createWebHistory } from 'vue-router';
import HomeRoute from '@/routes/HomeRoute.vue';
import MonthlyCountReport from '@/routes/MonthlyCountReport.vue';
import SkillAdd from '@/routes/SkillAdd.vue';
import SkillEdit from '@/routes/SkillEdit.vue';
import SkillList from '@/routes/SkillList.vue';
import SkillTypeAdd from '@/routes/SkillTypeAdd.vue';
import SkillTypeEdit from '@/routes/SkillTypeEdit.vue';
import SkillTypeList from '@/routes/SkillTypeList.vue';
const routes = [
    {
        path: '/',
        name: 'home',
        component: HomeRoute
    },
    {
        path: '/monthly-count-report',
        name: 'monthly-count-report',
        component: MonthlyCountReport
    },
    {
        path: '/skill-add/:skilltypeid?',
        name: 'skill-add',
        component: SkillAdd
    },
    {
        path: '/skill-edit/:skillid',
        name: 'skill-edit',
        component: SkillEdit
    },
    {
        path: '/skill-list',
        name: 'skill-list',
        component: SkillList
    },
    {
        path: '/skill-type-add',
        name: 'skill-type-add',
        component: SkillTypeAdd
    },
    {
        path: '/skill-type-edit/:skilltypeid',
        name: 'skill-type-edit',
        component: SkillTypeEdit
    },
    {
        path: '/skill-type-list/',
        name: 'skill-type-list',
        component: SkillTypeList
    },
];
const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes
});
app.use(router);

// Add bootstrap
import BootstrapVue3 from 'bootstrap-vue-3';
import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue-3/dist/bootstrap-vue-3.css'
app.use(BootstrapVue3);

//app.use(store);
app.mount('#app');
