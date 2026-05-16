# Tasks: alt-migrate

Work through tasks in order. Check each item when done. **Integration test task comes first — before any implementation code** (per CLAUDE.md testing mandate). All tests run on the **OpenBSD server**.

---

## Task 1 — Write integration tests [required, written first]

**File:** `pocketbaseserver/internal/altmigrate/altmigrate_test.go`

Write three test cases before any implementation code exists. Watch them fail at compile time — that is expected. Write the minimum implementation in Task 2 to make them pass.

### `TestAltMigrateRunCreatesJobPosts`
- [x] Start PocketBase with `t.TempDir()` data dir; import `_ "keybook/pocketbaseserver/migrations"` to apply all migrations.
- [x] Create a legacy SQLite file in `t.TempDir()` using `database/sql`; insert 1 site row and 2 job post rows.
- [x] Manually create the matching site record in PocketBase (same name as the legacy site).
- [x] Call `altmigrate.Run(app, legacyDbPath)`.
- [x] Assert PocketBase `jobPosts` collection has 2 records.
- [x] Assert each record has correct `jobSiteNumber`, `site` (PocketBase ID), `postedDate`, `content` JSON, and `location` JSON.
- [x] Assert returned Summary: `Attempted=2, Written=2, Skipped=0, Failed=0`.

### `TestAltMigrateIsIdempotent`
- [x] Same setup as above.
- [x] Call `altmigrate.Run` twice.
- [x] Assert exactly 2 records exist after second run (no duplicates).
- [x] Assert second Summary: `Attempted=2, Written=0, Skipped=2, Failed=0`.

### `TestAltMigrateSkipsJobPostWithUnknownSite`
- [x] Start PocketBase (no sites seeded).
- [x] Create legacy SQLite with 1 site and 1 job post; do NOT create a matching PocketBase site.
- [x] Call `altmigrate.Run`.
- [x] Assert 0 records in PocketBase `jobPosts`.
- [x] Assert Summary: `Attempted=1, Written=0, Skipped=0, Failed=1`.

### `TestAltMigrateSiteIdZeroCountsAsFailed`
- [x] Start PocketBase with 1 site seeded.
- [x] Create legacy SQLite with a job post where `site_id=0`.
- [x] Call `altmigrate.Run`.
- [x] Assert 0 records in PocketBase `jobPosts`.
- [x] Assert Summary: `Attempted=1, Written=0, Skipped=0, Failed=1`.

### `TestAltMigrateHandlesZeroDate`
- [x] Start PocketBase with 1 site seeded.
- [x] Create legacy SQLite with a job post where `posted_date='0001-01-01 00:00:00+00:00'`.
- [x] Call `altmigrate.Run`.
- [x] Assert 1 record written to PocketBase (zero date is stored, not treated as an error).
- [x] Assert Summary: `Attempted=1, Written=1, Skipped=0, Failed=0`.

**Expected outcome:** ✅ Confirmed on OpenBSD server — `undefined: Run` compile failure on all 5 test calls. Proceed to Task 2.

---

## Task 2 — Implement altmigrate package [required]

**File:** `pocketbaseserver/internal/altmigrate/altmigrate.go`

Implement `Run(app core.App, legacyDbPath string) (Summary, error)` following the design:

- [x] Open legacy SQLite with `database/sql` using the `"sqlite"` driver (registered by PocketBase on bootstrap — no extra import needed).
- [x] Load all legacy sites via `SELECT id, name FROM sites`.
- [x] Load all PocketBase site records via `app.FindRecordsByFilter("sites", "", "", -1, 0)` and build a `map[uint]string` from legacy site ID to PocketBase site ID (joined by site name).
- [x] Load all legacy job posts via `SELECT jp.id, jp.site_id, jp.job_site_number, jp.title, jp.body, jp.posted_date, jp.city, jp.country, jp.suburb FROM job_posts jp`.
- [x] For each job post: resolve `pbSiteID` from the map; if missing, log warning, increment `Failed`, continue.
- [x] For each job post: check existence via `app.FindRecordsByFilter("jobPosts", "jobSiteNumber={:n} && site={:s}", ...)`; if found, increment `Skipped`, continue.
- [x] For each job post: create record via `core.NewRecord(col)`, `record.Set(...)`, `app.Save(record)`; on error log and increment `Failed`, on success increment `Written`.
- [x] Return `Summary`.

