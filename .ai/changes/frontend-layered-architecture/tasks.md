# Tasks: Frontend Layered Architecture

Work order: infrastructure → schemas/store → test infra → (write tests then implement) per layer → views → cleanup → verify.
Mark each `[x]` as you complete it.

---

## Phase 0 — Foundation (no tests yet)

- [ ] **T1** Create `frontend/src/store/pocketbase.ts`
  - Module-level PocketBase singleton: init with `VUE_APP_POCKETBASE_URL`, load cookie, register onChange
  - Default-export the `pb` instance
  - _Outcome_: any file can `import pb from '@/store/pocketbase'` and get the same instance

- [ ] **T2** Create `frontend/src/schemas/monthly-count-report.ts`
  - Move `MonthlyCountRecord` interface out of `MonthlyCountReport.vue`; export it
  - _Outcome_: `MonthlyCountRecord` importable from `@/schemas/monthly-count-report`

---

## Phase 1 — Test infrastructure

Dependencies: T1

- [ ] **T3** Install test dev dependencies
  - Add to `frontend/package.json` devDependencies: `vitest`, `@vitest/coverage-v8`, `@vue/test-utils`, `happy-dom`, `@playwright/test`
  - Run `npm install` on the server (do **not** run `playwright install` — use system Chromium)
  - _Outcome_: `vitest` and `playwright` CLIs available

- [ ] **T4** Create `frontend/vitest.config.ts`
  - Configure: `environment: 'happy-dom'`, `resolve.alias` `@` → `src/`, `globalSetup` and `setupFiles` pointing at Phase 1 setup files
  - _Outcome_: `npm run test:unit` resolves imports and runs in happy-dom

- [ ] **T5** Create `frontend/playwright.config.ts`
  - Single `chromium` project; reads `CHROMIUM_PATH` env var for `executablePath`; sets `--no-sandbox` in `launchOptions`
  - `baseURL` reads `TEST_E2E_URL` env var (the URL where `npm run serve` listens during E2E)
  - _Outcome_: `CHROMIUM_PATH=/usr/local/bin/chromium npx playwright test` can launch a browser

- [ ] **T6** Create `frontend/tests/setup/vitest.global-setup.ts`
  - `setup()`: spawn `pocketbaseserver/build/pocketbaseserver serve --http 127.0.0.1:18090 --dir <tmpdir>`, poll `/api/health`, call `seedInitialData()`
  - `teardown()`: kill process, remove tmpdir
  - _Outcome_: contract and integration tests get a clean PocketBase instance per run

- [ ] **T7** Create `frontend/tests/setup/seed.ts`
  - `seedInitialData(baseUrl)`: create superadmin via `POST /api/collections/_superusers/records`, authenticate, create test user, insert seed `monthlyCountReports` and `userSettings` records
  - Write credentials to `process.env.TEST_USER_EMAIL` / `TEST_USER_PASSWORD`
  - _Outcome_: integration and E2E tests can authenticate and find test data

- [ ] **T8** Create `frontend/tests/setup/vitest.setup.ts`
  - `beforeEach`: clear `pb.authStore`; override `pb`'s base URL to `TEST_PB_URL` when set
  - _Outcome_: tests cannot share login state across files

- [ ] **T9** Add npm test scripts to `frontend/package.json`
  - `"test:unit"`, `"test:contract"`, `"test:integration"`, `"test:e2e"`, `"test"`
  - _Outcome_: each test scope runnable independently on the server

---

## Phase 2 — Services (unit tests first)

Dependencies: T2, T4

- [ ] **T10** Write unit tests for `monthly-count-report.service.ts` — **before implementing**
  - File: `tests/unit/services/monthly-count-report.service.spec.ts`
  - `getRecentMonths`: empty input → `[]`; < 12 months → all; > 12 months → last 12
  - `buildChartDatasets`: one dataset per skill; missing months filled with `0`; every dataset has `hidden: true`
  - Tests import from `@/services/monthly-count-report` — they will fail until T11
  - _Outcome_: failing tests that define the service contract

- [ ] **T11** Create `frontend/src/services/monthly-count-report.service.ts`
  - Implement `getRecentMonths` and `buildChartDatasets` to make T10 tests pass
  - Both functions are pure (no imports from store/repositories)
  - _Outcome_: `npm run test:unit` green for service tests

- [ ] **T12** Write unit tests for `arrays.ts`
  - File: `tests/unit/services/arrays.spec.ts`
  - `sortByProperty`: sorts ascending by string property; sorts ascending by numeric property; returns same array reference
  - _Outcome_: `arrays.ts` covered by tests

