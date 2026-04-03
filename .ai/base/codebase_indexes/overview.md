# SkillSurvey — System Overview

SkillSurvey tracks which technical skills appear in Australian job listings (Seek, Jora) and exposes the trend data through a web UI. The system has three sub-projects:

| Project | Purpose | Status |
|---|---|---|
| `backend/` | Old Go stack: scrapers + GORM/SQLite REST API | Legacy (being replaced) |
| `pocketbaseserver/` | New Go stack: PocketBase BaaS, auth, DB | Active |
| `frontend/` | Vue 3 + TypeScript SPA | Active (shared by both stacks) |

## Data flow (new stack)

```
Job Sites (Seek / Jora)
       │
       ▼ [runtask scraper — not in this repo yet]
pocketbaseserver  ─── SQLite (pb_data/) ──► REST API (:8090)
       │
       ▼ serves pb_public/
   frontend  ◄─────────────────────────────────────────────
       │ fetch() calls to :3000 (old) and PocketBase SDK (:8090)
       ▼
    Browser
```

## Port map

| Service | Port | Notes |
|---|---|---|
| `backend/results` (old REST API) | 3000 | Still called by frontend for skills/reports |
| `pocketbaseserver` | 8090 | Auth, userSettings, future collections |

## Cross-cutting concerns

- **OS target**: OpenBSD (all binaries must build/run on OpenBSD)
- **Secrets**: Never commit API keys or passwords
- **Tests**: All new features require integration tests; no DB mocking
- **Style**: Google Go style guide; Google TypeScript style guide
