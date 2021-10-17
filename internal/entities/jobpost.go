package entities

import (
	"time"
)

type JobPost struct {
	EntityBase
	Site          Site `gorm:"foreignKey:ID"`
	JobSiteNumber string
	Title         string
	Body          string
	PostedDate    time.Time
	City          string
	Country       string
	Suburb        string
}

type InboundJobPost struct {
	JobPost
	SiteName string
}