---

## Phase 3 — Repositories (contract tests first)

Dependencies: T1, T2, T6, T7

- [ ] **T13** Write contract tests for auth — **before implementing auth repository**
  - File: `tests/contract/auth.contract.spec.ts`
  - Unauthenticated `POST /api/collections/users/records` (self-register) → HTTP 400
  - Valid `POST /api/collections/users/auth-with-password` → HTTP 200 with token
  - Invalid credentials → HTTP 400
  - _Outcome_: failing tests that verify PocketBase access rules independently of our code

- [ ] **T14** Write contract tests for `monthlyCountReports` — **before implementing**
  - File: `tests/contract/monthly-count-report.contract.spec.ts`
  - Unauthenticated `GET /api/collections/monthlyCountReports/records` → HTTP 403
  - Authenticated → HTTP 200
  - _Outcome_: access rule verified at the HTTP layer

- [ ] **T15** Write contract tests for `userSettings` — **before implementing**
  - File: `tests/contract/user-settings.contract.spec.ts`
  - Authenticated user fetches their own record → HTTP 200
  - Authenticated user fetches another user's record → HTTP 403
  - Fetch when no record exists → HTTP 404
  - _Outcome_: per-user isolation rule verified

- [ ] **T16** Create `frontend/src/repositories/auth.repository.ts`
  - `isAuthenticated` getter, `currentUser` getter, `login(email, pw)`, `register(name, email, pw, pwConfirm)`, `logout()`
  - All contract tests in T13 pass
  - _Outcome_: auth operations in one place; no raw PocketBase calls in views

- [ ] **T17** Create `frontend/src/repositories/monthly-count-report.repository.ts`
  - `getAll()` → `Promise<MonthlyCountRecord[]>`; expands `skillName`, sorts `yearMonthDate`
  - T14 contract tests pass
  - _Outcome_: collection access encapsulated

- [ ] **T18** Create `frontend/src/repositories/user-settings.repository.ts`
  - `getOrCreate(userId)` → get existing, or create default (`portalTheme: 'white'`) and return
  - T15 contract tests pass; "not found" handled internally; all other errors re-thrown
  - _Outcome_: get-or-create logic in one place

---

## Phase 4 — Composables (unit tests first)

Dependencies: T11, T16, T17, T18, T4

- [ ] **T19** Write unit tests for `use-auth.ts` — **before implementing**
  - File: `tests/unit/composables/use-auth.spec.ts`
  - Mock `@/repositories/auth.repository` with `vi.mock`
  - `isAuthenticated` reflects mocked `authRepository.isAuthenticated`
  - `login` delegates to `authRepository.login`; propagates rejection
  - `logout` delegates to `authRepository.logout`
  - _Outcome_: failing tests defining composable contract

- [ ] **T20** Create `frontend/src/composables/use-auth.ts`
  - `isAuthenticated` and `currentUser` as `computed` (derived from repo)
  - `login` async, throws on failure; `logout` delegates
  - T19 tests pass
  - _Outcome_: layouts and Login.vue can use `useAuth()`

- [ ] **T21** Write unit tests for `use-monthly-count-report.ts` — **before implementing**
  - File: `tests/unit/composables/use-monthly-count-report.spec.ts`
  - Mock `@/repositories/monthly-count-report.repository` and `@/services/monthly-count-report`
  - On resolve: `chartData.labels` and `chartData.datasets` populated; `error` null
  - On reject: `error` set; `chartData` remains empty
  - _Outcome_: failing tests

- [ ] **T22** Create `frontend/src/composables/use-monthly-count-report.ts`
  - `Chart.register(...registerables)` at module level
  - Returns `{ chartData, chartHeight, error }`; calls `load()` on init
  - T21 tests pass
  - _Outcome_: MonthlyCountReport.vue delegates entirely to this composable

- [ ] **T23** Write unit tests for `use-user-settings.ts` — **before implementing**
  - File: `tests/unit/composables/use-user-settings.spec.ts`
  - Mock `@/repositories/auth.repository` and `@/repositories/user-settings.repository`
  - No current user → `userSetting` stays null; repository not called
  - With current user → repository called; `userSetting` set
  - _Outcome_: failing tests

- [ ] **T24** Create `frontend/src/composables/use-user-settings.ts`
  - Returns `{ userSetting, load }`; `load()` guards on `currentUser`
  - T23 tests pass
  - _Outcome_: Settings.vue delegates entirely to this composable

---

