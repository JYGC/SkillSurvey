package migrations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestDiagnosticRule(t *testing.T) {
	app, serverURL := startTestServer(t)

	// Create no-role user
	col, _ := app.FindCollectionByNameOrId("_pb_users_auth_")
	user := core.NewRecord(col)
	user.Set("email", "diag@example.com")
	user.Set("password", "testtest123")
	user.Set("passwordConfirm", "testtest123")
	user.Set("verified", true)
	if err := app.Save(user); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Auth
	body := `{"identity":"diag@example.com","password":"testtest123"}`
	resp, _ := http.Post(serverURL+"/api/collections/users/auth-with-password", "application/json", strings.NewReader(body))
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	var auth map[string]any
	json.Unmarshal(data, &auth)
	token, _ := auth["token"].(string)
	t.Logf("token: %s", token)

	// Create site
	siteCol, _ := app.FindCollectionByNameOrId("sites")
	site := core.NewRecord(siteCol)
	site.Set("name", "DiagSite")
	site.Set("url", "https://diag.example.com")
	app.Save(site)

	// POST to jobPosts
	postBody, _ := json.Marshal(map[string]any{
		"jobSiteNumber": "DIAG-001",
		"site":          site.Id,
		"content":       map[string]any{"title": "t", "body": "b"},
		"location":      map[string]any{"city": "c", "country": "au", "suburb": "s"},
	})
	req, _ := http.NewRequest(http.MethodPost, serverURL+"/api/collections/jobPosts/records", strings.NewReader(string(postBody)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	t.Logf("status: %d", resp2.StatusCode)
	t.Logf("body: %s", body2)
	fmt.Println("STATUS:", resp2.StatusCode)
	fmt.Println("BODY:", string(body2))
}
