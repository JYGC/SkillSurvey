# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is SkillSurvey

SkillSurvey tracks which technical skills appear in Australian job listings (Seek, Jora) and exposes monthly trend data through a web UI.

## Architecture

```
pocketbaseserver/   PocketBase BaaS — auth, collections, REST API, serves frontend
runtask/            Scheduled task runner — scrape → report → housekeeping
migrate/            One-shot legacy SQLite → PocketBase migration tool (temporary)
frontend/           Vue 3 + TypeScript SPA (served from pocketbaseserver/pb_public/)
backend/            Legacy GORM/SQLite stack — superseded, not in go.work
```

**Data flow:**
```
Seek / Jora  →  runtask scrape  →  jobPosts in PocketBase
                                ↓
             runtask report  →  monthlyCountReports in PocketBase
                                ↓
                         frontend (reads via PocketBase SDK)
```

**Go workspace** (`go.work`) includes `./migrate`, `./pocketbaseserver`, `./runtask`. The `backend/` module is standalone and excluded.

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

### migrate

One-shot tool that reads the legacy SQLite DB (via GORM) and writes into PocketBase. Migrates in dependency order: sites → skillTypes → skillNames → skillNameAliases → jobPosts → monthlyCountReports. All steps are idempotent (deduplication by natural key). The service account must have the `migration` role.

### frontend

Vue 3 SPA built to `frontend/dist/`, then copied to `pocketbaseserver/pb_public/`. Uses the PocketBase JS SDK for all data and auth. Base URL from `VUE_APP_POCKETBASE_URL` (`.env`).

## Changes (spec-driven work)

Non-trivial features and bug fixes are tracked as a **change** — a folder at `.ai/changes/<change-name>/` containing up to three spec files. Use specs for anything complex, costly to get wrong, or requiring iterative design. Skip specs for exploratory/prototype work.

### Spec files

**`requirements.md`** — the *what*. Organise by feature area (H2) and user story group (H3). Each requirement uses EARS notation:

```
WHEN <condition> THE SYSTEM SHALL <action>
```

Example:
```
## Report Generation

### Monthly count report
WHEN runtask report runs THE SYSTEM SHALL create one monthlyCountReport per skill per month found in jobPosts.
WHEN a report for a skill+month already exists THE SYSTEM SHALL update its count rather than create a duplicate.
```

Also cover edge cases and error-handling scenarios.

**`design.md`** — the *how*. Sections: system architecture and components, sequence diagrams, data models and interfaces, error-handling approach, testing strategy.

**`tasks.md`** — the *steps*. Discrete, trackable implementation tasks, each with a clear description, expected outcome, and any dependencies. Mark tasks required vs optional. Work through independent tasks first, then dependent ones in order.

**`bugfix.md`** — replaces `requirements.md` for bug fixes. Three sections using their own notation:

```
## Current Behavior (Defect)
WHEN <condition> THEN the system <incorrect behavior>

## Expected Behavior (Correct)
WHEN <condition> THEN the system SHALL <correct behavior>

## Unchanged Behavior (Regression Prevention)
WHEN <condition> THEN the system SHALL CONTINUE TO <existing behavior>
```

The "Unchanged Behavior" section is the key addition — explicitly locking down what must not change prevents regressions. The `design.md` for a bugfix includes root cause analysis; `tasks.md` includes tests that verify the bug is fixed and unchanged behavior is preserved.

### Workflow

**Feature:**
1. Create `requirements.md` and agree on it before writing `design.md`.
2. Create `design.md` and agree on it before writing `tasks.md`.
3. Execute `tasks.md` one task at a time, marking each done as you go.

**Bugfix:**
1. Create `bugfix.md` (current / expected / unchanged behavior) and agree on it.
2. Create `design.md` including root cause analysis.
3. Create and execute `tasks.md`, including tests for fix and regression prevention.

Before starting any non-trivial feature, refactor, or bug fix, check `.ai/changes/` for an existing change folder. If none exists, create one and start with `requirements.md` (feature) or `bugfix.md` (bug).

## Development rules

