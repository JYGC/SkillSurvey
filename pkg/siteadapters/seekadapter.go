package siteadapters

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/pkg/config"
	"github.com/gocolly/colly/v2"
)

const seekConfigPath = "./pkg/siteadapters/seek.json"

type SeekAdapter struct {
	UrlJobPathDayDateSite
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

func (s SeekAdapter) GetPostedDate(doc *colly.HTMLElement) time.Time {
	ageString := doc.ChildText(".yvsb870._14uh9942c._1qw3t4i0._1qw3t4ix._1qw3t4i1.xn3fpb4._1qw3t4i9")
	timeAgoIndex := strings.Index(ageString, "Posted ") + 7
	agoIndex := strings.Index(ageString, " ago")
	timeAgo := ageString[timeAgoIndex:agoIndex]
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
		//TODO: Error handling
		fmt.Println(err.Error())
	}
	return postedDate
}
