package migrator_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	pocketbaseclient "github.com/r--w/pocketbase"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"keybook/migrate/internal/legacyentities"
	"keybook/migrate/internal/migrator"
	_ "keybook/pocketbaseserver/migrations"
)

// startTestPocketBase starts a PocketBase HTTP server with all pocketbaseserver migrations
// applied. Returns the app, base URL, and a cleanup function.
func startTestPocketBase(t *testing.T) (core.App, string) {
	t.Helper()

	app := pocketbase.NewWithConfig(pocketbase.Config{DefaultDataDir: t.TempDir()})
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{})

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	serverURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	var httpServer *http.Server
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		httpServer = e.Server
		return e.Next()
	})

	app.RootCmd.SetArgs([]string{
		"serve",
		"--http", fmt.Sprintf("127.0.0.1:%d", port),
	})
	go func() { _ = app.Start() }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			t.Fatal("PocketBase test server did not start in time")
		default:
		}
		resp, err := http.Get(serverURL + "/api/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Cleanup(func() {
		if httpServer != nil {
			shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = httpServer.Shutdown(shutCtx)
		}
	})

	return app, serverURL
}

// legacyDB creates an in-memory SQLite database with the legacy schema seeded with one
// record per table. Returns the DB and the seed IDs used.
func legacyDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	if err := db.AutoMigrate(
		&legacyentities.Site{},
		&legacyentities.SkillType{},
		&legacyentities.SkillName{},
		&legacyentities.SkillNameAlias{},
		&legacyentities.JobPost{},
		&legacyentities.MonthlyCountReport{},
	); err != nil {
		t.Fatalf("auto-migrate legacy schema: %v", err)
	}

	site := legacyentities.Site{EntityBase: legacyentities.EntityBase{ID: 1}, Name: "TestSite", Url: "https://test.example.com"}
	db.Create(&site)

	st := legacyentities.SkillType{EntityBase: legacyentities.EntityBase{ID: 1}, Name: "Programming", Description: "Prog langs"}
	db.Create(&st)

	sn := legacyentities.SkillName{EntityBase: legacyentities.EntityBase{ID: 1}, SkillTypeID: 1, Name: "Go", IsEnabled: true}
	db.Create(&sn)

	alias := legacyentities.SkillNameAlias{EntityBase: legacyentities.EntityBase{ID: 1}, SkillNameID: 1, Alias: "golang"}
	db.Create(&alias)

	jp := legacyentities.JobPost{
		EntityBase:    legacyentities.EntityBase{ID: 1},
		SiteID:        1,
		JobSiteNumber: "JP-001",
		Title:         "Go Developer",
		Body:          "We need a Go developer",
		City:          "Sydney",
		Country:       "AU",
		Suburb:        "CBD",
	}
	db.Create(&jp)

	report := legacyentities.MonthlyCountReport{
		EntityBase:  legacyentities.EntityBase{ID: 1},
		SkillNameID: 1,
		YearMonth:   "2024-01",
		Count:       5,
	}
	db.Create(&report)

	return db
}

// createServiceAccount creates a PocketBase user with the migration role so the
// r--w/pocketbase client can authenticate and write records.
func createServiceAccount(t *testing.T, app core.App, serverURL string) (email, password string) {
	t.Helper()
	email = "migrate-svc@example.com"
	password = "testtest123"

	col, _ := app.FindCollectionByNameOrId("_pb_users_auth_")
	user := core.NewRecord(col)
	user.Set("email", email)
	user.Set("password", password)
	user.Set("passwordConfirm", password)
	user.Set("verified", true)
	if err := app.Save(user); err != nil {
		t.Fatalf("save service account: %v", err)
	}

	role, err := app.FindFirstRecordByData("roles", "name", "migration")
	if err != nil {
		t.Fatalf("find migration role: %v", err)
	}
	urCol, _ := app.FindCollectionByNameOrId("userRoles")
	ur := core.NewRecord(urCol)
	ur.Set("user", user.Id)
	ur.Set("role", role.Id)
	if err := app.Save(ur); err != nil {
		t.Fatalf("assign migration role: %v", err)
	}

	return email, password
}

func TestMigratorRunCreatesAllRecords(t *testing.T) {
	app, serverURL := startTestPocketBase(t)
	db := legacyDB(t)
	email, password := createServiceAccount(t, app, serverURL)

	pb := pocketbaseclient.NewClient(serverURL, pocketbaseclient.WithUserEmailPassword(email, password))

	m := migrator.New(db, pb)
	summaries, err := m.Run()
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	for _, s := range summaries {
		if s.Attempted != s.Written {
			t.Errorf("collection %s: attempted=%d written=%d", s.Collection, s.Attempted, s.Written)
		}
		if s.Written != 1 {
			t.Errorf("collection %s: expected 1 written record, got %d", s.Collection, s.Written)
		}
	}
}

