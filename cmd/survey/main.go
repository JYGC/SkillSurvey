package main

import (
	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/siteadapters"
	"github.com/JYGC/SkillSurvey/internal/webscraper"
)

const userAgent = "node-spider"

func main() {
	variableRef := make(map[string]interface{})
	defer exception.ReportErrorIfPanic(map[string]interface{}{"Variables": variableRef})
	// get jobposts from websites
	var newInboundJobPostSlice []entities.InboundJobPost
	for _, webScraperSite := range []webscraper.WebScraper{
		*webscraper.NewWebScraper(siteadapters.NewJoraAdapter(), userAgent),
		*webscraper.NewWebScraper(siteadapters.NewSeekAdapter(), userAgent),
	} {
		newInboundJobPostSlice = append(newInboundJobPostSlice, webScraperSite.Scrape()...)
	}
	existingSites, err := database.DbAdapter.Site.GetAll()
	if err != nil {
		panic(err)
	}
	// insert jobposts to database
	siteMap := entities.MakeSiteMap(existingSites)
	newJobPostSlice := entities.CreateJobPosts(siteMap, newInboundJobPostSlice)
	if err := database.DbAdapter.JobPost.BulkUpdateOrInsert(newJobPostSlice); err != nil {
		variableRef["existingSites"] = existingSites
		variableRef["siteMap"] = siteMap
		variableRef["newInboundJobPostSlice"] = newInboundJobPostSlice
		variableRef["newJobPostSlice"] = newJobPostSlice
		panic(err)
	}
}
