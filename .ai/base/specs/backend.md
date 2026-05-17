# backend — Spec

> **Status: Legacy.** Remains in service while `pocketbaseserver` migration is in progress. No new features should be added. Stabilise and maintain only.

## Goals

- Keep the REST API at `:3000` stable and correct so the frontend continues to function during migration.
- Scraping, report generation, and maintenance jobs must remain reliable on OpenBSD.

## Functional requirements

### survey binary
- Crawl Seek and Jora for Australian job listings using Chromedp.
- Insert or update `JobPost` records in SQLite, keyed by `(SiteID, JobSiteNumber)` to prevent duplicates.
- Rate-limit requests per scraper config files (`seek.json`, `jora.json`).

### reports binary
- Compute monthly skill-demand counts from `JobPost` bodies.
- Match skill aliases case-insensitively with word-boundary patterns (space, comma, period, newline).
- Write results to `MonthlyCountReport`, identified by `<SkillNameID>_<YearMonth>`.

### results binary (REST API, port 3000)
- Serve all endpoints listed in the codebase index over HTTP with CORS enabled for all origins.
- All responses must be JSON.
- Endpoints must not require authentication.

### housekeeping binary
- `cleanfs`: remove Chromium temporary files left by scrapers.
- `sendlog`: email `error.log` contents then clear the file.

## Non-functional requirements

- Must build and run on OpenBSD without CGO.
- All panics must be caught, written to `error.log` with a timestamp and stack trace, and the process must exit cleanly.
- No new dependencies may be introduced without review.
- Config must be read from JSON files next to the executable — no environment variables for runtime config.

## Out of scope

- New API endpoints or data models.
- Authentication or authorisation.
- Any migration of data to the new stack (handled separately).
