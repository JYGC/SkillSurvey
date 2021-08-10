package siteadapters

import (
	"strings"

	"github.com/JYGC/SkillSurvey/pkg/config"
)

const seekConfigPath = "./pkg/siteadapters/Seek.json"

type Seek struct {
	SiteAdapterBase
}

func NewSeek() *Seek {
	seekAdapter := new(Seek)
	config.JsonToConfig(&seekAdapter.ConfigSettings, seekConfigPath)
	seekAdapter.SiteName = "Seek.com.au"
	seekAdapter.JobPostLink = "[data-automation=\"jobTitle\"]"
	seekAdapter.TitleSelector = "[data-automation=\"job-detail-title\"]"
	seekAdapter.BodySelector = "[data-automation=\"jobAdDetails\"]"
	seekAdapter.PostedDateSelector = "span.FYwKg._2Bz3E.C6ZIU_4._6ufcS_4._3KSG8_4._29m7__4._2WTa0_4"
	seekAdapter.CitySelector = "div.FYwKg._3VxpE_4 > div:nth-child(1)"
	seekAdapter.Country = "Australia"
	seekAdapter.SuburbSelector = "div.FYwKg._3VxpE_4 > div:nth-child(2)"
	seekAdapter.TitleType = "text"
	seekAdapter.BodyType = seekAdapter.TitleType
	seekAdapter.PostedDateType = seekAdapter.TitleType
	seekAdapter.CityType = seekAdapter.TitleType
	seekAdapter.SuburbType = seekAdapter.TitleType
	return seekAdapter
}

func (s Seek) GetJobSiteNumber(url string, doc string) string {
	return url[strings.Index(url, "/job/")+5 : strings.LastIndex(url, "?")]
}

func (s Seek) GetPostedDate(url string, doc string) string {
	return ""
}
