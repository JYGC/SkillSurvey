# Design: Frontend Layered Architecture

## System architecture and components

The frontend adopts a four-layer structure. Each layer imports only from the layer directly below it; views never reach past the composable layer.

```
View / Layout
      │  imports
      ▼
Composable  (use-auth, use-monthly-count-report, use-user-settings)
      │  imports
      ▼
Service     (monthly-count-report.service, arrays)
Repository  (auth, monthly-count-report, user-settings)
      │  imports
      ▼
Store       (pocketbase singleton)
```

### Target directory structure

```
frontend/src/
  store/
    pocketbase.ts                           ← NEW (replaces getBackendClient)
  repositories/
    auth.repository.ts                      ← NEW
    monthly-count-report.repository.ts      ← NEW
    user-settings.repository.ts             ← NEW
  services/
    monthly-count-report.service.ts         ← NEW (logic extracted from view)
    arrays.ts                               ← unchanged
  composables/
    use-auth.ts                             ← NEW
    use-monthly-count-report.ts             ← NEW
    use-user-settings.ts                    ← NEW
  schemas/
    monthly-count-report.ts                 ← NEW (MonthlyCountRecord moved here)
    skills.ts                               ← unchanged
    users.ts                                ← unchanged
  layouts/
    PublicLayout.vue                        ← refactored (use-auth)
    UserLayout.vue                          ← refactored (use-auth)
  views/
    public/
      HomeRoute.vue                         ← unchanged
      Login.vue                             ← refactored (use-auth)
      MonthlyCountReport.vue                ← refactored (use-monthly-count-report)
      RegisterUser.vue                      ← refactored (auth.repository)
    user/
      Profile.vue                           ← unchanged
      Settings.vue                          ← refactored (use-user-settings)
  App.vue                                   ← refactored (script setup)
  main.ts                                   ← unchanged
  services/backend-client.ts               ← DELETED
```

## Interfaces and data models

### `schemas/monthly-count-report.ts` (new)

```typescript
export interface MonthlyCountRecord {
  YearMonth: string;
  yearMonthDate: string;
  count: number;
  expand?: { skillName?: { name: string } };
}
```

### `store/pocketbase.ts`

```typescript
import PocketBase from 'pocketbase';

const pb = new PocketBase(process.env.VUE_APP_POCKETBASE_URL);
pb.authStore.loadFromCookie(document.cookie);
pb.authStore.onChange(() => {
  document.cookie = pb.authStore.exportToCookie({ httpOnly: false, secure: false });
});

export default pb;
```

### `repositories/auth.repository.ts`

```typescript
import pb from '@/store/pocketbase';

export const authRepository = {
  get isAuthenticated(): boolean { return pb.authStore.isValid; },
  get currentUser() { return pb.authStore.record; },
  async login(email: string, password: string) {
    return pb.collection('users').authWithPassword(email, password);
  },
  async register(name: string, email: string, password: string, passwordConfirm: string) {
    return pb.collection('users').create({ name, email, password, passwordConfirm });
  },
  logout() { pb.authStore.clear(); },
};
```

Registration is kept on `authRepository` because creating a user account is an authentication concern (same collection, same auth context).

### `repositories/monthly-count-report.repository.ts`

```typescript
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
```

### `repositories/user-settings.repository.ts`

```typescript
import pb from '@/store/pocketbase';
import type { IUserSettings } from '@/schemas/users';

export const userSettingsRepository = {
  async getOrCreate(userId: string): Promise<IUserSettings> {
    try {
      return await pb.collection('userSettings').getFirstListItem<IUserSettings>(
        `user="${userId}"`, { fields: 'id,user,portalTheme' }
      );
    } catch (e: any) {
      if (!e?.message?.includes("wasn't found")) throw e;
      const defaults: IUserSettings = { id: '', user: userId, portalTheme: 'white' };
      return pb.collection('userSettings').create<IUserSettings>(defaults);
    }
  },
};
```

### `services/monthly-count-report.service.ts`

