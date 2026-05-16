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

Non-trivial features and bug fixes are tracked as a **change** — a folder at `.ai/changes/<change-name>/`. Use specs for anything complex or costly to get wrong.

**Spec files:**

- `requirements.md` — the *what*, using EARS notation: `WHEN <condition> THE SYSTEM SHALL <action>`
- `design.md` — the *how*: architecture, sequence diagrams, data models, error handling, test strategy
- `tasks.md` — discrete implementation tasks with expected outcomes and dependencies
- `bugfix.md` — replaces `requirements.md` for bugs; three sections: Current Behavior (Defect) / Expected Behavior (Correct) / Unchanged Behavior (Regression Prevention)

**Feature workflow:** requirements → design → tasks, each agreed before the next is written.  
**Bugfix workflow:** bugfix.md → design (with root cause) → tasks (includes regression tests).

Before starting non-trivial work, check `.ai/changes/` for an existing change folder. If none exists, create one.

## Development rules

- Always read a file before editing it.
- Schema changes via migrations only — never edit the PocketBase database directly.
- No speculative abstractions — only build what is needed now.
- Format Go code with `gofmt` and `goimports` before committing.
- Run `npm run lint` in `frontend/` to check TypeScript/Vue style.
- Code style: [Google Go style guide](https://google.github.io/styleguide/go/guide), [Google TypeScript style guide](https://google.github.io/styleguide/tsguide.html), [Vue style guide](https://vuejs.org/style-guide/) (Composition API).

## Testing

- **All Go tests run on the OpenBSD server** — not on Windows. Push changes, pull on server, run there.
- Integration tests use a real PocketBase instance (`t.TempDir()` data dir) started in-process — no mocking.
- Write integration tests before implementation code.
- `pocketbaseserver/migrations` tests start a full PocketBase server with all migrations applied; SMTP tests use a real TCP stub listener on a random port.
- Importing `_ "keybook/pocketbaseserver/migrations"` in a test pulls all migrations into the test binary.

```sh
# Run all tests in a module (on OpenBSD server)
go test ./...

# Run a specific package
go test ./runtask/internal/housekeeping/ -v

# Run a specific test
go test ./runtask/internal/scrape/ -run TestScrapeRunCreatesJobPosts -v -timeout 120s
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
