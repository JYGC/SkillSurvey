package apiclient

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/entities"
)

type ApiClient struct {
	configSettings config.SearchApiSiteAdapterConfig
	endpointPath   string
}

func NewApiClient(
	configSettings config.SearchApiSiteAdapterConfig,
) *ApiClient {
	apiClient := &ApiClient{
		configSettings: configSettings,
		endpointPath:   configSettings.SearchApiUrl,
	}
	return apiClient
}

func (a ApiClient) convertUrlParameterStructToString(
	urlParameterInterface interface{},
) (
	string,
	error,
) {
	parameterString := ""

	valueOfUrlParameterInterface := reflect.ValueOf(urlParameterInterface)
	typeOfUrlParameterInterface := reflect.TypeOf(urlParameterInterface)

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

func (a ApiClient) Get(
	urlParameterStruct interface{},
) (
	[]entities.InboundJobPost,
	error,
) {
	urlParameterString, urlParamStringErr := a.convertUrlParameterStructToString(
		urlParameterStruct,
	)
	if urlParamStringErr != nil {
		return nil, urlParamStringErr
	}
	test := fmt.Sprintf(
		"%s?%s",
		a.endpointPath,
		urlParameterString,
	)
	response, responseErr := http.Get(test)
	if responseErr != nil {
		return nil, responseErr
	}
	defer response.Body.Close()
	body, readBodyErr := io.ReadAll(response.Body)
	if readBodyErr != nil {
		return nil, readBodyErr
	}
	fmt.Printf("body: %v\n", string(body))

	return []entities.InboundJobPost{}, nil
}
