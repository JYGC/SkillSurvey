export interface MonthlyCountRecord {
  YearMonth: string;
  yearMonthDate: string;
  count: number;
  expand?: { skillName?: { name: string } };
}
