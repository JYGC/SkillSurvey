package database

import (
	"github.com/JYGC/SkillSurvey/internal/entities"
	"gorm.io/gorm"
)

type JobPostTableCall struct {
	DbTableCallBase
}

func NewJobPostTableCall(db *gorm.DB) (tableCall *JobPostTableCall) {
	tableCall = new(JobPostTableCall)
	tableCall.db = db
	tableCall.MigrateTable(&entities.JobPost{})
	return tableCall
}

func (j JobPostTableCall) BulkUpdateOrInsert(inboundJobPosts []entities.JobPost) {
	inboundJobPostWithSiteNumbers := make(map[string]entities.JobPost)
	var inboundJobPostSiteNumbers []string
	for _, jobPost := range inboundJobPosts {
		inboundJobPostSiteNumbers = append(inboundJobPostSiteNumbers, jobPost.JobSiteNumber)
		inboundJobPostWithSiteNumbers[jobPost.JobSiteNumber] = jobPost
	}
	var existingJobPosts []entities.JobPost
	j.db.Where("job_site_number IN ?", inboundJobPostSiteNumbers).Find(&existingJobPosts)
	//Update existing and save new

	j.db.Create(inboundJobPosts)
	// Get existing JobPost SiteNumbers
	//var query interface{}
	//json.Unmarshal([]byte(`{"eq":}`), &query)
}
