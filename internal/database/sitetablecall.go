package database

import (
	"fmt"

	"github.com/JYGC/SkillSurvey/internal/entities"
	"gorm.io/gorm"
)

type SiteTableCall struct {
	DbTableCallBase
}

func NewSiteTableCall(db *gorm.DB) (tableCall *SiteTableCall) {
	tableCall = new(SiteTableCall)
	tableCall.db = db
	tableCall.MigrateTable(&entities.Site{})
	return tableCall
}

func (s SiteTableCall) GetAll() {
	var results []entities.Site
	s.db.Find(&results)
	for _, res := range results {
		fmt.Println(res.Name)
	}
}

func (s SiteTableCall) InsertBulk() {
	siteSlice := make([]*entities.Site, 2)
	siteSlice[0] = new(entities.Site)
	siteSlice[0].Name = "https://www.seek.com.au"
	siteSlice[1] = new(entities.Site)
	siteSlice[1].Name = "https://au.jora.com"
	s.db.Create(siteSlice)
}