- Always read a file before editing it.
- Schema changes via migrations only — never edit the PocketBase database directly.
- No speculative abstractions — only build what is needed now.
- Format Go code with `gofmt` and `goimports` before committing.
- Run `npm run lint` in `frontend/` to check TypeScript/Vue style.
- Code style: [Google Go style guide](https://google.github.io/styleguide/go/guide), [Google TypeScript style guide](https://google.github.io/styleguide/tsguide.html), [Vue style guide](https://vuejs.org/style-guide/) (Composition API).

## Testing

### Mandate

**All tests run on the OpenBSD server — not on Windows.** Push changes, pull on the server, run tests there.

**Integration tests must be written before the implementation code they cover.** Write the test, watch it fail, then write the minimum code to make it pass. Frontend unit tests are not required; Go unit tests are not required where an integration test covers the same behaviour.

### Test types

| Type | Scope | When required |
|---|---|---|
| **Integration (Go)** | Real PocketBase instance + real stubs (SMTP, httptest) | Always for non-trivial Go features — written first |
| **Integration (frontend)** | API-connected Vue components exercised against PocketBase | Required when adding new API-connected UI features |
| **Contract** | PocketBase collection rules verified via HTTP (status codes, auth) | When adding or changing collections, roles, or access rules |

### Go tests

Every integration test starts a real PocketBase HTTP server using a `t.TempDir()` data directory — no database mocking. Stubs for external services use real TCP/HTTP:
- **SMTP**: `net.Listen("tcp", "127.0.0.1:0")` stub that speaks the SMTP protocol
- **HTTP scrape targets**: `httptest.NewServer` serving canned HTML/JSON

Import `_ "keybook/pocketbaseserver/migrations"` in any test binary that needs the full schema (including roles and seed data) applied automatically.

```sh
go test ./...                                                              # all tests in a module
go test ./runtask/internal/housekeeping/ -v -timeout 60s                   # specific package
go test ./runtask/internal/scrape/ -run TestScrapeRunCreatesJobPosts -v    # single test
go test ./pocketbaseserver/migrations/ -v -timeout 120s                    # migration + RBAC tests
```

Place test files alongside the Go source (`*_test.go`).

### In specs

`tasks.md` must include explicit test tasks. For features: the integration test task comes first, before any implementation task. For bugfixes: tasks must include a test that reproduces the bug before it is fixed, and regression tests drawn from `bugfix.md`'s "Unchanged Behavior" section.

## Build commands

Each module has a `Makefile`. Run from the module directory.

```sh
# pocketbaseserver
make build      # → pocketbaseserver/build/pocketbaseserver
make run_dev    # build + serve --dev (auto-migrate, verbose)

# runtask
make build      # → runtask/build/runtask
make run_dev    # build + run

# migrate
make build      # → migrate/build/migrate
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

Server connection details, credentials, and ready-to-run SSH commands are in `CLAUDE.local.md` (gitignored — not committed).

**Key facts (non-sensitive):**
- `sshpass` is installed in Cygwin at `/c/cygwin64/bin/sshpass` — **not** on the Git Bash PATH; always use the full path alongside Cygwin's `ssh`/`sftp`.
- PocketBase listens on a **different network interface** from the SSH host — both are on the same machine.
- PocketBase is started manually (not via rc.d); config files (`runtask.json`, `migrate.json`) live next to each binary in `build/`.
- SSL verification is disabled for `git push` on Windows (`git -c http.sslVerify=false push`).

### Deployment steps

1. Push from Windows: `git -c http.sslVerify=false push`
2. On server: `git fetch --all && git checkout origin/<branch-name>`
3. On server: `cd <module>/ && make build` (repeat for each changed module)
4. Restart pocketbaseserver if changed: `pkill pocketbaseserver`, then start with `nohup ... &` and capture the printed PID

### Running tests on the server

Push, pull on the server, then run from the repo root:

```sh
go test ./...                                                         # all modules
go test ./runtask/internal/housekeeping/ -v -timeout 60s             # specific package
go test ./runtask/internal/scrape/ -run TestScrapeRunCreatesJobPosts  # specific test
```