func TestMigratorSkillNameHasNewSkillTypeID(t *testing.T) {
	app, serverURL := startTestPocketBase(t)
	db := legacyDB(t)
	email, password := createServiceAccount(t, app, serverURL)

	pb := pocketbaseclient.NewClient(serverURL, pocketbaseclient.WithUserEmailPassword(email, password))

	m := migrator.New(db, pb)
	if _, err := m.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Find the skillType first so we can filter the skillName by both name and type.
	stList, err := pb.List("skillTypes", pocketbaseclient.ParamsList{Filters: `name = "Programming"`})
	if err != nil || len(stList.Items) == 0 {
		t.Fatalf("find skillType Programming: %v", err)
	}
	stID := stList.Items[0]["id"].(string)

	// The seed migration also contains a "Go" skill name (under "Back End Language"),
	// so filter by both name and the specific skillType to get the migrated record.
	snList, err := pb.List("skillNames", pocketbaseclient.ParamsList{
		Filters: fmt.Sprintf(`name = "Go" && skillType = %q`, stID),
	})
	if err != nil || len(snList.Items) == 0 {
		t.Fatalf("find skillName Go under Programming: %v", err)
	}
	snRecord := snList.Items[0]

	// The skillType field should be a PocketBase ID (string), not an integer.
	skillTypeField, ok := snRecord["skillType"].(string)
	if !ok || skillTypeField == "" {
		t.Errorf("skillName.skillType should be a non-empty PocketBase string ID, got %v", snRecord["skillType"])
	}
	if skillTypeField != stID {
		t.Errorf("skillName.skillType = %q, want PocketBase ID %q", skillTypeField, stID)
	}
}

func TestMigratorIsIdempotent(t *testing.T) {
	app, serverURL := startTestPocketBase(t)
	db := legacyDB(t)
	email, password := createServiceAccount(t, app, serverURL)

	pb := pocketbaseclient.NewClient(serverURL, pocketbaseclient.WithUserEmailPassword(email, password))

	m := migrator.New(db, pb)
	if _, err := m.Run(); err != nil {
		t.Fatalf("first Run: %v", err)
	}

	collections := []string{"sites", "skillTypes", "skillNames", "skillNameAliases", "jobPosts", "monthlyCountReports"}
	countAfterFirst := make(map[string]int, len(collections))
	for _, col := range collections {
		list, err := pb.List(col, pocketbaseclient.ParamsList{Size: 500})
		if err != nil {
			t.Fatalf("list %s after first run: %v", col, err)
		}
		countAfterFirst[col] = list.TotalItems
	}

	if _, err := m.Run(); err != nil {
		t.Fatalf("second Run: %v", err)
	}

	for _, col := range collections {
		list, err := pb.List(col, pocketbaseclient.ParamsList{Size: 500})
		if err != nil {
			t.Fatalf("list %s after second run: %v", col, err)
		}
		if list.TotalItems != countAfterFirst[col] {
			t.Errorf("collection %s: count after first run=%d, after second run=%d (expected no change)", col, countAfterFirst[col], list.TotalItems)
		}
	}
}

func TestMigratorMonthlyCountReportIdentifierUsesNewPBID(t *testing.T) {
	app, serverURL := startTestPocketBase(t)
	db := legacyDB(t)
	email, password := createServiceAccount(t, app, serverURL)

	pb := pocketbaseclient.NewClient(serverURL, pocketbaseclient.WithUserEmailPassword(email, password))

	m := migrator.New(db, pb)
	if _, err := m.Run(); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The seed migration also contains a "Go" skill name (under "Back End Language"),
	// so resolve "Programming" first to find the migrated "Go" specifically.
	stList, err := pb.List("skillTypes", pocketbaseclient.ParamsList{Filters: `name = "Programming"`})
	if err != nil || len(stList.Items) == 0 {
		t.Fatalf("find skillType Programming: %v", err)
	}
	stID := stList.Items[0]["id"].(string)

	snList, err := pb.List("skillNames", pocketbaseclient.ParamsList{
		Filters: fmt.Sprintf(`name = "Go" && skillType = %q`, stID),
	})
	if err != nil || len(snList.Items) == 0 {
		t.Fatalf("find skillName Go under Programming: %v", err)
	}
	newSkillNameID := snList.Items[0]["id"].(string)

	// Verify the identifier in the monthlyCountReport.
	reportList, err := pb.List("monthlyCountReports", pocketbaseclient.ParamsList{})
	if err != nil || len(reportList.Items) == 0 {
		t.Fatalf("find monthlyCountReports: %v", err)
	}
	identifier, _ := reportList.Items[0]["identifier"].(string)
	expectedIdentifier := newSkillNameID + "_2024-01"
	if identifier != expectedIdentifier {
		t.Errorf("identifier = %q, want %q", identifier, expectedIdentifier)
	}
}
