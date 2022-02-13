package database

import (
	"path/filepath"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/readonlysettings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseAdapter struct {
	database     *gorm.DB
	JobPost      *JobPostTableCall
	Site         *SiteTableCall
	SkillName    *SkillNameTableCall
	MonthlyCount *MonthlyCountReportTableCall
}

func (da DatabaseAdapter) Create(items interface{}) (tx *gorm.DB) {
	tx = da.database.Create(items)
	return tx
}

func (da DatabaseAdapter) Find(items interface{}) *gorm.DB {
	return da.database.Find(items)
}

const databaseFile = "SkillSurvey.db"

var DbAdapter *DatabaseAdapter

func init() {
	DbAdapter = new(DatabaseAdapter)
	configSettings := config.LoadMainConfig()
	var err error
	var appDataFolder string
	appDataFolder, err = readonlysettings.GetAppDataFolder(configSettings.IsProduction)
	if err != nil {
		exception.ErrorLogger.Println(err)
		panic(err)
	}
	DbAdapter.database, err = gorm.Open(sqlite.Open(filepath.Join(
		appDataFolder,
		databaseFile,
	)), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		exception.ErrorLogger.Println(err)
		panic(err)
	}
	DbAdapter.JobPost = NewJobPostTableCall(DbAdapter.database)
	DbAdapter.Site = NewSiteTableCall(DbAdapter.database)
	DbAdapter.SkillName = NewSkillNameTableCall(DbAdapter.database)
	DbAdapter.MonthlyCount = NewMonthlyCountReportTableCall(DbAdapter.database)
}
