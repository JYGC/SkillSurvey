package siteadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"keybook/runtask/internal/dynamiccontentextractor"
	"keybook/runtask/internal/exception"
)

// SeekAdapter scrapes job listings from the Seek API.
type SeekAdapter struct {
	configSettings          SeekAdapterConfig
	dynamicContentExtractor *dynamiccontentextractor.DynamicContentExtractor
}

// NewSeekAdapter loads adapter configuration from the provided JSON file path.
func NewSeekAdapter(configFilePath string) (*SeekAdapter, error) {
	seek := new(SeekAdapter)
	if err := loadJSON(configFilePath, &seek.configSettings); err != nil {
		return nil, fmt.Errorf("load seek config %s: %w", configFilePath, err)
	}
	seek.dynamicContentExtractor = dynamiccontentextractor.NewDynamicContentExtractor()
	return seek, nil
}

func (s SeekAdapter) RunSurvey() ([]InboundJobPost, error) {
	return scrapeAPI(
		s.configSettings.SearchApiUrl,
		s.configSettings.Pages,
		len(s.configSettings.ApiParameters),
		func(page int, apiParameterSetNumber int) any {
			newSince := time.Now().Add(-time.Hour * 24 * time.Duration(
				s.configSettings.ApiParameters[apiParameterSetNumber].NewSinceDaysAgo,
			))
			return SeekGetApiParameters{
				Page:                  strconv.Itoa(page),
				NewSince:              strconv.FormatInt(newSince.Unix(), 10),
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
		func(body []byte) ([]InboundJobPost, error) {
			var bodyJsonMap map[string]any
			json.Unmarshal(body, &bodyJsonMap)
			dataBytes, dataBytesErr := json.Marshal(bodyJsonMap["data"])
			if dataBytesErr != nil {
				return nil, dataBytesErr
			}
			var dataJsonMaps []map[string]any
			json.Unmarshal(dataBytes, &dataJsonMaps)
			var newInboundJobPosts []InboundJobPost
			var jobPostErrors []error
			for _, dataJsonMap := range dataJsonMaps {
				jobSiteNumber := dataJsonMap["id"].(string)
				url := fmt.Sprintf("%s/job/%s", s.configSettings.BaseUrl, jobSiteNumber)
				fmt.Printf("url: %v\n", url)

				newInboundJobPost := InboundJobPost{}
				newInboundJobPostErr := s.dynamicContentExtractor.ExtractDynamicContent(
					url,
					func(ctx context.Context) (err error) {
						var errParts []error
						if getTitleErr := dynamiccontentextractor.GetTextBySelector(s.configSettings.SiteSelectors.TitleSelector, &newInboundJobPost.Title, ctx); getTitleErr != nil {
							errParts = append(errParts, fmt.Errorf("getTitleErr: %v", getTitleErr))
						}
						if getBodyErr := dynamiccontentextractor.GetTextBySelector(s.configSettings.SiteSelectors.BodySelector, &newInboundJobPost.Body, ctx); getBodyErr != nil {
							errParts = append(errParts, fmt.Errorf("getBodyErr: %v", getBodyErr))
						}
						if len(errParts) > 0 {
							err = fmt.Errorf("%v", errParts)
						}
						return err
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
				newInboundJobPost.Country = locationJsonMap["countryCode"].(string)
				newInboundJobPost.Suburb = locationJsonMap["label"].(string)
				postedDate, postedDateErr := time.Parse(time.RFC3339, dataJsonMap["listingDate"].(string))
				if postedDateErr != nil {
					jobPostErrors = append(jobPostErrors, postedDateErr)
					exception.LogErrorWithLabel("postedDateErr", postedDateErr)
					continue
				}
				newInboundJobPost.PostedDate = postedDate
				newInboundJobPosts = append(newInboundJobPosts, newInboundJobPost)
			}
			var err error
			if len(jobPostErrors) > 0 {
				err = fmt.Errorf("jobPostErrors: %v", jobPostErrors)
			}
			return newInboundJobPosts, err
		},
	)
}

// ── internal API scraper ──────────────────────────────────────────────────────

func scrapeAPI(
	searchApiUrl string,
	numberOfPages int,
	numberOfApiParameterSets int,
	getApiParametersForPage func(int, int) any,
	getItemsFromBody func([]byte) ([]InboundJobPost, error),
) ([]InboundJobPost, error) {
	var allPosts []InboundJobPost
	var pageErrors []error

	for page := 1; page <= numberOfPages; page++ {
		for apiParamSet := range numberOfApiParameterSets {
			posts, err := fetchPage(searchApiUrl, page, apiParamSet, getApiParametersForPage, getItemsFromBody)
			if err != nil {
				pageErrors = append(pageErrors, err)
			}
			allPosts = append(allPosts, posts...)
		}
	}

	var err error
	if len(pageErrors) > 0 {
		err = fmt.Errorf("page errors: %v", pageErrors)
	}
	return allPosts, err
}

func fetchPage(
	searchApiUrl string,
	page, apiParamSet int,
	getApiParametersForPage func(int, int) any,
	getItemsFromBody func([]byte) ([]InboundJobPost, error),
) ([]InboundJobPost, error) {
	paramString, err := structToQueryString(getApiParametersForPage(page, apiParamSet))
	if err != nil {
		return nil, err
	}
	apiUrl := fmt.Sprintf("%s?%s", searchApiUrl, paramString)
	fmt.Printf("apiUrl: %v\n", apiUrl)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return getItemsFromBody(body)
}

func structToQueryString(params any) (string, error) {
	v := reflect.ValueOf(params)
	t := reflect.TypeOf(params)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("params must be a struct")
	}
	result := ""
	for i := range t.NumField() {
		val := v.Field(i).Interface().(string)
		if val != "" {
			result += fmt.Sprintf("%s=%s&", t.Field(i).Name, val)
		}
	}
	if len(result) > 0 {
		result = result[:len(result)-1]
	}
	return result, nil
}

// loadJSON decodes a JSON file at path into dst.
func loadJSON(path string, dst any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(dst)
}
