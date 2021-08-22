package siteadapters

import (
	"strings"

	"github.com/JYGC/SkillSurvey/pkg/config"
)

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
	jora.BodySelector = "#job-description-container"
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

func (j JoraAdapter) GetJobSiteNumber(url string, doc string) string {
	return url[strings.Index(url, "/job/")+5 : strings.LastIndex(url, "?")]
}

// Advertisement's post date can calculated by subtracting how old the advert is in days from
// the current date.
func (j JoraAdapter) GetPostedDate(url string, doc string) string {
	var ageString = doc.$('.date').text(); // .date contains either "N days" or "today"
	var daysOld = 0;
	var daysIndex = ageString.indexOf("days");
	var currentDate = new Date();
	var postedDate = new Date();

	if (daysIndex !== -1) {
		// If "today", the advert is 0 days old, leave daysOld as 0
		daysOld = parseInt(ageString.substring(0, daysIndex - 1));
	}

	// setDate changes postedDate to next month if argument is more then number of days in
	// current month
	postedDate.setDate(currentDate.getDate() - daysOld);
	
	return postedDate.toISOString();
}
