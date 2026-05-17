package migrator

import (
	"fmt"
	"log"

	pocketbase "github.com/r--w/pocketbase"
	"gorm.io/gorm"

	"keybook/migrate/internal/legacyentities"
)

func migrateMonthlyCountReports(db *gorm.DB, pb *pocketbase.Client, skillNameIdMap map[uint]string) (map[uint]string, Summary, error) {
	summary := Summary{Collection: "monthlyCountReports"}
	idMap := make(map[uint]string)

	var reports []legacyentities.MonthlyCountReport
	if err := db.Find(&reports).Error; err != nil {
		return idMap, summary, fmt.Errorf("read monthlyCountReports: %w", err)
	}

	for _, r := range reports {
		summary.Attempted++

		newSkillNameID, ok := skillNameIdMap[r.SkillNameID]
		if !ok {
			log.Printf("migrate monthlyCountReports id=%d: no mapping for legacy skillNameID=%d", r.ID, r.SkillNameID)
			continue
		}

		// Identifier uses the new PocketBase skill name ID, not the legacy integer.
		identifier := fmt.Sprintf("%s_%s", newSkillNameID, r.YearMonth)

		existing, err := pb.List("monthlyCountReports", pocketbase.ParamsList{
			Filters: fmt.Sprintf(`identifier = %q`, identifier),
			Size:    1,
		})
		if err != nil {
			log.Printf("migrate monthlyCountReports id=%d: list check: %v", r.ID, err)
			continue
		}
		if existing.TotalItems > 0 {
			idMap[r.ID] = existing.Items[0]["id"].(string)
			summary.Written++
			continue
		}

		created, err := pb.Create("monthlyCountReports", map[string]any{
			"identifier":    identifier,
			"YearMonth":     r.YearMonth,
			"yearMonthDate": fmt.Sprintf("%s-01 00:00:00.000Z", r.YearMonth),
			"count":         r.Count,
			"skillName":     newSkillNameID,
		})
		if err != nil {
			log.Printf("migrate monthlyCountReports id=%d: %v", r.ID, err)
			continue
		}
		idMap[r.ID] = created.ID
		summary.Written++
	}

	return idMap, summary, nil
}
