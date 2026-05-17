package siteadapters

import "time"

// InboundJobPost is the site-adapter output type, free of GORM/entity dependencies.
type InboundJobPost struct {
	Title         string
	Body          string
	JobSiteNumber string
	PostedDate    time.Time
	City          string
	Country       string
	Suburb        string
	SiteName      string
}

// ISiteAdapter is the common interface for all job-board adapters.
type ISiteAdapter interface {
	RunSurvey() ([]InboundJobPost, error)
}
