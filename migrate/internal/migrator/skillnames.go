package migrator

import (
	"fmt"
	"log"

	pocketbase "github.com/r--w/pocketbase"
	"gorm.io/gorm"

	"keybook/migrate/internal/legacyentities"
)

func migrateSkillNames(db *gorm.DB, pb *pocketbase.Client, skillTypeIdMap map[uint]string) (map[uint]string, Summary, error) {
	summary := Summary{Collection: "skillNames"}
	idMap := make(map[uint]string)

	var skillNames []legacyentities.SkillName
	if err := db.Find(&skillNames).Error; err != nil {
		return idMap, summary, fmt.Errorf("read skillNames: %w", err)
	}

	for _, sn := range skillNames {
		summary.Attempted++

		newSkillTypeID, ok := skillTypeIdMap[sn.SkillTypeID]
		if !ok {
			log.Printf("migrate skillNames id=%d: no mapping for legacy skillTypeID=%d", sn.ID, sn.SkillTypeID)
			continue
		}

		existing, err := pb.List("skillNames", pocketbase.ParamsList{
			Filters: fmt.Sprintf(`skillType = %q && name = %q`, newSkillTypeID, sn.Name),
			Size:    1,
		})
		if err != nil {
			log.Printf("migrate skillNames id=%d: list check: %v", sn.ID, err)
			continue
		}
		if existing.TotalItems > 0 {
			idMap[sn.ID] = existing.Items[0]["id"].(string)
			summary.Written++
			continue
		}

		created, err := pb.Create("skillNames", map[string]any{
			"name":      sn.Name,
			"isEnabled": sn.IsEnabled,
			"skillType": newSkillTypeID,
		})
		if err != nil {
			log.Printf("migrate skillNames id=%d: %v", sn.ID, err)
			continue
		}
		idMap[sn.ID] = created.ID
		summary.Written++
	}

	return idMap, summary, nil
}
