package config

type SiteAdapterConfig struct {
	ConfigBase
	BaseUrl         string
	AllowedDomains  []string
	SearchCriterias []SearchCriteria
	PageFlag        string
	Pages           int
}

type SearchCriteria struct {
	Name string
	Url  string
}