## Phase 5 — Views (integration tests first)

Dependencies: T20, T22, T24, T6, T7, T8

- [ ] **T25** Write integration test for `Login.vue` — **before refactoring**
  - File: `tests/integration/Login.spec.ts`
  - Mount `Login.vue` with a router stub pointing at the test PocketBase
  - Submit valid credentials → `router.currentRoute` is `/user/profile`
  - Submit invalid credentials → error `<p>` visible; no navigation
  - Tests will fail until T28 is complete
  - _Outcome_: failing tests defining expected Login behaviour

- [ ] **T26** Write integration test for `MonthlyCountReport.vue` — **before refactoring**
  - File: `tests/integration/MonthlyCountReport.spec.ts`
  - Mount `MonthlyCountReport.vue` against test PocketBase with seed data
  - Assert a `<canvas>` element is rendered after data loads
  - Mount with a repository that rejects → assert error text is visible
  - _Outcome_: failing tests

- [ ] **T27** Write integration test for `Settings.vue` — **before refactoring**
  - File: `tests/integration/Settings.spec.ts`
  - Mount `Settings.vue` as authenticated test user against test PocketBase
  - Assert `userSetting.portalTheme` value appears in the rendered output
  - _Outcome_: failing tests

- [ ] **T28** Refactor `App.vue` — Options API → `<script setup>`
  - Remove `defineComponent({ name: 'App' })` wrapper; script block can be empty
  - _Outcome_: no Options API remains

- [ ] **T29** Refactor `PublicLayout.vue`
  - Replace `getBackendClient()` check with `useAuth().isAuthenticated`
  - _Outcome_: no import from `services/backend-client`

- [ ] **T30** Refactor `UserLayout.vue`
  - Replace `getBackendClient()` auth check, record display, and logout with `useAuth()`
  - _Outcome_: no import from `services/backend-client`

- [ ] **T31** Refactor `Login.vue` — T25 integration tests pass after this
  - Replace `getBackendClient()` call with `await useAuth().login(...)`
  - Fix the async bug: navigation now runs only after `await login(...)` resolves
  - Replace `alert()` with an inline error `<p>`
  - _Outcome_: T25 integration tests green

- [ ] **T32** Refactor `RegisterUser.vue`
  - Replace `pb.collection('users').create(...)` with `authRepository.register(...)`
  - Replace `alert()` with inline error `<p>`
  - _Outcome_: no raw PocketBase calls

- [ ] **T33** Refactor `MonthlyCountReport.vue` — T26 integration tests pass after this
  - Replace inline IIFE, `createDatasets`, `Chart.register`, and PocketBase calls with `useMonthlyCountReport()`
  - Add error `<p>` driven by `error` ref; remove local `MonthlyCountRecord` interface
  - _Outcome_: T26 integration tests green; view ≤ 20 lines

- [ ] **T34** Refactor `Settings.vue` — T27 integration tests pass after this
  - Replace inline try/catch and repository call with `useUserSettings()`; call `load()` via `onMounted`
  - _Outcome_: T27 integration tests green

---

## Phase 6 — E2E tests

Dependencies: T29–T34, T5

- [ ] **T35** Write E2E test: login flow
  - File: `tests/e2e/login.e2e.spec.ts`
  - Navigate to `/login`; fill credentials; submit; assert URL becomes `/user/profile`
  - _Outcome_: full browser login flow verified on OpenBSD

- [ ] **T36** Write E2E test: monthly count report
  - File: `tests/e2e/monthly-count-report.e2e.spec.ts`
  - Login; navigate to `/monthly-count-report`; assert `canvas` element present
  - _Outcome_: chart render verified end-to-end

- [ ] **T37** Write E2E test: logout
  - From user layout, click Logout; assert redirect to `/`
  - Can be added to `login.e2e.spec.ts` as a follow-on step
  - _Outcome_: logout flow verified

---

## Phase 7 — Cleanup

Dependencies: T29–T34 all complete

- [ ] **T38** Delete `frontend/src/services/backend-client.ts`
  - Confirm zero imports of `getBackendClient` remain in `src/`
  - _Outcome_: singleton store is the sole PocketBase entry point

---

## Phase 8 — Verification

Dependencies: T35–T38

- [ ] **T39** Run `npm run lint` on the server — zero errors
- [ ] **T40** Run full test suite on the server:
  ```sh
  npm run test:unit
  npm run test:contract
  npm run test:integration
  CHROMIUM_PATH=/usr/local/bin/chromium npm run test:e2e
  ```
  All passes; no regressions
