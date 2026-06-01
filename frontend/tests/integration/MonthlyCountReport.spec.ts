import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';

vi.mock('@carbon/charts-vue', () => ({
  CcvLineChart: { template: '<svg />' },
}));

const mockGetAll = vi.hoisted(() => vi.fn());

vi.mock('@/repositories/monthly-count-report.repository', () => ({
  monthlyCountReportRepository: { getAll: mockGetAll },
}));

const stubs = { CcvLineChart: { template: '<svg />' } };

const seedRecords: MonthlyCountRecord[] = [
  { id: '1', YearMonth: '2024-10', yearMonthDate: '2024-10-01', count: 10, skillName: 's1', expand: { skillName: { name: 'TestSkill' } } },
  { id: '2', YearMonth: '2024-11', yearMonthDate: '2024-11-01', count: 15, skillName: 's1', expand: { skillName: { name: 'TestSkill' } } },
  { id: '3', YearMonth: '2024-12', yearMonthDate: '2024-12-01', count: 20, skillName: 's1', expand: { skillName: { name: 'TestSkill' } } },
];

describe('MonthlyCountReport', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('renders svg element after data loads', async () => {
    mockGetAll.mockResolvedValue(seedRecords);
    const { default: MonthlyCountReport } = await import('@/views/public/MonthlyCountReport.vue');
    const wrapper = mount(MonthlyCountReport, { global: { stubs } });
    await flushPromises();
    expect(wrapper.find('svg').exists()).toBe(true);
    expect(wrapper.find('p[data-testid="report-error"]').exists()).toBe(false);
  });

  it('shows error text when repository rejects', async () => {
    mockGetAll.mockRejectedValue(new Error('fetch failed'));
    const { default: MonthlyCountReport } = await import('@/views/public/MonthlyCountReport.vue');
    const wrapper = mount(MonthlyCountReport, { global: { stubs } });
    await flushPromises();
    expect(wrapper.find('p[data-testid="report-error"]').exists()).toBe(true);
  });
});
