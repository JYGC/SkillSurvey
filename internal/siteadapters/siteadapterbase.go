package siteadapters

import (
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/gocolly/colly/v2"
)

type ISiteAdapter interface {
	GetConfigSettings() config.SiteAdapterConfig
	GetJobSiteNumber(doc *colly.HTMLElement) string
	GetPostedDate(doc *colly.HTMLElement) time.Time
}

type SiteAdapterBase struct {
	ConfigSettings config.SiteAdapterConfig
}

func (s SiteAdapterBase) GetConfigSettings() config.SiteAdapterConfig {
	return s.ConfigSettings
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
