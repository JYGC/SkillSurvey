# SkillSurvey — System Overview

SkillSurvey tracks which technical skills appear in Australian job listings (Seek, Jora) and exposes the trend data through a web UI.

## Sub-projects

| Project | Purpose | Status |
|---|---|---|
| `pocketbaseserver/` | PocketBase BaaS — auth, DB, REST API | Active |
| `runtask/` | Scheduled task runner — scrape, report, housekeeping | Active |
| `migrate/` | One-shot legacy SQLite → PocketBase migration tool | Active (tool) |
| `frontend/` | Vue 3 + TypeScript SPA | Active |
| `backend/` | Old Go stack: GORM/SQLite scrapers + REST API | Legacy (standalone, not in go.work) |

## Go workspace

`go.work` includes three modules:

```
go 1.24.1

use (
    ./migrate
    ./pocketbaseserver
    ./runtask
)
```

`backend/` has its own `go.mod` but is excluded from the workspace.

## Data flow

```
Job Sites (Seek / Jora)
       │
       ▼  runtask scrape
pocketbaseserver  ─── SQLite (pb_data/) ──► REST API (:8090)
       │
       ▼  runtask report
monthlyCountReports in PocketBase
       │
       ▼  served via pb_public/
   frontend  ◄──────────────────────────────────────────
       │  PocketBase JS SDK + fetch() calls to :8090
       ▼
    Browser
```

## Port map

| Service | Port | Notes |
|---|---|---|
| `pocketbaseserver` | 8090 | All data, auth, and static frontend |
| `backend/results` (legacy) | 3000 | No longer the primary data source |

## Cross-cutting concerns

- **OS target**: OpenBSD (all binaries must build/run on OpenBSD)
- **Secrets**: Never commit API keys or passwords
- **Tests**: Integration tests use real PocketBase + real SMTP stubs; no mocking
- **Style**: Google Go style guide; Google TypeScript style guide
- **Config**: JSON files placed next to each binary (no env vars for runtime config)
