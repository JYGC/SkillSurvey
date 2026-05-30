<template>
  <div>
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
