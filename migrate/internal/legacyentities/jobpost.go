package legacyentities

import "time"

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
