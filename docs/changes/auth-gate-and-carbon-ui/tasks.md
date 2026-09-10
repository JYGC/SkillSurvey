# Tasks: Auth Gate, Sidebar Move, and Bootstrap-to-Carbon Migration

## Task list

- [x] **Task 1 — Move MonthlyCountReport component** _(Required)_

  Copy `src/views/public/MonthlyCountReport.vue` to `src/views/user/MonthlyCountReport.vue`, then delete the original.

  Expected outcome: file exists at the new path, old path is gone.

---

- [x] **Task 2 — Update router in `main.ts`** _(Required; depends on Task 1)_

  - Change the import of `MonthlyCountReport` to point to `./views/user/MonthlyCountReport.vue`
  - Remove `monthly-count-report` from the `PublicLayout` children
  - Add `{ path: 'monthly-count-report', name: 'user-monthly-count-report', component: MonthlyCountReport }` to the `UserLayout` children
  - Remove Bootstrap imports: `BootstrapVue3`, `bootstrap/dist/css/bootstrap.css`, `bootstrap-vue-3/dist/bootstrap-vue-3.css`, and `app.use(BootstrapVue3)`

  Expected outcome: the route `/user/monthly-count-report` is protected by `UserLayout`'s auth guard; Bootstrap is no longer loaded.

---

- [x] **Task 3 — Update post-login redirect targets** _(Required)_

  - In `PublicLayout.vue`: change `router.push('/user/profile')` → `router.push('/user/monthly-count-report')`
  - In `Login.vue`: change `router.push('/user/profile')` → `router.push('/user/monthly-count-report')`

  Expected outcome: after login (or when an authenticated user hits a public page) they land on the report.

---

- [x] **Task 4 — Simplify `PublicLayout.vue`** _(Required; depends on Task 3)_

  Replace the entire template and script with a minimal pass-through layout:
  - Remove `<b-nav>` / `<b-nav-item>` sidebar
  - Remove Bootstrap grid (`row`, `col-md-2`, `col-md-10`)
  - Keep only a plain wrapper div, `<router-view />`, and the auth-redirect script block

  Expected outcome: login and register pages render without a sidebar; no Bootstrap component references remain.

---

- [x] **Task 5 — Rewrite `UserLayout.vue` with Carbon UI Shell** _(Required; depends on Task 2)_

  Replace the current layout with Carbon UI Shell components:
  - `CvHeader` + `CvHeaderName` for the top bar
  - User email and `CvButton[kind=ghost]` Logout button in the `#header-global` slot; keep `data-testid="logout-btn"` on the logout button
  - `CvSideNav[fixed]` + `CvSideNavItems` for the sidebar
  - `CvSideNavLink` (with `to` prop) for "Monthly count report" and "Settings"
  - `CvContent` wrapping `<router-view />`
  - Retain the auth guard (`if (!isAuthenticated.value) router.push('/')`)
  - Retain the `v-if="!isAuthenticated"` error message

  Expected outcome: authenticated pages display a fixed Carbon sidebar and Carbon header; all Bootstrap classes are gone.

---

- [x] **Task 6 — Clean up `App.vue` styles** _(Required; depends on Tasks 4 and 5)_

  Remove Bootstrap-specific CSS from `App.vue`:
  - Delete rules targeting `.nav`, `.nav > .association`, `.nav > .new-association`
  - Delete rules targeting `.nav-div`, `.nav-div .dropdown-menu`, `.nav-div .nav-item`
  - Delete Bootstrap button overrides: `.btn`, `.btn-danger`, `.expand button`, `.expand .nav-link`
  - Delete `#app > .row { height: 100vh; width: 100vw; }` (Bootstrap grid rule)
  - Delete `.margin-left-10` (unused)
  - Retain: `@import './styles/carbon'`, `#app` font rules, `table`/`tbody` scroll rules, `.fill-parent`, `.vertical-padding`

  Expected outcome: no Bootstrap selector names remain; Carbon and custom utility styles are intact.

---

- [x] **Task 7 — Remove Bootstrap from `package.json`** _(Required)_

  Remove from `dependencies`:
  - `bootstrap`
  - `bootstrap-vue-3`
  - `@popperjs/core`

  Expected outcome: `package.json` no longer lists Bootstrap packages.

---

- [x] **Task 8 — Run `npm install`** _(Required; depends on Task 7)_

  Run `npm install` from `frontend/` to update `package-lock.json` and remove Bootstrap packages from `node_modules`.

  Expected outcome: lock file updated; `bootstrap` and `bootstrap-vue-3` absent from `node_modules`.

---

- [x] **Task 9 — Write contract test for `monthlyCountReports` auth rule** _(Required — written before Task 10)_

  Create `pocketbaseserver/migrations/1780963200_monthly_count_reports_auth_required_test.go` with two test cases:
  - Unauthenticated list: seed a record, expect HTTP 200 with 0 items (PocketBase filter rules hide data but do not return 403)
  - Authenticated list: seed a record, expect HTTP 200 with 1+ items

  Also update `TestMonthlyCountReportExpandSkillNameUnauthenticated` → `TestMonthlyCountReportExpandSkillName` in `1780185600_skill_names_public_read_test.go`: changed unauthenticated `http.Get` to authenticated request.

  Run the new tests on the OpenBSD server — they must **fail** at this point (migration not yet written).

  Expected outcome: contract tests exist and fail; updated existing test still passes.

---

- [x] **Task 10 — Add PocketBase migration: restrict `monthlyCountReports` to authenticated users** _(Required; depends on Task 9)_

  Create `pocketbaseserver/migrations/1780963200_monthly_count_reports_auth_required.go`.

  - Up: set `listRule` and `viewRule` to `@request.auth.id != ""`
  - Down: restore `listRule` and `viewRule` to `""` (public — the value set by the preceding `1780099200` migration)

  Run the contract tests on the OpenBSD server — they must now **pass**.

  Expected outcome: unauthenticated callers see 0 items from `monthlyCountReports`; authenticated callers see all items.
