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

func TestMigrationRoleCanListCollections(t *testing.T) {
	app, serverURL := startTestServer(t)

	// Seed a site (needed for jobPosts relation).
	siteCol, _ := app.FindCollectionByNameOrId("sites")
	site := core.NewRecord(siteCol)
	site.Set("name", "ReadTestSite")
	site.Set("url", "https://read.example.com")
	app.Save(site)

	// Seed a skillType.
	stCol, _ := app.FindCollectionByNameOrId("skillTypes")
	st := core.NewRecord(stCol)
	st.Set("name", "Programming")
	st.Set("description", "Programming languages")
	app.Save(st)

	// Seed a skillName.
	snCol, _ := app.FindCollectionByNameOrId("skillNames")
	sn := core.NewRecord(snCol)
	sn.Set("name", "Go")
	sn.Set("isEnabled", true)
	sn.Set("skillType", st.Id)
	app.Save(sn)

	// Seed a skillNameAlias.
	aliasCol, _ := app.FindCollectionByNameOrId("skillNameAliases")
	alias := core.NewRecord(aliasCol)
	alias.Set("skillName", sn.Id)
	alias.Set("alias", "golang")
	app.Save(alias)

	// Seed a jobPost.
	jpCol, _ := app.FindCollectionByNameOrId("jobPosts")
	jp := core.NewRecord(jpCol)
	jp.Set("jobSiteNumber", "READ-001")
	jp.Set("site", site.Id)
	jp.Set("content", map[string]any{"title": "t", "body": "b"})
	jp.Set("location", map[string]any{"city": "c", "country": "au", "suburb": "s"})
	app.Save(jp)

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
		"sites":             site.Id,
		"skillTypes":        st.Id,
		"skillNames":        sn.Id,
		"skillNameAliases":  alias.Id,
		"jobPosts":          jp.Id,
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
