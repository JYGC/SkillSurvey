package entities

import "time"

type JobPost struct {
	SiteName      string
	JobSiteNumber string
	Title         string
	Body          string
	PostedDate    time.Time
	City          string
	Country       string
	Suburb        string
}
