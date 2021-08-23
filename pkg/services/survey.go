package services

import (
	"github.com/JYGC/SkillSurvey/pkg/siteadapters"
	"github.com/JYGC/SkillSurvey/pkg/webscraper"
)

const userAgent = "node-spider"

type Survey struct {
	ServiceBase
}

func NewSurvey() *Survey {
	survey := new(Survey)
	return survey
}

func (s *Survey) Run() {
	for _, webScraperSite := range []webscraper.WebScraper{
		*webscraper.NewWebScraper(siteadapters.NewJoraAdapter(), userAgent),
		*webscraper.NewWebScraper(siteadapters.NewSeekAdapter(), userAgent),
	} {
		webScraperSite.Start()
	}
}
