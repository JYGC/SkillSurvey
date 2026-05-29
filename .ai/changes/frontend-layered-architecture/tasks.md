# Tasks: Frontend Layered Architecture

Rule: every unit of implementation is preceded by its test. Mark each `[x]` as you complete it.

---

## Phase 0 — Foundation (no tests; these are types and wiring, no logic)

- [x] **T1** Create `frontend/src/store/pocketbase.ts`
  - Module-level PocketBase singleton: init with `VUE_APP_POCKETBASE_URL`, load cookie, register onChange
  - Default-export the `pb` instance
  - _Outcome_: any file can `import pb from '@/store/pocketbase'` and get the same instance

- [x] **T2** Create `frontend/src/schemas/monthly-count-report.ts`
  - Move `MonthlyCountRecord` interface out of `MonthlyCountReport.vue`; export it
  - _Outcome_: `MonthlyCountRecord` importable from `@/schemas/monthly-count-report`

---

## Phase 1 — Test infrastructure (dependencies: T1)

- [x] **T3** Install test dev dependencies
  - Add to `frontend/package.json` devDependencies: `vitest`, `@vitest/coverage-v8`, `@vue/test-utils`, `happy-dom`, `@wdio/cli`, `@wdio/local-runner`, `@wdio/mocha-framework`, `@wdio/spec-reporter`, `@wdio/types`, `wdio-chromedriver-service`, `webdriverio`
  - Note: `@playwright/test` replaced by WebdriverIO — Playwright throws `Unsupported platform: openbsd` at module init and cannot be used on the OpenBSD server
  - Run `npm install` on the server; do **not** run `playwright install` — system Chrome + chromedriver are used
  - _Outcome_: `vitest` and `wdio` CLIs available

- [x] **T4** Create `frontend/vitest.config.ts`
  - `environment: 'happy-dom'`; `resolve.alias` `@` → `src/`; `globalSetup` and `setupFiles` pointing at Phase 1 files
  - _Outcome_: `npm run test:unit` resolves `@/` imports and runs in happy-dom

- [x] **T5** Create `frontend/wdio.conf.ts`
  - Single Chrome capability using system Chrome (`CHROMIUM_PATH`, default `/usr/local/bin/chrome`) and system chromedriver (`CHROMEDRIVER_PATH`, default `/usr/local/bin/chromedriver`)
  - `--headless`, `--no-sandbox`, `--disable-dev-shm-usage` Chrome args; Mocha framework; `baseUrl` from `TEST_E2E_URL`
  - _Outcome_: `npm run test:e2e` can launch Chrome on OpenBSD

- [x] **T6** Create `frontend/tests/setup/vitest.global-setup.ts`
  - `setup()`: spawn `pocketbaseserver/build/pocketbaseserver serve --http 127.0.0.1:18090 --dir <tmpdir>`, poll `/api/health`, call `seedInitialData()`
  - `teardown()`: kill process, remove tmpdir
  - _Outcome_: contract and integration tests get a clean PocketBase instance per run

- [x] **T7** Create `frontend/tests/setup/seed.ts`
  - `seedInitialData(baseUrl)`: create superadmin via `POST /api/collections/_superusers/records`, authenticate, create test user, insert seed `monthlyCountReports` and `userSettings` records
  - Write test credentials to `process.env.TEST_USER_EMAIL` / `TEST_USER_PASSWORD`
  - _Outcome_: integration and E2E tests can authenticate and find test data

- [x] **T8** Create `frontend/tests/setup/vitest.setup.ts`
  - `beforeEach`: clear `pb.authStore`; override `pb` base URL to `TEST_PB_URL` when set
  - _Outcome_: tests cannot share login state across files

- [x] **T9** Add npm test scripts to `frontend/package.json`
  - `"test:unit"`, `"test:contract"`, `"test:integration"`, `"test:e2e"`, `"test"`
  - _Outcome_: each scope runnable independently on the server

---

## Phase 2 — Services (unit tests before implementation)

Dependencies: T2, T4

