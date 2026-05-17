# pocketbaseserver — Codebase Index

## Purpose

PocketBase-based backend replacing the old `backend/` stack. Provides:
- Auto-generated REST API for all collections
- RBAC (role-based access control) via PocketBase rules
- User authentication (email/password); self-registration disabled
- Serves the compiled frontend from `./pb_public`

## Tech stack

- Go 1.24.1
- PocketBase v0.29.1 (BaaS framework + embedded SQLite)
- `modernc.org/sqlite` — pure-Go SQLite driver (no CGO, OpenBSD-compatible)
- Target OS: OpenBSD

## Directory map

```
pocketbaseserver/
├── cmd/pocketbaseserver/
│   └── main.go               # Entry point: init PocketBase, register migrations,
│                             #   static file route (./pb_public), start server
├── internal/models/          # Type-safe Go wrappers (RecordProxy) for each collection
│   ├── site.go
│   ├── skilltype.go
│   ├── skillname.go
│   ├── skillnamealias.go
│   ├── jobpost.go
│   ├── monthlycountreport.go
│   └── (no userSettings model)
├── migrations/
│   ├── 1743465600_init_collections.go       # Creates all 7 custom collections
│   ├── 1743552000_add_roles.go              # RBAC: roles + userRoles collections + write rules
│   ├── 1743552000_add_roles_test.go
│   ├── 1743638400_migration_role_read_access.go  # Role-based read rules
│   ├── 1743638400_migration_role_read_access_test.go
│   ├── 1746748800_seed_runtask_user.go      # Creates runtask service account
│   ├── 1746748800_seed_runtask_user_test.go
│   ├── 1778284800_seed_sites.go             # Seeds www.seek.com.au and au.jora.com
│   └── 1778544000_seed_skill_data.go        # Seeds 22 skill types, 94 skill names, 107 aliases
├── Makefile                  # build | run_dev
├── go.mod
└── go.sum
```

## Migrations (in order)

### 1743465600_init_collections
Creates 7 custom collections in dependency order. All collections include `created` and `updated` AutodateFields.

### 1743552000_add_roles
- Creates `roles` collection (name, description, created, updated) with unique index on name
- Seeds 3 roles: `webscraper`, `reporting`, `migration`
- Creates `userRoles` junction collection (user → roles, unique index on user+role pair)
- Applies write rules to existing collections based on roles
- Sets `users.CreateRule = nil` (disables self-registration; superadmin only)

### 1743638400_migration_role_read_access
Applies read rules to restrict list/view access by role (see Permissions table below).

### 1746748800_seed_runtask_user
Creates service account `runtask@skillsurvey.com` with a random password and assigns it the `webscraper` and `reporting` roles. Idempotent.

### 1778284800_seed_sites
Seeds two site records: `www.seek.com.au` and `au.jora.com`. Idempotent (skips if name already exists).

