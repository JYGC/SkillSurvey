package models

import (
	"time"

	"github.com/pocketbase/pocketbase/core"
)

var _ core.RecordProxy = (*JobPost)(nil)

type JobPost struct {
	core.BaseRecordProxy
	// EntityBase
	// SiteID        uint
	// Site          Site `gorm:"foreignKey:SiteID"`
	// JobSiteNumber string
	// Title         string
	// Body          string
	// PostedDate    time.Time
	// City          string
	// Country       string
	// Suburb        string
	// CreateDate    time.Time
}

func (j *JobPost) JobSiteNumber() string {
	return j.GetString("job_site_number")
}

func (j *JobPost) SetJobSiteNumber(jobSiteNumber string) {
	j.Set("job_site_number", jobSiteNumber)
}

func (j *JobPost) Content() string {
	return j.GetString("content")
}

func (j *JobPost) SetContent(content string) {
	j.Set("content", content)
}

func (j *JobPost) Body() string {
	return j.GetString("title")
}

func (j *JobPost) SetBody(title string) {
	j.Set("title", title)
}
