# CLAUDE-project.md

Repository-specific guidance for SkillSurvey: what it is, how it is built, and how it is tested and deployed. General engineering standards — layered architecture, spec-driven changes, and the test-first mandate — are in `CLAUDE.md`.

Server credentials, addresses, and ready-to-run SSH commands are in `CLAUDE.local.md` (gitignored — not committed).

## What is SkillSurvey

SkillSurvey tracks which technical skills appear in Australian job listings (Seek, Jora) and exposes monthly trend data through a web UI.

## Architecture

```
pocketbaseserver/   PocketBase BaaS — auth, collections, REST API, serves frontend
runtask/            Scheduled task runner — scrape → report → housekeeping
frontend/           Vue 3 + TypeScript SPA (served from pocketbaseserver/pb_public/)
```

**Data flow:**
```
Seek / Jora  →  runtask scrape  →  jobPosts in PocketBase
                                ↓
             runtask report  →  monthlyCountReports in PocketBase
                                ↓
                         frontend (reads via PocketBase SDK)
```

**Go workspace** (`go.work`) includes `./pocketbaseserver`, `./runtask`.

### pocketbaseserver

All data lives in PocketBase collections. Schema changes go in `pocketbaseserver/migrations/` — never edit the database directly. Migrations run automatically on `serve`.

**Roles and permissions:**

| Role | Can read | Can write |
|---|---|---|
| `webscraper` | sites, jobPosts | jobPosts |
| `reporting` | jobPosts, skillTypes, skillNames, skillNameAliases | monthlyCountReports |
| `migration` | all (except users/roles/userRoles) | all (except users/roles/userRoles) |

Self-registration on `users` is disabled — superadmin creates accounts. Service account `runtask@skillsurvey.com` holds `webscraper` + `reporting` roles.

### runtask

Four commands dispatched from `cmd/runtask/main.go`:

| Command | Description |
|---|---|
| `runtask scrape` | Crawl Seek + Jora; upsert jobPosts into PocketBase |
| `runtask report` | Count alias occurrences in jobPosts; write monthlyCountReports |
| `runtask housekeeping cleanfs` | Remove Chromium temp dirs under `/tmp` |
| `runtask housekeeping sendlog` | Email `ErrorLogFile` via SMTP PlainAuth; truncate log |

Config loaded from `runtask.json` next to the binary (no env vars). Call `exception.Init(cfg.ErrorLogFile)` once at startup before any `exception.LogErrorWithLabel` / `LogExtraData` / `ReportErrorIfPanic` calls — the logger is nil until initialised.

When using `chromedp.Nodes`, always pass `chromedp.AtLeast(0)` — without it the call blocks forever when a selector is not found.

### frontend

Vue 3 SPA built to `frontend/dist/`, then copied to `pocketbaseserver/pb_public/`. Uses the PocketBase JS SDK for all data and auth. Base URL from `VUE_APP_POCKETBASE_URL` (`.env`).

## Development rules

- Schema changes via migrations only — never edit the PocketBase database directly.
- Format Go code with `gofmt` and `goimports` before committing.
- Run `npm run lint` in `frontend/` to check TypeScript/Vue style.
- Code style: [Google Go style guide](https://google.github.io/styleguide/go/guide), [Google TypeScript style guide](https://google.github.io/styleguide/tsguide.html), [Vue style guide](https://vuejs.org/style-guide/) (Composition API).

## Testing

### Where tests run

**All tests run on the OpenBSD server.** Push changes, pull on the server, run tests there (see `CLAUDE.local.md` for connection details).

Frontend unit tests are not required; Go unit tests are not required where an integration test covers the same behaviour.

### Test types

| Type | Scope | When required |
|---|---|---|
| **Unit** | Single function or package in isolation (no external dependencies) | When logic is complex enough to warrant isolated testing — written first |
| **Integration (Go)** | Real PocketBase instance + real stubs (SMTP, httptest) | Always for non-trivial Go features — written first |
| **Integration (frontend)** | API-connected Vue components exercised against PocketBase | Required when adding new API-connected UI features |
| **Contract** | PocketBase collection rules verified via HTTP (status codes, auth) | When adding or changing collections, roles, or access rules |

### Go tests

Every integration test starts a real PocketBase HTTP server using a `t.TempDir()` data directory — no database mocking. Stubs for external services use real TCP/HTTP:
- **SMTP**: `net.Listen("tcp", "127.0.0.1:0")` stub that speaks the SMTP protocol
- **HTTP scrape targets**: `httptest.NewServer` serving canned HTML/JSON

Import `_ "keybook/pocketbaseserver/migrations"` in any test binary that needs the full schema (including roles and seed data) applied automatically.

Place test files alongside the Go source (`*_test.go`).

Run from the repo root on the server:

```sh
go test ./...                                                              # all modules
go test ./runtask/internal/housekeeping/ -v -timeout 60s                   # specific package
go test ./runtask/internal/scrape/ -run TestScrapeRunCreatesJobPosts -v    # single test
go test ./pocketbaseserver/migrations/ -v -timeout 120s                    # migration + RBAC tests
```

## Build commands

Each module has a `Makefile`. Run from the module directory.

```sh
# pocketbaseserver
make build      # → pocketbaseserver/build/pocketbaseserver
make run_dev    # build + serve --dev (auto-migrate, verbose)

# runtask
make build      # → runtask/build/runtask
make run_dev    # build + run
```

## Frontend commands

Run from `frontend/`:

```sh
npm install
npm run serve   # dev server (hot-reload)
npm run build   # production build → dist/
npm run lint    # ESLint + style check
```

After building, copy `dist/` contents to `pocketbaseserver/pb_public/` to deploy the frontend.

## Deploying to the OpenBSD server

**Key facts (non-sensitive):**
- Connect with the OpenSSH client using key auth: `ssh -i ~/.ssh/openbsd_key junying@192.168.8.145`.
- The SSH address and the address services listen on differ, but both reach the same machine.
- `runtask.json` lives next to the runtask binary in `build/`.
- The **production** PocketBase runs as user `skillsurvey` from its own checkout. Deployment steps below act on the `junying` checkout only and must not stop, restart, or rebuild the production instance.

### Deployment steps

1. Push: `git push`
2. On server: `git fetch --all && git checkout --detach origin/<branch-name>`
3. On server: `cd <module>/ && make build` (repeat for each changed module)
4. Do **not** `pkill pocketbaseserver` — that would target the production instance. Restarting production is a deliberate manual act performed as `skillsurvey`.

Run tests on the server the same way — push, pull on the server, then use the commands under [Go tests](#go-tests).
