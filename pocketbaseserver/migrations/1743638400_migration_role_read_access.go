package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// readRuleMigration allows authenticated users with the migration role to list and view records.
const readRuleMigration = `@request.auth.id != "" && @collection.userRoles.user = @request.auth.id && @collection.userRoles.role.name = 'migration'`

func init() {
	m.Register(func(app core.App) error {
		return applyMigrationRoleReadRules(app)
	}, func(app core.App) error {
		return revertMigrationRoleReadRules(app)
	})
}

func applyMigrationRoleReadRules(app core.App) error {
	type readUpdate struct {
		name     string
		listRule *string
		viewRule *string // nil means leave unchanged
	}

	rule := types.Pointer(readRuleMigration)

	updates := []readUpdate{
		// sites and jobPosts already have a public viewRule ("") — leave viewRule unchanged.
		{name: "sites", listRule: rule, viewRule: nil},
		{name: "jobPosts", listRule: rule, viewRule: nil},
		// skill collections have no viewRule — set both.
		{name: "skillTypes", listRule: rule, viewRule: rule},
		{name: "skillNames", listRule: rule, viewRule: rule},
		{name: "skillNameAliases", listRule: rule, viewRule: rule},
	}

	for _, u := range updates {
		col, err := app.FindCollectionByNameOrId(u.name)
		if err != nil {
			return err
		}
		col.ListRule = u.listRule
		if u.viewRule != nil {
			col.ViewRule = u.viewRule
		}
		if err := app.Save(col); err != nil {
			return err
		}
	}
	return nil
}

func revertMigrationRoleReadRules(app core.App) error {
	type revertUpdate struct {
		name     string
		listRule *string // nil = superadmin only
		viewRule *string // nil = superadmin only; types.Pointer("") = public
	}

	empty := types.Pointer("")

	reverts := []revertUpdate{
		{name: "sites", listRule: nil, viewRule: empty},
		{name: "jobPosts", listRule: nil, viewRule: empty},
		{name: "skillTypes", listRule: nil, viewRule: nil},
		{name: "skillNames", listRule: nil, viewRule: nil},
		{name: "skillNameAliases", listRule: nil, viewRule: nil},
	}

	for _, u := range reverts {
		col, err := app.FindCollectionByNameOrId(u.name)
		if err != nil {
			return err
		}
		col.ListRule = u.listRule
		col.ViewRule = u.viewRule
		if err := app.Save(col); err != nil {
			return err
		}
	}
	return nil
}
