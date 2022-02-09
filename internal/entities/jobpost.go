package entities

import (
	"time"
)

type JobPost struct {
	EntityBase
	SiteID        uint
	Site          Site `gorm:"foreignKey:SiteID"`
	JobSiteNumber string
	Title         string
	Body          string
	PostedDate    time.Time
	City          string
	Country       string
	Suburb        string
	CreateDate    time.Time
}

type InboundJobPost struct {
	JobPost
	SiteName string
}

func CreateJobPosts(siteMap map[string]Site, newInboundJobPostSlice []InboundJobPost) (newJobPostSlice []JobPost) {
	for _, inboundJobPost := range newInboundJobPostSlice {
		newJobPostSlice = append(newJobPostSlice, JobPost{
			Site:          siteMap[inboundJobPost.SiteName],
			JobSiteNumber: inboundJobPost.JobSiteNumber,
			Title:         inboundJobPost.Title,
			Body:          inboundJobPost.Body,
			PostedDate:    inboundJobPost.PostedDate,
			City:          inboundJobPost.City,
			Country:       inboundJobPost.Country,
			Suburb:        inboundJobPost.Suburb,
		})
	}
	return newJobPostSlice
}
