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

SELECT jp.id, jp.site_id, jp.job_site_number,
       jp.title, jp.body, jp.posted_date,
       jp.city, jp.country, jp.suburb
FROM   job_posts jp;
```

Column names follow GORM's default snake_case convention for the legacy schema.

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
| Site not found in PocketBase | Log warning with legacy job post ID and site name; increment Failed; continue |
| `app.Save()` error | Log error with legacy job post ID; increment Failed; continue |
| PocketBase Bootstrap failure | Print error, exit non-zero |

Processing continues record-by-record on individual failures. A non-zero exit at the end signals partial failure so the operator can investigate and re-run (safe due to idempotency).

---

## Testing strategy

Integration test in `altmigrate_test.go`. Starts a real PocketBase instance using `t.TempDir()` with all migrations applied via `import _ "keybook/pocketbaseserver/migrations"`. Creates a real legacy SQLite file using `database/sql`.

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
