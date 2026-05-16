package altmigrate

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"

	_ "keybook/pocketbaseserver/migrations"
)

// startTestApp bootstraps a real PocketBase instance with all migrations applied.
// Does not start an HTTP server — only the internal app API is needed for altmigrate.
func startTestApp(t *testing.T) core.App {
	t.Helper()
	app := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: t.TempDir()})
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{})
	if err := app.Bootstrap(); err != nil {
		t.Fatalf("bootstrap PocketBase: %v", err)
	}
	return app
}

// createSiteInPB inserts a site record into PocketBase and returns its ID.
func createSiteInPB(t *testing.T, app core.App, name string) string {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("sites")
	if err != nil {
		t.Fatalf("find sites collection: %v", err)
	}
	rec := core.NewRecord(col)
	rec.Set("name", name)
	rec.Set("url", "https://"+name)
	if err := app.Save(rec); err != nil {
		t.Fatalf("save site %q: %v", name, err)
	}
	return rec.Id
}

// setupLegacyDB creates a temporary SQLite file with the legacy schema (sites + job_posts).
// The returned *sql.DB is closed automatically at end of test.
func setupLegacyDB(t *testing.T) (*sql.DB, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "legacy.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	if _, err = db.Exec("CREATE TABLE sites (id INTEGER PRIMARY KEY, name TEXT)"); err != nil {
		t.Fatalf("create sites table: %v", err)
	}
	if _, err = db.Exec(`CREATE TABLE job_posts (
		id             INTEGER PRIMARY KEY,
		site_id        INTEGER,
		job_site_number TEXT,
		title          TEXT,
		body           TEXT,
		posted_date    TEXT,
		city           TEXT,
		country        TEXT,
		suburb         TEXT
	)`); err != nil {
		t.Fatalf("create job_posts table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, dbPath
}

func insertSite(t *testing.T, db *sql.DB, id int, name string) {
	t.Helper()
	if _, err := db.Exec("INSERT INTO sites (id, name) VALUES (?, ?)", id, name); err != nil {
		t.Fatalf("insert legacy site %q: %v", name, err)
	}
}

func insertJobPost(t *testing.T, db *sql.DB, id, siteID int, jobSiteNumber, title, body, postedDate, city, country, suburb string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO job_posts
			(id, site_id, job_site_number, title, body, posted_date, city, country, suburb)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, siteID, jobSiteNumber, title, body, postedDate, city, country, suburb,
	)
	if err != nil {
		t.Fatalf("insert legacy job post %q: %v", jobSiteNumber, err)
	}
}

func countJobPosts(t *testing.T, app core.App) int {
	t.Helper()
	records, err := app.FindRecordsByFilter("jobPosts", "", "", -1, 0)
	if err != nil {
		t.Fatalf("count jobPosts: %v", err)
	}
	return len(records)
}

// TestAltMigrateRunCreatesJobPosts verifies that Run writes all legacy job posts
// to PocketBase with correct field values.
func TestAltMigrateRunCreatesJobPosts(t *testing.T) {
	app := startTestApp(t)
	pbSiteID := createSiteInPB(t, app, "seek.com.au")

	db, dbPath := setupLegacyDB(t)
	insertSite(t, db, 1, "seek.com.au")
	insertJobPost(t, db, 1, 1, "JP-001", "Go Developer", "Great Go role", "2024-01-15 10:00:00+00:00", "Sydney", "Australia", "CBD")
	insertJobPost(t, db, 2, 1, "JP-002", "Python Dev", "Great Python role", "2024-02-20 09:00:00+00:00", "Melbourne", "Australia", "Inner East")

	summary, err := Run(app, dbPath)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if summary.Attempted != 2 {
		t.Errorf("Attempted: want 2, got %d", summary.Attempted)
	}
	if summary.Written != 2 {
		t.Errorf("Written: want 2, got %d", summary.Written)
	}
	if summary.Skipped != 0 {
		t.Errorf("Skipped: want 0, got %d", summary.Skipped)
	}
	if summary.Failed != 0 {
		t.Errorf("Failed: want 0, got %d", summary.Failed)
	}
	if n := countJobPosts(t, app); n != 2 {
		t.Fatalf("expected 2 jobPosts in PocketBase, got %d", n)
	}

	// Verify field values on the first record.
	recs, err := app.FindRecordsByFilter("jobPosts", "jobSiteNumber='JP-001'", "", 1, 0)
	if err != nil || len(recs) == 0 {
		t.Fatalf("record JP-001 not found: %v", err)
	}
	r := recs[0]
	if r.GetString("site") != pbSiteID {
		t.Errorf("site: want %q, got %q", pbSiteID, r.GetString("site"))
	}
	if !strings.Contains(r.GetString("content"), "Go Developer") {
		t.Errorf("content JSON missing title: %s", r.GetString("content"))
	}
	if !strings.Contains(r.GetString("location"), "Sydney") {
		t.Errorf("location JSON missing city: %s", r.GetString("location"))
	}
	if r.GetString("postedDate") == "" {
		t.Error("postedDate is empty")
	}
}

