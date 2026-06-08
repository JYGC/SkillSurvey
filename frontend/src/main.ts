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
import MonthlyCountReport from './views/user/MonthlyCountReport.vue';
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
        ]
    },
    {
        path: '/user',
        component: UserLayout,
        children: [
            {
                path: 'monthly-count-report',
                name: 'user-monthly-count-report',
                component: MonthlyCountReport
            },
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

//app.use(store);
app.mount('#app');
