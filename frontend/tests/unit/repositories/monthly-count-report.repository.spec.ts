import { describe, it, expect, vi, beforeEach } from 'vitest';

const { mockPb, mockCollection } = vi.hoisted(() => {
  const mockCollection = { getFullList: vi.fn().mockResolvedValue([]) };
  const mockPb = { collection: vi.fn().mockReturnValue(mockCollection) };
  return { mockPb, mockCollection };
});

vi.mock('@/store/pocketbase', () => ({ default: mockPb }));

import { monthlyCountReportRepository } from '@/repositories/monthly-count-report.repository';

describe('monthlyCountReportRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPb.collection.mockReturnValue(mockCollection);
    mockCollection.getFullList.mockResolvedValue([]);
  });

  it('getAll calls getFullList on monthlyCountReports with expand and sort', async () => {
    await monthlyCountReportRepository.getAll();
    expect(mockPb.collection).toHaveBeenCalledWith('monthlyCountReports');
    expect(mockCollection.getFullList).toHaveBeenCalledWith({
      expand: 'skillName',
      sort: 'yearMonthDate',
    });
  });

  it('getAll returns the records returned by pb', async () => {
    const records = [{ YearMonth: '2024-10', yearMonthDate: '2024-10-01', count: 5 }];
    mockCollection.getFullList.mockResolvedValue(records);
    const result = await monthlyCountReportRepository.getAll();
    expect(result).toEqual(records);
  });
});
