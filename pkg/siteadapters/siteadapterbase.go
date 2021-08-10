package siteadapters

import (
	"fmt"

	"github.com/JYGC/SkillSurvey/pkg/config"
	"github.com/gocolly/colly/v2"
)

type ISiteAdapter interface{}

type SiteAdapterBase struct {
	ConfigSettings     config.SiteAdapterConfig
	SiteName           string
	JobPostLink        string
	TitleSelector      string
	BodySelector       string
	PostedDateSelector string
	CitySelector       string
	Country            string
	SuburbSelector     string
	TitleType          string
	BodyType           string
	PostedDateType     string
	CityType           string
	SuburbType         string
}

func (s *SiteAdapterBase) FetchJobPost(e *colly.HTMLElement) {
	link := e.Attr("href")
	fmt.Printf("Got link: %q -> %s\n", e.Text, link)
}
