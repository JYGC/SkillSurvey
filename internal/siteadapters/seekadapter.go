package siteadapters

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/getapiscraper"

	"github.com/gocolly/colly/v2"
)

const seekConfigFilename = "./seek.json"

type SeekAdapter struct {
	ConfigSettings config.SearchApiSiteAdapterConfig
	ApiClient      *getapiscraper.GetApiScraper
}

func NewSeekAdapter() *SeekAdapter {
	seek := new(SeekAdapter)
	config.JsonToConfig(&seek.ConfigSettings, seekConfigFilename)
	seek.ApiClient = getapiscraper.NewGetApiScraper(
		seek.ConfigSettings,
	)
	return seek
}

func (s SeekAdapter) RunSurvey() []entities.InboundJobPost {
	s.ApiClient.Scrape(func(page int) any {
		seekApiParameters := SeekGetApiParameters{
			Page:                  1,
			NewSince:              "1742971081",
			SiteKey:               "AU-Main",
			SourceSystem:          "houstob",
			UserQueryId:           "aeb5109edbfc379e2a97d0dd748fd81f-1099727",
			UserId:                "bd4c5bde-f33f-4ea4-9257-eb590762f52e",
			UserSessionId:         "bd4c5bde-f33f-4ea4-9257-eb590762f52e",
			EventCaptureSessionId: "bd4c5bde-f33f-4ea4-9257-eb590762f52e",
			Where:                 "All+Melbourne+VIC",
			Classification:        "6281",
			PageSize:              "10",
			Include:               "seodata,relatedsearches,joracrosslink,gptTargeting,pills",
			Locale:                "en-AU",
			SolId:                 "78fc4265-7367-48f8-b9b4-dae834474999",
			RelatedSearchesCount:  "12",
			BaseKeywords:          "",
		}
		return seekApiParameters
	})

	// opts := []chromedp.ExecAllocatorOption{
	// 	chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"),
	// 	chromedp.WindowSize(1920, 1080),
	// 	chromedp.NoFirstRun,
	// 	chromedp.NoDefaultBrowserCheck,
	// 	chromedp.Flag("headless", true),                                 // Headless mode; set to false for headful if needed
	// 	chromedp.Flag("disable-blink-features", "AutomationControlled"), // Hide automation signals
	// }
	// allocatorCtx, AllocatorCancel := chromedp.NewExecAllocator(context.Background(), opts...)
	// defer AllocatorCancel()

	// ctx, cancel := chromedp.NewContext(allocatorCtx)
	// defer cancel()

	// ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	// defer cancel()

	// url := "https://www.seek.com.au/job/84647179"
	// var html string
	// err := chromedp.Run(
	// 	ctx,
	// 	chromedp.Evaluate(`Object.defineProperty(navigator, 'webdriver', {get: () => undefined});`, nil),
	// 	chromedp.Evaluate(`Object.defineProperty(navigator, 'plugins', {get: () => [1, 2, 3, 4, 5]});`, nil),
	// 	chromedp.Navigate(url),
	// 	chromedp.WaitVisible("body", chromedp.ByQueryAll),
	// 	chromedp.Text("[data-automation=\"job-detail-title\"]", &html),
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(strings.TrimSpace(html))

	return []entities.InboundJobPost{}
}

func (s SeekAdapter) getPostedDate(doc *colly.HTMLElement) time.Time {
	variableRef := make(map[string]interface{})
	defer exception.ReportErrorIfPanic(map[string]interface{}{
		"Url":       doc.Request.URL.String(),
		"Variables": variableRef,
	})

	const startString = "Posted "
	const endString = " ago"

	textFromDocument := doc.Text
	startStringStartIndex := strings.Index(textFromDocument, startString)
	variableRef["startStringStartIndex"] = startStringStartIndex
	if startStringStartIndex == -1 {
		panic(fmt.Errorf("can't find %v", startString))
	}
	startStringEndIndex := startStringStartIndex + len(startString)
	variableRef["startStringEndIndex"] = startStringEndIndex
	endStringStartIndex := startStringEndIndex + strings.Index(textFromDocument[startStringEndIndex:], endString)
	variableRef["endStringStartIndex"] = endStringStartIndex

	if endStringStartIndex == -1 {
		// when ageString: Posted 4 Apr 2023
		return turnAgeStringToTime(textFromDocument, startStringEndIndex)
	}
	return turnTimeAgoFormatAgeStringToTime(textFromDocument, startStringEndIndex, endStringStartIndex)
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

func turnAgeStringToTime(textFromDocument string, timeStringIndex int) time.Time {
	stringPartsToTurnToTime := strings.Split(textFromDocument[timeStringIndex:], " ")

	var err error
	day, err := strconv.Atoi(stringPartsToTurnToTime[1])
	if err != nil {
		panic(err)
	}
	monthCode := stringPartsToTurnToTime[1]
	year, err := strconv.Atoi(stringPartsToTurnToTime[2][:4])
	if err != nil {
		panic(err)
	}
	return time.Date(year, time.Month(shortMonthNumMap[monthCode]), day, 0, 0, 0, 0, time.Local)
}

func turnTimeAgoFormatAgeStringToTime(textFromDocument string, startStringEndIndex int, endStringStartIndex int) time.Time {
	postedDate, err := getPostedDateFromDateTimeCalculation(textFromDocument, startStringEndIndex, endStringStartIndex)
	if err != nil {
		panic(err)
	}
	return postedDate
}

func getPostedDateFromDateTimeCalculation(textFromDocument string, startStringEndIndex int, endStringStartIndex int) (postedDate time.Time, err error) {
	timeAgo := textFromDocument[startStringEndIndex:endStringStartIndex]
	currentDate := time.Now()
	switch timeAgoUnit := textFromDocument[endStringStartIndex-1 : endStringStartIndex]; timeAgoUnit {
	case "d":
		var day int
		day, err = strconv.Atoi(timeAgo[:len(timeAgo)-1])
		postedDate = currentDate.AddDate(0, 0, -day)
	case "h", "m", "s":
		var timeAgoDuration time.Duration
		timeAgoDuration, err = time.ParseDuration(timeAgo)
		postedDate = currentDate.Add(-timeAgoDuration)
	case "+":
		postedDate, err = getPostedDateFromDateTimeCalculation(textFromDocument, startStringEndIndex, endStringStartIndex-1)
	default:
		err = errors.New("cannot determine posted time")
	}

	return postedDate, err
}
