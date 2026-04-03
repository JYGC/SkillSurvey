// Declare app
import { createApp } from 'vue';
import App from './App.vue';
const app = createApp(App);

// Declare routes
import { createRouter, createWebHistory } from 'vue-router';
import PublicLayout from './layouts/PublicLayout.vue';
import UserLayout from './layouts/UserLayout.vue';
import Login from './views/public/Login.vue';
import RegisterUser from './views/public/RegisterUser.vue';
import HomeRoute from './views/public/HomeRoute.vue';
import MonthlyCountReport from './views/public/MonthlyCountReport.vue';
import SkillAdd from './views/public/SkillAdd.vue';
import SkillEdit from './views/public/SkillEdit.vue';
import SkillList from './views/public/SkillList.vue';
import SkillTypeAdd from './views/public/SkillTypeAdd.vue';
import SkillTypeEdit from './views/public/SkillTypeEdit.vue';
import SkillTypeList from './views/public/SkillTypeList.vue';
import Profile from './views/user/Profile.vue';
import Settings from './views/user/Settings.vue';
const routes = [
    {
        path: '/',
        component: PublicLayout,
        children: [
            {
                path: '',
                name: 'home',
                component: HomeRoute
            },
            {
                path: 'login',
                name: 'login',
                component: Login
            },
            {
                path: 'register',
                name: 'register-user',
                component: RegisterUser
            },
            {
                path: 'monthly-count-report',
                name: 'monthly-count-report',
                component: MonthlyCountReport
            },
            {
                path: 'skill-add/:skilltypeid?',
                name: 'skill-add',
                component: SkillAdd
            },
            {
                path: 'skill-edit/:skillid',
                name: 'skill-edit',
                component: SkillEdit
            },
            {
                path: 'skill-list',
                name: 'skill-list',
                component: SkillList
            },
            {
                path: 'skill-type-add',
                name: 'skill-type-add',
                component: SkillTypeAdd
            },
            {
                path: 'skill-type-edit/:skilltypeid',
                name: 'skill-type-edit',
                component: SkillTypeEdit
            },
            {
                path: 'skill-type-list/',
                name: 'skill-type-list',
                component: SkillTypeList
            }
        ]
    },
    {
        path: '/user',
        component: UserLayout,
        children: [
            {
                path: 'profile',
                name: 'user-profile',
                component: Profile
            },
            {
                path: 'settings',
                name: 'user-settings',
                component: Settings
            }
        ]
    },
];
const router = createRouter({
    history: createWebHistory(process.env.BASE_URL),
    routes
});
app.use(router);

// Add Carbon
import CarbonVue3 from '@carbon/vue';
app.use(CarbonVue3);

// Add bootstrap
import BootstrapVue3 from 'bootstrap-vue-3';
import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap-vue-3/dist/bootstrap-vue-3.css';
app.use(BootstrapVue3);

//app.use(store);
app.mount('#app');
