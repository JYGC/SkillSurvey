package database

import (
	"path/filepath"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseAdapter struct {
	database  *gorm.DB
	JobPost   *JobPostTableCall
	Site      *SiteTableCall
	SkillName *SkillNameTableCall
}

func (da DatabaseAdapter) Create(items interface{}) (tx *gorm.DB) {
	tx = da.database.Create(items)
	return tx
}

func (da DatabaseAdapter) Find(items interface{}) *gorm.DB {
	return da.database.Find(items)
}

var DbAdapter *DatabaseAdapter

func init() {
	DbAdapter = new(DatabaseAdapter)
	configSettings := config.LoadMainConfig()
	var err error
	DbAdapter.database, err = gorm.Open(sqlite.Open(filepath.Join(
		configSettings.AppDataFolder,
		configSettings.DatabaseFile,
	)), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		exception.ErrorLogger.Println(err)
		panic(err)
	}
	DbAdapter.JobPost = NewJobPostTableCall(DbAdapter.database)
	DbAdapter.Site = NewSiteTableCall(DbAdapter.database)
	DbAdapter.SkillName = NewSkillNameTableCall(DbAdapter.database)
}
