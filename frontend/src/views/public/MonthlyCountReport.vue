<template>
  <div>
    <LineChart :chartData="getData" :height="chartHeight" />
  </div>
</template>
<script lang="ts" setup>
import { Chart, registerables, type ChartDataset } from "chart.js";
import { reactive } from "vue";
import { LineChart } from "vue-chart-3";
import { getBackendClient } from "../../services/backend-client";

interface MonthlyCountRecord {
  YearMonth: string;
  yearMonthDate: string;
  count: number;
  expand?: { skillName?: { name: string } };
}

// make chart fill window height-wise
const chartHeight = window.innerHeight;
// make chart
Chart.register(...registerables);
let getData = reactive({
  labels: new Array<string>(),
  datasets: new Array<ChartDataset<"line">>(),
});

(async function () {
  try {
    const allRecords = await getBackendClient()
      .collection('monthlyCountReports')
      .getFullList<MonthlyCountRecord>({
        expand: 'skillName',
        sort: 'yearMonthDate',
      });

    // Filter to the most recent 12 distinct YearMonth values.
    const recentMonths = [...new Set(allRecords.map((r) => r.YearMonth))].slice(-12);
    const filtered = allRecords.filter((r) => recentMonths.includes(r.YearMonth));

    getData.labels = recentMonths;
    getData.datasets = createDatasets(filtered, recentMonths);
  } catch (error) {
    alert(error);
    console.log(error);
  }
})();

function createDatasets(records: MonthlyCountRecord[], months: string[]): ChartDataset<"line">[] {
  // Group records by skill name.
  const bySkill: Record<string, Record<string, number>> = {};
  for (const record of records) {
    const skillName = record.expand?.skillName?.name ?? 'Unknown';
    if (!bySkill[skillName]) {
      bySkill[skillName] = {};
    }
    bySkill[skillName][record.YearMonth] = record.count;
  }

  return Object.entries(bySkill).map(([skillName, monthCounts]) => ({
    label: skillName,
    data: months.map((m) => monthCounts[m] ?? 0),
    fill: false,
    borderColor: `#${Math.floor(Math.random() * 16777215).toString(16).padStart(6, '0')}`,
    hidden: true,
  } as ChartDataset<"line">));
}
</script>
