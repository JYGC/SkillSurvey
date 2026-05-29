# Tasks: Frontend Layered Architecture

Work bottom-up: store → repositories → services → schemas → composables → views.
Each task is independent of the ones above it in the same layer.
Mark each `[x]` as you complete it.

---

## Layer 1 — Store

- [ ] **T1** Create `frontend/src/store/pocketbase.ts`
  - Module-level PocketBase singleton initialised with `VUE_APP_POCKETBASE_URL`
  - Loads auth from cookie on creation; registers `onChange` to write cookie back
  - Default-exports the `pb` instance
  - _Outcome_: any file can `import pb from '@/store/pocketbase'` and get the same instance

---

## Layer 2 — Schemas

- [ ] **T2** Create `frontend/src/schemas/monthly-count-report.ts`
  - Move `MonthlyCountRecord` interface out of `MonthlyCountReport.vue` into this file
  - Export the interface
  - _Outcome_: `MonthlyCountRecord` is importable from `@/schemas/monthly-count-report`

---

## Layer 3 — Repositories

Dependencies: T1, T2

- [ ] **T3** Create `frontend/src/repositories/auth.repository.ts`
  - Imports singleton from `@/store/pocketbase`
  - Exposes `authRepository` object with: `isAuthenticated` (getter), `currentUser` (getter), `login(email, password)`, `register(name, email, password, passwordConfirm)`, `logout()`
  - `login` and `register` are async; `logout` clears `pb.authStore`
  - _Outcome_: all auth operations accessible via one import, no raw PocketBase calls in views

- [ ] **T4** Create `frontend/src/repositories/monthly-count-report.repository.ts`
  - Imports singleton and `MonthlyCountRecord`
  - Exposes `monthlyCountReportRepository.getAll()` returning `Promise<MonthlyCountRecord[]>`
  - Fetches `monthlyCountReports` expanded with `skillName`, sorted by `yearMonthDate`
  - _Outcome_: collection access is encapsulated; composable calls `getAll()`, not PocketBase directly

- [ ] **T5** Create `frontend/src/repositories/user-settings.repository.ts`
  - Imports singleton and `IUserSettings`
  - Exposes `userSettingsRepository.getOrCreate(userId)` returning `Promise<IUserSettings>`
  - On "not found" creates a default record (`portalTheme: 'white'`) and returns it; rethrows any other error
  - _Outcome_: Settings.vue / composable no longer contains try/catch for "not found"

---

## Layer 4 — Services

Dependencies: T2

- [ ] **T6** Create `frontend/src/services/monthly-count-report.service.ts`
  - Extract `getRecentMonths(records)` — returns last 12 distinct `YearMonth` strings
  - Extract `buildChartDatasets(records, months)` — groups by skill name, fills zeros, assigns random colour, sets `hidden: true`
  - Both functions are pure (no side effects, no imports from store/repositories)
  - _Outcome_: business logic is testable in isolation; view contains no transformation code

---

## Layer 5 — Composables

Dependencies: T3, T4, T5, T6

- [ ] **T7** Create `frontend/src/composables/use-auth.ts`
  - Returns `{ isAuthenticated, currentUser, login, logout }`
  - `isAuthenticated` and `currentUser` are `computed` (derived from `authRepository`)
  - `login(email, password)` is async, throws on failure
  - `logout()` calls `authRepository.logout()`
  - _Outcome_: layouts and Login.vue use `useAuth()` instead of calling the store/repo directly

- [ ] **T8** Create `frontend/src/composables/use-monthly-count-report.ts`
  - Calls `Chart.register(...registerables)` once at module level
  - Returns `{ chartData, chartHeight, error }`
  - `chartData` is `reactive({ labels: [], datasets: [] })`; `error` is `ref(null)`
  - Calls `load()` immediately on composable invocation
  - On success populates `chartData`; on failure sets `error`, leaves `chartData` empty
  - _Outcome_: MonthlyCountReport.vue contains no data-fetching or transformation logic

- [ ] **T9** Create `frontend/src/composables/use-user-settings.ts`
  - Returns `{ userSetting, load }`
  - `load()` returns early if `authRepository.currentUser` is null
  - On success sets `userSetting.value`; propagates unexpected errors
  - _Outcome_: Settings.vue contains no repository calls

---

## Layer 6 — Views and layouts

Dependencies: T7, T8, T9

- [ ] **T10** Refactor `App.vue` — Options API → Composition API
  - Replace `defineComponent({ name: 'App' })` block with `<script lang="ts" setup>`
  - No logic needed; the script block can be empty or omitted entirely
  - _Outcome_: no Options API usage remains in the codebase

- [ ] **T11** Refactor `PublicLayout.vue`
  - Replace `getBackendClient()` auth check with `useAuth().isAuthenticated`
  - _Outcome_: layout imports nothing from `services/backend-client`

- [ ] **T12** Refactor `UserLayout.vue`
  - Replace `getBackendClient()` auth check with `useAuth().isAuthenticated`
  - Replace `backendClient.authStore.record` display with `useAuth().currentUser`
  - Replace `backendClient.authStore.clear()` in logout with `useAuth().logout()`
  - _Outcome_: layout imports nothing from `services/backend-client`

- [ ] **T13** Refactor `Login.vue`
  - Replace `getBackendClient().collection('users').authWithPassword(...)` with `await useAuth().login(...)`
  - Fix the existing bug: `login()` was not awaited, so navigation ran before auth completed
  - On catch, display `error.message` in a `<p>` element instead of `alert()`
  - _Outcome_: login is properly async; no raw PocketBase calls; no `alert()`

- [ ] **T14** Refactor `RegisterUser.vue`
  - Replace `pb.collection('users').create(...)` with `authRepository.register(...)`
  - On catch, display error in a `<p>` element instead of `alert()`
  - _Outcome_: no raw PocketBase calls; no `alert()`

- [ ] **T15** Refactor `MonthlyCountReport.vue`
  - Replace inline IIFE, `createDatasets`, `Chart.register`, and all PocketBase calls with `useMonthlyCountReport()`
  - Bind `chartData` and `chartHeight` from composable to `<LineChart>`
  - Display `error` ref value in a `<p>` element when set
  - Remove the local `MonthlyCountRecord` interface (now in schemas)
  - _Outcome_: view is < 20 lines of template + script

- [ ] **T16** Refactor `Settings.vue`
  - Replace inline repository call and try/catch with `useUserSettings()`
  - Call `load()` via `onMounted`
  - _Outcome_: no raw PocketBase calls; no direct repository imports in view

---

## Layer 7 — Cleanup

Dependencies: T11–T16 all complete

- [ ] **T17** Delete `frontend/src/services/backend-client.ts`
  - Confirm no remaining imports of `getBackendClient` anywhere in `src/`
  - Delete the file
  - _Outcome_: `getBackendClient` no longer exists; store singleton is the only PocketBase entry point

---

## Verification

Dependencies: T17

- [ ] **T18** Run `npm run lint` on the server — zero errors, zero warnings introduced by this change
- [ ] **T19** Run `/frontend-openbsd serve` and manually verify:
  - `/login` renders and login succeeds
  - `/monthly-count-report` renders the line chart
  - `/user/settings` renders user settings
  - Logout from `/user/profile` redirects to `/`
