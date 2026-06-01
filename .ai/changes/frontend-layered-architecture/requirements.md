# Requirements: Frontend Layered Architecture

## Store layer — PocketBase singleton

### Single client instance
WHEN any module imports the PocketBase client THE SYSTEM SHALL return the same singleton instance throughout the application lifetime.
WHEN the singleton is created THE SYSTEM SHALL initialise it with `VUE_APP_POCKETBASE_URL`, load auth from cookie, and register an `onChange` handler that writes back to cookie.

## Repository layer — collection access

### Monthly count report repository
WHEN the monthly-count-report repository is called THE SYSTEM SHALL fetch all `monthlyCountReports` records expanded with `skillName`, sorted by `yearMonthDate`, using the PocketBase store.
WHEN the fetch succeeds THE SYSTEM SHALL return typed `MonthlyCountRecord[]`.
WHEN the fetch fails THE SYSTEM SHALL propagate the error to the caller.

### Auth repository
WHEN the auth repository login method is called with email and password THE SYSTEM SHALL authenticate against the `users` collection.
WHEN authentication succeeds THE SYSTEM SHALL return the auth record.
WHEN authentication fails THE SYSTEM SHALL propagate the error to the caller.
WHEN the auth repository logout method is called THE SYSTEM SHALL clear the PocketBase auth store.
WHEN the auth repository `isAuthenticated` property is read THE SYSTEM SHALL return the current `authStore.isValid` value.

### User settings repository
WHEN the user settings repository is called with a user ID THE SYSTEM SHALL fetch the matching `userSettings` record.
WHEN no settings record exists THE SYSTEM SHALL create a default record (`portalTheme: "white"`) and return it.
WHEN a settings record exists THE SYSTEM SHALL return it without creating a duplicate.
WHEN the fetch or create fails THE SYSTEM SHALL propagate the error to the caller.

## Service layer — business logic

### Monthly count report service
WHEN the service receives a list of `MonthlyCountRecord` values THE SYSTEM SHALL extract the most recent 12 distinct `YearMonth` values.
WHEN the service builds chart datasets THE SYSTEM SHALL group records by skill name and produce a flat array of `{ group, date, value }` data points; missing months SHALL have a value of `0`.

## Composable layer — use-case orchestration

### useAuth composable
WHEN `useAuth` is called THE SYSTEM SHALL return reactive `isAuthenticated` state, a `login(email, password)` async function, and a `logout()` function.
WHEN `login` is awaited and succeeds THE SYSTEM SHALL update `isAuthenticated` to `true`.
WHEN `login` is awaited and fails THE SYSTEM SHALL expose the error for the caller to handle.
WHEN `logout` is called THE SYSTEM SHALL call the auth repository and update `isAuthenticated` to `false`.

### useMonthlyCountReport composable
WHEN `useMonthlyCountReport` is called THE SYSTEM SHALL return reactive `chartData` (flat `CarbonChartDataPoint[]`), `chartOptions` (axis configuration object), and an `error` ref.
WHEN the composable initialises THE SYSTEM SHALL fetch records via the repository, transform them via the service, and populate `chartData`.
WHEN the fetch or transform fails THE SYSTEM SHALL set `error` and leave `chartData` as an empty array.

### useUserSettings composable
WHEN `useUserSettings` is called THE SYSTEM SHALL return a reactive `userSetting` ref and a `load()` async function.
WHEN `load()` is called and the auth store has no current user THE SYSTEM SHALL return without fetching.
WHEN `load()` is called and the auth store has a current user THE SYSTEM SHALL fetch settings via the repository and populate `userSetting`.

## Component / View layer — thin components

### App.vue
WHEN `App.vue` is rendered THE SYSTEM SHALL use `<script setup>` (Composition API) with no Options API `defineComponent` wrapper.

### Login.vue
WHEN the login form is submitted THE SYSTEM SHALL await the `useAuth` login function before navigating.
WHEN login fails THE SYSTEM SHALL display the error to the user.

### RegisterUser.vue
WHEN the registration form is submitted THE SYSTEM SHALL call the PocketBase `users` collection create via a repository method.
WHEN registration succeeds THE SYSTEM SHALL navigate to `/`.
WHEN registration fails THE SYSTEM SHALL display the error to the user.

### MonthlyCountReport.vue
WHEN the component is mounted THE SYSTEM SHALL delegate all data fetching and transformation to `useMonthlyCountReport`.
WHEN `useMonthlyCountReport` reports an error THE SYSTEM SHALL display it to the user.

### Settings.vue
WHEN the component is mounted THE SYSTEM SHALL delegate settings loading to `useUserSettings`.

### PublicLayout.vue
WHEN the layout is mounted and the user is authenticated THE SYSTEM SHALL redirect to `/user/profile` via `useAuth`.