```typescript
import type { ChartDataset } from 'chart.js';
import type { MonthlyCountRecord } from '@/schemas/monthly-count-report';

export function getRecentMonths(records: MonthlyCountRecord[]): string[] {
  return [...new Set(records.map(r => r.YearMonth))].slice(-12);
}

export function buildChartDatasets(
  records: MonthlyCountRecord[],
  months: string[]
): ChartDataset<'line'>[] {
  const bySkill: Record<string, Record<string, number>> = {};
  for (const r of records) {
    const name = r.expand?.skillName?.name ?? 'Unknown';
    if (!bySkill[name]) bySkill[name] = {};
    bySkill[name][r.YearMonth] = r.count;
  }
  return Object.entries(bySkill).map(([label, counts]) => ({
    label,
    data: months.map(m => counts[m] ?? 0),
    fill: false,
    borderColor: `#${Math.floor(Math.random() * 16777215).toString(16).padStart(6, '0')}`,
    hidden: true,
  } as ChartDataset<'line'>));
}
```

### `composables/use-auth.ts`

```typescript
import { computed } from 'vue';
import { authRepository } from '@/repositories/auth.repository';

export function useAuth() {
  const isAuthenticated = computed(() => authRepository.isAuthenticated);
  const currentUser = computed(() => authRepository.currentUser);

  async function login(email: string, password: string) {
    await authRepository.login(email, password);
  }

  function logout() {
    authRepository.logout();
  }

  return { isAuthenticated, currentUser, login, logout };
}
```

`isAuthenticated` and `currentUser` are `computed` (not `ref`) so they stay in sync with the PocketBase auth store without extra watchers.

### `composables/use-monthly-count-report.ts`

```typescript
import { reactive, ref } from 'vue';
import { Chart, registerables } from 'chart.js';
import type { ChartDataset } from 'chart.js';
import { monthlyCountReportRepository } from '@/repositories/monthly-count-report.repository';
import { getRecentMonths, buildChartDatasets } from '@/services/monthly-count-report.service';

Chart.register(...registerables);

export function useMonthlyCountReport() {
  const chartHeight = window.innerHeight;
  const chartData = reactive({ labels: [] as string[], datasets: [] as ChartDataset<'line'>[] });
  const error = ref<unknown>(null);

  async function load() {
    try {
      const records = await monthlyCountReportRepository.getAll();
      const months = getRecentMonths(records);
      chartData.labels = months;
      chartData.datasets = buildChartDatasets(records, months);
    } catch (e) {
      error.value = e;
    }
  }

  load();

  return { chartData, chartHeight, error };
}
```

### `composables/use-user-settings.ts`

```typescript
import { ref } from 'vue';
import { authRepository } from '@/repositories/auth.repository';
import { userSettingsRepository } from '@/repositories/user-settings.repository';
import type { IUserSettings } from '@/schemas/users';

export function useUserSettings() {
  const userSetting = ref<IUserSettings | null>(null);

  async function load() {
    if (!authRepository.currentUser) return;
    userSetting.value = await userSettingsRepository.getOrCreate(authRepository.currentUser.id);
  }

  return { userSetting, load };
}
```

## Sequence diagrams

### Login flow

```
Login.vue  →  useAuth.login(email, pw)
           →  authRepository.login(email, pw)
           →  pb.collection('users').authWithPassword(...)
           ←  AuthRecord
           ←  (resolves)
           →  router.push('/user/profile')
```

### MonthlyCountReport load flow

```
MonthlyCountReport.vue  mounts
  →  useMonthlyCountReport()  [calls load() immediately]
  →  monthlyCountReportRepository.getAll()
  →  pb.collection('monthlyCountReports').getFullList(...)
  ←  MonthlyCountRecord[]
  →  getRecentMonths(records)        [service]
  →  buildChartDatasets(records, months)  [service]
  ←  ChartDataset<'line'>[]
  ←  chartData populated  →  view re-renders
```

### Settings load flow

```
Settings.vue  onMounted
  →  useUserSettings().load()
  →  authRepository.currentUser  (check)
  →  userSettingsRepository.getOrCreate(userId)
  →  pb.collection('userSettings').getFirstListItem(...)
  ←  IUserSettings  (or create default + return)
  ←  userSetting.value set  →  view re-renders
```

## Error-handling approach

- Repositories propagate errors (throw). They do not alert or log.
- Composables catch errors from repositories/services, store them in an `error` ref, and leave reactive state at its safe initial value.
- Views read the `error` ref and display it. No `alert()` calls remain after refactoring.
- The one exception is `userSettingsRepository.getOrCreate`, which internally handles the "not found" case as normal control flow (not an error).

## Testing strategy

This is a pure refactor — no new API-connected behaviour is introduced. No new integration tests are required by the mandate. Correctness is verified by:

1. `npm run lint` — no type or style errors after refactoring.
2. Manual serve smoke-test: all existing routes render and behave identically to before.
