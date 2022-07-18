import { Chart, registerables } from "chart.js";
Chart.register(...registerables);

// Declare routes
// import { createRouter, createWebHistory } from 'vue-router';
// import HelloWorld from '@/components/HelloWorld.vue';
// import MonthlyCountReport from '@/components/MonthlyCountReport.vue';
// const routes = [
//     {
//         path: '/',
//         name: 'home',
//         component: HelloWorld
//     },
//     {
//         path: '/monthly-count-report',
//         name: 'monthly-count-report',
//         component: MonthlyCountReport
//     },
// ]
// const router = createRouter({
//     history: createWebHistory(process.env.BASE_URL),
//     routes
// });

// Declare app
import { createApp } from 'vue';
import App from './App.vue';
createApp(App).mount('#app');
//app.use(router);