**PocketBase field mapping:**

| Field | Value |
|---|---|
| `jobSiteNumber` | `legacyJobPost.JobSiteNumber` |
| `site` | Resolved PocketBase site ID string |
| `postedDate` | Parse source string (`"2006-01-02 15:04:05-07:00"`) → convert to UTC → format as `"2006-01-02 15:04:05.000Z"` |
| `content` | `map[string]any{"title": ..., "body": ...}` |
| `location` | `map[string]any{"city": ..., "country": ..., "suburb": ...}` |

**Expected outcome:** ✅ Confirmed on OpenBSD server — all 5 tests pass (`ok keybook/pocketbaseserver/internal/altmigrate 8.272s`).

---

## Task 3 — Register cobra command in main.go [required]

**File:** `pocketbaseserver/cmd/pocketbaseserver/main.go`

Before `app.Start()`, add:

```go
altMigrateCmd := &cobra.Command{
    Use:   "alt-migrate",
    Short: "Migrate jobPosts from legacy SQLite directly into PocketBase",
    RunE: func(cmd *cobra.Command, args []string) error {
        dbPath, _ := cmd.Flags().GetString("db")
        if dbPath == "" {
            return errors.New("--db flag is required")
        }
        if err := app.Bootstrap(); err != nil {
            return err
        }
        summary, err := altmigrate.Run(app, dbPath)
        fmt.Printf("jobPosts:   attempted=%d  written=%d  skipped=%d  failed=%d\n",
            summary.Attempted, summary.Written, summary.Skipped, summary.Failed)
        if err != nil {
            return err
        }
        if summary.Failed > 0 {
            return errors.New("migration completed with failures — see log for details")
        }
        return nil
    },
}
altMigrateCmd.Flags().String("db", "", "Path to legacy SkillSurvey.db SQLite file")
app.RootCmd().AddCommand(altMigrateCmd)
```

- [ ] Add `alt-migrate` cobra command to `main.go`.
- [ ] Confirm `pocketbaseserver alt-migrate --help` prints usage on the server.

**Expected outcome:** `pocketbaseserver alt-migrate --db /path/to/SkillSurvey.db` runs and prints a summary.

---

## Task 4 — Build and run on OpenBSD server [required]

- [ ] Push branch; pull on server.
- [ ] `cd pocketbaseserver && make build`
- [ ] Verify the command is registered: `./build/pocketbaseserver alt-migrate --help`
- [ ] Run against a copy of the real `SkillSurvey.db`: `./build/pocketbaseserver alt-migrate --db /path/to/SkillSurvey.db`
- [ ] Confirm summary shows `failed=2746` (the `site_id=0` orphans — expected) and `written + skipped + failed = attempted` (totals must balance). Written count will exceed 432,296 — the backup inspected was from 2026-05-08; the live DB will have more records scraped since then.
- [ ] Spot-check a few records in PocketBase admin UI (verify `postedDate`, `content`, `location` fields).
- [ ] Run a second time to verify idempotency: `written=0`, all previously written records appear as `skipped`, `failed=2746`.

**Expected outcome:** All valid job posts migrated (432,296+ depending on scraping since 2026-05-08); second run is a no-op for all successfully written records.

---

## Task 5 — Remove alt-migrate after migration is confirmed [required]

Once the migration is confirmed complete on the production PocketBase instance:

- [ ] Delete `pocketbaseserver/internal/altmigrate/` (both files).
- [ ] Remove the `alt-migrate` command registration block from `main.go`.
- [ ] Build and verify server still starts normally.
- [ ] Commit with message referencing this change.

**Expected outcome:** No trace of alt-migrate in the codebase.
