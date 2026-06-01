export interface MonthlyCountRecord {
  id: string;
  YearMonth: string;
  yearMonthDate: string;
  count: number;
  skillName: string;
  expand?: { skillName?: { name: string } };
}
