package migrations

import (
	"testing"
)

func TestRuntaskUserExistsAfterMigration(t *testing.T) {
	app, _ := startTestServer(t)

	user, err := app.FindFirstRecordByData("users", "email", runtaskUserEmail)
	if err != nil {
		t.Fatalf("runtask user not found: %v", err)
	}
	if !user.GetBool("verified") {
		t.Error("runtask user should be verified")
	}

	userRoles, err := app.FindRecordsByFilter(
		"userRoles", "user='"+user.Id+"'", "", 0, 0,
	)
	if err != nil {
		t.Fatalf("find userRoles: %v", err)
	}

	roleNames := map[string]bool{}
	for _, ur := range userRoles {
		role, err := app.FindRecordById("roles", ur.GetString("role"))
		if err == nil {
			roleNames[role.GetString("name")] = true
		}
	}

	for _, expected := range []string{"webscraper", "reporting"} {
		if !roleNames[expected] {
			t.Errorf("runtask user missing role %q", expected)
		}
	}
}

func TestRuntaskUserSeedIsIdempotent(t *testing.T) {
	app, _ := startTestServer(t)

	// Running the up function a second time must not create duplicates.
	if err := seedRuntaskUser(app); err != nil {
		t.Fatalf("second seedRuntaskUser call: %v", err)
	}

	users, err := app.FindRecordsByFilter(
		"users", "email='"+runtaskUserEmail+"'", "", 0, 0,
	)
	if err != nil {
		t.Fatalf("find users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("expected 1 runtask user, got %d", len(users))
	}

	userRoles, err := app.FindRecordsByFilter(
		"userRoles", "user='"+users[0].Id+"'", "", 0, 0,
	)
	if err != nil {
		t.Fatalf("find userRoles: %v", err)
	}
	if len(userRoles) != 2 {
		t.Errorf("expected 2 userRoles for runtask user, got %d", len(userRoles))
	}
}