### UserLayout.vue
WHEN the layout is mounted and the user is not authenticated THE SYSTEM SHALL redirect to `/` via `useAuth`.
WHEN the logout button is clicked THE SYSTEM SHALL call `useAuth` logout then navigate to `/`.

## Chart library — Carbon Charts

### Chart rendering
WHEN the monthly count report is rendered THE SYSTEM SHALL use `CcvLineChart` from `@carbon/charts-vue` instead of `chart.js` and `vue-chart-3`.
WHEN the frontend is built THE SYSTEM SHALL NOT depend on `chart.js` or `vue-chart-3`.

## Deleted / removed code

### getBackendClient factory removed
WHEN any code previously importing `getBackendClient` is refactored THE SYSTEM SHALL import the singleton store instead.
WHEN `backend-client.ts` no longer contains any exports consumed by views or composables THE SYSTEM SHALL be replaced by the singleton store module.

## Testing

### Test infrastructure
WHEN tests run on OpenBSD THE SYSTEM SHALL start a dedicated pocketbaseserver process using a temporary data directory so each run is isolated.
WHEN the test run completes THE SYSTEM SHALL stop the pocketbaseserver process and remove the temporary directory.
WHEN E2E tests run on OpenBSD THE SYSTEM SHALL locate the Chromium binary via the `CHROMIUM_PATH` environment variable rather than downloading a browser.

### Unit tests — services
WHEN `getRecentMonths` is called with records spanning more than 12 months THE SYSTEM SHALL return only the last 12 distinct `YearMonth` values.
WHEN `getRecentMonths` is called with records spanning fewer than 12 months THE SYSTEM SHALL return all distinct `YearMonth` values.
WHEN `getRecentMonths` is called with an empty array THE SYSTEM SHALL return an empty array.
WHEN `buildChartDatasets` is called THE SYSTEM SHALL return one `CarbonChartDataPoint` per skill per month within the provided month window.
WHEN `buildChartDatasets` is called with records that omit a month for a skill THE SYSTEM SHALL include a data point for that month with `value: 0`.
WHEN `buildChartDatasets` is called THE SYSTEM SHALL set `group` to the skill name and `date` to the `YearMonth` string.

### Unit tests — composables (repositories mocked)
WHEN `useAuth` is called and `authRepository.isAuthenticated` is false THE SYSTEM SHALL expose `isAuthenticated` as false.
WHEN `useAuth.login` is awaited and the repository resolves THE SYSTEM SHALL not throw.
WHEN `useAuth.login` is awaited and the repository rejects THE SYSTEM SHALL propagate the error to the caller.
WHEN `useAuth.logout` is called THE SYSTEM SHALL delegate to `authRepository.logout`.
WHEN `useMonthlyCountReport` initialises and the repository resolves THE SYSTEM SHALL populate `chartData` with a non-empty `CarbonChartDataPoint[]`.
WHEN `useMonthlyCountReport` initialises and the repository rejects THE SYSTEM SHALL set `error` and leave `chartData` as an empty array.
WHEN `useUserSettings.load` is called with no authenticated user THE SYSTEM SHALL leave `userSetting` as null.
WHEN `useUserSettings.load` is called with an authenticated user THE SYSTEM SHALL call the repository and set `userSetting`.

### Contract tests — PocketBase access rules
WHEN an unauthenticated request fetches `monthlyCountReports` THE SYSTEM SHALL return HTTP 403.
WHEN an authenticated user fetches `monthlyCountReports` THE SYSTEM SHALL return HTTP 200.
WHEN an unauthenticated request attempts to create a `users` record (self-registration) THE SYSTEM SHALL return HTTP 400.
WHEN a user fetches another user's `userSettings` record THE SYSTEM SHALL return HTTP 403.
WHEN a user fetches their own `userSettings` record THE SYSTEM SHALL return HTTP 200.

### Integration tests — component mounting
WHEN `Login.vue` is mounted and the form is submitted with valid credentials THE SYSTEM SHALL navigate to `/user/profile`.
WHEN `Login.vue` is mounted and the form is submitted with invalid credentials THE SYSTEM SHALL display an error message without navigating.
WHEN `MonthlyCountReport.vue` is mounted and the repository returns records THE SYSTEM SHALL render an `<svg>` element.
WHEN `MonthlyCountReport.vue` is mounted and the repository rejects THE SYSTEM SHALL render the error message text.
WHEN `Settings.vue` is mounted with an authenticated user THE SYSTEM SHALL display the user's `portalTheme`.

### E2E tests — full user flows
WHEN a user navigates to `/login` and submits valid credentials THE SYSTEM SHALL redirect to the user layout page.
WHEN a user navigates to `/monthly-count-report` THE SYSTEM SHALL render a chart SVG element.
WHEN a logged-in user clicks Logout THE SYSTEM SHALL redirect to `/`.
