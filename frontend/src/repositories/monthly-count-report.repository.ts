import pb from '@/store/pocketbase';
import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';

export const monthlyCountReportRepository = {
  async getAll(): Promise<MonthlyCountRecord[]> {
    return pb.collection('monthlyCountReports').getFullList<MonthlyCountRecord>({
      expand: 'skillName',
      sort: 'yearMonthDate',
    });
  },
};
