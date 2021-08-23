package siteadapters

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/pkg/config"
	"github.com/gocolly/colly/v2"
)

const joraConfigPath = "./pkg/siteadapters/jora.json"

type JoraAdapter struct {
	UrlJobPathDayDateSite
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

// Advertisement's post date can calculated by subtracting how old the advert is in days from
// the current date.
func (j JoraAdapter) GetPostedDate(doc *colly.HTMLElement) time.Time {
	ageString := doc.ChildText(".listed-date") // .date contains either "N days" or "today"
	daysOld := 0
	daysIndex := strings.Index(ageString, "days")
	currentDate := time.Now()
	postedDate := time.Now()
	// TODO: support hours and minutes ago
	if daysIndex != -1 {
		// If "today", the advert is 0 days old, leave daysOld as 0
		var err error
		daysOld, err = strconv.Atoi(ageString[0 : daysIndex-1])
		if err != nil {
			//TODO: Error handling
			fmt.Println("Failed to get days")
		}
	}
	// set to next month if argument is more then number of days in
	// current month
	postedDate = currentDate.AddDate(0, 0, -daysOld)
	return postedDate
}
