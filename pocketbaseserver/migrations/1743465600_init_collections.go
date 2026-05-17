package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	m.Register(func(app core.App) error {
		// Collections are created in dependency order so relation fields resolve correctly.
		// Note: "users" is already created by PocketBase's own system migration — do not recreate it.

		// 1. sites (no dependencies)
		if err := createSites(app); err != nil {
			return err
		}

		// 2. skillTypes (no dependencies)
		if err := createSkillTypes(app); err != nil {
			return err
		}

		// 3. skillNames (-> skillTypes)
		if err := createSkillNames(app); err != nil {
			return err
		}

		// 4. skillNameAliases (-> skillNames)
		if err := createSkillNameAliases(app); err != nil {
			return err
		}

		// 5. jobPosts (-> sites)
		if err := createJobPosts(app); err != nil {
			return err
		}

		// 6. monthlyCountReports (-> skillNames)
		if err := createMonthlyCountReports(app); err != nil {
			return err
		}

		// 7. userSettings (-> users)
		if err := createUserSettings(app); err != nil {
			return err
		}

		return nil
	}, func(app core.App) error {
		// Downgrade: delete in reverse dependency order.
		// Note: "users" is owned by PocketBase's system migration — do not delete it here.
		for _, name := range []string{
			"userSettings",
			"monthlyCountReports",
			"jobPosts",
			"skillNameAliases",
			"skillNames",
			"skillTypes",
			"sites",
		} {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				return err
			}
			if err := app.Delete(collection); err != nil {
				return err
			}
		}
		return nil
	})
}

func createSites(app core.App) error {
	sites := core.NewBaseCollection("sites")
	sites.Id = "pbc_1313762900"
	sites.ViewRule = types.Pointer("")
	sites.Fields.Add(
		&core.TextField{
			Id:       "text1579384326",
			Name:     "name",
			Required: true,
		},
		&core.TextField{
			Id:       "text4101391790",
			Name:     "url",
			Required: true,
		},
		&core.AutodateField{Id: "autodate2990389176", Name: "created", OnCreate: true},
		&core.AutodateField{Id: "autodate3332085495", Name: "updated", OnCreate: true, OnUpdate: true},
	)
	return app.Save(sites)
}

func createSkillTypes(app core.App) error {
	skillTypes := core.NewBaseCollection("skillTypes")
	skillTypes.Id = "pbc_2364094840"
	skillTypes.Fields.Add(
		&core.TextField{
			Id:       "text1579384326",
			Name:     "name",
			Required: true,
		},
		&core.TextField{
			Id:       "text1843675174",
			Name:     "description",
			Required: true,
		},
		&core.AutodateField{Id: "autodate2990389176", Name: "created", OnCreate: true},
		&core.AutodateField{Id: "autodate3332085495", Name: "updated", OnCreate: true, OnUpdate: true},
	)
	return app.Save(skillTypes)
}

func createSkillNames(app core.App) error {
	skillNames := core.NewBaseCollection("skillNames")
	skillNames.Id = "pbc_669710600"
	skillNames.Fields.Add(
		&core.TextField{
			Id:       "text1579384326",
			Name:     "name",
			Required: true,
		},
		&core.BoolField{
			Id:   "bool1187331404",
			Name: "isEnabled",
		},
		&core.RelationField{
			Id:           "relation3416247195",
			Name:         "skillType",
			CollectionId: "pbc_2364094840",
			MaxSelect:    1,
			Required:     true,
		},
		&core.AutodateField{Id: "autodate2990389176", Name: "created", OnCreate: true},
		&core.AutodateField{Id: "autodate3332085495", Name: "updated", OnCreate: true, OnUpdate: true},
	)
	return app.Save(skillNames)
}

