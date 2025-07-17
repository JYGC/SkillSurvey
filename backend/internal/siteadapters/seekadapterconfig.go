package siteadapters

type Selectors struct {
	SiteName      string
	TitleSelector string
	BodySelector  string
}

type SearchCriteria struct {
	Name string
	Url  string
}

type ApiParameters struct {
	NewSinceDaysAgo       int
	SiteKey               string
	SourceSystem          string
	UserQueryId           string
	UserId                string
	UserSessionId         string
	EventCaptureSessionId string
	Where                 string
	Classification        string
	PageSize              string
	Include               string
	Locale                string
	SolId                 string
	RelatedSearchesCount  string
	BaseKeywords          string
}

type SeekAdapterConfig struct {
	BaseUrl         string
	AllowedDomains  []string
	SearchCriterias []SearchCriteria
	PageFlag        string
	Pages           int
	SiteSelectors   Selectors
	SearchApiUrl    string
	ApiParameters   []ApiParameters
}
