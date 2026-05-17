# frontend — Spec

## Goals

- Provide a single-page application for managing skills and viewing monthly skill-demand reports.
- Support user authentication and per-user settings via PocketBase.
- Remain usable during the backend migration (currently calls both `:3000` and `:8090`).

## Functional requirements

### Authentication
- Users can register with name, email, and password.
- Users can log in and log out.
- Unauthenticated users are redirected to `/login`; authenticated users are redirected away from public-only routes.
- Auth state is stored in PocketBase's cookie-based `authStore`.

### Skill management
- List all skill types with their skill count.
- Create, edit, and delete skill types (name + description).
- List all skills with their aliases.
- Create, edit, and delete skills (name, skill type, one or more aliases).
- Deleting a skill type must be blocked if it has associated skills (enforced by the API).

### Reports
- Display a line chart of monthly skill-demand counts for the past 12 months, grouped by skill type.
- Data is fetched from the backend REST API (`/report/getmonthlycount`).

### User settings
- Authenticated users can select a portal theme: `white`, `g10`, `g90`, or `g100`.
- Settings are persisted in the `userSettings` PocketBase collection (one record per user).

## Non-functional requirements

- Must run as a static SPA served from `pocketbaseserver/pb_public/`.
- PocketBase URL must be configurable via `VUE_APP_POCKETBASE_URL` (not hardcoded).
- The skill API base URL (currently `http://localhost:3000`) must be configurable via an environment variable — it must not remain hardcoded in components.
- Must target modern browsers (ESNext output); no IE11 support required.
- ESLint must pass with no errors before a build is considered releasable.

## Constraints

- No global state library (Vuex/Pinia); use component-local `reactive()`/`ref()` and PocketBase `authStore`.
- Use IBM Carbon Design System and Bootstrap Vue 3 for UI components; do not introduce additional UI libraries.

## Out of scope

- Server-side rendering.
- Mobile-native builds.
