package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// skillNames is expanded on the public MonthlyCountReport route.  PocketBase
// silently omits expand results when the requesting user lacks read access to
// the related collection, causing skill names to appear as "Unknown".
// Opening list/view to unauthenticated requests fixes the expand.
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("skillNames")
		if err != nil {
			return err
		}
		col.ListRule = types.Pointer("")
		col.ViewRule = types.Pointer("")
		return app.Save(col)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("skillNames")
		if err != nil {
			return err
		}
		col.ListRule = types.Pointer(readRuleSkillData)
		col.ViewRule = types.Pointer(readRuleSkillData)
		return app.Save(col)
	})
}
