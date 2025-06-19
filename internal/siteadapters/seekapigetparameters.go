package siteadapters

type SeekGetApiParameters struct {
	Page                  int
	NewSince              string
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
