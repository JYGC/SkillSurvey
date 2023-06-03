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

const seekConfigPath = "./seek.json"

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
	timeStringIndex := strings.Index(ageString, "Posted ") + 7
	variableRef["timeAgoIndex"] = timeStringIndex
	agoWordIndex := strings.Index(ageString, " ago")
	variableRef["agoWordIndex"] = agoWordIndex

	if agoWordIndex == -1 {
		// when ageString: Posted 4 Apr 2023
		return turnAgeStringToTime(ageString, timeStringIndex)
	}
	return turnTimeAgoFormatAgeStringToTime(ageString, timeStringIndex, agoWordIndex)
}

var shortMonthNumMap map[string]int = map[string]int{
	"Jan": 1,
	"Feb": 2,
	"Mar": 3,
	"Apr": 4,
	"May": 5,
	"Jun": 6,
	"Jul": 7,
	"Aug": 8,
	"Sep": 9,
	"Oct": 10,
	"Nov": 11,
	"Dec": 12,
}

func turnAgeStringToTime(ageString string, timeStringIndex int) time.Time {
	variableRef := make(map[string]interface{})
	defer exception.ReportErrorIfPanic(map[string]interface{}{
		"func": "getPostDateFromTimeAgoFormat",
		"parameters": map[string]interface{}{
			"ageString":       ageString,
			"timeStringIndex": timeStringIndex,
		},
		"Variables": variableRef,
	})

	stringPartsToTurnToTime := strings.Split(ageString[timeStringIndex:], " ")
	variableRef["stringToTurnToTimeParts"] = strings.Join(stringPartsToTurnToTime, ",")

	var err error
	day, err := strconv.Atoi(stringPartsToTurnToTime[1])
	if err != nil {
		panic(err)
	}
	monthCode := stringPartsToTurnToTime[1]
	year, err := strconv.Atoi(stringPartsToTurnToTime[2])
	if err != nil {
		panic(err)
	}
	return time.Date(year, time.Month(shortMonthNumMap[monthCode]), day, 0, 0, 0, 0, time.Local)
}

func turnTimeAgoFormatAgeStringToTime(ageString string, timeStringIndex int, agoWordIndex int) time.Time {
	variableRef := make(map[string]interface{})
	defer exception.ReportErrorIfPanic(map[string]interface{}{
		"func": "getPostDateFromTimeAgoFormat",
		"parameters": map[string]interface{}{
			"ageString":       ageString,
			"timeStringIndex": timeStringIndex,
			"agoWordIndex":    agoWordIndex,
		},
		"Variables": variableRef,
	})

	timeAgo := ageString[timeStringIndex:agoWordIndex]
	variableRef["timeAgo"] = timeAgo
	currentDate := time.Now()
	var postedDate time.Time
	var err error
	switch timeAgoUnit := ageString[agoWordIndex-1 : agoWordIndex]; timeAgoUnit {
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
