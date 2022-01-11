package siteadapters

import (
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/gocolly/colly/v2"
)

const joraConfigPath = "./internal/siteadapters/jora.json"

type JoraAdapter struct {
	UrlJobPathDayDateSite
}

func NewJoraAdapter() *JoraAdapter {
	jora := new(JoraAdapter)
	config.JsonToConfig(&jora.ConfigSettings, joraConfigPath)
	return jora
}

// Advertisement's post date can calculated by subtracting how old the advert is in days from
// the current date.
func (j JoraAdapter) GetPostedDate(doc *colly.HTMLElement) time.Time {
	variableRef := make(map[string]interface{})
	defer exception.ReportErrorIfPanic(map[string]interface{}{
		"Url":       doc.Request.URL.String(),
		"Variables": variableRef,
	})
	ageString := doc.ChildText(j.ConfigSettings.SiteSelectors.PostedDateSelector) // .date contains either "N days" or "today"
	variableRef["ageString"] = ageString
	daysOld := 0
	daysIndex := strings.Index(ageString, "days")
	variableRef["daysIndex"] = daysIndex
	currentDate := time.Now()
	postedDate := time.Now()
	// TODO: support hours and minutes ago
	if daysIndex != -1 {
		// If "today", the advert is 0 days old, leave daysOld as 0
		var err error
		daysOld, err = strconv.Atoi(ageString[0 : daysIndex-1])
		if err != nil {
			panic(err)
		}
	}
	// set to next month if argument is more then number of days in
	// current month
	postedDate = currentDate.AddDate(0, 0, -daysOld)
	return postedDate
}
