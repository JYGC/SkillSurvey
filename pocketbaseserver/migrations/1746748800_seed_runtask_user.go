package migrations

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const runtaskUserEmail = "runtask@skillsurvey.com"

func init() {
	m.Register(func(app core.App) error {
		return seedRuntaskUser(app)
	}, func(app core.App) error {
		return removeRuntaskUser(app)
	})
}

// ── up helpers ───────────────────────────────────────────────────────────────

func seedRuntaskUser(app core.App) error {
	user, err := app.FindFirstRecordByData("users", "email", runtaskUserEmail)
	if err != nil {
		// Not found — create the service account.
		usersCol, err := app.FindCollectionByNameOrId("_pb_users_auth_")
		if err != nil {
			return err
		}
		password, err := runtaskRandomPassword(32)
		if err != nil {
			return err
		}
		user = core.NewRecord(usersCol)
		user.Set("email", runtaskUserEmail)
		user.Set("password", password)
		user.Set("passwordConfirm", password)
		user.Set("verified", true)
		if err := app.Save(user); err != nil {
			return err
		}
	}

	for _, roleName := range []string{"webscraper", "reporting"} {
		if err := ensureRuntaskUserRole(app, user.Id, roleName); err != nil {
			return err
		}
	}
	return nil
}

func ensureRuntaskUserRole(app core.App, userID, roleName string) error {
	role, err := app.FindFirstRecordByData("roles", "name", roleName)
	if err != nil {
		return fmt.Errorf("find role %q: %w", roleName, err)
	}

	// Skip if already linked.
	existing, err := app.FindRecordsByFilter(
		"userRoles",
		"user='"+userID+"' && role='"+role.Id+"'",
		"", 1, 0,
	)
	if err == nil && len(existing) > 0 {
		return nil
	}

	userRolesCol, err := app.FindCollectionByNameOrId("userRoles")
	if err != nil {
		return err
	}
	ur := core.NewRecord(userRolesCol)
	ur.Set("user", userID)
	ur.Set("role", role.Id)
	return app.Save(ur)
}

func runtaskRandomPassword(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}

// ── down helper ──────────────────────────────────────────────────────────────

func removeRuntaskUser(app core.App) error {
	user, err := app.FindFirstRecordByData("users", "email", runtaskUserEmail)
	if err != nil {
		return nil // already absent
	}

	userRoles, err := app.FindRecordsByFilter(
		"userRoles", "user='"+user.Id+"'", "", 0, 0,
	)
	if err == nil {
		for _, ur := range userRoles {
			if err := app.Delete(ur); err != nil {
				return err
			}
		}
	}

	return app.Delete(user)
}
