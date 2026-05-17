package report_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	pocketbaseclient "github.com/r--w/pocketbase"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/pbclient"
	"keybook/runtask/internal/report"
	_ "keybook/pocketbaseserver/migrations"
)

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

// createReportingAccount creates a user with both webscraper and reporting roles.
func createReportingAccount(t *testing.T, app core.App) (email, password string) {
	t.Helper()
	email = "reporting-svc@example.com"
	password = "testtest123"

	col, _ := app.FindCollectionByNameOrId("_pb_users_auth_")
	user := core.NewRecord(col)
	user.Set("email", email)
	user.Set("password", password)
	user.Set("passwordConfirm", password)
	user.Set("verified", true)
	if err := app.Save(user); err != nil {
		t.Fatalf("save reporting account: %v", err)
	}

	urCol, _ := app.FindCollectionByNameOrId("userRoles")
	for _, roleName := range []string{"webscraper", "reporting"} {
		role, _ := app.FindFirstRecordByData("roles", "name", roleName)
		ur := core.NewRecord(urCol)
		ur.Set("user", user.Id)
		ur.Set("role", role.Id)
		app.Save(ur)
	}
	return email, password
}

// seedSkillData creates a skillType, skillName, and alias; returns their PocketBase IDs.
func seedSkillData(t *testing.T, app core.App) (skillTypeID, skillNameID string) {
	t.Helper()

	stCol, _ := app.FindCollectionByNameOrId("skillTypes")
	st := core.NewRecord(stCol)
	st.Set("name", "Programming")
	st.Set("description", "Languages")
	app.Save(st)
	skillTypeID = st.Id

	snCol, _ := app.FindCollectionByNameOrId("skillNames")
	sn := core.NewRecord(snCol)
	sn.Set("name", "Go")
	sn.Set("isEnabled", true)
	sn.Set("skillType", skillTypeID)
	app.Save(sn)
	skillNameID = sn.Id

	aliasCol, _ := app.FindCollectionByNameOrId("skillNameAliases")
	alias := core.NewRecord(aliasCol)
	alias.Set("skillName", skillNameID)
	alias.Set("alias", "golang")
	app.Save(alias)

	return skillTypeID, skillNameID
}

// seedJobPost creates a jobPost with a body that mentions an alias.
func seedJobPost(t *testing.T, app core.App, siteID, jobSiteNumber, body string, postedDate time.Time) {
	t.Helper()
	col, _ := app.FindCollectionByNameOrId("jobPosts")
	jp := core.NewRecord(col)
	jp.Set("jobSiteNumber", jobSiteNumber)
	jp.Set("site", siteID)
	contentJSON, _ := json.Marshal(map[string]string{"title": "Dev Role", "body": body})
	jp.Set("content", json.RawMessage(contentJSON))
	locationJSON, _ := json.Marshal(map[string]string{"city": "Sydney", "country": "AU", "suburb": "CBD"})
	jp.Set("location", json.RawMessage(locationJSON))
	jp.Set("postedDate", postedDate.UTC().Format("2006-01-02 15:04:05.000Z"))
	app.Save(jp)
}

func TestReportRunExcludesOldJobPosts(t *testing.T) {
	pbApp, pbURL := startTestPocketBase(t)
	email, password := createReportingAccount(t, pbApp)

	siteCol, _ := pbApp.FindCollectionByNameOrId("sites")
	site := core.NewRecord(siteCol)
	site.Set("name", "TestSite")
	site.Set("url", "https://test.example.com")
	pbApp.Save(site)

	_, skillNameID := seedSkillData(t, pbApp)

	// Post dated 2 years ago — outside the 13-month window.
	seedJobPost(t, pbApp, site.Id, "JP-OLD-001", "We need a golang developer.", time.Now().AddDate(-2, 0, 0))

	pb, err := pbclient.New(pbURL, email, password)
	if err != nil {
		t.Fatalf("pbclient.New: %v", err)
	}
	if err := report.Run(config.Config{}, pb); err != nil {
		t.Fatalf("report.Run: %v", err)
	}

	rawPb := pocketbaseclient.NewClient(pbURL, pocketbaseclient.WithUserEmailPassword(email, password))
	if err := rawPb.Authorize(); err != nil {
		t.Fatalf("authenticate rawPb: %v", err)
	}
	list, err := rawPb.List("monthlyCountReports", pocketbaseclient.ParamsList{
		Filters: fmt.Sprintf(`skillName = %q`, skillNameID),
	})
	if err != nil {
		t.Fatalf("list monthlyCountReports: %v", err)
	}
	if list.TotalItems != 0 {
		t.Errorf("expected 0 monthlyCountReports for a 2-year-old post, got %d", list.TotalItems)
	}
}

func TestReportRunCreatesMonthlyCountReports(t *testing.T) {
	pbApp, pbURL := startTestPocketBase(t)
	email, password := createReportingAccount(t, pbApp)

	// Seed a site.
	siteCol, _ := pbApp.FindCollectionByNameOrId("sites")
	site := core.NewRecord(siteCol)
	site.Set("name", "TestSite")
	site.Set("url", "https://test.example.com")
	pbApp.Save(site)

	_, skillNameID := seedSkillData(t, pbApp)

	// Seed a job post whose body contains " golang " (word-boundary match).
	// Use a date within the 13-month window so the report filter includes it.
	seedJobPost(t, pbApp, site.Id, "JP-R-001", "We need a golang developer for our team.", time.Now().AddDate(0, -1, 0))

	pb, err := pbclient.New(pbURL, email, password)
	if err != nil {
		t.Fatalf("pbclient.New: %v", err)
	}

	if err := report.Run(config.Config{}, pb); err != nil {
		t.Fatalf("report.Run: %v", err)
	}

	rawPb := pocketbaseclient.NewClient(pbURL, pocketbaseclient.WithUserEmailPassword(email, password))
	if err := rawPb.Authorize(); err != nil {
		t.Fatalf("authenticate rawPb: %v", err)
	}
	list, err := rawPb.List("monthlyCountReports", pocketbaseclient.ParamsList{
		Filters: fmt.Sprintf(`skillName = %q`, skillNameID),
	})
	if err != nil {
		t.Fatalf("list monthlyCountReports: %v", err)
	}
	if list.TotalItems == 0 {
		t.Error("expected at least one monthlyCountReport for skill 'Go', got 0")
	}
	count, _ := list.Items[0]["count"].(float64)
	if int(count) != 1 {
		t.Errorf("expected count=1, got %v", count)
	}
}
