package migrations

import (
	"net/http"
	"testing"
)

func TestMonthlyCountReportsUnauthenticatedIsDenied(t *testing.T) {
	_, serverURL := startTestServer(t)

	status := apiGet(t, serverURL, "", "monthlyCountReports")
	if status != http.StatusForbidden {
		t.Errorf("unauthenticated GET monthlyCountReports: expected 403, got %d", status)
	}
}

func TestMonthlyCountReportsAuthenticatedCanList(t *testing.T) {
	app, serverURL := startTestServer(t)

	user := createTestUser(t, app, "mcr-reader@example.com", "testtest123")
	token := authToken(t, serverURL, user.GetString("email"), "testtest123")

	status := apiGet(t, serverURL, token, "monthlyCountReports")
	if status != http.StatusOK {
		t.Errorf("authenticated GET monthlyCountReports: expected 200, got %d", status)
	}
}
