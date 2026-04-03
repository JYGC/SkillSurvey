package migrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

// startTestServer starts a real PocketBase HTTP server in a background goroutine.
// All migrations in this package have already been registered via init() functions.
// Returns the base URL and a cleanup function.
func startTestServer(t *testing.T) (core.App, string) {
	t.Helper()

	app := pocketbase.New()
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{})

	// Find a free port
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("find free port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()

	serverURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Capture the http.Server so we can shut it down cleanly.
	var httpServer *http.Server
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		httpServer = e.Server
		return e.Next()
	})

	app.RootCmd.SetArgs([]string{
		"--dir", t.TempDir(),
		"serve",
		"--http", fmt.Sprintf("127.0.0.1:%d", port),
	})
	go func() { _ = app.Start() }()

	// Wait up to 10s for the server to respond.
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

// createTestUser creates a regular (non-admin) user via the app's internal API.
func createTestUser(t *testing.T, app core.App, email, password string) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("_pb_users_auth_")
	if err != nil {
		t.Fatalf("find users collection: %v", err)
	}
	user := core.NewRecord(col)
	user.Set("email", email)
	user.Set("password", password)
	user.Set("passwordConfirm", password)
	user.Set("verified", true)
	if err := app.Save(user); err != nil {
		t.Fatalf("save user %s: %v", email, err)
	}
	return user
}

// assignRole assigns a named role to a user via the app's internal API.
func assignRole(t *testing.T, app core.App, userID, roleName string) {
	t.Helper()
	role, err := app.FindFirstRecordByData("roles", "name", roleName)
	if err != nil {
		t.Fatalf("find role %q: %v", roleName, err)
	}
	col, err := app.FindCollectionByNameOrId("userRoles")
	if err != nil {
		t.Fatalf("find userRoles collection: %v", err)
	}
	ur := core.NewRecord(col)
	ur.Set("user", userID)
	ur.Set("role", role.Id)
	if err := app.Save(ur); err != nil {
		t.Fatalf("assign role %q to user %s: %v", roleName, userID, err)
	}
}

// authToken authenticates as a user via HTTP and returns the bearer token.
func authToken(t *testing.T, serverURL, email, password string) string {
	t.Helper()
	body := fmt.Sprintf(`{"identity":%q,"password":%q}`, email, password)
	resp, err := http.Post(
		serverURL+"/api/collections/users/auth-with-password",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		t.Fatalf("auth request: %v", err)
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode auth response: %v", err)
	}
	token, ok := result["token"].(string)
	if !ok || token == "" {
		t.Fatalf("no token in auth response for %s: %v", email, result)
	}
	return token
}

// apiPost makes an authenticated POST to a PocketBase collection API.
// Returns the HTTP status code.
func apiPost(t *testing.T, serverURL, token, collection string, body map[string]any) int {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, serverURL+"/api/collections/"+collection+"/records", strings.NewReader(string(b)))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", collection, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func TestNoRoleCannotWriteJobPostsOrMonthlyCountReports(t *testing.T) {
	app, serverURL := startTestServer(t)
	_ = app

	user := createTestUser(t, app, "norole@example.com", "testtest123")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	// Must find a valid site ID to POST a jobPost.
	site, err := app.FindFirstRecordByData("sites", "name", "")
	if err != nil {
		// No sites seeded — create one via admin (internal API bypasses rules).
		siteCol, _ := app.FindCollectionByNameOrId("sites")
		site = core.NewRecord(siteCol)
		site.Set("name", "TestSite")
		site.Set("url", "https://test.example.com")
		_ = app.Save(site)
	}

	status := apiPost(t, serverURL, token, "jobPosts", map[string]any{
		"jobSiteNumber": "NOROLE-001",
		"site":          site.Id,
		"content":       map[string]any{"title": "t", "body": "b"},
		"location":      map[string]any{"city": "c", "country": "au", "suburb": "s"},
	})
	if status != http.StatusForbidden {
		t.Errorf("expected 403 for jobPosts POST without role, got %d", status)
	}

	status = apiPost(t, serverURL, token, "monthlyCountReports", map[string]any{
		"identifier":    "norole_2024-01",
		"YearMonth":     "2024-01",
		"yearMonthDate": "2024-01-01 00:00:00",
		"count":         1,
	})
	if status != http.StatusForbidden {
		t.Errorf("expected 403 for monthlyCountReports POST without role, got %d", status)
	}
}

