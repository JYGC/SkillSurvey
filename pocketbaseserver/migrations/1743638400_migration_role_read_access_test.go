package migrations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func apiGet(t *testing.T, serverURL, token, collection string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, serverURL+"/api/collections/"+collection+"/records", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", collection, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func apiGetOne(t *testing.T, serverURL, token, collection, id string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/collections/%s/records/%s", serverURL, collection, id), nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s/%s: %v", collection, id, err)
	}
	defer resp.Body.Close()
	return resp.StatusCode
}

func listTotalItems(t *testing.T, serverURL, token, collection string) int {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, serverURL+"/api/collections/"+collection+"/records", nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", collection, err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(b, &result)
	total, _ := result["totalItems"].(float64)
	return int(total)
}

// seedSite creates a sites record and returns it.
func seedSite(t *testing.T, app core.App, name, url string) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("sites")
	if err != nil {
		t.Fatalf("find sites collection: %v", err)
	}
	rec := core.NewRecord(col)
	rec.Set("name", name)
	rec.Set("url", url)
	if err := app.Save(rec); err != nil {
		t.Fatalf("save site: %v", err)
	}
	return rec
}

// seedSkillType creates a skillTypes record and returns it.
func seedSkillType(t *testing.T, app core.App, name, description string) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("skillTypes")
	if err != nil {
		t.Fatalf("find skillTypes collection: %v", err)
	}
	rec := core.NewRecord(col)
	rec.Set("name", name)
	rec.Set("description", description)
	if err := app.Save(rec); err != nil {
		t.Fatalf("save skillType: %v", err)
	}
	return rec
}

// seedSkillName creates an enabled skillNames record under skillTypeID and returns it.
func seedSkillName(t *testing.T, app core.App, name, skillTypeID string) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("skillNames")
	if err != nil {
		t.Fatalf("find skillNames collection: %v", err)
	}
	rec := core.NewRecord(col)
	rec.Set("name", name)
	rec.Set("isEnabled", true)
	rec.Set("skillType", skillTypeID)
	if err := app.Save(rec); err != nil {
		t.Fatalf("save skillName: %v", err)
	}
	return rec
}

// seedSkillNameAlias creates a skillNameAliases record under skillNameID and returns it.
func seedSkillNameAlias(t *testing.T, app core.App, skillNameID, alias string) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("skillNameAliases")
	if err != nil {
		t.Fatalf("find skillNameAliases collection: %v", err)
	}
	rec := core.NewRecord(col)
	rec.Set("skillName", skillNameID)
	rec.Set("alias", alias)
	if err := app.Save(rec); err != nil {
		t.Fatalf("save skillNameAlias: %v", err)
	}
	return rec
}

// seedJobPost creates a jobPosts record under siteID with placeholder content/location and returns it.
func seedJobPost(t *testing.T, app core.App, jobSiteNumber, siteID string) *core.Record {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("jobPosts")
	if err != nil {
		t.Fatalf("find jobPosts collection: %v", err)
	}
	rec := core.NewRecord(col)
	rec.Set("jobSiteNumber", jobSiteNumber)
	rec.Set("site", siteID)
	rec.Set("content", map[string]any{"title": "t", "body": "b"})
	rec.Set("location", map[string]any{"city": "c", "country": "au", "suburb": "s"})
	if err := app.Save(rec); err != nil {
		t.Fatalf("save jobPost: %v", err)
	}
	return rec
}

func TestMigrationRoleCanListCollections(t *testing.T) {
	app, serverURL := startTestServer(t)

	site := seedSite(t, app, "ReadTestSite", "https://read.example.com")
	st := seedSkillType(t, app, "Programming", "Programming languages")
	sn := seedSkillName(t, app, "Go", st.Id)
	alias := seedSkillNameAlias(t, app, sn.Id, "golang")
	jp := seedJobPost(t, app, "READ-001", site.Id)

	user := createTestUser(t, app, "migration-reader@example.com", "testtest123")
	assignRole(t, app, user.Id, "migration")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	// List access.
	for _, col := range []string{"sites", "skillTypes", "skillNames", "skillNameAliases", "jobPosts"} {
		status := apiGet(t, serverURL, token, col)
		if status != http.StatusOK {
			t.Errorf("expected 200 for GET %s with migration role, got %d", col, status)
		}
	}

	// Verify seeded records are visible.
	for col, id := range map[string]string{
		"sites":            site.Id,
		"skillTypes":       st.Id,
		"skillNames":       sn.Id,
		"skillNameAliases": alias.Id,
		"jobPosts":         jp.Id,
	} {
		status := apiGetOne(t, serverURL, token, col, id)
		if status != http.StatusOK {
			t.Errorf("expected 200 for view %s/%s with migration role, got %d", col, id, status)
		}
	}
}

func TestMigrationRoleListReturnsSeededRecords(t *testing.T) {
	app, serverURL := startTestServer(t)

	stCol, _ := app.FindCollectionByNameOrId("skillTypes")
	st := core.NewRecord(stCol)
	st.Set("name", "DevOps")
	st.Set("description", "DevOps tools")
	app.Save(st)

	user := createTestUser(t, app, "migration-list@example.com", "testtest123")
	assignRole(t, app, user.Id, "migration")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	total := listTotalItems(t, serverURL, token, "skillTypes")
	if total == 0 {
		t.Error("expected migration role to see seeded skillTypes, got 0 items")
	}
}

func TestNoRoleCannotListSkillCollections(t *testing.T) {
	app, serverURL := startTestServer(t)

	stCol, _ := app.FindCollectionByNameOrId("skillTypes")
	st := core.NewRecord(stCol)
	st.Set("name", "Testing")
	st.Set("description", "Testing tools")
	app.Save(st)

	user := createTestUser(t, app, "norole-reader@example.com", "testtest123")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	// skillTypes and skillNameAliases require migration role to list/view.
	for _, col := range []string{"skillTypes", "skillNameAliases"} {
		total := listTotalItems(t, serverURL, token, col)
		if total != 0 {
			t.Errorf("expected no-role user to see 0 %s, got %d", col, total)
		}
	}

	// View of a specific skillType record should return 404 (rule filters it out).
	status := apiGetOne(t, serverURL, token, "skillTypes", st.Id)
	if status != http.StatusNotFound {
		t.Errorf("expected 404 viewing skillTypes record without migration role, got %d", status)
	}
}
