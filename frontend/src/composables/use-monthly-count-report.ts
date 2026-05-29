import { ref } from 'vue';
import { Chart, registerables } from 'chart.js';
import type { ChartData } from 'chart.js';
import { monthlyCountReportRepository } from '@/repositories/monthly-count-report.repository';
import { getRecentMonths, buildChartDatasets } from '@/services/monthly-count-report.service';

Chart.register(...registerables);

export function useMonthlyCountReport() {
  const chartData = ref<ChartData<'line'>>({ labels: [], datasets: [] });
  const chartHeight = ref(400);
  const error = ref<Error | null>(null);

  async function load() {
    try {
      const records = await monthlyCountReportRepository.getAll();
      const months = getRecentMonths(records);
      chartData.value = {
        labels: months,
        datasets: buildChartDatasets(records, months),
      };
      error.value = null;
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e));
    }
  }

  return { chartData, chartHeight, error, load };
}
