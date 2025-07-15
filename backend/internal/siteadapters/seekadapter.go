package siteadapters

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/dynamiccontentextractor"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/environment"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/getapiscraper"
)

const seekConfigFilename = "./seek.json"

type SeekAdapter struct {
	configSettings          SeekAdapterConfig
	apiScraper              *getapiscraper.GetApiScraper
	dynamicContentExtractor dynamiccontentextractor.DynamicContentExtractor
}

func NewSeekAdapter() *SeekAdapter {
	seek := new(SeekAdapter)
	config.JsonToConfig(
		&seek.configSettings,
		environment.AttachToExecutableDir(seekConfigFilename),
	)
	seek.dynamicContentExtractor =
		dynamiccontentextractor.NewDynamicContentExtractor()
	seek.apiScraper = getapiscraper.NewGetApiScraper(
		seek.configSettings.SearchApiUrl,
	)
	return seek
}

func (s SeekAdapter) RunSurvey() (
	[]entities.InboundJobPost,
	error,
) {
	return s.apiScraper.Scrape(
		s.configSettings.Pages,
		len(s.configSettings.ApiParameters),
		func(page int, apiParameterSetNumber int) any {
			newSince := time.Now().Add(-time.Hour * 24 * time.Duration(
				s.configSettings.ApiParameters[apiParameterSetNumber].NewSinceDaysAgo,
			))
			newSinceUnix := newSince.Unix()
			return SeekGetApiParameters{
				Page:                  strconv.Itoa(page),
				NewSince:              strconv.FormatInt(newSinceUnix, 10),
				SiteKey:               s.configSettings.ApiParameters[apiParameterSetNumber].SiteKey,
				SourceSystem:          s.configSettings.ApiParameters[apiParameterSetNumber].SourceSystem,
				UserQueryId:           s.configSettings.ApiParameters[apiParameterSetNumber].UserQueryId,
				UserId:                s.configSettings.ApiParameters[apiParameterSetNumber].UserId,
				UserSessionId:         s.configSettings.ApiParameters[apiParameterSetNumber].UserSessionId,
				EventCaptureSessionId: s.configSettings.ApiParameters[apiParameterSetNumber].EventCaptureSessionId,
				Where:                 s.configSettings.ApiParameters[apiParameterSetNumber].Where,
				Classification:        s.configSettings.ApiParameters[apiParameterSetNumber].Classification,
				PageSize:              s.configSettings.ApiParameters[apiParameterSetNumber].PageSize,
				Include:               s.configSettings.ApiParameters[apiParameterSetNumber].Include,
				Locale:                s.configSettings.ApiParameters[apiParameterSetNumber].Locale,
				SolId:                 s.configSettings.ApiParameters[apiParameterSetNumber].SolId,
				RelatedSearchesCount:  s.configSettings.ApiParameters[apiParameterSetNumber].RelatedSearchesCount,
				BaseKeywords:          s.configSettings.ApiParameters[apiParameterSetNumber].BaseKeywords,
			}
		},
		func(body []byte) (
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
			var jobPostErrors []error
			for _, dataJsonMap := range dataJsonMaps {
				jobSiteNumber := dataJsonMap["id"].(string)
				url := fmt.Sprintf(
					"%s/job/%s",
					s.configSettings.BaseUrl,
					jobSiteNumber,
				)
				fmt.Printf("url: %v\n", url)
				//fmt.Printf("dataJsonMap: %v\n", dataJsonMap)

				newInboundJobPost, newInboundJobPostErr :=
					s.dynamicContentExtractor.GetInboundJobPost(
						url,
						func(
							newInboundJobPost *entities.InboundJobPost,
						) map[string]*string {
							return map[string]*string{
								s.configSettings.SiteSelectors.TitleSelector: &newInboundJobPost.Title,
								s.configSettings.SiteSelectors.BodySelector:  &newInboundJobPost.Body,
							}
						},
					)
				if newInboundJobPostErr != nil {
					jobPostErrors = append(jobPostErrors, newInboundJobPostErr)
					exception.LogErrorWithLabel("newInboundJobPostErr", newInboundJobPostErr)
					continue
				}
				newInboundJobPost.SiteName = s.configSettings.SiteSelectors.SiteName
				newInboundJobPost.JobSiteNumber = jobSiteNumber
				locationJsonMap := dataJsonMap["locations"].([]any)[0].(map[string]any)
				countryCode := locationJsonMap["countryCode"]
				label := locationJsonMap["label"]
				newInboundJobPost.Country = countryCode.(string)
				newInboundJobPost.Suburb = label.(string)
				postedDate, postedDateErr := time.Parse(
					time.RFC3339,
					dataJsonMap["listingDate"].(string),
				)
				if postedDateErr != nil {
					jobPostErrors = append(jobPostErrors, postedDateErr)
					exception.LogErrorWithLabel("postedDateErr", postedDateErr)
					continue
				}
				newInboundJobPost.PostedDate = postedDate
				newInboundJobPosts = append(newInboundJobPosts, newInboundJobPost)
			}
			if len(jobPostErrors) > 0 {
				err = fmt.Errorf("jobPostErrors: %v", jobPostErrors)
			}
			return newInboundJobPosts, err
		},
	)
}
