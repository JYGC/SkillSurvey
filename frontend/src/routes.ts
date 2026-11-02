import type { RouteRecordRaw } from 'vue-router';
import PublicLayout from './layouts/PublicLayout.vue';
import UserLayout from './layouts/UserLayout.vue';
import Login from './views/public/Login.vue';
import RegisterUser from './views/public/RegisterUser.vue';
import HomeRoute from './views/public/HomeRoute.vue';
import MonthlyCountReport from './views/user/MonthlyCountReport.vue';
import Profile from './views/user/Profile.vue';
import Settings from './views/user/Settings.vue';

export const routes: RouteRecordRaw[] = [
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
    {
        path: '/monthly-count-report',
        redirect: { name: 'user-monthly-count-report' }
    },
];
