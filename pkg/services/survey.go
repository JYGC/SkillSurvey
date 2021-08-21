package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/pkg/siteadapters"
	"github.com/gocolly/colly"
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
	for _, currentSiteAdpter := range []siteadapters.ISiteAdapter{
		siteadapters.NewSeekAdapter(),
		siteadapters.NewJoraAdapter(),
	} {
		webSpider := colly.NewCollector(
			colly.UserAgent(userAgent),
			colly.AllowedDomains(currentSiteAdpter.GetConfigSettings().AllowedDomains...),
		)
		webSpider.OnHTML(currentSiteAdpter.GetJobPostLink(), func(e *colly.HTMLElement) {
			link := e.Attr("href")
			fmt.Printf("Got link: %q -> %s\n", e.Text, link)
		})
		webSpider.OnRequest(func(r *colly.Request) {
			fmt.Println("Visiting", r.URL.String())
		})
		for _, searchCriteria := range currentSiteAdpter.GetConfigSettings().SearchCriterias {
			for searchPage := 1; searchPage <= currentSiteAdpter.GetConfigSettings().Pages; searchPage++ {
				fullUrl := strings.ReplaceAll(
					searchCriteria.Url,
					currentSiteAdpter.GetConfigSettings().PageFlag,
					strconv.Itoa(searchPage),
				)
				webSpider.Visit(fullUrl)
			}
		}
	}
}
