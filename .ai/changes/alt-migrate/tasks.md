# Tasks: alt-migrate

Work through tasks in order. Check each item when done. **Integration test task comes first — before any implementation code** (per CLAUDE.md testing mandate). All tests run on the **OpenBSD server**.

---

## Task 1 — Write integration tests [required, written first]

**File:** `pocketbaseserver/internal/altmigrate/altmigrate_test.go`

Write three test cases before any implementation code exists. Watch them fail at compile time — that is expected. Write the minimum implementation in Task 2 to make them pass.

### `TestAltMigrateRunCreatesJobPosts`
- [ ] Start PocketBase with `t.TempDir()` data dir; import `_ "keybook/pocketbaseserver/migrations"` to apply all migrations.
- [ ] Create a legacy SQLite file in `t.TempDir()` using `database/sql`; insert 1 site row and 2 job post rows.
- [ ] Manually create the matching site record in PocketBase (same name as the legacy site).
- [ ] Call `altmigrate.Run(app, legacyDbPath)`.
- [ ] Assert PocketBase `jobPosts` collection has 2 records.
- [ ] Assert each record has correct `jobSiteNumber`, `site` (PocketBase ID), `postedDate`, `content` JSON, and `location` JSON.
- [ ] Assert returned Summary: `Attempted=2, Written=2, Skipped=0, Failed=0`.

### `TestAltMigrateIsIdempotent`
- [ ] Same setup as above.
- [ ] Call `altmigrate.Run` twice.
- [ ] Assert exactly 2 records exist after second run (no duplicates).
- [ ] Assert second Summary: `Attempted=2, Written=0, Skipped=2, Failed=0`.

### `TestAltMigrateSkipsJobPostWithUnknownSite`
- [ ] Start PocketBase (no sites seeded).
- [ ] Create legacy SQLite with 1 site and 1 job post; do NOT create a matching PocketBase site.
- [ ] Call `altmigrate.Run`.
- [ ] Assert 0 records in PocketBase `jobPosts`.
- [ ] Assert Summary: `Attempted=1, Written=0, Skipped=0, Failed=1`.

**Expected outcome:** Tests compile and fail because `altmigrate` package does not exist yet. Push to OpenBSD server and confirm compile failure there before proceeding.

---

## Task 2 — Implement altmigrate package [required]

**File:** `pocketbaseserver/internal/altmigrate/altmigrate.go`

Implement `Run(app core.App, legacyDbPath string) (Summary, error)` following the design:

- [ ] Open legacy SQLite with `database/sql` using the `"sqlite"` driver (registered by PocketBase on bootstrap — no extra import needed).
- [ ] Load all legacy sites via `SELECT id, name FROM sites`.
- [ ] Load all PocketBase site records via `app.FindRecordsByFilter("sites", "", "", -1, 0)` and build a `map[uint]string` from legacy site ID to PocketBase site ID (joined by site name).
- [ ] Load all legacy job posts via `SELECT jp.id, jp.site_id, jp.job_site_number, jp.title, jp.body, jp.posted_date, jp.city, jp.country, jp.suburb FROM job_posts jp`.
- [ ] For each job post: resolve `pbSiteID` from the map; if missing, log warning, increment `Failed`, continue.
- [ ] For each job post: check existence via `app.FindRecordsByFilter("jobPosts", "jobSiteNumber={:n} && site={:s}", ...)`; if found, increment `Skipped`, continue.
- [ ] For each job post: create record via `core.NewRecord(col)`, `record.Set(...)`, `app.Save(record)`; on error log and increment `Failed`, on success increment `Written`.
- [ ] Return `Summary`.

**PocketBase field mapping:**

| Field | Value |
|---|---|
| `jobSiteNumber` | `legacyJobPost.JobSiteNumber` |
| `site` | Resolved PocketBase site ID string |
| `postedDate` | `time.Time` formatted as `"2006-01-02 15:04:05.000Z"` |
| `content` | `map[string]any{"title": ..., "body": ...}` |
| `location` | `map[string]any{"city": ..., "country": ..., "suburb": ...}` |

**Expected outcome:** Push to OpenBSD server; run `go test ./pocketbaseserver/internal/altmigrate/ -v -timeout 60s`; all three tests pass.

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
- [ ] Confirm summary counts match expected totals; spot-check a few records in PocketBase admin UI.
- [ ] Run a second time to verify idempotency (`written=0, skipped=N, failed=0`).

**Expected outcome:** All job posts migrated; second run is a no-op.

---

## Task 5 — Remove alt-migrate after migration is confirmed [required]

Once the migration is confirmed complete on the production PocketBase instance:

- [ ] Delete `pocketbaseserver/internal/altmigrate/` (both files).
- [ ] Remove the `alt-migrate` command registration block from `main.go`.
- [ ] Build and verify server still starts normally.
- [ ] Commit with message referencing this change.

**Expected outcome:** No trace of alt-migrate in the codebase.
