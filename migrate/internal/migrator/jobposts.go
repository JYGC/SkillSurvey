package migrator

import (
	"encoding/json"
	"fmt"
	"log"

	pocketbase "github.com/r--w/pocketbase"
	"gorm.io/gorm"

	"keybook/migrate/internal/legacyentities"
)

func migrateJobPosts(db *gorm.DB, pb *pocketbase.Client, siteIdMap map[uint]string) (map[uint]string, Summary, error) {
	summary := Summary{Collection: "jobPosts"}
	idMap := make(map[uint]string)

	var jobPosts []legacyentities.JobPost
	if err := db.Find(&jobPosts).Error; err != nil {
		return idMap, summary, fmt.Errorf("read jobPosts: %w", err)
	}

	for _, jp := range jobPosts {
		summary.Attempted++

		newSiteID, ok := siteIdMap[jp.SiteID]
		if !ok {
			log.Printf("migrate jobPosts id=%d: no mapping for legacy siteID=%d", jp.ID, jp.SiteID)
			continue
		}

		existing, err := pb.List("jobPosts", pocketbase.ParamsList{
			Filters: fmt.Sprintf(`site = %q && jobSiteNumber = %q`, newSiteID, jp.JobSiteNumber),
			Size:    1,
		})
		if err != nil {
			log.Printf("migrate jobPosts id=%d: list check: %v", jp.ID, err)
			continue
		}
		if existing.TotalItems > 0 {
			idMap[jp.ID] = existing.Items[0]["id"].(string)
			summary.Written++
			continue
		}

		contentJSON, _ := json.Marshal(map[string]string{
			"title": jp.Title,
			"body":  jp.Body,
		})
		locationJSON, _ := json.Marshal(map[string]string{
			"city":    jp.City,
			"country": jp.Country,
			"suburb":  jp.Suburb,
		})

		created, err := pb.Create("jobPosts", map[string]any{
			"jobSiteNumber": jp.JobSiteNumber,
			"site":          newSiteID,
			"content":       json.RawMessage(contentJSON),
			"location":      json.RawMessage(locationJSON),
			"postedDate":    jp.PostedDate,
		})
		if err != nil {
			log.Printf("migrate jobPosts id=%d: %v", jp.ID, err)
			continue
		}
		idMap[jp.ID] = created.ID
		summary.Written++
	}

	return idMap, summary, nil
}
