package migrator

import (
	"fmt"
	"log"

	pocketbase "github.com/r--w/pocketbase"
	"gorm.io/gorm"

	"keybook/migrate/internal/legacyentities"
)

func migrateSkillNameAliases(db *gorm.DB, pb *pocketbase.Client, skillNameIdMap map[uint]string) (map[uint]string, Summary, error) {
	summary := Summary{Collection: "skillNameAliases"}
	idMap := make(map[uint]string)

	var aliases []legacyentities.SkillNameAlias
	if err := db.Find(&aliases).Error; err != nil {
		return idMap, summary, fmt.Errorf("read skillNameAliases: %w", err)
	}

	for _, alias := range aliases {
		summary.Attempted++

		newSkillNameID, ok := skillNameIdMap[alias.SkillNameID]
		if !ok {
			log.Printf("migrate skillNameAliases id=%d: no mapping for legacy skillNameID=%d", alias.ID, alias.SkillNameID)
			continue
		}

		existing, err := pb.List("skillNameAliases", pocketbase.ParamsList{
			Filters: fmt.Sprintf(`skillName = %q && alias = %q`, newSkillNameID, alias.Alias),
			Size:    1,
		})
		if err != nil {
			log.Printf("migrate skillNameAliases id=%d: list check: %v", alias.ID, err)
			continue
		}
		if existing.TotalItems > 0 {
			idMap[alias.ID] = existing.Items[0]["id"].(string)
			summary.Written++
			continue
		}

		created, err := pb.Create("skillNameAliases", map[string]any{
			"skillName": newSkillNameID,
			"alias":     alias.Alias,
		})
		if err != nil {
			log.Printf("migrate skillNameAliases id=%d: %v", alias.ID, err)
			continue
		}
		idMap[alias.ID] = created.ID
		summary.Written++
	}

	return idMap, summary, nil
}
