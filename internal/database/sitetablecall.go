package database

import (
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

func (s SiteTableCall) GetAll() (results []entities.Site, err error) {
	err = s.db.Find(&results).Error
	return results, err
}

// func (s SiteTableCall) InsertBulk() {
// 	siteSlice := make([]*entities.Site, 2)
// 	siteSlice[0] = new(entities.Site)
// 	siteSlice[0].Name = "https://www.seek.com.au"
// 	siteSlice[1] = new(entities.Site)
// 	siteSlice[1].Name = "https://au.jora.com"
// 	s.db.Create(siteSlice)
// }
