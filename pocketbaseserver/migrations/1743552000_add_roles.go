package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// Step 1: create roles collection.
		if err := createRoles(app); err != nil {
			return err
		}

		// Step 2: seed role records.
		if err := seedRoles(app); err != nil {
			return err
		}

		// Step 3: create userRoles collection.
		if err := createUserRoles(app); err != nil {
			return err
		}

		// Step 4: apply access rules to existing collections.
		return applyCollectionRules(app)
	}, func(app core.App) error {
		// Down step 1: revert access rules on existing collections.
		if err := revertCollectionRules(app); err != nil {
			return err
		}

		// Down step 2: delete seed records from roles.
		if err := deleteSeedRoles(app); err != nil {
			return err
		}

		// Down step 3: delete userRoles collection.
		if err := dropCollection(app, "userRoles"); err != nil {
			return err
		}

		// Down step 4: delete roles collection.
		return dropCollection(app, "roles")
	})
}

// ── up helpers ───────────────────────────────────────────────────────────────

func createRoles(app core.App) error {
	roles := core.NewBaseCollection("roles")
	roles.ListRule = types.Pointer(`@request.auth.id != ""`)
	roles.ViewRule = types.Pointer(`@request.auth.id != ""`)
	// CreateRule / UpdateRule / DeleteRule left nil → superadmin only
	roles.Indexes = []string{
		"CREATE UNIQUE INDEX idx_roles_name ON roles (name ASC)",
	}
	roles.Fields.Add(
		&core.TextField{
			Name:     "name",
			Required: true,
		},
		&core.TextField{
			Name:     "description",
			Required: true,
		},
	)
	return app.Save(roles)
}

func seedRoles(app core.App) error {
	rolesCol, err := app.FindCollectionByNameOrId("roles")
	if err != nil {
		return err
	}

	seeds := []struct{ name, description string }{
		{"webscraper", "Write access to jobPosts"},
		{"reporting", "Write access to monthlyCountReports"},
		{"migration", "Write access to all collections except users, userRoles, and roles"},
	}

	for _, seed := range seeds {
		rec := core.NewRecord(rolesCol)
		rec.Set("name", seed.name)
		rec.Set("description", seed.description)
		if err := app.Save(rec); err != nil {
			return err
		}
	}
	return nil
}

func createUserRoles(app core.App) error {
	rolesCol, err := app.FindCollectionByNameOrId("roles")
	if err != nil {
		return err
	}

	userRoles := core.NewBaseCollection("userRoles")
	userRoles.ListRule = types.Pointer(`@request.auth.id != ""`)
	userRoles.ViewRule = types.Pointer(`@request.auth.id != ""`)
	// CreateRule / UpdateRule / DeleteRule left nil → superadmin only
	userRoles.Indexes = []string{
		"CREATE UNIQUE INDEX idx_userRoles_user_role ON userRoles (user ASC, role ASC)",
	}
	userRoles.Fields.Add(
		&core.RelationField{
			Name:         "user",
			CollectionId: "_pb_users_auth_",
			MaxSelect:    1,
			Required:     true,
		},
		&core.RelationField{
			Name:         "role",
			CollectionId: rolesCol.Id,
			MaxSelect:    1,
			Required:     true,
		},
	)
	return app.Save(userRoles)
}

const (
	writeRuleReporting  = `@request.auth.id != "" && @collection.userRoles.user = @request.auth.id && (@collection.userRoles.role.name = 'reporting' || @collection.userRoles.role.name = 'migration')`
	writeRuleWebscraper = `@request.auth.id != "" && @collection.userRoles.user = @request.auth.id && (@collection.userRoles.role.name = 'webscraper' || @collection.userRoles.role.name = 'migration')`
	writeRuleMigration  = `@request.auth.id != "" && @collection.userRoles.user = @request.auth.id && @collection.userRoles.role.name = 'migration'`
)

func applyCollectionRules(app core.App) error {
	type ruleSet struct {
		name       string
		listRule   *string // nil = unchanged
		viewRule   *string
		createRule *string
		updateRule *string
		deleteRule *string
	}

	empty := types.Pointer("")

	rules := []ruleSet{
		{
			name:       "monthlyCountReports",
			listRule:   empty,
			viewRule:   empty,
			createRule: types.Pointer(writeRuleReporting),
			updateRule: types.Pointer(writeRuleReporting),
			deleteRule: types.Pointer(writeRuleReporting),
		},
		{
			name:       "jobPosts",
			createRule: types.Pointer(writeRuleWebscraper),
			updateRule: types.Pointer(writeRuleWebscraper),
			deleteRule: types.Pointer(writeRuleWebscraper),
		},
		{
			name:       "skillTypes",
			createRule: types.Pointer(writeRuleMigration),
			updateRule: types.Pointer(writeRuleMigration),
			deleteRule: types.Pointer(writeRuleMigration),
		},
		{
			name:       "skillNames",
			createRule: types.Pointer(writeRuleMigration),
			updateRule: types.Pointer(writeRuleMigration),
			deleteRule: types.Pointer(writeRuleMigration),
		},
		{
			name:       "skillNameAliases",
			createRule: types.Pointer(writeRuleMigration),
			updateRule: types.Pointer(writeRuleMigration),
			deleteRule: types.Pointer(writeRuleMigration),
		},
		{
			name:       "sites",
			createRule: types.Pointer(writeRuleMigration),
			updateRule: types.Pointer(writeRuleMigration),
			deleteRule: types.Pointer(writeRuleMigration),
		},
	}

	for _, rs := range rules {
		col, err := app.FindCollectionByNameOrId(rs.name)
		if err != nil {
			return err
		}
		if rs.listRule != nil {
			col.ListRule = rs.listRule
		}
		if rs.viewRule != nil {
			col.ViewRule = rs.viewRule
		}
		if rs.createRule != nil {
			col.CreateRule = rs.createRule
		}
		if rs.updateRule != nil {
			col.UpdateRule = rs.updateRule
		}
		if rs.deleteRule != nil {
			col.DeleteRule = rs.deleteRule
		}
		if err := app.Save(col); err != nil {
			return err
		}
	}
	return nil
}

// ── down helpers ─────────────────────────────────────────────────────────────

func revertCollectionRules(app core.App) error {
	collections := []string{
		"monthlyCountReports",
		"jobPosts",
		"skillTypes",
		"skillNames",
		"skillNameAliases",
		"sites",
	}
	for _, name := range collections {
		col, err := app.FindCollectionByNameOrId(name)
		if err != nil {
			return err
		}
		col.CreateRule = nil
		col.UpdateRule = nil
		col.DeleteRule = nil
		if name == "monthlyCountReports" {
			col.ListRule = nil
			col.ViewRule = nil
		}
		if err := app.Save(col); err != nil {
			return err
		}
	}
	return nil
}

func deleteSeedRoles(app core.App) error {
	for _, name := range []string{"webscraper", "reporting", "migration"} {
		rec, err := app.FindFirstRecordByData("roles", "name", name)
		if err != nil {
			return err
		}
		if err := app.Delete(rec); err != nil {
			return err
		}
	}
	return nil
}

func dropCollection(app core.App, name string) error {
	col, err := app.FindCollectionByNameOrId(name)
	if err != nil {
		return err
	}
	return app.Delete(col)
}
