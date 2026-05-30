import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';

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

const seedRecords: MonthlyCountRecord[] = [
  { id: '1', YearMonth: '2024-01', yearMonthDate: '2024-01-01', count: 5, skillName: 's1', expand: { skillName: { name: 'TypeScript' } } },
];
const seedMonths = ['2024-01'];
const seedDataPoints = [{ group: 'TypeScript', date: '2024-01', value: 5 }];

describe('useMonthlyCountReport', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.resetModules();
    mockRepository.getAll.mockResolvedValue(seedRecords);
    mockService.getRecentMonths.mockReturnValue(seedMonths);
    mockService.buildChartDatasets.mockReturnValue(seedDataPoints);
  });

  it('populates chartData with data points after successful load', async () => {
    const { useMonthlyCountReport } = await import('@/composables/use-monthly-count-report');
    const { chartData, error, load } = useMonthlyCountReport();
    await load();
    expect(chartData.value).toEqual(seedDataPoints);
    expect(chartData.value.length).toBeGreaterThan(0);
    expect(error.value).toBeNull();
  });

  it('sets error and leaves chartData empty when repository rejects', async () => {
    mockRepository.getAll.mockRejectedValue(new Error('network error'));
    const { useMonthlyCountReport } = await import('@/composables/use-monthly-count-report');
    const { chartData, error, load } = useMonthlyCountReport();
    await load();
    expect(error.value).toBeInstanceOf(Error);
    expect(chartData.value).toHaveLength(0);
  });
});
