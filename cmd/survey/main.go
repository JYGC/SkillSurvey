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
	var newInboundJobPostSlice []entities.InboundJobPost
	for _, webScraperSite := range []webscraper.WebScraper{
		*webscraper.NewWebScraper(siteadapters.NewJoraAdapter(), userAgent),
		*webscraper.NewWebScraper(siteadapters.NewSeekAdapter(), userAgent),
	} {
		newInboundJobPostSlice = append(newInboundJobPostSlice, webScraperSite.Scrape()...)
	}
	existingSites, getSitesErr := database.DbAdapter.Site.GetAll()
	if getSitesErr != nil {
		exception.ReportError(map[string]string{
			"Details": getSitesErr.Error(),
		})
	}
	siteMap := entities.MakeSiteMap(existingSites)
	newJobPostSlice := entities.CreateJobPosts(siteMap, newInboundJobPostSlice)
	database.DbAdapter.JobPost.BulkUpdateOrInsert(newJobPostSlice)
}
