package migrator

import (
	"fmt"
	"log"

	pocketbase "github.com/r--w/pocketbase"
	"gorm.io/gorm"

	"keybook/migrate/internal/legacyentities"
)

func migrateSkillTypes(db *gorm.DB, pb *pocketbase.Client) (map[uint]string, Summary, error) {
	summary := Summary{Collection: "skillTypes"}
	idMap := make(map[uint]string)

	var skillTypes []legacyentities.SkillType
	if err := db.Find(&skillTypes).Error; err != nil {
		return idMap, summary, fmt.Errorf("read skillTypes: %w", err)
	}

	for _, st := range skillTypes {
		summary.Attempted++

		existing, err := pb.List("skillTypes", pocketbase.ParamsList{
			Filters: fmt.Sprintf(`name = %q`, st.Name),
			Size:    1,
		})
		if err != nil {
			log.Printf("migrate skillTypes id=%d: list check: %v", st.ID, err)
			continue
		}
		if existing.TotalItems > 0 {
			idMap[st.ID] = existing.Items[0]["id"].(string)
			summary.Written++
			continue
		}

		created, err := pb.Create("skillTypes", map[string]any{
			"name":        st.Name,
			"description": st.Description,
		})
		if err != nil {
			log.Printf("migrate skillTypes id=%d: %v", st.ID, err)
			continue
		}
		idMap[st.ID] = created.ID
		summary.Written++
	}

	return idMap, summary, nil
}