- [x] **T10** Write unit tests for `monthly-count-report.service.ts` — **tests first**
  - File: `tests/unit/services/monthly-count-report.service.spec.ts`
  - `getRecentMonths`: empty → `[]`; < 12 months → all; > 12 months → last 12 only
  - `buildChartDatasets`: one dataset per skill; missing months → `0`; every dataset has `hidden: true`
  - Imports will fail to resolve until T11 — that is expected (red phase)
  - _Outcome_: failing tests that define the service contract

- [x] **T11** Implement `frontend/src/services/monthly-count-report.service.ts`
  - `getRecentMonths` and `buildChartDatasets` — pure functions, no store/repository imports
  - _Outcome_: T10 tests pass (green)

- [x] **T12** Write unit tests for `frontend/src/services/arrays.ts` — **tests first**
  - File: `tests/unit/services/arrays.spec.ts`
  - `sortByProperty`: sorts ascending by string; by number; returns mutated array
  - `arrays.ts` already exists; tests are written against its current interface
  - _Outcome_: `arrays.ts` covered; any future change caught

---

## Phase 3 — Repositories (unit tests + contract tests before each implementation)

Dependencies: T1, T2, T6, T7, T8

### Auth repository

- [x] **T13** Write unit tests for `auth.repository.ts` — **tests first**
  - File: `tests/unit/repositories/auth.repository.spec.ts`
  - Mock `@/store/pocketbase` with `vi.mock`
  - `isAuthenticated` returns `pb.authStore.isValid`
  - `login(email, pw)` calls `pb.collection('users').authWithPassword(email, pw)`
  - `register(...)` calls `pb.collection('users').create(...)`
  - `logout()` calls `pb.authStore.clear()`
  - _Outcome_: failing tests

- [x] **T14** Write contract tests for auth — **tests first**
  - File: `tests/contract/auth.contract.spec.ts`
  - Unauthenticated `POST /api/collections/users/records` → HTTP 400 (self-registration disabled)
  - Valid `POST /api/collections/users/auth-with-password` → HTTP 200 with token
  - Invalid credentials → HTTP 400
  - _Outcome_: PocketBase access rules verified at HTTP layer (failing until real server available)

- [x] **T15** Implement `frontend/src/repositories/auth.repository.ts`
  - `isAuthenticated` getter, `currentUser` getter, `login`, `register`, `logout`
  - _Outcome_: T13 unit tests pass; T14 contract tests pass

### Monthly count report repository

- [x] **T16** Write unit tests for `monthly-count-report.repository.ts` — **tests first**
  - File: `tests/unit/repositories/monthly-count-report.repository.spec.ts`
  - Mock `@/store/pocketbase`
  - `getAll()` calls `pb.collection('monthlyCountReports').getFullList` with `expand: 'skillName'` and `sort: 'yearMonthDate'`
  - _Outcome_: failing tests

- [x] **T17** Write contract tests for `monthlyCountReports` — **tests first**
  - File: `tests/contract/monthly-count-report.contract.spec.ts`
  - Unauthenticated fetch → HTTP 200 (public collection — `MonthlyCountReport` is a public route)
  - Authenticated fetch → HTTP 200
  - Note: `monthlyCountReports` had nil list/view rules (superadmin-only) in the initial migration.
    Added `pocketbaseserver/migrations/1780099200_monthly_count_reports_public_read.go` to set
    `ListRule = ""` and `ViewRule = ""` so the public frontend route can read it.
  - _Outcome_: access rule verified at HTTP layer

- [x] **T18** Implement `frontend/src/repositories/monthly-count-report.repository.ts`
  - `getAll()` → `Promise<MonthlyCountRecord[]>`
  - _Outcome_: T16 unit tests pass; T17 contract tests pass

### User settings repository

- [x] **T19** Write unit tests for `user-settings.repository.ts` — **tests first**
  - File: `tests/unit/repositories/user-settings.repository.spec.ts`
  - Mock `@/store/pocketbase`
  - `getOrCreate(userId)`: when `getFirstListItem` resolves → returns the record
  - `getOrCreate(userId)`: when `getFirstListItem` throws "wasn't found" → calls `create` with defaults and returns result
  - `getOrCreate(userId)`: when `getFirstListItem` throws another error → rethrows without calling `create`
  - _Outcome_: failing tests (get-or-create logic explicitly covered)

