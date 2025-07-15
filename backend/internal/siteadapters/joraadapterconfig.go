package siteadapters

type JoraSearchCriteria struct {
	Name string
	Url  string
}

type JoraSelectors struct {
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

type JoraAdapterConfig struct {
	BaseUrl                string
	AllowedDomains         []string
	SearchCriterias        []JoraSearchCriteria
	PageFlag               string
	Pages                  int
	SiteSelectors          JoraSelectors
	SecondsBetweenJobPosts int
}
