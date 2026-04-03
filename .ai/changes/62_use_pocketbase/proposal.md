# Proposal: Migrate to PocketBase (#62)

## Problem

The current system runs three concerns in a single Go codebase (`backend/`): web scraping, report computation, and a hand-rolled REST API backed by GORM/SQLite. This creates unnecessary coupling and means all features must be deployed together. The REST API also lacks authentication and has no row-level security.

## Proposed change

Decompose `backend/` into purpose-built services:

| Concern | Old home | New home |
|---|---|---|
| Data storage + REST API | `backend/` (GORM/SQLite) | `pocketbaseserver/` (PocketBase) |
| Skill/report API handlers | `backend/internal/handlers/` | `pocketbaseserver/` (auto-generated) |
| Role-based access control | — | `pocketbaseserver/` (new `roles` and `userRoles` collections) |
| Web scraping | `backend/internal/siteadapters/` | `runtask/` (new service) |
| Report computation | `backend/cmd/reports/` | `runtask/` (new service) |
| Maintenance (cleanup, email log) | `backend/cmd/housekeeping/` | `runtask/` (new service) |
| Data migration | — | `migrate/` (temporary one-shot binary, to be deleted after use) |
| Frontend API calls | `fetch()` to `:3000` | PocketBase JS SDK to `:8090` |

`pocketbaseserver` has already been built with migrations and typed models. This proposal covers wiring it up and migrating the clients.

## Scope

### In scope
- Add `roles` and `userRoles` collections to `pocketbaseserver` via a new migration. `roles` has three seed records: `webscraper` (write access to `jobPosts`), `reporting` (write access to `monthlyCountReports`), and `migration` (write access to all collections except `users`, `userRoles`, and `roles`). A user with no `userRoles` entry is a normal login user. `runtask` authenticates with a single service account holding both the `webscraper` and `reporting` roles; `migrate` authenticates with a service account holding the `migration` role. Service account passwords are chosen by the operator during manual configuration of `runtask` and `migrate` and are never generated or stored by the migration. No passwords are hardcoded or committed to source control. After production deployment and account setup, the operator adds the passwords to the config files for `runtask` and `migrate`.
- Update remaining frontend views to call `pocketbaseserver` via the PocketBase JS SDK and auto-generated collection API instead of raw `fetch()` to `:3000`. Monthly count report data is fetched from the auto-generated `monthlyCountReports` collection endpoint.
- Remove `SkillAdd.vue`, `SkillEdit.vue`, `SkillTypeAdd.vue`, `SkillTypeEdit.vue`, `SkillList.vue`, and `SkillTypeList.vue` from the frontend — all skill and skill-type management (list, create, edit, delete) is handled via the PocketBase admin page instead.
- Create the `runtask/` module by extracting the Chromedp scraping code, report computation, and maintenance tasks from `backend/` and writing to `pocketbaseserver` via its REST API.
- Build a temporary `migrate/` binary (Go, GORM for reading the old SQLite database, [`r--w/pocketbase`](https://github.com/r--w/pocketbase) Go client for writing to `pocketbaseserver`) that reads all tables from the old `SkillSurvey.db` SQLite database and writes them to `pocketbaseserver` via its REST API. The operator manually triggers the migration and manually deletes the binary after verifying the result. Unit tests must be written before implementation.

### Out of scope
- Ongoing sync between old and new databases — migration is a one-shot operation. If migration fails or produces incorrect results, rollback is to wipe the `pocketbaseserver` data directory and re-run `migrate/`.

## Dependencies

- `pocketbaseserver` must be running and accessible before the updated frontend, `runtask`, or `migrate` can function.
- `runtask` requires a service account with the `webscraper` and `reporting` roles in `pocketbaseserver`.
- `migrate` requires a service account with the `migration` role in `pocketbaseserver`.

## Risks

| Risk | Mitigation |
|---|---|
| PocketBase string IDs vs GORM numeric IDs break frontend schema | Update `skills.ts` interfaces as part of the frontend migration |
| Alias sync (separate collection in PocketBase, embedded array in GORM) is error-prone | Centralise alias sync logic in a dedicated service module |
| Duplicate job posts if `runtask` runs more than once | Upsert by `(site, jobSiteNumber)` — check before creating |
| `runtask` must build on OpenBSD without CGO | Use the same pure-Go approach already established in `pocketbaseserver`; Chrome is available on the production OpenBSD host |

## Acceptance criteria

- Monthly count report chart loads from the auto-generated `monthlyCountReports` collection API with no calls to `:3000`.
- `runtask` builds on OpenBSD and writes `jobPosts` records to `pocketbaseserver`.
- Running `runtask` twice does not create duplicate job posts.
- `runtask` computes and writes monthly count reports to `pocketbaseserver`.
- `runtask` performs maintenance tasks (temp file cleanup, email log).
- `migrate` successfully copies all records from `SkillSurvey.db` to `pocketbaseserver`.
- Integration tests are written before implementation; no database mocking.
