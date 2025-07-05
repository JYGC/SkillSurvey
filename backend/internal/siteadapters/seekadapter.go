package siteadapters

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/dynamiccontentextractor"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/getapiscraper"

	"github.com/gocolly/colly/v2"
)

const seekConfigFilename = "./seek.json"

type SeekAdapter struct {
	configSettings          config.SearchApiSiteAdapterConfig
	apiScraper              *getapiscraper.GetApiScraper
	dynamicContentExtractor dynamiccontentextractor.DynamicContentExtractor
}

func NewSeekAdapter() *SeekAdapter {
	seek := new(SeekAdapter)
	config.JsonToConfig(&seek.configSettings, seekConfigFilename)
	seek.dynamicContentExtractor =
		dynamiccontentextractor.NewDynamicContentExtractor(seek.configSettings)
	seek.apiScraper = getapiscraper.NewGetApiScraper(
		seek.configSettings,
		func(
			body []byte,
		) (
			newInboundJobPosts []entities.InboundJobPost,
			err error,
		) {
			var bodyJsonMap map[string]any
			json.Unmarshal(body, &bodyJsonMap)
			dataBytes, dataBytesErr := json.Marshal(bodyJsonMap["data"])
			if dataBytesErr != nil {
				return nil, dataBytesErr
			}
			var dataJsonMaps []map[string]any
			json.Unmarshal(dataBytes, &dataJsonMaps)
			for _, dataJsonMap := range dataJsonMaps {
				jobSiteNumber := dataJsonMap["id"].(string)
				url := fmt.Sprintf(
					"%s/job/%s",
					seek.configSettings.BaseUrl,
					jobSiteNumber,
				)
				fmt.Printf("url: %v\n", url)
				//fmt.Printf("dataJsonMap: %v\n", dataJsonMap)

				newInboundJobPost, newInboundJobPostErr :=
					seek.dynamicContentExtractor.GetInboundJobPost(url)
				if newInboundJobPostErr != nil {
					return nil, newInboundJobPostErr
				}
				newInboundJobPost.SiteName = seek.configSettings.SiteSelectors.SiteName
				newInboundJobPost.JobSiteNumber = jobSiteNumber
				locationJsonMap := dataJsonMap["locations"].([]any)[0].(map[string]any)
				countryCode := locationJsonMap["countryCode"]
				label := locationJsonMap["label"]
				newInboundJobPost.Country = countryCode.(string)
				newInboundJobPost.Suburb = label.(string)
				postedDate, postedDateErr := time.Parse(time.RFC3339, dataJsonMap["listingDate"].(string))
				if postedDateErr != nil {
					return nil, postedDateErr
				}
				newInboundJobPost.PostedDate = postedDate
				newInboundJobPosts = append(newInboundJobPosts, newInboundJobPost)
			}
			return newInboundJobPosts, nil
		},
	)
	return seek
}

func (s SeekAdapter) RunSurvey() []entities.InboundJobPost {
	inboundJobPosts, scrapeErr := s.apiScraper.Scrape(func(page int) any {
		newSince := time.Now().Add(-time.Hour * 24 * 90)
		newSinceUnix := newSince.Unix()
		return SeekGetApiParameters{
			// Move to config
			Page:                  strconv.Itoa(page),
			NewSince:              strconv.FormatInt(newSinceUnix, 10),
			SiteKey:               "AU-Main",
			SourceSystem:          "houston",
			UserQueryId:           "aeb5109edbfc379e2a97d0dd748fd81f-1099727",
			UserId:                "bd4c5bde-f33f-4ea4-9257-eb590762f52e",
			UserSessionId:         "bd4c5bde-f33f-4ea4-9257-eb590762f52e",
			EventCaptureSessionId: "bd4c5bde-f33f-4ea4-9257-eb590762f52e",
			Where:                 "All+Melbourne+VIC",
			Classification:        "6281",
			PageSize:              "10",
			//Include:               "", //"seodata,relatedsearches,joracrosslink,gptTargeting,pills",
			Locale:               "en-AU",
			SolId:                "78fc4265-7367-48f8-b9b4-dae834474999",
			RelatedSearchesCount: "12",
			BaseKeywords:         "",
		}
	})
	if scrapeErr != nil {
		panic(scrapeErr)
	}
	return inboundJobPosts
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
