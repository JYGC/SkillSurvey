import { describe, it, expect } from 'vitest';
import { getRecentMonths, buildChartDatasets } from '@/services/monthly-count-report.service';
import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';

const rec = (ym: string, skill: string, count: number): MonthlyCountRecord => ({
  YearMonth: ym,
  yearMonthDate: `${ym}-01`,
  count,
  expand: { skillName: { name: skill } },
});

describe('getRecentMonths', () => {
  it('returns [] for empty input', () => {
    expect(getRecentMonths([])).toEqual([]);
  });

  it('returns all months when fewer than 12 unique months', () => {
    const records = [
      rec('2024-10', 'TypeScript', 5),
      rec('2024-11', 'TypeScript', 8),
      rec('2024-12', 'TypeScript', 3),
    ];
    expect(getRecentMonths(records)).toEqual(['2024-10', '2024-11', '2024-12']);
  });

  it('returns last 12 months when more than 12 unique months', () => {
    const months = [
      '2023-01', '2023-02', '2023-03', '2023-04', '2023-05', '2023-06',
      '2023-07', '2023-08', '2023-09', '2023-10', '2023-11', '2023-12',
      '2024-01', '2024-02',
    ];
    const records = months.map(ym => rec(ym, 'TypeScript', 1));
    const result = getRecentMonths(records);
    expect(result).toHaveLength(12);
    expect(result[0]).toBe('2023-03');
    expect(result[11]).toBe('2024-02');
  });

  it('deduplicates months across skills', () => {
    const records = [
      rec('2024-10', 'TypeScript', 5),
      rec('2024-10', 'Python', 3),
      rec('2024-11', 'TypeScript', 8),
    ];
    expect(getRecentMonths(records)).toEqual(['2024-10', '2024-11']);
  });
});

describe('buildChartDatasets', () => {
  it('produces one dataset per skill', () => {
    const months = ['2024-10', '2024-11'];
    const records = [
      rec('2024-10', 'TypeScript', 5),
      rec('2024-11', 'TypeScript', 8),
      rec('2024-10', 'Python', 3),
    ];
    const datasets = buildChartDatasets(records, months);
    expect(datasets).toHaveLength(2);
    expect(datasets.map(d => d.label)).toContain('TypeScript');
    expect(datasets.map(d => d.label)).toContain('Python');
  });

  it('fills missing months with 0', () => {
    const months = ['2024-10', '2024-11', '2024-12'];
    const records = [
      rec('2024-10', 'TypeScript', 5),
      rec('2024-12', 'TypeScript', 7),
    ];
    const datasets = buildChartDatasets(records, months);
    expect(datasets).toHaveLength(1);
    expect(datasets[0].data).toEqual([5, 0, 7]);
  });

  it('sets hidden: true on every dataset', () => {
    const datasets = buildChartDatasets(
      [rec('2024-10', 'TypeScript', 5), rec('2024-10', 'Python', 2)],
      ['2024-10'],
    );
    expect(datasets.every(d => d.hidden === true)).toBe(true);
  });

  it('labels missing expand as "Unknown"', () => {
    const records: MonthlyCountRecord[] = [
      { YearMonth: '2024-10', yearMonthDate: '2024-10-01', count: 3 },
    ];
    const datasets = buildChartDatasets(records, ['2024-10']);
    expect(datasets[0].label).toBe('Unknown');
  });
});