- [x] **T20** Write contract tests for `userSettings` — **tests first**
  - File: `tests/contract/user-settings.contract.spec.ts`
  - Authenticated user fetches own record → HTTP 200
  - Authenticated user fetches another user's record → HTTP 403
  - Fetch when no record exists → HTTP 404
  - _Outcome_: per-user isolation rule verified at HTTP layer

- [x] **T21** Implement `frontend/src/repositories/user-settings.repository.ts`
  - `getOrCreate(userId)` with internal "not found" handling
  - _Outcome_: T19 unit tests pass; T20 contract tests pass

---

## Phase 4 — Composables (unit tests before implementation)

Dependencies: T15, T18, T21, T4

- [x] **T22** Write unit tests for `use-auth.ts` — **tests first**
  - File: `tests/unit/composables/use-auth.spec.ts`
  - `vi.mock('@/repositories/auth.repository')`
  - `isAuthenticated` mirrors `authRepository.isAuthenticated`
  - `login` delegates to `authRepository.login`; propagates rejection
  - `logout` delegates to `authRepository.logout`
  - _Outcome_: failing tests

- [x] **T23** Implement `frontend/src/composables/use-auth.ts`
  - `isAuthenticated` and `currentUser` as `computed`; `login` async; `logout` delegates
  - _Outcome_: T22 tests pass

- [x] **T24** Write unit tests for `use-monthly-count-report.ts` — **tests first**
  - File: `tests/unit/composables/use-monthly-count-report.spec.ts`
  - `vi.mock` repository and service modules
  - On resolve: `chartData.labels` and `chartData.datasets` populated; `error` is null
  - On reject: `error` set; `chartData` remains empty
  - _Outcome_: failing tests

- [x] **T25** Implement `frontend/src/composables/use-monthly-count-report.ts`
  - `Chart.register(...registerables)` at module level; returns `{ chartData, chartHeight, error }`; calls `load()` on init
  - _Outcome_: T24 tests pass

- [x] **T26** Write unit tests for `use-user-settings.ts` — **tests first**
  - File: `tests/unit/composables/use-user-settings.spec.ts`
  - `vi.mock` auth repository and user-settings repository
  - `load()` with no current user → `userSetting` stays null; repository not called
  - `load()` with current user → repository called; `userSetting` set to result
  - _Outcome_: failing tests

- [x] **T27** Implement `frontend/src/composables/use-user-settings.ts`
  - Returns `{ userSetting, load }`; `load()` guards on `currentUser`
  - _Outcome_: T26 tests pass

---

## Phase 5 — Views (integration tests before each refactor)

Dependencies: T23, T25, T27, T6, T7, T8

- [x] **T28** Write integration test for `PublicLayout.vue` — **tests first**
  - File: `tests/integration/PublicLayout.spec.ts`
  - Mount with unauthenticated state → no redirect; nav links rendered
  - Mount with authenticated state → `router.push('/user/profile')` called
  - _Outcome_: failing tests (current component calls `getBackendClient()` which will not resolve correctly in test environment)

- [x] **T29** Refactor `PublicLayout.vue`
  - Replace `getBackendClient()` check with `useAuth().isAuthenticated`
  - _Outcome_: T28 tests pass; no import from `services/backend-client`

- [x] **T30** Write integration test for `UserLayout.vue` — **tests first**
  - File: `tests/integration/UserLayout.spec.ts`
  - Mount with unauthenticated state → redirect to `/`
  - Mount with authenticated state → displays `currentUser` data; Logout button present
  - Click Logout → `authRepository.logout` called; redirect to `/`
  - _Outcome_: failing tests

- [x] **T31** Refactor `UserLayout.vue`
  - Replace `getBackendClient()` calls with `useAuth()`; drive display from `currentUser`
  - _Outcome_: T30 tests pass; no import from `services/backend-client`