// TestAltMigrateIsIdempotent verifies that running Run twice produces no duplicates
// and reports all records as skipped on the second run.
func TestAltMigrateIsIdempotent(t *testing.T) {
	app := startTestApp(t)
	createSiteInPB(t, app, "seek.com.au")

	db, dbPath := setupLegacyDB(t)
	insertSite(t, db, 1, "seek.com.au")
	insertJobPost(t, db, 1, 1, "JP-001", "Go Developer", "Great Go role", "2024-01-15 10:00:00+00:00", "Sydney", "Australia", "CBD")
	insertJobPost(t, db, 2, 1, "JP-002", "Python Dev", "Great Python role", "2024-02-20 09:00:00+00:00", "Melbourne", "Australia", "Inner East")

	if _, err := Run(app, dbPath); err != nil {
		t.Fatalf("first Run: %v", err)
	}

	summary, err := Run(app, dbPath)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}

	if summary.Attempted != 2 {
		t.Errorf("Attempted: want 2, got %d", summary.Attempted)
	}
	if summary.Written != 0 {
		t.Errorf("Written: want 0, got %d", summary.Written)
	}
	if summary.Skipped != 2 {
		t.Errorf("Skipped: want 2, got %d", summary.Skipped)
	}
	if summary.Failed != 0 {
		t.Errorf("Failed: want 0, got %d", summary.Failed)
	}
	if n := countJobPosts(t, app); n != 2 {
		t.Errorf("expected 2 jobPosts after second run (no duplicates), got %d", n)
	}
}

// TestAltMigrateSkipsJobPostWithUnknownSite verifies that job posts referencing a
// site name not found in PocketBase are counted as failed and not written.
func TestAltMigrateSkipsJobPostWithUnknownSite(t *testing.T) {
	app := startTestApp(t)
	// No PocketBase site created — site lookup will fail.

	db, dbPath := setupLegacyDB(t)
	insertSite(t, db, 1, "seek.com.au")
	insertJobPost(t, db, 1, 1, "JP-001", "Go Developer", "role body", "2024-01-15 10:00:00+00:00", "Sydney", "Australia", "CBD")

	summary, err := Run(app, dbPath)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if summary.Attempted != 1 {
		t.Errorf("Attempted: want 1, got %d", summary.Attempted)
	}
	if summary.Written != 0 {
		t.Errorf("Written: want 0, got %d", summary.Written)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed: want 1, got %d", summary.Failed)
	}
	if n := countJobPosts(t, app); n != 0 {
		t.Errorf("expected 0 jobPosts, got %d", n)
	}
}

// TestAltMigrateSiteIdZeroCountsAsFailed verifies that job posts with site_id=0
// (orphaned records in the legacy DB) are counted as failed and not written.
func TestAltMigrateSiteIdZeroCountsAsFailed(t *testing.T) {
	app := startTestApp(t)
	createSiteInPB(t, app, "seek.com.au")

	db, dbPath := setupLegacyDB(t)
	insertSite(t, db, 1, "seek.com.au")
	// site_id=0 has no matching entry in the legacy sites table.
	insertJobPost(t, db, 1, 0, "JP-ZERO", "Orphan Job", "no valid site", "2024-01-15 10:00:00+00:00", "Sydney", "Australia", "CBD")

	summary, err := Run(app, dbPath)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if summary.Attempted != 1 {
		t.Errorf("Attempted: want 1, got %d", summary.Attempted)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed: want 1, got %d", summary.Failed)
	}
	if n := countJobPosts(t, app); n != 0 {
		t.Errorf("expected 0 jobPosts, got %d", n)
	}
}

// TestAltMigrateHandlesZeroDate verifies that job posts with a Go zero date
// (0001-01-01) are written successfully and not treated as an error.
func TestAltMigrateHandlesZeroDate(t *testing.T) {
	app := startTestApp(t)
	createSiteInPB(t, app, "seek.com.au")

	db, dbPath := setupLegacyDB(t)
	insertSite(t, db, 1, "seek.com.au")
	insertJobPost(t, db, 1, 1, "JP-ZERODATE", "Zero Date Job", "body", "0001-01-01 00:00:00+00:00", "Sydney", "Australia", "CBD")

	summary, err := Run(app, dbPath)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if summary.Written != 1 {
		t.Errorf("Written: want 1, got %d", summary.Written)
	}
	if summary.Failed != 0 {
		t.Errorf("Failed: want 0, got %d", summary.Failed)
	}
	if n := countJobPosts(t, app); n != 1 {
		t.Errorf("expected 1 jobPost, got %d", n)
	}
}
