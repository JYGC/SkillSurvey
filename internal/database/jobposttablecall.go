package database

import (
	"strings"

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

func (j JobPostTableCall) BulkUpdateOrInsert(jobPosts []entities.JobPost) {
	jobPostMap := make(map[string]entities.JobPost)
	var jobPostSiteNumbers []string
	for _, jobPost := range jobPosts {
		jobPostSiteNumbers = append(jobPostSiteNumbers, jobPost.JobSiteNumber)
		jobPostMap[jobPost.JobSiteNumber] = jobPost
	}
	var existingJobPosts []entities.JobPost
	j.db.Where("job_site_number IN ?", jobPostSiteNumbers).Find(&existingJobPosts)
	for _, jobPost := range existingJobPosts {
		jobPostMapElement := jobPostMap[jobPost.JobSiteNumber]
		if len(strings.TrimSpace(jobPostMapElement.Title)) != 0 {
			jobPost.Title = jobPostMapElement.Title
		}
		if len(strings.TrimSpace(jobPostMapElement.Body)) != 0 {
			jobPost.Body = jobPostMapElement.Body
		}
		jobPost.PostedDate = jobPostMapElement.PostedDate
		if len(strings.TrimSpace(jobPostMapElement.City)) != 0 {
			jobPost.City = jobPostMapElement.City
		}
		if len(strings.TrimSpace(jobPostMapElement.Country)) != 0 {
			jobPost.Country = jobPostMapElement.Country
		}
		if len(strings.TrimSpace(jobPostMapElement.Suburb)) != 0 {
			jobPost.Suburb = jobPostMapElement.Suburb
		}
		delete(jobPostMap, jobPost.JobSiteNumber)
		j.db.Save(&jobPost)
	}
	var newJobPosts []entities.JobPost
	for _, value := range jobPostMap {
		newJobPosts = append(newJobPosts, value)
	}
	j.db.Create(newJobPosts)
}