- [x] **T32** Refactor `App.vue` — Options API → `<script setup>`
  - Remove `defineComponent({ name: 'App' })` wrapper; script block can be empty or omitted
  - No test required: zero logic, cannot meaningfully fail
  - _Outcome_: no Options API remains in the codebase

- [x] **T33** Write integration test for `Login.vue` — **tests first**
  - File: `tests/integration/Login.spec.ts`
  - Submit valid credentials → `router.currentRoute` is `/user/profile`
  - Submit invalid credentials → error `<p>` visible; no navigation
  - _Outcome_: failing tests (current component has async bug and calls `getBackendClient()`)

- [x] **T34** Refactor `Login.vue`
  - Replace `getBackendClient()` with `await useAuth().login(...)`; fix async bug (navigation after await); replace `alert()` with inline error `<p>`
  - _Outcome_: T33 tests pass

- [x] **T35** Write integration test for `RegisterUser.vue` — **tests first**
  - File: `tests/integration/RegisterUser.spec.ts`
  - Submit valid form → `authRepository.register` called; router navigates to `/`
  - Password mismatch → `authRepository.register` not called; error message shown
  - Submit fails → error message shown; no navigation
  - _Outcome_: failing tests

- [x] **T36** Refactor `RegisterUser.vue`
  - Replace `pb.collection('users').create(...)` with `authRepository.register(...)`; replace `alert()` with inline error `<p>`
  - _Outcome_: T35 tests pass; no raw PocketBase calls

- [x] **T37** Write integration test for `MonthlyCountReport.vue` — **tests first**
  - File: `tests/integration/MonthlyCountReport.spec.ts`
  - Mount against test PocketBase with seed data → `<canvas>` element rendered after data loads
  - Mount with repository that rejects → error text visible
  - _Outcome_: failing tests

- [x] **T38** Refactor `MonthlyCountReport.vue`
  - Replace inline IIFE, `createDatasets`, `Chart.register`, and PocketBase calls with `useMonthlyCountReport()`; add error `<p>`; remove local `MonthlyCountRecord` interface
  - _Outcome_: T37 tests pass; view ≤ 20 lines

- [x] **T39** Write integration test for `Settings.vue` — **tests first**
  - File: `tests/integration/Settings.spec.ts`
  - Mount as authenticated test user against test PocketBase → `userSetting.portalTheme` appears in rendered output
  - _Outcome_: failing tests

- [x] **T40** Refactor `Settings.vue`
  - Replace inline try/catch and repository call with `useUserSettings()`; call `load()` via `onMounted`
  - _Outcome_: T39 tests pass; no raw PocketBase calls in view

---

## Phase 6 — E2E tests (written after views are stable)

Dependencies: T29, T31, T34, T36, T38, T40, T5

- [x] **T41** Write E2E test: login flow
  - File: `tests/e2e/login.e2e.spec.ts`
  - Navigate to `/login`; fill valid credentials; submit; assert URL is `/user/profile`
  - _Outcome_: full browser login flow verified on OpenBSD

- [x] **T42** Write E2E test: monthly count report
  - File: `tests/e2e/monthly-count-report.e2e.spec.ts`
  - Login; navigate to `/monthly-count-report`; assert `canvas` element is present
  - _Outcome_: chart render verified end-to-end

- [x] **T43** Write E2E test: logout
  - From user layout, click Logout; assert redirect to `/`
  - Can be appended to `login.e2e.spec.ts` as a continuation step
  - _Outcome_: logout flow verified

---

## Phase 7 — Cleanup (dependencies: T29–T40 all complete)

- [x] **T44** Delete `frontend/src/services/backend-client.ts`
  - Confirm zero `getBackendClient` imports remain in `src/`
  - _Outcome_: store singleton is the sole PocketBase entry point

---

## Phase 8 — Verification (dependencies: T41–T44)

- [x] **T45** Run `npm run lint` on the server — zero errors

- [x] **T46** Run full test suite on the server:
  ```sh
  npm run test:unit
  npm run test:contract
  npm run test:integration
  CHROMIUM_PATH=/usr/local/bin/chromium npm run test:e2e
  ```
  All pass; no regressions
