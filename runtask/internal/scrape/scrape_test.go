package scrape_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
	pocketbaseclient "github.com/r--w/pocketbase"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/pbclient"
	"keybook/runtask/internal/scrape"
	_ "keybook/pocketbaseserver/migrations"
)

// startTestPocketBase starts a real PocketBase HTTP server with all migrations applied.
func startTestPocketBase(t *testing.T) (core.App, string) {
	t.Helper()

	app := pocketbase.New()
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
		"--dir", t.TempDir(),
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

// createWebscraperAccount creates a service account with the webscraper role.
func createWebscraperAccount(t *testing.T, app core.App) (email, password string) {
	t.Helper()
	email = "scraper-svc@example.com"
	password = "testtest123"

	col, _ := app.FindCollectionByNameOrId("_pb_users_auth_")
	user := core.NewRecord(col)
	user.Set("email", email)
	user.Set("password", password)
	user.Set("passwordConfirm", password)
	user.Set("verified", true)
	if err := app.Save(user); err != nil {
		t.Fatalf("save webscraper account: %v", err)
	}

	role, err := app.FindFirstRecordByData("roles", "name", "webscraper")
	if err != nil {
		t.Fatalf("find webscraper role: %v", err)
	}
	urCol, _ := app.FindCollectionByNameOrId("userRoles")
	ur := core.NewRecord(urCol)
	ur.Set("user", user.Id)
	ur.Set("role", role.Id)
	if err := app.Save(ur); err != nil {
		t.Fatalf("assign webscraper role: %v", err)
	}
	return email, password
}

// seedSite creates a sites record and returns its ID.
func seedSite(t *testing.T, app core.App, name, url string) string {
	t.Helper()
	col, _ := app.FindCollectionByNameOrId("sites")
	site := core.NewRecord(col)
	site.Set("name", name)
	site.Set("url", url)
	if err := app.Save(site); err != nil {
		t.Fatalf("seed site %s: %v", name, err)
	}
	return site.Id
}

// seekStubResponse returns a minimal Seek API response containing one job listing.
func seekStubResponse(stubBaseURL, jobID string) []byte {
	resp := map[string]any{
		"data": []map[string]any{
			{
				"id":          jobID,
				"listingDate": "2024-01-15T10:00:00Z",
				"locations": []map[string]any{
					{"countryCode": "AU", "label": "Sydney CBD"},
				},
			},
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

// makeSeekConfig writes a minimal seek adapter config JSON to a temp file and returns its path.
func makeSeekConfig(t *testing.T, stubAPIURL, stubBaseURL, siteName string) string {
	t.Helper()
	cfg := map[string]any{
		"BaseUrl":        stubBaseURL,
		"SearchApiUrl":   stubAPIURL,
		"Pages":          1,
		"SiteSelectors":  map[string]any{"SiteName": siteName, "TitleSelector": "h1", "BodySelector": "p"},
		"ApiParameters":  []map[string]any{{"NewSinceDaysAgo": 1}},
		"AllowedDomains": []string{},
	}
	b, _ := json.Marshal(cfg)
	p := filepath.Join(t.TempDir(), "seek.json")
	os.WriteFile(p, b, 0644)
	return p
}

func TestScrapeRunCreatesJobPosts(t *testing.T) {
	pbApp, pbURL := startTestPocketBase(t)
	email, password := createWebscraperAccount(t, pbApp)

	const siteName = "Seek"
	const jobID = "SEEK-TEST-001"

	// Stub server returning a seek-style job listing.
	stubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(seekStubResponse(r.Host, jobID))
	}))
	defer stubServer.Close()

	seedSite(t, pbApp, siteName, stubServer.URL)
	seekCfgPath := makeSeekConfig(t, stubServer.URL+"/search", stubServer.URL, siteName)

	rawPb := pocketbaseclient.NewClient(pbURL)
	if _, err := rawPb.Authenticate(email, password); err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	pb, err := pbclient.New(pbURL, email, password)
	if err != nil {
		t.Fatalf("pbclient.New: %v", err)
	}

	cfg := config.Config{
		SeekConfigFile: seekCfgPath,
		ErrorLogFile:   filepath.Join(t.TempDir(), "error.log"),
	}

	if err := scrape.Run(cfg, pb); err != nil {
		t.Logf("scrape.Run returned (non-fatal) error: %v", err)
	}

	list, err := rawPb.List("jobPosts", pocketbaseclient.ParamsList{})
	if err != nil {
		t.Fatalf("list jobPosts: %v", err)
	}
	if list.TotalItems == 0 {
		t.Error("expected at least one jobPost record after scrape, got 0")
	}
}

func TestScrapeRunIsIdempotent(t *testing.T) {
	pbApp, pbURL := startTestPocketBase(t)
	email, password := createWebscraperAccount(t, pbApp)

	const siteName = "Seek"
	const jobID = "SEEK-TEST-002"

	stubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(seekStubResponse(r.Host, jobID))
	}))
	defer stubServer.Close()

	seedSite(t, pbApp, siteName, stubServer.URL)
	seekCfgPath := makeSeekConfig(t, stubServer.URL+"/search", stubServer.URL, siteName)

	pb, err := pbclient.New(pbURL, email, password)
	if err != nil {
		t.Fatalf("pbclient.New: %v", err)
	}

	cfg := config.Config{
		SeekConfigFile: seekCfgPath,
		ErrorLogFile:   filepath.Join(t.TempDir(), "error.log"),
	}

	scrape.Run(cfg, pb)
	scrape.Run(cfg, pb)

	rawPb := pocketbaseclient.NewClient(pbURL)
	rawPb.Authenticate(email, password)
	list, err := rawPb.List("jobPosts", pocketbaseclient.ParamsList{})
	if err != nil {
		t.Fatalf("list jobPosts: %v", err)
	}
	if list.TotalItems != 1 {
		t.Errorf("expected exactly 1 jobPost after two scrape runs, got %d", list.TotalItems)
	}
}
