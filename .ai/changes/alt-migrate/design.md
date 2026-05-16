# Design: alt-migrate

## System architecture and components

```
pocketbaseserver binary
└── cobra command: "alt-migrate --db <path>"
        │
        ├── app.Bootstrap()            initialise PocketBase (apply migrations, open DB)
        │
        └── altmigrate.Run(app, path)  migration package
                │
                ├── database/sql ──── legacy SQLite (read-only)
                │
                └── core.App API ──── PocketBase SQLite (write)
```

### Layered Architecture mapping

This command is temporary and minimal, so only the relevant layers apply:

| Layer | Component | Role |
|---|---|---|
| **API** | cobra command in `main.go` | Parses `--db` flag; calls `app.Bootstrap()` and `altmigrate.Run()` |
| **Application** | `altmigrate.Run()` | Orchestrates the migration: loads sites, iterates job posts, coordinates reads and writes |
| **Store** | `database/sql` (legacy SQLite) + `core.App` (PocketBase SQLite) | Raw persistence — reads from legacy DB, writes through PocketBase's internal API |

No Service or Repository layer is warranted for this temporary one-shot command.

### New files

| Path | Purpose |
|---|---|
| `pocketbaseserver/internal/altmigrate/altmigrate.go` | Migration logic: read legacy → write PocketBase |
| `pocketbaseserver/internal/altmigrate/altmigrate_test.go` | Integration test |

### Changed files

| Path | Change |
|---|---|
| `pocketbaseserver/cmd/pocketbaseserver/main.go` | Register `alt-migrate` cobra command |

---

## Sequence diagram

```
user
  │
  │  pocketbaseserver alt-migrate --db /path/to/SkillSurvey.db
  ▼
main.go
  │  app.Bootstrap()          ← applies all migrations; PocketBase SQLite ready
  │
  │  altmigrate.Run(app, dbPath)
  ▼
altmigrate.Run
  │
  │  sql.Open("sqlite", dbPath)         open legacy DB (read-only)
  │
  │  SELECT id, name FROM sites         load legacy sites
  │  app.FindRecordsByFilter("sites")   load PocketBase sites
  │  build: legacySiteID → pbSiteID     join by site name
  │
  │  SELECT * FROM job_posts            load all legacy job posts
  │
  │  for each job post:
  │    pbSiteID = siteMap[jp.SiteID]    resolve site
  │    if not found → log warning, skip
  │
  │    app.FindRecordsByFilter(         existence check
  │      "jobPosts",
  │      "jobSiteNumber='X' && site='Y'"
  │    )
  │    if found → skip (idempotent)
  │
  │    record = core.NewRecord(col)
  │    record.Set(...)
  │    app.Save(record)
  │
  └── return Summary{Attempted, Written, Skipped, Failed}

main.go
  │  print summary
  └── os.Exit(0 or 1)
```

---

## Data models and interfaces

### Summary (returned by Run)

```go
type Summary struct {
    Attempted int
    Written   int
    Skipped   int  // already existed
    Failed    int
}
```

### Legacy entities (read via database/sql)

Two structs used only within `altmigrate` — no GORM, plain `sql.Rows` scan:

```go
type legacySite struct {
    ID   uint
    Name string
}

type legacyJobPost struct {
    ID            uint
    SiteID        uint
    JobSiteNumber string
    Title         string
    Body          string
    PostedDate    time.Time
    City          string
    Country       string
    Suburb        string
}
```

### Legacy SQL queries

```sql
SELECT id, name FROM sites;

SELECT id, site_id, job_site_number,
       title, body, posted_date,
       city, country, suburb
FROM   job_posts;
```

Confirmed column names from the production backup — all snake_case, no `url` column on `sites`.

### Known data characteristics (from production backup)

| Fact | Value |
|---|---|
| Total job_posts rows | 435,042 |
| site_id=1 (seek.com.au) | 233,924 |
| site_id=2 (au.jora.com) | 198,372 |
| site_id=0 (orphaned) | 2,746 |
| NULL values in any key field | None |
| Legacy site names | `seek.com.au`, `au.jora.com` |
| Seed migration site name | `www.seek.com.au` (different from legacy `seek.com.au`) — the existing `migrate` tool creates `seek.com.au` separately; both exist in PocketBase alongside each other |
| posted_date format | RFC3339 with offset, e.g. `2018-09-10 14:00:00+00:00` |
| Zero dates present | Yes — min date is `0001-01-01 00:00:00+00:00` |

