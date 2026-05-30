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

### Frameworks and tools

| Scope | Framework | Notes |
|---|---|---|
| Unit (services) | Vitest 2.x + happy-dom | Pure functions; no browser or HTTP |
| Unit (composables) | Vitest + @vue/test-utils | `vi.mock` for repository modules |
| Contract | Vitest + `fetch` | HTTP assertions against real PocketBase; no browser |
| Integration (component) | Vitest + @vue/test-utils | Full component mount against real PocketBase |
| E2E | WebdriverIO 9.x + Mocha | Full browser flows; uses system Chrome + chromedriver on OpenBSD |

### OpenBSD-specific configuration

Vitest and @vue/test-utils run in Node.js with no OS-specific requirements.

Playwright throws `Unsupported platform: openbsd` at module-load time (hard-coded registry check) and cannot be used on OpenBSD. WebdriverIO uses the system `chromedriver` binary directly and has no platform restrictions.

On OpenBSD, Chrome is at `/usr/local/bin/chrome` and chromedriver at `/usr/local/bin/chromedriver`. The `wdio.conf.ts` reads `CHROMIUM_PATH` and `CHROMEDRIVER_PATH` env vars:

```typescript
// wdio.conf.ts (relevant excerpt)
capabilities: [{
  browserName: 'chrome',
  'goog:chromeOptions': {
    binary: process.env.CHROMIUM_PATH ?? '/usr/local/bin/chrome',
    args: ['--headless', '--no-sandbox', '--disable-dev-shm-usage'],
  },
}],
services: [['chromedriver', {
  chromedriverCustomPath: process.env.CHROMEDRIVER_PATH ?? '/usr/local/bin/chromedriver',
}]],
```

`--no-sandbox` is required when running Chromium as a non-root user on OpenBSD.

### New dev dependencies

```
vitest
@vitest/coverage-v8
@vue/test-utils
happy-dom
@wdio/cli
@wdio/local-runner
@wdio/mocha-framework
@wdio/spec-reporter
@wdio/types
wdio-chromedriver-service
webdriverio
```

Add to `frontend/package.json` `devDependencies`. Do not run `playwright install` — system Chrome + chromedriver are used instead.

### Directory structure

```
frontend/
  tests/
    unit/
      services/
        monthly-count-report.service.spec.ts
        arrays.spec.ts
      composables/
        use-auth.spec.ts
        use-monthly-count-report.spec.ts
        use-user-settings.spec.ts
    contract/
      auth.contract.spec.ts
      monthly-count-report.contract.spec.ts
      user-settings.contract.spec.ts
    integration/
      Login.spec.ts
      MonthlyCountReport.spec.ts
      Settings.spec.ts
    e2e/
      login.e2e.spec.ts
      monthly-count-report.e2e.spec.ts
    setup/
      vitest.global-setup.ts   ← starts/stops real pocketbaseserver binary
      vitest.setup.ts          ← per-file setup (auth state reset)
      seed.ts                  ← helpers to create test users and records via API
  vitest.config.ts
  wdio.conf.ts
```

### `vitest.config.ts`

```typescript
import { defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';
import { resolve } from 'path';

export default defineConfig({
  plugins: [vue()],
  resolve: { alias: { '@': resolve(__dirname, 'src') } },
  test: {
    environment: 'happy-dom',
    globalSetup: './tests/setup/vitest.global-setup.ts',
    setupFiles:  './tests/setup/vitest.setup.ts',
  },
});
```

### `tests/setup/vitest.global-setup.ts`

Starts and tears down a real pocketbaseserver binary. Runs once per test suite.

```typescript
import { spawn, ChildProcess } from 'child_process';
import * as fs from 'fs';
import * as os from 'os';
import * as path from 'path';

let server: ChildProcess;
let tmpDir: string;

export async function setup() {
  tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'pb-test-'));
  server = spawn(
    path.resolve(__dirname, '../../../pocketbaseserver/build/pocketbaseserver'),
    ['serve', '--http', '127.0.0.1:18090', '--dir', tmpDir],
    { stdio: 'pipe' }
  );
  await waitForUrl('http://127.0.0.1:18090/api/health');
  process.env.TEST_PB_URL = 'http://127.0.0.1:18090';
  await seedInitialData(process.env.TEST_PB_URL);
}

export async function teardown() {
  server.kill();
  fs.rmSync(tmpDir, { recursive: true, force: true });
}

async function waitForUrl(url: string, timeoutMs = 10_000) {
  const deadline = Date.now() + timeoutMs;
  while (Date.now() < deadline) {
    try { const r = await fetch(url); if (r.ok) return; } catch {}
    await new Promise(r => setTimeout(r, 200));
  }
  throw new Error(`Timed out waiting for ${url}`);
}
```

`seedInitialData` (in `tests/setup/seed.ts`):
1. Creates the first superadmin via `POST /api/collections/_superusers/records` (allowed only when no admins exist on a fresh instance).
2. Authenticates as superadmin to obtain an admin token.
3. Creates a test regular user via `POST /api/collections/users/records` (admin-only because self-registration is disabled).
4. Creates seed `monthlyCountReports` and `userSettings` records needed by integration and E2E tests.
5. Writes test credentials to `process.env.TEST_USER_EMAIL` and `process.env.TEST_USER_PASSWORD`.

### `tests/setup/vitest.setup.ts`

Runs before each test file. Resets the PocketBase singleton's auth state so tests don't share login sessions.

```typescript
import pb from '@/store/pocketbase';
import { beforeEach } from 'vitest';

beforeEach(() => {
  pb.authStore.clear();
  // Override base URL so the singleton points at the test server
  (pb as any)._baseUrl = process.env.TEST_PB_URL ?? pb.baseUrl;
});
```

