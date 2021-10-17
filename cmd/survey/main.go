package main

import (
	"github.com/JYGC/SkillSurvey/internal/siteadapters"
	"github.com/JYGC/SkillSurvey/internal/webscraper"
)

const userAgent = "node-spider"

func main() {
	for _, webScraperSite := range []webscraper.WebScraper{
		*webscraper.NewWebScraper(siteadapters.NewJoraAdapter(), userAgent),
		*webscraper.NewWebScraper(siteadapters.NewSeekAdapter(), userAgent),
	} {
		webScraperSite.Start()
	}
}
