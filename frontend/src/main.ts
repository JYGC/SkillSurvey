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
