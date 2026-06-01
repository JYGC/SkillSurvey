package migrations

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestSkillNamesPublicRead(t *testing.T) {
	app, serverURL := startTestServer(t)

	// Seed a skill type and skill name.
	stCol, _ := app.FindCollectionByNameOrId("skillTypes")
	st := core.NewRecord(stCol)
	st.Set("name", "Language")
	st.Set("description", "Programming languages")
	app.Save(st)

	snCol, _ := app.FindCollectionByNameOrId("skillNames")
	sn := core.NewRecord(snCol)
	sn.Set("name", "TypeScript")
	sn.Set("isEnabled", true)
	sn.Set("skillType", st.Id)
	app.Save(sn)

	// Unauthenticated list of skillNames must return 200.
	status := apiGet(t, serverURL, "", "skillNames")
	if status != http.StatusOK {
		t.Errorf("unauthenticated GET skillNames: expected 200, got %d", status)
	}

	// Unauthenticated view of a specific skillName record must return 200.
	status = apiGetOne(t, serverURL, "", "skillNames", sn.Id)
	if status != http.StatusOK {
		t.Errorf("unauthenticated GET skillNames/%s: expected 200, got %d", sn.Id, status)
	}
}

func TestMonthlyCountReportExpandSkillNameUnauthenticated(t *testing.T) {
	app, serverURL := startTestServer(t)

	// Seed skill type, skill name, and a monthly count report.
	stCol, _ := app.FindCollectionByNameOrId("skillTypes")
	st := core.NewRecord(stCol)
	st.Set("name", "Language")
	st.Set("description", "Programming languages")
	app.Save(st)

	snCol, _ := app.FindCollectionByNameOrId("skillNames")
	sn := core.NewRecord(snCol)
	sn.Set("name", "Go")
	sn.Set("isEnabled", true)
	sn.Set("skillType", st.Id)
	app.Save(sn)

	mcrCol, _ := app.FindCollectionByNameOrId("monthlyCountReports")
	mcr := core.NewRecord(mcrCol)
	mcr.Set("identifier", "Go-2024-10")
	mcr.Set("YearMonth", "2024-10")
	mcr.Set("yearMonthDate", "2024-10-01 00:00:00.000Z")
	mcr.Set("count", 42)
	mcr.Set("skillName", sn.Id)
	app.Save(mcr)

	// Unauthenticated GET with expand=skillName must include skill name in response.
	url := fmt.Sprintf("%s/api/collections/monthlyCountReports/records?expand=skillName", serverURL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET monthlyCountReports: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Items []struct {
			Expand struct {
				SkillName struct {
					Name string `json:"name"`
				} `json:"skillName"`
			} `json:"expand"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(result.Items) == 0 {
		t.Fatal("expected at least one monthlyCountReport item")
	}
	gotName := result.Items[0].Expand.SkillName.Name
	if gotName != "Go" {
		t.Errorf("expected expand.skillName.name = %q, got %q", "Go", gotName)
	}
}
