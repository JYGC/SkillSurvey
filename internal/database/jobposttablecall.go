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
		jobPost.Title = jobPostMap[jobPost.JobSiteNumber].Title
		jobPost.Body = jobPostMap[jobPost.JobSiteNumber].Body
		jobPost.PostedDate = jobPostMap[jobPost.JobSiteNumber].PostedDate
		jobPost.City = jobPostMap[jobPost.JobSiteNumber].City
		jobPost.Country = jobPostMap[jobPost.JobSiteNumber].Country
		jobPost.Suburb = jobPostMap[jobPost.JobSiteNumber].Suburb
		delete(jobPostMap, jobPost.JobSiteNumber)
		j.db.Save(&jobPost)
	}
	newJobPosts := make([]entities.JobPost, 0, len(jobPostMap))
	for _, value := range jobPostMap {
		newJobPosts = append(newJobPosts, value)
	}
	j.db.Create(newJobPosts)
}