### posted_date parsing

The `posted_date` column stores timestamps as strings in the format `2018-09-10 14:00:00+00:00` (RFC3339-like, space separator instead of `T`). Timezone offsets vary (`+00:00`, `+10:00`).

Parse in Go using:
```go
t, err := time.Parse("2006-01-02 15:04:05.999999999-07:00", row.PostedDate)
```

If parsing fails, fall back to:
```go
t, err = time.Parse("2006-01-02 15:04:05-07:00", row.PostedDate)
```

Write to PocketBase as UTC: `t.UTC().Format("2006-01-02 15:04:05.000Z")`

### Zero-date handling

Records with `posted_date = '0001-01-01 00:00:00+00:00'` (Go zero time) exist in the backup. These are data quality issues in the source. The system shall store them as-is — PocketBase accepts the zero date — and not treat them as an error.

### PocketBase record mapping

| PocketBase field | Source |
|---|---|
| `jobSiteNumber` | `legacyJobPost.JobSiteNumber` |
| `site` | Resolved PocketBase site ID |
| `postedDate` | `legacyJobPost.PostedDate` formatted as `"2006-01-02 15:04:05.000Z"` |
| `content` | `{"title": "...", "body": "..."}` JSON |
| `location` | `{"city": "...", "country": "...", "suburb": "..."}` JSON |

### SQLite driver

`database/sql` with the `"sqlite"` driver. PocketBase registers this driver during `app.Bootstrap()` via its own use of `modernc.org/sqlite` — no additional import or dependency is needed in pocketbaseserver.

---

## Error-handling approach

| Scenario | Handling |
|---|---|
| `--db` flag missing | cobra validation; exit before Bootstrap |
| Legacy DB file not found | `sql.Open` / first query fails; print error, exit non-zero |
| `site_id=0` (orphaned record) | Site lookup returns nothing; log warning with job post ID; increment Failed; continue |
| Site name not found in PocketBase | Log warning with legacy job post ID and site name; increment Failed; continue |
| `posted_date` parse failure | Log warning with job post ID and raw value; store empty string; continue |
| Zero date (`0001-01-01`) | Store as-is — not treated as an error |
| `app.Save()` error | Log error with legacy job post ID; increment Failed; continue |
| PocketBase Bootstrap failure | Print error, exit non-zero |

Processing continues record-by-record on individual failures. A non-zero exit at the end signals partial failure so the operator can investigate and re-run (safe due to idempotency).

**Expected production run outcome:** `failed=2746` (the `site_id=0` orphaned records — expected). Total attempted will exceed 435,042 (the 2026-05-08 backup count) as scraping has continued since then.

---

## Testing strategy

Integration tests are written **before** implementation code (per the CLAUDE.md testing mandate). All tests run on the **OpenBSD server** — not on Windows.

Integration test in `altmigrate_test.go`. Starts a real PocketBase instance using `t.TempDir()` with all migrations applied via `import _ "keybook/pocketbaseserver/migrations"`. Creates a real legacy SQLite file using `database/sql`. No mocking.

### Test cases

**`TestAltMigrateRunCreatesJobPosts`**
- Seed: 1 site in PocketBase; 1 matching site + 2 job posts in legacy DB
- Run `altmigrate.Run`
- Assert: 2 jobPost records in PocketBase with correct field values (jobSiteNumber, site relation, content JSON, location JSON, postedDate)
- Assert: Summary = `{Attempted:2, Written:2, Skipped:0, Failed:0}`

**`TestAltMigrateIsIdempotent`**
- Same seed as above
- Run `altmigrate.Run` twice
- Assert: still 2 jobPost records after second run (no duplicates)
- Assert: second Summary = `{Attempted:2, Written:0, Skipped:2, Failed:0}`

**`TestAltMigrateSkipsJobPostWithUnknownSite`**
- Seed: 0 sites in PocketBase; 1 job post in legacy DB referencing a site
- Run `altmigrate.Run`
- Assert: 0 jobPost records in PocketBase
- Assert: Summary = `{Attempted:1, Written:0, Skipped:0, Failed:1}`
