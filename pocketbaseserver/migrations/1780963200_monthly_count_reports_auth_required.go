package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

// monthlyCountReports is now auth-gated in the frontend (moved to the user
// route tree). This migration matches the API rules to the UI intent so that
// direct unauthenticated API calls are also denied.
func init() {
	m.Register(func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("monthlyCountReports")
		if err != nil {
			return err
		}
		col.ListRule = types.Pointer("@request.auth.id != \"\"")
		col.ViewRule = types.Pointer("@request.auth.id != \"\"")
		return app.Save(col)
	}, func(app core.App) error {
		col, err := app.FindCollectionByNameOrId("monthlyCountReports")
		if err != nil {
			return err
		}
		col.ListRule = types.Pointer("")
		col.ViewRule = types.Pointer("")
		return app.Save(col)
	})
}
