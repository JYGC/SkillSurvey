package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// monthlyCountReports is served on the public MonthlyCountReport route — it
// must be readable without authentication.  The initial migration left the
// list/view rules as nil (superadmin-only); this migration opens them up.
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("monthlyCountReports")
		if err != nil {
			return err
		}
		col.ListRule = types.Pointer("")
		col.ViewRule = types.Pointer("")
		return app.Save(col)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("monthlyCountReports")
		if err != nil {
			return err
		}
		col.ListRule = nil
		col.ViewRule = nil
		return app.Save(col)
	})
}
