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
WHEN the service builds chart datasets THE SYSTEM SHALL group records by skill name, fill missing months with zero, and assign a random hex border colour per skill with `hidden: true`.

## Composable layer — use-case orchestration

### useAuth composable
WHEN `useAuth` is called THE SYSTEM SHALL return reactive `isAuthenticated` state, a `login(email, password)` async function, and a `logout()` function.
WHEN `login` is awaited and succeeds THE SYSTEM SHALL update `isAuthenticated` to `true`.
WHEN `login` is awaited and fails THE SYSTEM SHALL expose the error for the caller to handle.
WHEN `logout` is called THE SYSTEM SHALL call the auth repository and update `isAuthenticated` to `false`.

### useMonthlyCountReport composable
WHEN `useMonthlyCountReport` is called THE SYSTEM SHALL return reactive `chartData` (labels + datasets), `chartHeight`, and an `error` ref.
WHEN the composable initialises THE SYSTEM SHALL fetch records via the repository, transform them via the service, and populate `chartData`.
WHEN the fetch or transform fails THE SYSTEM SHALL set `error` and leave `chartData` in its empty initial state.

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

## Deleted / removed code

### getBackendClient factory removed
WHEN any code previously importing `getBackendClient` is refactored THE SYSTEM SHALL import the singleton store instead.
WHEN `backend-client.ts` no longer contains any exports consumed by views or composables THE SYSTEM SHALL be replaced by the singleton store module.
