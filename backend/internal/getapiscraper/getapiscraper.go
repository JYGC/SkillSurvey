package getapiscraper

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/JYGC/SkillSurvey/internal/entities"
)

type GetApiScraper struct {
	SearchApiUrl string
}

func NewGetApiScraper(
	searchApiUrl string,

) *GetApiScraper {
	apiClient := &GetApiScraper{
		SearchApiUrl: searchApiUrl,
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
		parameterValue := valueOfUrlParameterInterface.Field(i).Interface().(string)
		if len(parameterValue) > 0 {
			parameterString = fmt.Sprintf(
				"%s%s=%s&",
				parameterString,
				typeOfUrlParameterInterface.Field(i).Name,
				parameterValue,
			)
		}
	}
	if len(parameterString) > 0 {
		parameterString = parameterString[:len(parameterString)-1]
	}
	return parameterString, nil
}

func (a GetApiScraper) getInboundJobPostsFromPage(
	page int,
	apiParameterSetNumber int,
	getApiParametersForPage func(int, int) any,
	getInboundJobPostsFromResponseBody func([]byte) (
		[]entities.InboundJobPost,
		error,
	),
) (
	inboundJobPosts []entities.InboundJobPost,
	err error,
) {
	urlParameterString, urlParamStringErr :=
		a.convertUrlParameterStructToString(
			getApiParametersForPage(page, apiParameterSetNumber),
		)
	if urlParamStringErr != nil {
		return nil, urlParamStringErr
	}
	apiUrl := fmt.Sprintf(
		"%s?%s",
		a.SearchApiUrl,
		urlParameterString,
	)
	fmt.Printf("apiUrl: %v\n", apiUrl)
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
		getInboundJobPostsFromResponseBody(body)
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
	numberOfPages int,
	numberOfApiParameterSets int,
	getApiParametersForPage func(int, int) any,
	getInboundJobPostsFromResponseBody func([]byte) (
		[]entities.InboundJobPost,
		error,
	),
) (
	inboundJobPosts []entities.InboundJobPost,
	err error,
) {
	var pageErrors []error
	for page := 1; page <= numberOfPages; page++ {
		for apiParameterSetNumber := range numberOfApiParameterSets {
			pageResults, pageError := a.getInboundJobPostsFromPage(
				page,
				apiParameterSetNumber,
				getApiParametersForPage,
				getInboundJobPostsFromResponseBody,
			)
			if pageError != nil {
				pageErrors = append(pageErrors, pageError)
			}
			inboundJobPosts = append(inboundJobPosts, pageResults...)
		}
	}
	if len(pageErrors) > 0 {
		err = fmt.Errorf("Page errors: %v", pageErrors)
	}

	return inboundJobPosts, err
}
