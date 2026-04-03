# frontend — Codebase Index

## Purpose

Single-page application for managing skills/skill types and viewing monthly skill-demand reports. Uses PocketBase for auth and a separate Go REST API (`:3000`) for skill data.

## Tech stack

- Vue 3.5 (Composition API) + TypeScript 5.8
- Vue Router 4 (hash/history routing)
- IBM Carbon Design System (`@carbon/vue` 3.0) + Bootstrap Vue 3 0.5
- Chart.js 4 + vue-chart-3 (line chart for reports)
- PocketBase JS SDK 0.26 (auth + userSettings collection)
- Build: `@vue/cli-service` + Vite-compatible Babel
- Target: Modern browsers (ESNext output)

## Directory map

```
frontend/
├── public/
│   └── index.html              # SPA entry; mounts #app
├── src/
│   ├── main.ts                 # createApp, router, Carbon + BootstrapVue3 plugins
│   ├── App.vue                 # Root: <router-view> only
│   ├── assets/logo.png
│   ├── schemas/
│   │   ├── skills.ts           # SkillType, SkillName, SkillNameAlias interfaces
│   │   └── users.ts            # IUserSettings interface
│   ├── services/
│   │   ├── backend-client.ts   # getBackendClient() → PocketBase instance (cookie auth store)
│   │   └── arrays.ts           # sortByProperty<T>() generic sort helper
│   ├── styles/
│   │   └── _carbon.scss        # Carbon component SCSS imports
│   ├── layouts/
│   │   ├── PublicLayout.vue    # Sidebar nav; redirects authenticated users → /user/profile
│   │   └── UserLayout.vue      # Top bar with logout; redirects unauthenticated → /
│   ├── components/
│   │   ├── SkillView.vue       # v-model form: name, type dropdown, alias list (add/delete)
│   │   └── SkillTypeView.vue   # v-model form: name + description
│   └── views/
│       ├── public/
│       │   ├── HomeRoute.vue            # Redirects → /login
│       │   ├── Login.vue                # Email/password → PocketBase authWithPassword()
│       │   ├── RegisterUser.vue         # Name/email/password registration
│       │   ├── MonthlyCountReport.vue   # Line chart: monthly count per skill type (12 mo)
│       │   ├── SkillList.vue            # Table: all skills + aliases; link → edit
│       │   ├── SkillAdd.vue             # POST /skill/add via SkillView component
│       │   ├── SkillEdit.vue            # GET /skill/getbyid → form; save/delete with confirm modal
│       │   ├── SkillTypeList.vue        # Table: skill types + skill count; link → edit
│       │   ├── SkillTypeAdd.vue         # POST /skilltype/add via SkillTypeView component
│       │   └── SkillTypeEdit.vue        # GET /skilltype/getbyid → form + skills sidebar; save/delete
│       └── user/
│           ├── Profile.vue              # Placeholder
│           └── Settings.vue            # Load/create userSettings in PocketBase; theme select
├── .env                        # VUE_APP_POCKETBASE_URL=http://192.168.8.147:8090
├── .env.local                  # Local overrides (same key)
├── babel.config.js
├── tsconfig.json               # strict, ESNext, path alias @/* → src/*
└── package.json
```

## Routing

Defined in `src/main.ts`.

### Public routes (wrapped in `PublicLayout`)

| Path | Component | Notes |
|---|---|---|
| `/` | HomeRoute | Redirects → `/login` |
| `/login` | Login | PocketBase auth |
| `/register` | RegisterUser | PocketBase user create |
| `/monthly-count-report` | MonthlyCountReport | Chart.js line chart |
| `/skill-list` | SkillList | GET :3000/skill/getall |
| `/skill-add/:skilltypeid?` | SkillAdd | POST :3000/skill/add |
| `/skill-edit/:skillid` | SkillEdit | GET+POST :3000/skill/* |
| `/skill-type-list` | SkillTypeList | GET :3000/skilltype/getall |
| `/skill-type-add` | SkillTypeAdd | POST :3000/skilltype/add |
| `/skill-type-edit/:skilltypeid` | SkillTypeEdit | GET+POST :3000/skilltype/* |

### User routes (wrapped in `UserLayout`, require auth)

| Path | Component |
|---|---|
| `/user/profile` | Profile (placeholder) |
| `/user/settings` | Settings (theme) |

## API calls

### Backend REST API (`:3000`, old `backend/results`)
Hardcoded base URL `http://localhost:3000` in each component. Raw `fetch()` calls, no shared client.

| Endpoint | Used by |
|---|---|
| GET `/skill/getall` | SkillList |
| GET `/skill/getbyid?skillid=` | SkillEdit |
| POST `/skill/add` | SkillAdd |
| POST `/skill/save` | SkillEdit |
| DELETE `/skill/delete` | SkillEdit |
| GET `/skilltype/getall` | SkillTypeList |
| GET `/skilltype/getbyid?skilltypeid=` | SkillTypeEdit |
| GET `/skilltype/getallidandname` | SkillView (dropdown) |
| POST `/skilltype/add` | SkillTypeAdd |
| POST `/skilltype/save` | SkillTypeEdit |
| DELETE `/skilltype/delete` | SkillTypeEdit |
| GET `/report/getmonthlycount` | MonthlyCountReport |

### PocketBase SDK (`:8090`)
Via `getBackendClient()` in `services/backend-client.ts`.

| Operation | Used by |
|---|---|
| `authWithPassword()` | Login |
| `create()` on `users` | RegisterUser |
| GET/create `userSettings` | Settings |
| `authStore.isValid` / `clear()` | Layout guards, logout |

## State management

No Vuex/Pinia. State is:
- **Component-local** `reactive()` / `ref()` variables
- **Props + emits** between `SkillView`/`SkillTypeView` and parent pages
- **PocketBase `authStore`** for global auth state (stored in cookies)

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `VUE_APP_POCKETBASE_URL` | `http://192.168.8.147:8090` | PocketBase server URL |

The skill API base URL (`http://localhost:3000`) is hardcoded in components — there is no env var for it.

## Build & run

```bash
npm install          # install dependencies

npm run serve        # dev server (usually :8080, hot-reload)
npm run build        # production build → dist/
npm run lint         # ESLint check
```

To deploy: copy `dist/` contents to `pocketbaseserver/pb_public/`.

**Prerequisites for dev:**
- PocketBase running at `VUE_APP_POCKETBASE_URL`
- `backend/results` running at `:3000`
