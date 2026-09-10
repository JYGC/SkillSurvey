package migrations

import (
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

// seedMonthlyCountReport creates a monthlyCountReport record directly via the
// internal API (bypassing rules) so the list tests have a record to work with.
func seedMonthlyCountReport(t *testing.T, app core.App) {
	t.Helper()
	col, err := app.FindCollectionByNameOrId("monthlyCountReports")
	if err != nil {
		t.Fatalf("find monthlyCountReports collection: %v", err)
	}
	rec := core.NewRecord(col)
	rec.Set("identifier", "auth-test-2024-10")
	rec.Set("YearMonth", "2024-10")
	rec.Set("yearMonthDate", "2024-10-01 00:00:00.000Z")
	rec.Set("count", 1)
	if err := app.Save(rec); err != nil {
		t.Fatalf("seed monthlyCountReport: %v", err)
	}
}

// TestMonthlyCountReportsUnauthenticatedSeesNoData verifies that the listRule
// hides all records from unauthenticated callers. PocketBase returns HTTP 200
// with 0 items (not 403) when a filter rule evaluates to false — so we check
// the item count rather than the status code.
func TestMonthlyCountReportsUnauthenticatedSeesNoData(t *testing.T) {
	app, serverURL := startTestServer(t)
	seedMonthlyCountReport(t, app)

	total := listTotalItems(t, serverURL, "", "monthlyCountReports")
	if total != 0 {
		t.Errorf("unauthenticated GET monthlyCountReports: expected 0 items, got %d", total)
	}
}

func TestMonthlyCountReportsAuthenticatedCanList(t *testing.T) {
	app, serverURL := startTestServer(t)
	seedMonthlyCountReport(t, app)

	user := createTestUser(t, app, "mcr-reader@example.com", "testtest123")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	status := apiGet(t, serverURL, token, "monthlyCountReports")
	if status != http.StatusOK {
		t.Errorf("authenticated GET monthlyCountReports: expected 200, got %d", status)
	}

	total := listTotalItems(t, serverURL, token, "monthlyCountReports")
	if total == 0 {
		t.Errorf("authenticated GET monthlyCountReports: expected 1+ items, got 0")
	}
}
