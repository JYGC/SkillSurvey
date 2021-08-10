package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/pkg/siteadapters"
	"github.com/gocolly/colly/v2"
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
	// webSpider := colly.NewCollector(
	// 	colly.UserAgent(userAgent),
	// 	colly.AllowedDomains("www.wikipedia.org", "wikipedia.org"),
	// )

	// webSpider.OnHTML("a[href]", func(e *colly.HTMLElement) {
	// 	link := e.Attr("href")
	// 	fmt.Printf("Got link: %q -> %s\n", e.Text, link)
	// 	webSpider.Visit(e.Request.AbsoluteURL(link))
	// })

	// webSpider.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL.String())
	// })

	// webSpider.Visit("https://www.wikipedia.org")
	currentSiteAdpter := siteadapters.NewSeek()
	webSpider := colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains(currentSiteAdpter.ConfigSettings.AllowedDomains...),
	)
	webSpider.OnHTML(currentSiteAdpter.JobPostLink, currentSiteAdpter.FetchJobPost)
	webSpider.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})
	for _, searchCriteria := range currentSiteAdpter.ConfigSettings.SearchCriterias {
		for searchPage := 1; searchPage <= currentSiteAdpter.ConfigSettings.Pages; searchPage++ {
			fullUrl := strings.ReplaceAll(
				searchCriteria.Url,
				currentSiteAdpter.ConfigSettings.PageFlag,
				strconv.Itoa(searchPage),
			)
			webSpider.Visit(fullUrl)
		}
	}
}
