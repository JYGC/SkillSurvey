package services

import (
	"fmt"

	"github.com/JYGC/SkillSurvey/pkg/config"
	"github.com/gocolly/colly/v2"
)

const userAgent = "node-spider"

type Survey struct {
	ServiceBase
	webSpider *colly.Collector
}

func NewSurvey() *Survey {
	survey := new(Survey)
	survey.webSpider = colly.NewCollector()
	survey.webSpider.UserAgent = userAgent
	//Need to ad more function
	//survey.webSpider.WithTransport(&http.Transport{})
	return survey
}

func (s *Survey) Run() {
	config := config.LoadConfiguration()
	fmt.Println(config)
	// c := colly.NewCollector(
	// 	colly.AllowedDomains("www.halopedia.org"),
	// )

	// c.OnHTML("a[href]", func(e *colly.HTMLElement) {
	// 	link := e.Attr("href")
	// 	fmt.Printf("Got link: %q -> %s\n", e.Text, link)
	// 	c.Visit(e.Request.AbsoluteURL(link))
	// })

	// c.OnRequest(func(r *colly.Request) {
	// 	fmt.Println("Visiting", r.URL.String())
	// })

	// c.Visit("https://www.halopedia.org")
}
