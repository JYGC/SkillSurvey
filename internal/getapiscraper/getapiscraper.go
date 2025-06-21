package getapiscraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/entities"
)

type GetApiScraper struct {
	configSettings                     config.SearchApiSiteAdapterConfig
	getInboundJobPostsFromResponseBody func([]byte) (
		[]entities.InboundJobPost,
		error,
	)
}

func NewGetApiScraper(
	configSettings config.SearchApiSiteAdapterConfig,
	getInboundJobPostsFromResponseBody func([]byte) (
		[]entities.InboundJobPost,
		error,
	),
) *GetApiScraper {
	apiClient := &GetApiScraper{
		configSettings:                     configSettings,
		getInboundJobPostsFromResponseBody: getInboundJobPostsFromResponseBody,
	}
	return apiClient
}

func (a GetApiScraper) convertUrlParameterStructToString(
	apiParameters any,
) (
	string,
	error,
) {
	parameterString := ""

	valueOfUrlParameterInterface := reflect.ValueOf(apiParameters)
	typeOfUrlParameterInterface := reflect.TypeOf(apiParameters)

	if valueOfUrlParameterInterface.Kind() == reflect.Ptr {
		valueOfUrlParameterInterface = valueOfUrlParameterInterface.Elem()
		typeOfUrlParameterInterface = typeOfUrlParameterInterface.Elem()
	}

	if valueOfUrlParameterInterface.Kind() != reflect.Struct {
		return "", errors.New("url parameter must be a struct")
	}

	for i := range typeOfUrlParameterInterface.NumField() {
		parameterString = fmt.Sprintf(
			"%s%s=%s&",
			parameterString,
			typeOfUrlParameterInterface.Field(i).Name,
			valueOfUrlParameterInterface.Field(i).Interface(),
		)
	}
	if len(parameterString) > 0 {
		parameterString = parameterString[:len(parameterString)-1]
	}
	return parameterString, nil
}

func (a GetApiScraper) getInboundJobPostsFromPage(
	getApiParameters func(int) any,
	page int,
) (
	inboundJobPosts []entities.InboundJobPost,
	err error,
) {
	urlParameterString, urlParamStringErr :=
		a.convertUrlParameterStructToString(
			getApiParameters(page),
		)
	if urlParamStringErr != nil {
		return nil, urlParamStringErr
	}
	apiUrl := fmt.Sprintf(
		"%s?%s",
		a.configSettings.SearchApiUrl,
		urlParameterString,
	)
	response, responseErr := http.Get(apiUrl)
	if responseErr != nil {
		return nil, responseErr
	}
	defer response.Body.Close()
	body, readBodyErr := io.ReadAll(response.Body)
	if readBodyErr != nil {
		return nil, readBodyErr
	}
	inboundJobPostsFromResponseBody, fromResponseBodyErr :=
		a.getInboundJobPostsFromResponseBody(body)
	if fromResponseBodyErr != nil {
		return nil, fromResponseBodyErr
	}
	inboundJobPosts = append(
		inboundJobPosts,
		inboundJobPostsFromResponseBody...,
	)

	return inboundJobPosts, err
}

func (a GetApiScraper) Scrape(
	getApiParameters func(int) any,
) (
	inboundJobPosts []entities.InboundJobPost,
	err error,
) {
	var pageErrors []error
	for page := 1; page <= a.configSettings.Pages; page++ {
		pageResults, pageError := a.getInboundJobPostsFromPage(
			getApiParameters,
			page,
		)

		pageErrors = append(pageErrors, pageError)
		inboundJobPosts = append(inboundJobPosts, pageResults...)
	}

	if len(pageErrors) > 0 {
		err = fmt.Errorf("Page errors: %v", pageErrors)
	}

	return inboundJobPosts, err
}
