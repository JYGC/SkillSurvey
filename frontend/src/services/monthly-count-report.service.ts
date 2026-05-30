import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';

export interface CarbonChartDataPoint {
  group: string;
  date: string;
  value: number;
}

export function getRecentMonths(records: MonthlyCountRecord[]): string[] {
  return [...new Set(records.map(r => r.YearMonth))].slice(-12);
}

export function buildChartDatasets(
  records: MonthlyCountRecord[],
  months: string[],
): CarbonChartDataPoint[] {
  const bySkill: Record<string, Record<string, number>> = {};
  for (const r of records) {
    const name = r.expand?.skillName?.name ?? 'Unknown';
    if (!bySkill[name]) bySkill[name] = {};
    bySkill[name][r.YearMonth] = r.count;
  }
  return Object.entries(bySkill).flatMap(([group, counts]) =>
    months.map(month => ({
      group,
      date: month,
      value: counts[month] ?? 0,
    }))
  );
}
