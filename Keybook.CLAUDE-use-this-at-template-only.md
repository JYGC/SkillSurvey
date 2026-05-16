# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Changes (spec-driven work)

Non-trivial features and bug fixes are tracked as a **change** — a folder at `.claude/changes/<change-name>/` containing up to three spec files. Use specs for anything complex, costly to get wrong, or requiring iterative design. Skip specs for exploratory/prototype work.

### Spec files

**`requirements.md`** — the *what*. Organise by feature area (H2) and user story group (H3). Each requirement uses EARS notation:

```
WHEN <condition> THE SYSTEM SHALL <action>
```

Example:
```
## Device Management

### Add device
WHEN a user submits a valid new-device form THE SYSTEM SHALL create the device record and record a creation history entry.
WHEN a user submits a device name that already exists THE SYSTEM SHALL display an "Name already taken" error.
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

Before starting any non-trivial feature, refactor, or bug fix, check `.claude/changes/` for an existing change folder. If none exists, create one and start with `requirements.md` (feature) or `bugfix.md` (bug).

## What is KeyBook

KeyBook is a web app for managing devices, persons, and properties, with automatic audit history for all changes. The frontend is a SvelteKit static site; the backend is a Go binary that embeds PocketBase (SQLite + REST API). The compiled frontend is served by the backend from `backend/internal/frontend/build/`.

## Architecture

```
frontend/          SvelteKit (Svelte 5, TypeScript, Carbon Design System)
backend/
  cmd/keybook.go   Entry point — wires DI container, registers PocketBase hooks
  internal/
    repositories/  Data access layer (PocketBase DAO queries)
    services/      Business logic; history-tracking services called by hooks
    dtos/          Data transfer objects for all entities
    helpers/       PocketBase DAO error utilities
    frontend/build/ Gitignored — populated from frontend build output
database/
  pb_schema.json   PocketBase collection definitions
```

### Key patterns

**Frontend:** Business logic lives in `.svelte.ts` module files under `src/lib/modules/` (one per entity: device, person, property, persondevice, user). These implement typed interfaces and use Svelte 5 reactive primitives (`$state`, `$derived.by`). Shared state is distributed via Svelte's context API (set in the user layout, consumed via `getContext()`), not stores.

**Backend:** Dependency injection via `go.uber.org/dig`. Repository pattern for data access; service layer for business logic. Audit history is recorded automatically — PocketBase `OnModelAfterCreate` and `OnModelBeforeUpdate` hooks call history services for every entity type.

**Frontend → backend:** The PocketBase JS SDK (`pocketbase` npm package) is the only HTTP client. Base URL comes from the `PUBLIC_POCKETBASE_URL` env var. Auth state is persisted in cookies via `src/lib/api/backend-client.ts`. The user layout (`src/routes/user/+layout.ts`) guards all `/user/*` routes and redirects to `/auth` if unauthenticated.

## Testing

### Mandate

**Unit tests must be written before the implementation code they cover (TDD).** Write the test, watch it fail, then write the minimum code to make it pass.

### Test types

| Type | Scope | When required |
|---|---|---|
| **Unit** | Single function, module, or component in isolation | Always — written first |
| **Integration** | Multiple components or layers working together (e.g. repository + service, component + store) | Always for non-trivial interactions |
| **E2E** | Full user flow through the running app via browser | Always for user-facing features |
| **Contract** | API shape between frontend and backend (request/response structure) | When adding or changing PocketBase collection endpoints |

### Frontend tests

| Command | Tool | What it covers |
|---|---|---|
| `npm run test:unit` | Vitest | Unit and integration tests in `src/` |
| `npm run test:integration` | Playwright | E2E flows against the running app |

Place unit/integration test files alongside the code they test (`*.test.ts` or `*.spec.ts`). E2E tests live in `tests/`.

### Backend tests

```sh
go test ./...               # all tests
go test ./internal/...      # specific package tree
```

Place test files alongside the Go source (`*_test.go`). Use table-driven tests for repository and service logic.

### In specs

`tasks.md` must include explicit test tasks. For features: unit tests as the first task for each component, followed by integration and E2E tasks. For bugfixes: tasks must include a test that reproduces the bug before it is fixed, and regression tests drawn from `bugfix.md`'s "Unchanged Behavior" section.

## Frontend commands

Run from `frontend/`:

```sh
npm run dev          # dev server — address in CLAUDE.local.md
npm run build        # production build → frontend/build/
npm run check        # svelte-check + TypeScript
npm run lint         # prettier + eslint (check only)
npm run format       # prettier --write
npm run test:unit    # vitest
npm run test:integration  # playwright
```

Code style: tabs, single quotes, 100-char line width (Prettier). TypeScript strict mode.

## Backend commands

Run from `backend/`:

```sh
go build -o build/keybook ./cmd/     # compile
./build/keybook serve --dev --http <address>  # address in CLAUDE.local.md
go test ./...                         # tests
```

## Deploying to the OpenBSD server

Server addresses, credentials, and deployment steps are in `CLAUDE.local.md` (gitignored).

@CLAUDE.local.md
