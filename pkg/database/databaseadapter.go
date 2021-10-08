package database

import (
	"fmt"
	"path/filepath"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/JYGC/SkillSurvey/pkg/config"
)

type DatabaseAdapter struct {
	database *db.DB
	JobPost  *JobPostCollection
	Site     *SiteCollection
}

func NewDatabaseAdapter() *DatabaseAdapter {
	databaseAdapter := new(DatabaseAdapter)
	configSettings := config.LoadMainConfig()
	var err error
	databaseAdapter.database, err = db.OpenDB(filepath.Join(configSettings.AppDataFolder, configSettings.DatabaseFile))
	if err != nil {
		fmt.Printf("Failed to get database: %s\n", err.Error())
		//TODO: Error handling
	}
	databaseAdapter.JobPost = NewJobPostCollection(databaseAdapter.database)
	databaseAdapter.Site = NewSiteCollection(databaseAdapter.database)
	return databaseAdapter
}

var DbAdapter *DatabaseAdapter = NewDatabaseAdapter()
