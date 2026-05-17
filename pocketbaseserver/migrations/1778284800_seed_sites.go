package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		return seedSites(app)
	}, func(app core.App) error {
		return removeSeedSites(app)
	})
}

var seedSiteRecords = []struct{ name, url string }{
	{"www.seek.com.au", "https://www.seek.com.au"},
	{"au.jora.com", "https://au.jora.com"},
}

func seedSites(app core.App) error {
	col, err := app.FindCollectionByNameOrId("sites")
	if err != nil {
		return err
	}

	for _, s := range seedSiteRecords {
		if _, err := app.FindFirstRecordByData("sites", "name", s.name); err == nil {
			continue // already exists
		}
		rec := core.NewRecord(col)
		rec.Set("name", s.name)
		rec.Set("url", s.url)
		if err := app.Save(rec); err != nil {
			return err
		}
	}
	return nil
}

func removeSeedSites(app core.App) error {
	for _, s := range seedSiteRecords {
		rec, err := app.FindFirstRecordByData("sites", "name", s.name)
		if err != nil {
			continue // not found — nothing to remove
		}
		if err := app.Delete(rec); err != nil {
			return err
		}
	}
	return nil
}
