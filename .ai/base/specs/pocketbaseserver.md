# pocketbaseserver — Spec

## Goals

- Replace the legacy `backend/` REST API with a PocketBase-based backend.
- Provide authentication, row-level security, and a REST API for all collections.
- Serve the compiled frontend as a static SPA.
- Run on OpenBSD without CGO.

## Functional requirements

### Collections
All collections listed below must exist with the fields and access rules described. Schema changes must be applied via migrations — never by editing the database directly.

#### sites
- Fields: `name` (text, required), `url` (text, required).
- Access: publicly readable.

#### skillTypes
- Fields: `name` (text, required), `description` (text, required).
- Access: public read; authenticated write.

#### skillNames
- Fields: `name` (text, required), `isEnabled` (bool), `skillType` (relation → skillTypes, required).
- Access: public read; authenticated write.

#### skillNameAliases
- Fields: `skillName` (relation → skillNames, required), `alias` (text, required).
- Access: public read; authenticated write.

#### jobPosts
- Fields: `jobSiteNumber` (text, required), `site` (relation → sites, required), `content` (JSON, required), `postedDate` (date), `location` (JSON, required — `{city, country, suburb}`).
- `(site, jobSiteNumber)` must be unique to prevent duplicate imports.
- Access: publicly readable.

#### monthlyCountReports
- Fields: `identifier` (text, required — `<skillNameID>_<YearMonth>`), `YearMonth` (text, required — YYYY-MM), `yearMonthDate` (date, required — first day of period), `count` (int, required), `skillName` (relation → skillNames, optional).
- Access: public.

#### userSettings
- Fields: `user` (relation → users, required, unique), `portalTheme` (select, required — `white | g10 | g90 | g100`).
- Access: owner only (`@request.auth.id != "" && user = @request.auth.id`).
- Cascade delete when the related user is deleted.

#### users (system collection)
- Standard PocketBase email/password auth collection.
- No custom fields required.

### Static file serving
- The compiled frontend (`dist/`) must be served from `./pb_public` at `/`.
- The SPA route must fall through to `index.html` for all unmatched paths.

### Migrations
- All schema changes must be implemented as numbered migration files in `migrations/`.
- Each migration must implement both `up` and `down`.
- Migrations run automatically on `serve` in both dev and production modes.

## Non-functional requirements

- Must use the pure-Go SQLite driver (`modernc.org/sqlite`) — no CGO, no cgo-based sqlite3.
- Must build and run on OpenBSD.
- Default port: `8090`.
- `pb_data/` (SQLite database and uploads) must be git-ignored.
- All new features require integration tests against a real PocketBase instance — no mocking.

## Out of scope

- Job scraping (handled by `runtask`, a separate service not yet in this repo).
- Report computation (to be triggered externally or via a future scheduled task).
- Custom REST endpoints beyond the single SPA static-file route.
