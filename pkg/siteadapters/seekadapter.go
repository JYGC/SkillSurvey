package siteadapters

import (
	"strings"

	"github.com/JYGC/SkillSurvey/pkg/config"
)

const seekConfigPath = "./pkg/siteadapters/seek.json"

type SeekAdapter struct {
	SiteAdapterBase
}

func NewSeekAdapter() *SeekAdapter {
	seek := new(SeekAdapter)
	config.JsonToConfig(&seek.ConfigSettings, seekConfigPath)
	seek.SiteName = seek.ConfigSettings.BaseUrl
	seek.JobPostLink = "[data-automation=\"jobTitle\"]"
	seek.TitleSelector = "[data-automation=\"job-detail-title\"]"
	seek.BodySelector = "[data-automation=\"jobAdDetails\"]"
	seek.PostedDateSelector = "span.FYwKg._2Bz3E.C6ZIU_4._6ufcS_4._3KSG8_4._29m7__4._2WTa0_4"
	seek.CitySelector = "div.FYwKg._3VxpE_4 > div:nth-child(1)"
	seek.Country = "Australia"
	seek.SuburbSelector = "div.FYwKg._3VxpE_4 > div:nth-child(2)"
	seek.TitleType = "text"
	seek.BodyType = seek.TitleType
	seek.PostedDateType = seek.TitleType
	seek.CityType = seek.TitleType
	seek.SuburbType = seek.TitleType
	return seek
}

func (s SeekAdapter) GetJobSiteNumber(url string, doc string) string {
	return url[strings.Index(url, "/job/")+5 : strings.LastIndex(url, "?")]
}

func (s SeekAdapter) GetPostedDate(url string, doc string) string {
	return ""
}