### 1778544000_seed_skill_data
Seeds the full skill taxonomy from the legacy database:
- 22 `skillTypes` (Back End Language, Front End Framework, Database, …, AI Editors)
- 94 `skillNames` (C#, Java, Python, React, Vue.js, Docker, AWS, Firebase, Claude, ChatGPT, Cursor, Windsurf, …)
- 107 `skillNameAliases` (e.g. "JS" → JavaScript, ".NET" → .NET Framework, "TDD" → Test Driven Development)

**Note:** Temporarily sets `skillTypes.description` to `Required: false` before seeding (many legacy entries have empty descriptions). The schema remains non-required after the migration completes.

## Collections

All collections have `created` (autodate on create) and `updated` (autodate on create + update) fields.

### sites
| Field | Type | Notes |
|---|---|---|
| `name` | Text, required | e.g. "www.seek.com.au" |
| `url` | Text, required | Base URL |

### skillTypes
| Field | Type | Notes |
|---|---|---|
| `name` | Text, required | Category name |
| `description` | Text | Category description (not required after seed migration) |

### skillNames
| Field | Type | Notes |
|---|---|---|
| `name` | Text, required | Skill name |
| `isEnabled` | Bool | Include in reports |
| `skillType` | Relation → skillTypes, required | |

### skillNameAliases
| Field | Type | Notes |
|---|---|---|
| `skillName` | Relation → skillNames, required | |
| `alias` | Text, required | Alternative search term |

### jobPosts
| Field | Type | Notes |
|---|---|---|
| `jobSiteNumber` | Text, required | Job ID on source site |
| `site` | Relation → sites, required | |
| `content` | JSON, required | `{title, body}` |
| `postedDate` | Date | When posted on source site |
| `location` | JSON, required | `{city, country, suburb}` |

### monthlyCountReports
| Field | Type | Notes |
|---|---|---|
| `identifier` | Text, required | `<skillNamePBID>_<YearMonth>` |
| `YearMonth` | Text, required | YYYY-MM |
| `yearMonthDate` | Date, required | First day of period |
| `count` | Number (int) | Occurrences in that month |
| `skillName` | Relation → skillNames, optional | |

### userSettings
| Field | Type | Notes |
|---|---|---|
| `user` | Relation → users, required, unique, cascade-delete | |
| `portalTheme` | Select, required | "white" \| "g10" \| "g90" \| "g100" |

Access: `@request.auth.id != "" && user = @request.auth.id` (own record only).

### roles
| Field | Type | Notes |
|---|---|---|
| `name` | Text, required, unique | "webscraper" \| "reporting" \| "migration" |
| `description` | Text, required | |

Access: any authenticated user can list/view; only superadmin can write.

### userRoles
| Field | Type | Notes |
|---|---|---|
| `user` | Relation → users, required | |
| `role` | Relation → roles, required | |

Unique index on (user, role). Any authenticated user can list/view; only superadmin can write.

### users (system)
Standard PocketBase auth collection. Self-registration disabled (`CreateRule = nil`).

## Permissions matrix

| Collection | No role | webscraper | reporting | migration | superadmin |
|---|---|---|---|---|---|
| sites | — | list/view | — | list/view/write | all |
| skillTypes | — | — | list/view | list/view/write | all |
| skillNames | — | — | list/view | list/view/write | all |
| skillNameAliases | — | — | list/view | list/view/write | all |
| jobPosts | view | list/view/write | list/view | list/view/write | all |
| monthlyCountReports | list/view | — | list/view/write | list/view/write | all |
| users | — | — | — | — | all |
| userRoles | list/view | list/view | list/view | list/view | all |
| roles | list/view | list/view | list/view | list/view | all |
| userSettings | own record only | own record only | own record only | own record only | all |

## Service account

`runtask@skillsurvey.com` — has `webscraper` + `reporting` roles. Password set at first migration and stored in the `runtask.json` config file on the server.

## Collection dependency order

```
users (system)
├── userRoles ── roles
└── userSettings

sites
└── jobPosts

skillTypes
└── skillNames
    ├── skillNameAliases
    └── monthlyCountReports
```

## Build & run

```bash
# Development (auto-migrate, verbose errors)
make run_dev       # builds then: ./build/pocketbaseserver serve --dev

# Production build
make build         # → ./build/pocketbaseserver

# Run production binary
./build/pocketbaseserver serve --http 192.168.8.147:8090
```

Default port: **8090**
Admin dashboard: `http://localhost:8090/_/`
REST API base: `http://localhost:8090/api/`

On first launch, visit the dashboard to create a superuser account. `pb_data/` (SQLite + uploads) is git-ignored and created automatically.

## Custom routes

One custom route serves the compiled SPA:

```go
se.Router.GET("/{path...}", apis.Static(os.DirFS("./pb_public"), false))
```

All other endpoints are auto-generated by PocketBase.

## Frontend static serving

Build the frontend (`npm run build`) and copy `dist/` contents to `pocketbaseserver/pb_public/`. PocketBase serves it at `/`.
