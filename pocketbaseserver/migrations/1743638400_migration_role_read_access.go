package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// Read rules use the back-relation @request.auth.userRoles_via_user so all
// conditions apply to the same userRoles record for the authenticated user.
const (
	// readRuleSites: webscraper lists sites to know which boards to scrape.
	readRuleSites = `@request.auth.id != "" && (@request.auth.userRoles_via_user.role.name ?= 'webscraper' || @request.auth.userRoles_via_user.role.name ?= 'migration')`

	// readRuleJobPosts: webscraper checks for duplicates; reporting counts posts.
	readRuleJobPosts = `@request.auth.id != "" && (@request.auth.userRoles_via_user.role.name ?= 'webscraper' || @request.auth.userRoles_via_user.role.name ?= 'reporting' || @request.auth.userRoles_via_user.role.name ?= 'migration')`

	// readRuleSkillData: reporting reads skill names/aliases to build counts.
	readRuleSkillData = `@request.auth.id != "" && (@request.auth.userRoles_via_user.role.name ?= 'reporting' || @request.auth.userRoles_via_user.role.name ?= 'migration')`
)

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

	updates := []readUpdate{
		{name: "sites", listRule: types.Pointer(readRuleSites), viewRule: nil},
		{name: "jobPosts", listRule: types.Pointer(readRuleJobPosts), viewRule: nil},
		{name: "skillTypes", listRule: types.Pointer(readRuleSkillData), viewRule: types.Pointer(readRuleSkillData)},
		{name: "skillNames", listRule: types.Pointer(readRuleSkillData), viewRule: types.Pointer(readRuleSkillData)},
		{name: "skillNameAliases", listRule: types.Pointer(readRuleSkillData), viewRule: types.Pointer(readRuleSkillData)},
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
