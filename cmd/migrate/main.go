package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
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
	CreateDate    time.Time // CONITUE HERE: DATE COULD NOT BE READ
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
	oldDb, _ := gorm.Open(
		sqlite.Open(filepath.Join(configSettings.AppDataFolder, "SkillSurvey.old.db")),
		&gorm.Config{DisableForeignKeyConstraintWhenMigrating: true},
	)
	//Get data of old format
	var sites []Site
	var jobPosts []JobPost
	var skillTypes []SkillType
	var skillNames []SkillName
	var skillWordAliases []SkillWordAlias
	oldDb.Raw("SELECT * FROM Site").Scan(&sites)
	oldDb.Raw("SELECT Id, SiteId, JobSiteNumber, Title, Body, PostedDate, City, Country, Suburb, CreateDate FROM JobPost").Scan(&jobPosts)
	oldDb.Raw("SELECT * FROM SkillType").Scan(&skillTypes)
	oldDb.Raw("SELECT * FROM SkillName").Scan(&skillNames)
	oldDb.Raw("SELECT * FROM SkillWordAlias").Scan(&skillWordAliases)
	// create new tables
	_, err := database.DbAdapter.Site.GetAll()
	if err == nil {
		// convert to new format, make associations and commit ot database
		var newSites []entities.Site
		for _, site := range sites {
			newSites = append(newSites, entities.Site{
				Name: strings.ToLower(site.Name),
			})
		}
		database.DbAdapter.Create(newSites)
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
			database.DbAdapter.Create(newJobPostsDivided)
		}
		// migrate skill type
		var newSkillTypes []entities.SkillType
		for _, skillType := range skillTypes {
			newSkillTypes = append(newSkillTypes, entities.SkillType{
				Name: skillType.Name,
			})
		}
		database.DbAdapter.Create(newSkillTypes)
		newSkillTypesMap := make(map[uint]entities.SkillType)
		for _, skillType := range newSkillTypes {
			newSkillTypesMap[skillType.ID] = skillType
		}
		// migrate skill name
		var newSkillNames []entities.SkillName
		for _, skillName := range skillNames {
			newSkillNames = append(newSkillNames, entities.SkillName{
				Name:      skillName.Name,
				SkillType: newSkillTypesMap[uint(skillName.SkillTypeId)],
				IsEnabled: skillName.IsEnabled,
			})
		}
		database.DbAdapter.Create(newSkillNames)
		newSkillNamesMap := make(map[uint]entities.SkillName)
		for _, skillName := range newSkillNames {
			newSkillNamesMap[skillName.ID] = skillName
		}
		// migrate skill word alias
		var newSkillWordAliases []entities.SkillNameAlias
		for _, skillWordAlias := range skillWordAliases {
			newSkillWordAliases = append(newSkillWordAliases, entities.SkillNameAlias{
				SkillName: newSkillNamesMap[uint(skillWordAlias.SkillNameId)],
				Alias:     skillWordAlias.Alias,
			})
		}
		database.DbAdapter.Create(newSkillWordAliases)

	} else {
		fmt.Println(err)
	}
}
