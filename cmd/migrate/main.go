package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/readonlysettings"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type IdEntity struct {
	Id int
}

type Site struct {
	IdEntity
	Name string
}

type JobPost struct {
	IdEntity
	SiteId        int
	JobSiteNumber string
	Title         string
	Body          string
	PostedDate    string
	City          string
	Country       string
	Suburb        string
	CreateDate    time.Time
}

type SkillType struct {
	IdEntity
	Name string
}

type SkillName struct {
	IdEntity
	Name        string
	SkillTypeId int
	IsEnabled   bool
}

type SkillWordAlias struct {
	IdEntity
	SkillNameId int
	Alias       string
}

func main() {
	configSettings := config.LoadMainConfig()
	appDataFolder, err := readonlysettings.GetAppDataFolder(configSettings.IsProduction)
	if err != nil {
		panic(err)
	}
	oldDb, _ := gorm.Open(
		sqlite.Open(filepath.Join(appDataFolder, "SkillSurvey.old.db")),
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true},
	)
	//Get data of old format
	var sites []Site
	var jobPosts []JobPost
	oldDb.Raw("SELECT * FROM Site").Scan(&sites)
	oldDb.Raw("SELECT Id, SiteId, JobSiteNumber, Title, Body, PostedDate, City, Country, Suburb, CreateDate FROM JobPost").Scan(&jobPosts)
	// create new tables
	//_, err := database.DbAdapter.Site.GetAll()
	if err == nil {
		// convert to new format, make associations and commit ot database
		var newSites []entities.Site
		for _, site := range sites {
			newSites = append(newSites, entities.Site{
				Name: strings.ToLower(site.Name),
			})
		}
		newSitesMap := make(map[uint]entities.Site)
		for _, site := range newSites {
			newSitesMap[site.ID] = site
		}
		// migrate job posts
		var newJobPosts []entities.JobPost
		for _, jobPost := range jobPosts {
			postedDate, tperr := time.Parse(
				time.RFC3339Nano,
				strings.Replace(jobPost.PostedDate, " ", "T", 1)+"Z",
			)
			if tperr != nil {
				fmt.Println(tperr)
			}
			newJobPosts = append(newJobPosts, entities.JobPost{
				Site:          newSitesMap[uint(jobPost.SiteId)],
				JobSiteNumber: jobPost.JobSiteNumber,
				Title:         jobPost.Title,
				Body:          jobPost.Body,
				PostedDate:    postedDate,
				City:          jobPost.City,
				Country:       jobPost.Country,
				Suburb:        jobPost.Suburb,
				CreateDate:    jobPost.CreateDate,
			})
		}
		chunkSize := 1000
		for i := 0; i < len(newJobPosts); i += chunkSize {
			var newJobPostsDivided []entities.JobPost
			end := i + chunkSize
			if end > len(newJobPosts) {
				end = len(newJobPosts)
			}
			newJobPostsDivided = newJobPosts[i:end]
			if err := database.DbAdapter.JobPost.BulkUpdateOrInsert(newJobPostsDivided); err != nil {
				panic(err)
			}
		}
	} else {
		fmt.Println(err)
	}
}
