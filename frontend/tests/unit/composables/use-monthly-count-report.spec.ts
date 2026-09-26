import { describe, it, expect, vi, beforeEach } from 'vitest';

const mockLoadRecentMonthlyCountChartData = vi.hoisted(() => vi.fn());

vi.mock('@/services/monthly-count-report.service', () => ({
  loadRecentMonthlyCountChartData: mockLoadRecentMonthlyCountChartData,
}));

const seedDataPoints = [{ group: 'TypeScript', date: '2024-01', value: 5 }];

describe('useMonthlyCountReport', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.resetModules();
    mockLoadRecentMonthlyCountChartData.mockResolvedValue(seedDataPoints);
  });

  it('populates chartData with data points after successful load', async () => {
    const { useMonthlyCountReport } = await import('@/composables/use-monthly-count-report');
    const { chartData, error, load } = useMonthlyCountReport();
    await load();
    expect(chartData.value).toEqual(seedDataPoints);
    expect(chartData.value.length).toBeGreaterThan(0);
    expect(error.value).toBeNull();
  });

  it('sets error and leaves chartData empty when the service rejects', async () => {
    mockLoadRecentMonthlyCountChartData.mockRejectedValue(new Error('network error'));
    const { useMonthlyCountReport } = await import('@/composables/use-monthly-count-report');
    const { chartData, error, load } = useMonthlyCountReport();
    await load();
    expect(error.value).toBeInstanceOf(Error);
    expect(chartData.value).toHaveLength(0);
  });
});
