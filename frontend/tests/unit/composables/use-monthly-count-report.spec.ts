import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';
import type { ChartDataset } from 'chart.js';

const mockRepository = vi.hoisted(() => ({
  getAll: vi.fn(),
}));

const mockService = vi.hoisted(() => ({
  getRecentMonths: vi.fn(),
  buildChartDatasets: vi.fn(),
}));

vi.mock('@/repositories/monthly-count-report.repository', () => ({
  monthlyCountReportRepository: mockRepository,
}));

vi.mock('@/services/monthly-count-report.service', () => mockService);

vi.mock('chart.js', () => ({
  Chart: { register: vi.fn() },
  registerables: [],
}));

const seedRecords: MonthlyCountRecord[] = [
  { id: '1', YearMonth: '2024-01', yearMonthDate: '2024-01-01', count: 5, skillName: 's1', expand: { skillName: { name: 'TypeScript' } } },
];
const seedMonths = ['2024-01'];
const seedDatasets: ChartDataset<'line'>[] = [{ label: 'TypeScript', data: [5], hidden: true, fill: false, borderColor: '#abc123' }];

describe('useMonthlyCountReport', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.resetModules();
    mockRepository.getAll.mockResolvedValue(seedRecords);
    mockService.getRecentMonths.mockReturnValue(seedMonths);
    mockService.buildChartDatasets.mockReturnValue(seedDatasets);
  });

  it('populates chartData.labels and datasets after successful load', async () => {
    const { useMonthlyCountReport } = await import('@/composables/use-monthly-count-report');
    const { chartData, error, load } = useMonthlyCountReport();
    await load();
    expect(chartData.value.labels).toEqual(seedMonths);
    expect(chartData.value.datasets).toEqual(seedDatasets);
    expect(error.value).toBeNull();
  });

  it('sets error and leaves chartData empty when repository rejects', async () => {
    mockRepository.getAll.mockRejectedValue(new Error('network error'));
    const { useMonthlyCountReport } = await import('@/composables/use-monthly-count-report');
    const { chartData, error, load } = useMonthlyCountReport();
    await load();
    expect(error.value).toBeInstanceOf(Error);
    expect(chartData.value.datasets).toHaveLength(0);
  });
});