func TestWebscraperRoleCanPostJobPost(t *testing.T) {
	app, serverURL := startTestServer(t)

	user := createTestUser(t, app, "webscraper@example.com", "testtest123")
	assignRole(t, app, user.Id, "webscraper")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	siteCol, _ := app.FindCollectionByNameOrId("sites")
	site := core.NewRecord(siteCol)
	site.Set("name", "SeekSite")
	site.Set("url", "https://seek.example.com")
	_ = app.Save(site)

	status := apiPost(t, serverURL, token, "jobPosts", map[string]any{
		"jobSiteNumber": "SEEK-001",
		"site":          site.Id,
		"content":       map[string]any{"title": "Engineer", "body": "Great role"},
		"location":      map[string]any{"city": "Sydney", "country": "AU", "suburb": "CBD"},
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for jobPosts POST with webscraper role, got %d", status)
	}
}

func TestReportingRoleCanPostMonthlyCountReport(t *testing.T) {
	app, serverURL := startTestServer(t)

	user := createTestUser(t, app, "reporting@example.com", "testtest123")
	assignRole(t, app, user.Id, "reporting")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	status := apiPost(t, serverURL, token, "monthlyCountReports", map[string]any{
		"identifier":    "reporting_2024-01",
		"YearMonth":     "2024-01",
		"yearMonthDate": "2024-01-01 00:00:00",
		"count":         5,
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for monthlyCountReports POST with reporting role, got %d", status)
	}
}

func TestMigrationRoleCanPostToAllCollections(t *testing.T) {
	app, serverURL := startTestServer(t)

	user := createTestUser(t, app, "migration@example.com", "testtest123")
	assignRole(t, app, user.Id, "migration")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	// sites
	status := apiPost(t, serverURL, token, "sites", map[string]any{
		"name": "MigSite",
		"url":  "https://mig.example.com",
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for sites POST with migration role, got %d", status)
	}

	// skillTypes
	status = apiPost(t, serverURL, token, "skillTypes", map[string]any{
		"name":        "Programming",
		"description": "Programming languages",
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for skillTypes POST with migration role, got %d", status)
	}

	skillTypeID := ""
	stRecs, err := app.FindRecordsByFilter("skillTypes", "name='Programming'", "-created", 1, 0)
	if err == nil && len(stRecs) > 0 {
		skillTypeID = stRecs[0].Id
	}

	// skillNames
	status = apiPost(t, serverURL, token, "skillNames", map[string]any{
		"name":      "Go",
		"isEnabled": true,
		"skillType": skillTypeID,
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for skillNames POST with migration role, got %d", status)
	}

	skillNameID := ""
	snRecs, err := app.FindRecordsByFilter("skillNames", "name='Go'", "-created", 1, 0)
	if err == nil && len(snRecs) > 0 {
		skillNameID = snRecs[0].Id
	}

	// skillNameAliases
	status = apiPost(t, serverURL, token, "skillNameAliases", map[string]any{
		"skillName": skillNameID,
		"alias":     "golang",
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for skillNameAliases POST with migration role, got %d", status)
	}

	// jobPosts — need a site
	siteRecs, _ := app.FindRecordsByFilter("sites", "name='MigSite'", "-created", 1, 0)
	siteID := siteRecs[0].Id
	status = apiPost(t, serverURL, token, "jobPosts", map[string]any{
		"jobSiteNumber": "MIG-001",
		"site":          siteID,
		"content":       map[string]any{"title": "Dev", "body": "role body"},
		"location":      map[string]any{"city": "Melbourne", "country": "AU", "suburb": "CBD"},
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for jobPosts POST with migration role, got %d", status)
	}

	// monthlyCountReports
	status = apiPost(t, serverURL, token, "monthlyCountReports", map[string]any{
		"identifier":    "mig_2024-01",
		"YearMonth":     "2024-01",
		"yearMonthDate": "2024-01-01 00:00:00",
		"count":         10,
	})
	if status != http.StatusOK {
		t.Errorf("expected 200 for monthlyCountReports POST with migration role, got %d", status)
	}
}

func TestMigrationRoleCannotWriteUsersUserRolesOrRoles(t *testing.T) {
	app, serverURL := startTestServer(t)

	user := createTestUser(t, app, "mignorestricted@example.com", "testtest123")
	assignRole(t, app, user.Id, "migration")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	// Cannot POST to users
	status := apiPost(t, serverURL, token, "users", map[string]any{
		"email":           "new@example.com",
		"password":        "testtest123",
		"passwordConfirm": "testtest123",
	})
	if status != http.StatusForbidden {
		t.Errorf("expected 403 for users POST with migration role, got %d", status)
	}

	// Cannot POST to userRoles
	status = apiPost(t, serverURL, token, "userRoles", map[string]any{
		"user": user.Id,
		"role": user.Id, // arbitrary; rule check fires before validation
	})
	if status != http.StatusForbidden {
		t.Errorf("expected 403 for userRoles POST with migration role, got %d", status)
	}

	// Cannot POST to roles
	status = apiPost(t, serverURL, token, "roles", map[string]any{
		"name":        "hacker",
		"description": "bad actor",
	})
	if status != http.StatusForbidden {
		t.Errorf("expected 403 for roles POST with migration role, got %d", status)
	}
}

func TestDuplicateUserRoleIsRejected(t *testing.T) {
	app, _ := startTestServer(t)

	user := createTestUser(t, app, "dup@example.com", "testtest123")

	// Assign the same role twice — second save must fail.
	assignRole(t, app, user.Id, "webscraper")

	role, _ := app.FindFirstRecordByData("roles", "name", "webscraper")
	col, _ := app.FindCollectionByNameOrId("userRoles")
	ur2 := core.NewRecord(col)
	ur2.Set("user", user.Id)
	ur2.Set("role", role.Id)
	err := app.Save(ur2)
	if err == nil {
		t.Error("expected error when inserting duplicate (user, role) pair, got nil")
	}
}