### npm test scripts

Add to `frontend/package.json`:

```json
"test:unit":        "vitest run tests/unit",
"test:contract":    "vitest run tests/contract",
"test:integration": "vitest run tests/integration",
"test:e2e":         "playwright test",
"test":             "npm run test:unit && npm run test:contract && npm run test:integration && npm run test:e2e"
```

Unit tests do not require the PocketBase server; the global setup is guarded so it only starts the server when running contract, integration, or E2E tests.

## Phase 9 — Chart library migration (chart.js → @carbon/charts-vue)

### Overview

`chart.js` and `vue-chart-3` are replaced by `CcvLineChart` from `@carbon/charts-vue`, the charting package in the Carbon Design System family. `@carbon/vue` (already installed) does not include chart components — charts are in the separate `@carbon/charts-vue` package.


### Data format change

Chart.js uses a `{ labels, datasets }` structure:

```typescript
{
  labels: ['2024-10', '2024-11'],
  datasets: [{ label: 'Python', data: [10, 15], fill: false, borderColor: '...', hidden: true }]
}
```

Carbon Charts uses a flat array of data points:

```typescript
[
  { group: 'Python', date: '2024-10', value: 10 },
  { group: 'Python', date: '2024-11', value: 15 },
]
```

Chart options (axis config, title) are passed separately as a `:options` prop.

### New type

Add `CarbonChartDataPoint` to `services/monthly-count-report.service.ts`:

```typescript
export interface CarbonChartDataPoint {
  group: string;
  date: string;
  value: number;
}
```

### Files affected

| File | Change |
|---|---|
| `package.json` | Add `@carbon/charts`, `@carbon/charts-vue`; remove `vue-chart-3` |
| `services/monthly-count-report.service.ts` | `buildChartDatasets` returns `CarbonChartDataPoint[]`; remove Chart.js types |
| `composables/use-monthly-count-report.ts` | `chartData` type → `CarbonChartDataPoint[]`; remove `Chart.register`; add `chartOptions` |
| `views/public/MonthlyCountReport.vue` | Replace `<LineChart>` with `<CcvLineChart>`; import styles |
| `tests/unit/services/monthly-count-report.service.spec.ts` | Update `buildChartDatasets` assertions (no `hidden`/`borderColor`; assert `group`/`date`/`value`) |
| `tests/unit/composables/use-monthly-count-report.spec.ts` | Assert `chartData` is `CarbonChartDataPoint[]`; remove `chartData.labels`/`datasets` assertions |
| `tests/integration/MonthlyCountReport.spec.ts` | Mock `@carbon/charts-vue` instead of `chart.js`; stub `CcvLineChart`; check `svg` not `canvas` |
| `tests/e2e/monthly-count-report.e2e.ts` | Check `svg` instead of `canvas` |

### Updated service interface

```typescript
export function buildChartDatasets(
  records: MonthlyCountRecord[],
  months: string[],
): CarbonChartDataPoint[] {
  return months.flatMap(month =>
    Object.entries(groupBySkill(records)).map(([name, counts]) => ({
      group: name,
      date: month,
      value: counts[month] ?? 0,
    }))
  );
}
```

`getRecentMonths` is unchanged — still used to extract and limit to the last 12 months before passing to `buildChartDatasets`.

### Updated composable interface

```typescript
export function useMonthlyCountReport() {
  const chartData = ref<CarbonChartDataPoint[]>([]);
  const chartOptions = {
    axes: {
      bottom: { title: 'Month', mapsTo: 'date', scaleType: 'labels' },
      left: { title: 'Job Listings', mapsTo: 'value', scaleType: 'linear' },
    },
    height: '400px',
  };
  const error = ref<Error | null>(null);

  async function load() {
    try {
      const records = await monthlyCountReportRepository.getAll();
      const months = getRecentMonths(records);
      chartData.value = buildChartDatasets(records, months);
      error.value = null;
    } catch (e) {
      error.value = e instanceof Error ? e : new Error(String(e));
    }
  }

  return { chartData, chartOptions, error, load };
}
```

No `Chart.register` call. `chartHeight` is removed — height is expressed in `chartOptions.height` as a CSS string.

### Updated view

```vue
<template>
  <div>
    <p v-if="error" data-testid="report-error">{{ error.message }}</p>
    <CcvLineChart v-else :data="chartData" :options="chartOptions" />
  </div>
</template>
<script lang="ts" setup>
import { onMounted } from 'vue';
import { CcvLineChart } from '@carbon/charts-vue';
import '@carbon/charts-vue/styles.css';
import { useMonthlyCountReport } from '@/composables/use-monthly-count-report';

const { chartData, chartOptions, error, load } = useMonthlyCountReport();
onMounted(load);
</script>
```

### Updated integration test mock

```typescript
vi.mock('@carbon/charts-vue', () => ({
  CcvLineChart: { template: '<svg />' },
}));
const stubs = { CcvLineChart: { template: '<svg />' } };
// assertion changes: wrapper.find('svg').exists() → true
```

### Updated E2E selector

```typescript
await $('svg').waitForExist({ timeout: 15000 });
expect(await $('svg').isExisting()).toBe(true);
```

### Running on the OpenBSD server

```sh
# Unit only (no server needed)
npm run test:unit

# Contract + integration (starts/stops pocketbaseserver automatically)
npm run test:contract
npm run test:integration

# E2E (requires CHROMIUM_PATH)
CHROMIUM_PATH=/usr/local/bin/chromium npm run test:e2e

# All
CHROMIUM_PATH=/usr/local/bin/chromium npm test
```
