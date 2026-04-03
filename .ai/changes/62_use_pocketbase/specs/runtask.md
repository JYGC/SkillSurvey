# runtask — Spec (issue #62)

## Goals

- Replace `backend/cmd/survey`, `backend/cmd/reports`, and `backend/cmd/housekeeping`
  with a single Go module that writes to `pocketbaseserver` instead of SQLite.
- Authenticate with `pocketbaseserver` as a service account that holds both the
  `webscraper` and `reporting` roles.
- Build and run on OpenBSD without CGO.

## Module location

`runtask/` at the repository root (sibling of `pocketbaseserver/`).

## Commands

The module exposes three sub-commands, selected via a CLI argument:

| Command | Replaces | Description |
|---|---|---|
| `scrape` | `backend/cmd/survey` | Crawl Seek and Jora; write `jobPosts` to pocketbaseserver |
| `report` | `backend/cmd/reports` | Compute monthly skill-demand counts; write `monthlyCountReports` |
| `housekeeping` | `backend/cmd/housekeeping` | Clean Chromium temp files; email `error.log` |

Entry point: `runtask/cmd/runtask/main.go`.

## Configuration

Read from a JSON file (`runtask.json`) located next to the executable. No environment
variables for runtime config.

Required fields:

```json
{
  "PocketBaseUrl": "http://...:8090",
  "ServiceAccountEmail": "...",
  "ServiceAccountPassword": "...",
  "SeekConfigFile": "seek.json",
  "JoraConfigFile": "jora.json",
  "ErrorLogFile": "error.log",
  "SmtpHost": "...",
  "SmtpPort": 25,
  "EmailRecipient": "..."
}
```

Passwords are set by the operator; never hardcoded or committed.

## scrape command

### Behaviour

1. Authenticate with PocketBase using `ServiceAccountEmail`/`ServiceAccountPassword`.
2. Load the list of sites from the `sites` collection.
3. For each site, run the corresponding adapter (Seek or Jora) via Chromedp.
4. For each scraped job post, upsert to `jobPosts`:
   - Filter by `site = "<siteId>" && jobSiteNumber = "<number>"`.
   - If a record exists, skip (no update needed — content does not change after posting).
   - If no record exists, create it.
5. Log all errors to `ErrorLogFile`; do not panic.

### Upsert rule

`(site, jobSiteNumber)` is the natural key. A second run must not create duplicates.
Use `getFullList` with a filter before attempting create, or catch the unique-constraint
error from the PocketBase API and treat it as a no-op.

### Adapters

Extract `backend/internal/siteadapters/` verbatim into `runtask/internal/siteadapters/`.
The adapter interface remains the same; only the persistence layer changes (PocketBase
API instead of GORM).

The `content` field in `jobPosts` stores the full job post body as a JSON string
(`{"title":"...","body":"..."}`). The `location` field stores
`{"city":"...","country":"...","suburb":"..."}`.

### Chromedp

Copy `backend/internal/dynamiccontentextractor/` into
`runtask/internal/dynamiccontentextractor/` unchanged.

## report command

### Behaviour

1. Authenticate with PocketBase.
2. Fetch all enabled skill names with their aliases:
   `collection('skillNames').getFullList({ filter: 'isEnabled = true', expand: 'skillNameAliases_via_skillName' })`.
3. Fetch all job posts: `collection('jobPosts').getFullList()`.
4. For each `(skillName, yearMonth)` pair, count job posts whose `content.body`
   contains any alias (case-insensitive, word-boundary match using the same patterns
   as the legacy backend: space, comma, period, newline).
5. Upsert each result to `monthlyCountReports`:
   - `identifier` = `<skillNameId>_<YYYY-MM>`.
   - Filter by `identifier = "<value>"` before deciding to create or update.

### Word-boundary matching

Port the logic from `backend/internal/database/jobposttablecall.go`
`GetMonthlyCountBySkill` directly. Patterns: ` alias `, `,alias,`, `.alias.`,
`\nalias\n`, and leading/trailing variants.

## housekeeping command

### cleanfs task

Remove Chromium temporary files. Port the logic from
`backend/cmd/housekeeping/main.go` cleanfs block verbatim.

### sendlog task

Email the contents of `ErrorLogFile` to `EmailRecipient` via SMTP, then truncate
the file. Port the logic from `backend/cmd/housekeeping/main.go` sendlog block.

## Error handling

- All panics must be recovered, written to `ErrorLogFile` with a timestamp and
  stack trace, and the process must exit with a non-zero status.
- Port `backend/internal/exception/` into `runtask/internal/exception/` unchanged.

## Build

- `runtask/Makefile` must provide `build` and `run_dev` targets matching the
  pattern used in `pocketbaseserver/Makefile`.
- Use `modernc.org/sqlite` (pure-Go) if SQLite is ever needed; no CGO.
- Must compile on OpenBSD: `GOOS=openbsd go build ./...` must succeed.
- Chromedp connects to the system Chrome binary already installed on the OpenBSD
  host; no bundled browser.

## Integration tests

Write tests before implementation:

- `scrape`: after running the scraper against a stub HTTP server (not a mock of
  PocketBase — use a real test PocketBase instance), `jobPosts` contains the
  expected records.
- `scrape` twice: record count does not increase (upsert is idempotent).
- `report`: given known `jobPosts` and `skillNames` in PocketBase, the computed
  `monthlyCountReports` match the expected counts.
- `housekeeping cleanfs`: after creating known Chromium temp files in a temp
  directory and pointing the task at it, running `cleanfs` removes them and leaves
  no unexpected files behind.
- `housekeeping sendlog`: given a non-empty `ErrorLogFile` and a local SMTP test
  server (e.g. `github.com/emersion/go-smtp` in-process), running `sendlog`
  delivers one email containing the log contents to `EmailRecipient` and truncates
  the file to zero bytes.

## Out of scope

- Scheduling (operator responsibility — cron or equivalent).
- PocketBase admin UI interaction.
- Any changes to `backend/` (legacy, maintenance-only).
