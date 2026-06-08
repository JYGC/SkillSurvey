# Design: Auth Gate, Sidebar Move, and Bootstrap-to-Carbon Migration

## System Architecture

### Route changes

The `monthly-count-report` route moves from the public route tree to the user route tree:

**Before:**
```
/ (PublicLayout)
  /monthly-count-report  → views/public/MonthlyCountReport.vue
  /login
  /register

/user (UserLayout)
  /user/profile
  /user/settings
```

**After:**
```
/ (PublicLayout)
  /login
  /register

/user (UserLayout)
  /user/monthly-count-report  → views/user/MonthlyCountReport.vue  (moved)
  /user/profile
  /user/settings
```

`UserLayout` auth guard already protects all `/user/*` routes by redirecting to `/` when unauthenticated, so moving the route there is sufficient to auth-gate it.

Post-login redirect changes from `/user/profile` to `/user/monthly-count-report` in:
- `PublicLayout.vue` (redirect for already-authenticated users)
- `Login.vue` (redirect after successful login)

### Layout changes

**PublicLayout.vue** — sidebar is removed entirely. The layout becomes a plain pass-through wrapper:
```
<div class="public-layout">
  <router-view />
</div>
```
The existing auth redirect (send authenticated users away from public routes) is kept.

**UserLayout.vue** — receives the sidebar, rebuilt with Carbon UI Shell:
```
CvHeader
  CvHeaderName  (app name)
  #header-global slot
    user email (span)
    CvButton[kind=ghost] "Logout"
CvSideNav[fixed]
  CvSideNavItems
    CvSideNavLink[to=user-monthly-count-report]
    CvSideNavLink[to=user-settings]
CvContent
  router-view
```

`CvSideNav` with `fixed` prop renders as an always-visible, non-collapsing sidebar. `CvContent` automatically provides the correct left offset when a fixed side nav is present.

### Bootstrap removal

All Bootstrap imports are removed from `main.ts`:
```ts
// Remove:
import BootstrapVue3 from 'bootstrap-vue-3';
import 'bootstrap/dist/css/bootstrap.css';
import 'bootstrap-vue-3/dist/bootstrap-vue-3.css';
app.use(BootstrapVue3);
```

Bootstrap packages removed from `package.json`:
- `bootstrap`
- `bootstrap-vue-3`
- `@popperjs/core` (Bootstrap tooltip/dropdown dependency, unused by Carbon)

`PublicLayout.vue` currently uses `<b-nav>` / `<b-nav-item>` — these are replaced with the simplified layout above (no nav at all in public layout).

### App.vue style cleanup

`App.vue` contains Bootstrap-specific nav styles targeting `.nav`, `.nav-div`, `.btn`, and Bootstrap grid classes (`.row`, `#app > .row`). These are removed. The Carbon UI Shell's own CSS handles all navigation styling.

Retained in `App.vue`:
- `@import './styles/carbon'` — Carbon SCSS import
- Basic `#app` font/anti-aliasing rules
- `table` and `tbody` scroll styles (used by other views)
- `.fill-parent`, `.vertical-padding` utility classes

## Component file moves

| Old path | New path |
|---|---|
| `src/views/public/MonthlyCountReport.vue` | `src/views/user/MonthlyCountReport.vue` |

`main.ts` import path updates accordingly.

## Carbon components used

| Element | Component |
|---|---|
| Top header bar | `CvHeader`, `CvHeaderName` |
| Sidebar | `CvSideNav`, `CvSideNavItems`, `CvSideNavLink` |
| Main content area | `CvContent` |
| Logout button | `CvButton` (kind="ghost") |
| Login/register buttons | `CvButton` (existing, unchanged) |
| Form inputs | `CvTextInput`, `CvFluidForm` (existing, unchanged) |
| Links | `CvLink` (existing, unchanged) |
| Charts | `CcvLineChart` (existing, unchanged) |

## PocketBase migration

A new migration restricts `monthlyCountReports` list and view access to authenticated users, matching the frontend auth gate at the API layer.

**File:** `pocketbaseserver/migrations/1780963200_monthly_count_reports_auth_required.go`

| Direction | listRule | viewRule |
|---|---|---|
| Up | `@request.auth.id != ""` | `@request.auth.id != ""` |
| Down | `""` (public — restores the previous migration's setting) | `""` |

The `runtask report` service account authenticates via the `reporting` role before writing, so it satisfies `@request.auth.id != ""` on any read it also performs. Write rules are untouched.

`skillNames` public read (`1780185600`) is left as-is — skill names are non-sensitive reference data and the expand on `monthlyCountReports` will succeed because the requesting user is now always authenticated.

## Error handling

The `UserLayout` auth failure message (`Failure to get authenticate user`) is retained for the case where `isAuthenticated` is false inside the user layout (defensive check, per existing behaviour).

## Testing strategy

No new automated tests are required for this change. The changes are routing configuration and layout composition. Existing integration and e2e tests that verify:
- Login redirects to user area
- Logout redirects to `/`
- `data-testid="logout-btn"` on the logout button
- Monthly count report renders chart data

...remain valid and must continue to pass. Run on the OpenBSD server after deployment.
