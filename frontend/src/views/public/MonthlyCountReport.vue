<template>
  <div>
    <p v-if="error" data-testid="report-error">{{ error.message }}</p>
    <LineChart v-else :chartData="chartData" :height="chartHeight" />
  </div>
</template>
<script lang="ts" setup>
import { onMounted } from 'vue';
import { LineChart } from 'vue-chart-3';
import { useMonthlyCountReport } from '@/composables/use-monthly-count-report';

const { chartData, chartHeight, error, load } = useMonthlyCountReport();

onMounted(load);
</script>
