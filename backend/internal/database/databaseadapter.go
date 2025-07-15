package database

import (
	"github.com/JYGC/SkillSurvey/internal/environment"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DatabaseAdapter struct {
	database     *gorm.DB
	JobPost      *JobPostTableCall
	Site         *SiteTableCall
	SkillType    *SkillTypeTableCall
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

const databaseFile = "./SkillSurvey.db"

var DbAdapter *DatabaseAdapter

func init() {
	DbAdapter = new(DatabaseAdapter)
	var err error
	DbAdapter.database, err = gorm.Open(
		sqlite.Open(environment.AttachToExecutableDir(databaseFile)),
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true},
	)
	if err != nil {
		exception.ErrorLogger.Println(err)
		panic(err)
	}
	DbAdapter.JobPost = NewJobPostTableCall(DbAdapter.database)
	DbAdapter.Site = NewSiteTableCall(DbAdapter.database)
	DbAdapter.SkillType = NewSkillTypeTableCall(DbAdapter.database)
	DbAdapter.SkillName = NewSkillNameTableCall(DbAdapter.database)
	DbAdapter.MonthlyCount = NewMonthlyCountReportTableCall(DbAdapter.database)
}