func createSkillNameAliases(app core.App) error {
	skillNameAliases := core.NewBaseCollection("skillNameAliases")
	skillNameAliases.Id = "pbc_2939716207"
	skillNameAliases.Fields.Add(
		&core.RelationField{
			Id:           "relation425910964",
			Name:         "skillName",
			CollectionId: "pbc_669710600",
			MaxSelect:    1,
			Required:     true,
		},
		&core.TextField{
			Id:       "text3781979028",
			Name:     "alias",
			Required: true,
		},
		&core.AutodateField{Id: "autodate2990389176", Name: "created", OnCreate: true},
		&core.AutodateField{Id: "autodate3332085495", Name: "updated", OnCreate: true, OnUpdate: true},
	)
	return app.Save(skillNameAliases)
}

func createJobPosts(app core.App) error {
	jobPosts := core.NewBaseCollection("jobPosts")
	jobPosts.Id = "pbc_3979010322"
	jobPosts.ViewRule = types.Pointer("")
	jobPosts.Fields.Add(
		&core.TextField{
			Id:       "text2236787266",
			Name:     "jobSiteNumber",
			Required: true,
		},
		&core.RelationField{
			Id:           "relation1766001124",
			Name:         "site",
			CollectionId: "pbc_1313762900",
			MaxSelect:    1,
			Required:     true,
		},
		&core.JSONField{
			Id:       "json4274335913",
			Name:     "content",
			Required: true,
		},
		&core.DateField{
			Id:          "date789529322",
			Name:        "postedDate",
			Presentable: true,
		},
		&core.JSONField{
			Id:       "json1587448267",
			Name:     "location",
			Required: true,
		},
		&core.AutodateField{Id: "autodate2990389176", Name: "created", OnCreate: true},
		&core.AutodateField{Id: "autodate3332085495", Name: "updated", OnCreate: true, OnUpdate: true},
	)
	return app.Save(jobPosts)
}

func createMonthlyCountReports(app core.App) error {
	monthlyCountReports := core.NewBaseCollection("monthlyCountReports")
	monthlyCountReports.Id = "pbc_428779972"
	monthlyCountReports.Fields.Add(
		&core.TextField{
			Id:       "text1999537002",
			Name:     "identifier",
			Required: true,
		},
		&core.TextField{
			Id:       "text1116565208",
			Name:     "YearMonth",
			Required: true,
		},
		&core.DateField{
			Id:       "date4033519741",
			Name:     "yearMonthDate",
			Required: true,
		},
		&core.NumberField{
			Id:      "number2245608546",
			Name:    "count",
			OnlyInt: true,
		},
		&core.RelationField{
			Id:           "relation425910964",
			Name:         "skillName",
			CollectionId: "pbc_669710600",
			MaxSelect:    1,
		},
		&core.AutodateField{Id: "autodate2990389176", Name: "created", OnCreate: true},
		&core.AutodateField{Id: "autodate3332085495", Name: "updated", OnCreate: true, OnUpdate: true},
	)
	return app.Save(monthlyCountReports)
}

func createUserSettings(app core.App) error {
	userSettings := core.NewBaseCollection("userSettings")
	userSettings.Id = "pbc_3975969204"
	userSettings.ListRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	userSettings.ViewRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	userSettings.CreateRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	userSettings.UpdateRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	userSettings.DeleteRule = types.Pointer(`@request.auth.id != "" && user = @request.auth.id`)
	userSettings.Indexes = []string{
		"CREATE UNIQUE INDEX `idx_awMFnb4lWc` ON `userSettings` (`user` ASC)",
	}
	userSettings.Fields.Add(
		&core.RelationField{
			Id:            "relation2375276105",
			Name:          "user",
			CollectionId:  "_pb_users_auth_",
			MaxSelect:     1,
			Required:      true,
			CascadeDelete: true,
		},
		&core.SelectField{
			Id:        "select809235174",
			Name:      "portalTheme",
			MaxSelect: 1,
			Required:  true,
			Values:    []string{"white", "g10", "g90", "g100"},
		},
		&core.AutodateField{Id: "autodate2990389176", Name: "created", OnCreate: true},
		&core.AutodateField{Id: "autodate3332085495", Name: "updated", OnCreate: true, OnUpdate: true},
	)
	return app.Save(userSettings)
}
