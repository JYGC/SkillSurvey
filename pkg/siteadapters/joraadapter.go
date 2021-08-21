package siteadapters

import "github.com/JYGC/SkillSurvey/pkg/config"

const joraConfigPath = "./pkg/siteadapters/jora.json"

type JoraAdapter struct {
	SiteAdapterBase
}

func NewJoraAdapter() *JoraAdapter {
	jora := new(JoraAdapter)
	config.JsonToConfig(&jora.ConfigSettings, joraConfigPath)
	jora.SiteName = jora.ConfigSettings.BaseUrl
	jora.JobPostLink = ".job-item"
	jora.TitleSelector = "h3.job-title.heading-xxlarge"
	// jora.PostedDateSelector not used here
	jora.CitySelector = ".location"
	jora.Country = "Australia"
	jora.SuburbSelector = ".location"
	jora.TitleType = "text"
	jora.BodyType = jora.TitleType
	jora.PostedDateType = jora.TitleType
	jora.CityType = jora.TitleType
	jora.SuburbType = jora.TitleType
	return jora
}
