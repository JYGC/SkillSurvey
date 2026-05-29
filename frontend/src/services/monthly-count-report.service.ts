import type { ChartDataset } from 'chart.js';
import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';

export function getRecentMonths(records: MonthlyCountRecord[]): string[] {
  return [...new Set(records.map(r => r.YearMonth))].slice(-12);
}

export function buildChartDatasets(
  records: MonthlyCountRecord[],
  months: string[],
): ChartDataset<'line'>[] {
  const bySkill: Record<string, Record<string, number>> = {};
  for (const r of records) {
    const name = r.expand?.skillName?.name ?? 'Unknown';
    if (!bySkill[name]) bySkill[name] = {};
    bySkill[name][r.YearMonth] = r.count;
  }
  return Object.entries(bySkill).map(([label, counts]) => ({
    label,
    data: months.map(m => counts[m] ?? 0),
    fill: false,
    borderColor: `#${Math.floor(Math.random() * 16777215).toString(16).padStart(6, '0')}`,
    hidden: true,
  } as ChartDataset<'line'>));
}
