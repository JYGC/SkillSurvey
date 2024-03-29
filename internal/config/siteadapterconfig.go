package config

type SiteAdapterConfig struct {
	ConfigBase
	BaseUrl         string
	AllowedDomains  []string
	SearchCriterias []SearchCriteria
	PageFlag        string
	Pages           int
	SiteSelectors   Selectors
}

type SearchCriteria struct {
	Name string
	Url  string
}

type Selectors struct {
	SiteName           string
	JobPostLink        string
	TitleSelector      string
	BodySelector       string
	PostedDateSelector string
	CitySelector       string
	Country            string
	SuburbSelector     string
	TitleType          string
	BodyType           string
	PostedDateType     string
	CityType           string
	SuburbType         string
}
