package siteadapters

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/gocolly/colly/v2"
)

const seekConfigPath = "./internal/siteadapters/seek.json"

type SeekAdapter struct {
	UrlJobPathDayDateSite
}

func NewSeekAdapter() *SeekAdapter {
	seek := new(SeekAdapter)
	config.JsonToConfig(&seek.ConfigSettings, seekConfigPath)
	return seek
}

func (s SeekAdapter) GetPostedDate(doc *colly.HTMLElement) time.Time {
	variableRef := make(map[string]interface{})
	defer exception.ReportErrorIfPanic(map[string]interface{}{
		"Url":       doc.Request.URL.String(),
		"Variables": variableRef,
	})
	ageString := doc.ChildText(s.ConfigSettings.SiteSelectors.PostedDateSelector)
	variableRef["ageString"] = ageString
	timeAgoIndex := strings.Index(ageString, "Posted ") + 7
	variableRef["timeAgoIndex"] = timeAgoIndex
	agoIndex := strings.Index(ageString, " ago")
	variableRef["agoIndex"] = agoIndex
	timeAgo := ageString[timeAgoIndex:agoIndex]
	variableRef["timeAgo"] = timeAgo
	currentDate := time.Now()
	var postedDate time.Time
	var err error
	switch timeAgoUnit := ageString[agoIndex-1 : agoIndex]; timeAgoUnit {
	case "d":
		var day int
		day, err = strconv.Atoi(timeAgo[:len(timeAgo)-1])
		postedDate = currentDate.AddDate(0, 0, -day)
	case "h", "m", "s":
		var timeAgoDuration time.Duration
		timeAgoDuration, err = time.ParseDuration(timeAgo)
		postedDate = currentDate.Add(-timeAgoDuration)
	default:
		err = errors.New("cannot determine posted time")
	}
	if err != nil {
		panic(err)
	}
	return postedDate
}
