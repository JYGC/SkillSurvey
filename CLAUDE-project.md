# CLAUDE-project.md

Repository-specific guidance for SkillSurvey: what it is, how it is built, and how it is tested and deployed. General engineering standards — layered architecture, spec-driven changes, and the test-first mandate — are in `CLAUDE.md`.

Server credentials, addresses, and ready-to-run SSH commands are in `CLAUDE.local.md` (gitignored — not committed).

## What is SkillSurvey

SkillSurvey tracks which technical skills appear in Australian job listings (Seek, Jora) and exposes monthly trend data through a web UI.

## Architecture

```
pocketbaseserver/   PocketBase BaaS — auth, collections, REST API
runtask/            Scheduled task runner — scrape → report → housekeeping
frontend/           Vue 3 + TypeScript SPA (served by httpd on hosts, from pocketbaseserver/pb_public/ in E2E)
deploy/             Bundle build script, idempotent install scripts for OpenBSD hosts, and their tests
```

**Data flow:**
```
Seek / Jora  →  runtask scrape  →  jobPosts in PocketBase
                                ↓
             runtask report  →  monthlyCountReports in PocketBase
                                ↓
                         frontend (reads via PocketBase SDK)
```

**Go workspace** (`go.work`) includes `./deploy`, `./pocketbaseserver`, `./runtask`.

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

Five commands dispatched from `cmd/runtask/main.go`:

| Command | Description |
|---|---|
| `runtask scrape` | Crawl Seek + Jora; upsert jobPosts into PocketBase |
| `runtask report` | Count alias occurrences in jobPosts; write monthlyCountReports |
| `runtask housekeeping cleanfs` | Remove Chromium temp dirs under `/tmp` |
| `runtask housekeeping sendlog` | Email `ErrorLogFile` via SMTP PlainAuth; truncate log |
| `runtask check-setup` | Check that both site files load, the service account can sign in and read `sites`, and Chromium starts on a blank page. Runs every check, prints one line each, and exits 1 if any failed. Writes nothing and scrapes nothing |

Config loaded from `runtask.json` next to the binary (no env vars); malformed values (bad `PocketBaseUrl`, out-of-range `SmtpPort`, unparsable email addresses, a negative or overflow-causing `DynamicContentExtractionTimeoutSeconds`) are rejected at startup, naming every failing field. `DynamicContentExtractionTimeoutSeconds` governs chromedp's per-page-load timeout for both adapters and defaults to 60 seconds when absent or `0`. Call `exception.Init(cfg.ErrorLogFile)` once at startup before any `exception.LogErrorWithLabel` / `LogExtraData` / `ReportErrorIfPanic` calls — the logger is nil until initialised.

When using `chromedp.Nodes`, always pass `chromedp.AtLeast(0)` — without it the call blocks forever when a selector is not found.

### frontend

Vue 3 SPA built to `frontend/dist/`. Uses the PocketBase JS SDK for all data and auth. Dev and E2E builds take the base URL from `VUE_APP_POCKETBASE_URL` (`.env`, or `build:e2e`'s own value).

`npm run build:hosted` leaves `VUE_APP_POCKETBASE_URL` empty, so one build works on any host. At startup, `src/store/host-config.ts` fetches `/host-config.json` (`{"PocketBaseUrl": "http://<host>:8090"}`) uncached, and points the SDK at that address before the router starts. The file is written per host and is never part of a build. If it is missing or has no address, the page shows `SkillSurvey cannot start: its settings could not be read.` instead of a blank screen.

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
go test ./pocketbaseserver/... ./runtask/...                               # app and runtask modules (./... fails at the go.work root)
go test ./deploy/...                                                       # install script tests, run on their own: alongside the others they slow PocketBase and Chromium start-up enough to fail tests
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
npm run build   # build → dist/, with the address from .env
npm run build:hosted  # build → dist/, reading the address from /host-config.json at startup
npm run lint    # ESLint + style check
```

Hosts serve a `build:hosted` build through `httpd`. `build:e2e` writes into `pocketbaseserver/pb_public/`, which PocketBase serves for E2E only.

## Deploying

Every deployment is manual. CI builds two bundles as its last step on every passing run on `master` (`build-bundles` in `ci-steps.json`, which runs `deploy/build-bundles.sh`), and the operator copies them to each host and runs the install script as root. **`deploy/README.md` is the operator guide**: getting bundles, first installs, the steps left by hand, the schedule, upgrading and going back.

**Key facts:**
- A bundle is one commit, built from `git archive` with `go build -trimpath -buildvcs=false` and `npm run build:hosted`, so the same commit gives byte-identical programs wherever it is built. A plain `make build` does not, and would make every commit look like a new `pocketbaseserver` to the install script.
- The server bundle holds `pocketbaseserver`, the frontend and `install-server.sh`. The runtask bundle holds `runtask`, `install-runtask.sh` and example settings holding `REPLACE_ME`. Neither holds a real setting.
- The install scripts are idempotent, refuse before changing anything when a check fails, never overwrite a settings file and never remove `pb_data`. Their tests in `deploy/*_test.go` run each script against a throwaway root through `DESTDIR`, with stubs for the commands that need root. They need OpenBSD's `sha256`, so they skip elsewhere.
- Nothing under `deploy/` may name a real host or a private address, because the repository is public. `deploy/publicrepo_test.go` enforces it, and the README uses `<host>` placeholders.

### The current production server

Until production moves to the new hosts, it still runs on the old OpenBSD server as user `skillsurvey`, from that account's own checkout. Connection details are in `CLAUDE.local.md`. Work on that server acts on your own checkout only and must not stop, restart or rebuild the production instance. Do **not** `pkill pocketbaseserver`, which would target it. Restarting production there is a deliberate manual act performed as `skillsurvey`.

To try a branch on that server: push, then `git fetch --all && git checkout --detach origin/<branch-name>` in your checkout, then `make build` in each changed module. Run tests the same way, using the commands under [Go tests](#go-tests).
