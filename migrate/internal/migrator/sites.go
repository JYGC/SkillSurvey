package migrator

import (
	"fmt"
	"log"

	pocketbase "github.com/r--w/pocketbase"
	"gorm.io/gorm"

	"keybook/migrate/internal/legacyentities"
)

func migrateSites(db *gorm.DB, pb *pocketbase.Client) (map[uint]string, Summary, error) {
	summary := Summary{Collection: "sites"}
	idMap := make(map[uint]string)

	var sites []legacyentities.Site
	if err := db.Find(&sites).Error; err != nil {
		return idMap, summary, fmt.Errorf("read sites: %w", err)
	}

	for _, site := range sites {
		summary.Attempted++

		// Check if already exists by natural key.
		existing, err := pb.List("sites", pocketbase.ParamsList{
			Filters: fmt.Sprintf(`name = %q`, site.Name),
			Size:    1,
		})
		if err != nil {
			log.Printf("migrate sites id=%d: list check: %v", site.ID, err)
			continue
		}
		if existing.TotalItems > 0 {
			idMap[site.ID] = existing.Items[0]["id"].(string)
			summary.Written++
			continue
		}

		created, err := pb.Create("sites", map[string]any{
			"name": site.Name,
			"url":  site.Name,
		})
		if err != nil {
			log.Printf("migrate sites id=%d: %v", site.ID, err)
			continue
		}
		idMap[site.ID] = created.ID
		summary.Written++
	}

	return idMap, summary, nil
}
