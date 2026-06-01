import { ref } from 'vue';
import type { CarbonChartDataPoint } from '@/services/monthly-count-report.service';
import { monthlyCountReportRepository } from '@/repositories/monthly-count-report.repository';
import { getRecentMonths, buildChartDatasets } from '@/services/monthly-count-report.service';

export function useMonthlyCountReport() {
  const chartData = ref<CarbonChartDataPoint[]>([]);
  const chartOptions = {
    axes: {
      bottom: { title: 'Month', mapsTo: 'date', scaleType: 'labels' },
      left: { title: 'Job Listings', mapsTo: 'value', scaleType: 'linear' },
    },
    height: '400px',
  };
  const error = ref<Error | null>(null);

  async function load() {
    try {
      const records = await monthlyCountReportRepository.getAll();
      const months = getRecentMonths(records);
      chartData.value = buildChartDatasets(records, months);
      error.value = null;
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e));
    }
  }

  return { chartData, chartOptions, error, load };
}
