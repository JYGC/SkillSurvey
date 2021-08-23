package siteadapters

import (
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/pkg/config"
	"github.com/gocolly/colly/v2"
)

type ISiteAdapter interface {
	GetConfigSettings() config.SiteAdapterConfig
	GetSiteName() string
	GetJobPostLink() string
	GetTitleSelector() string
	GetBodySelector() string
	GetPostedDateSelector() string
	GetCitySelector() string
	GetCountry() string
	GetSuburbSelector() string
	GetTitleType() string
	GetBodyType() string
	GetPostedDateType() string
	GetCityType() string
	GetSuburbType() string
	GetJobSiteNumber(doc *colly.HTMLElement) string
	GetPostedDate(doc *colly.HTMLElement) time.Time
}

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

func (s SiteAdapterBase) GetConfigSettings() config.SiteAdapterConfig {
	return s.ConfigSettings
}

func (s SiteAdapterBase) GetSiteName() string {
	return s.SiteName
}

func (s SiteAdapterBase) GetJobPostLink() string {
	return s.JobPostLink
}

func (s SiteAdapterBase) GetTitleSelector() string {
	return s.TitleSelector
}

func (s SiteAdapterBase) GetBodySelector() string {
	return s.BodySelector
}

func (s SiteAdapterBase) GetPostedDateSelector() string {
	return s.PostedDateSelector
}

func (s SiteAdapterBase) GetCitySelector() string {
	return s.CitySelector
}

func (s SiteAdapterBase) GetCountry() string {
	return s.Country
}

func (s SiteAdapterBase) GetSuburbSelector() string {
	return s.SuburbSelector
}

func (s SiteAdapterBase) GetTitleType() string {
	return s.TitleType
}

func (s SiteAdapterBase) GetBodyType() string {
	return s.BodyType
}

func (s SiteAdapterBase) GetPostedDateType() string {
	return s.PostedDateType
}

func (s SiteAdapterBase) GetCityType() string {
	return s.CityType
}

func (s SiteAdapterBase) GetSuburbType() string {
	return s.SuburbType
}

func (s SiteAdapterBase) GetJobSiteNumber(doc *colly.HTMLElement) string {
	return ""
}

func (s SiteAdapterBase) GetPostedDate(doc *colly.HTMLElement) time.Time {
	return time.Now()
}

type UrlJobPathDayDateSite struct {
	SiteAdapterBase
}

func (u UrlJobPathDayDateSite) GetJobSiteNumber(doc *colly.HTMLElement) string {
	url := doc.Request.URL.String()
	return url[strings.Index(url, "/job/")+5 : strings.LastIndex(url, "?")]
}
