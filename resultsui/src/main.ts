import { Chart, registerables } from "chart.js";
Chart.register(...registerables);

import { createApp } from 'vue'
import App from './App.vue'
createApp(App).mount('#app')
