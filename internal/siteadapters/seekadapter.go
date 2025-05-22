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
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/gocolly/colly/v2"
)

const seekConfigFilename = "./seek.json"

type SeekAdapter struct {
	SiteAdapterBase
}

func NewSeekAdapter() *SeekAdapter {
	seek := new(SeekAdapter)
	config.JsonToConfig(&seek.ConfigSettings, seekConfigFilename)
	return seek
}

func (s SeekAdapter) RunSurvey() []entities.InboundJobPost {
	// httpResponse, httpResponseError := http.Get("https://www.seek.com.au/api/jobsearch/v5/search?newSince=1742971081&siteKey=AU-Main&sourcesystem=houston&userqueryid=aeb5109edbfc379e2a97d0dd748fd81f-1099727&userid=bd4c5bde-f33f-4ea4-9257-eb590762f52e&usersessionid=bd4c5bde-f33f-4ea4-9257-eb590762f52e&eventCaptureSessionId=bd4c5bde-f33f-4ea4-9257-eb590762f52e&where=All+Melbourne+VIC&page=1&classification=6281&pageSize=10&include=seodata,relatedsearches,joracrosslink,gptTargeting,pills&locale=en-AU&solId=78fc4265-7367-48f8-b9b4-dae834474999&relatedSearchesCount=12&baseKeywords=")
	// if httpResponseError != nil {
	// 	fmt.Printf("httpResponseError: %v\n", httpResponseError)
	// }
	// defer httpResponse.Body.Close()
	// body, readAllErr := io.ReadAll(httpResponse.Body)
	// if readAllErr != nil {
	// 	fmt.Printf("readAllErr: %v\n", readAllErr)
	// }
	// fmt.Printf("body: %v\n", string(body))

	// var query struct {
	// 	JobDetails struct {
	// 		JobId     graphql.ID
	// 		SessionId graphql.String
	// 	}
	// }
	// client := graphql.NewClient("https://www.seek.com.au/graphql", nil)

	// requestPayload := map[string]interface{}{
	// 	"jobId":                         graphql.ID("83066340"),
	// 	"jobDetailsViewedCorrelationId": graphql.String("95098f02-628a-4e11-ab02-822719e48b51"),
	// 	"sessionId":                     graphql.String("bd4c5bde-f33f-4ea4-9257-eb590762f52e"),
	// 	"zone":                          graphql.String("anz-1"),
	// 	"locale":                        graphql.String("en-AU"),
	// 	"languageCode":                  graphql.String("en"),
	// 	"countryCode":                   graphql.String("AU"),
	// 	"timezone":                      graphql.String("Australia/Melbourne"),
	// }

	// err := client.Query(context.Background(), &query, requestPayload)
	// if err != nil {
	// 	fmt.Printf("err: %v\n", err)
	// }

	// bin, err := launcher.NewBrowser().Get()
	// args := launcher.New().Leakless(false).FormatArgs()

	// parser := launcher.NewURLParser()

	// cmd := exec.Command(bin, args...)
	// cmd.Stderr = parser
	// err = cmd.Start()
	// utils.E(err)

	browser := rod.New().MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("https://www.seek.com.au/job/83878134").MustWaitStable()
	page.Keyboard.MustType(input.Slash)
	//page.MustElement("[data-automation=\"job-detail-title\"]")

	text := page.MustElement("[data-automation=\"job-detail-title\"]").MustText()
	fmt.Println(text)

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
