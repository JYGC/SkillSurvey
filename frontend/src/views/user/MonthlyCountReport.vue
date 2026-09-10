<template>
  <div class="chart-wrapper">
    <p v-if="error" data-testid="report-error">{{ error.message }}</p>
    <CcvLineChart v-else :data="chartData" :options="chartOptions" />
  </div>
</template>
<script lang="ts" setup>
import { onMounted } from 'vue';
import { CcvLineChart } from '@carbon/charts-vue';
import '@carbon/charts-vue/styles.css';
import { useMonthlyCountReport } from '@/composables/use-monthly-count-report';

const { chartData, chartOptions, error, load } = useMonthlyCountReport();
onMounted(load);
</script>
<style scoped>
.chart-wrapper {
  height: 100%;
}

.chart-wrapper :deep(.cds--chart-holder) {
  height: 100%;
}
</style>
