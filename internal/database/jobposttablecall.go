package database

import (
	"strings"
	"time"

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
	// update existing jobposts
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
	// insert new jobposts
	var newJobPosts []entities.JobPost
	jobPostMapIndex := 0
	for _, value := range jobPostMap {
		chunkSize := 1000
		value.CreateDate = time.Now()
		newJobPosts = append(newJobPosts, value)
		if len(newJobPosts) >= chunkSize || jobPostMapIndex >= len(jobPostMap)-1 {
			j.db.Create(newJobPosts)
			newJobPosts = nil
		}
		jobPostMapIndex++
	}
}

func (j JobPostTableCall) GetMonthlyCountBySkill(
	skillName string,
	skillNameAliases []entities.SkillNameAlias,
) (
	result []entities.MonthlyCountReport,
	err error,
) {
	bodyLike := "job_posts.body LIKE ?"
	query := j.db.Table("job_posts").Select(
		"strftime('%Y-%m', job_posts.posted_date) `[YearMonth]`, COUNT(job_posts.id) [Count]",
	).Group("[YearMonth]").Where(
		bodyLike, "%"+skillName+" %",
	).Or(
		bodyLike, "%"+skillName+",%",
	).Or(
		bodyLike, "%"+skillName+".%",
	).Or(
		bodyLike, "%"+skillName+"\n%",
	)
	for _, skillNameAlias := range skillNameAliases {
		query = query.Or(
			bodyLike, "%"+skillNameAlias.Alias+" %",
		).Or(
			bodyLike, "%"+skillNameAlias.Alias+",%",
		).Or(
			bodyLike, "%"+skillNameAlias.Alias+".%",
		).Or(
			bodyLike, "%"+skillNameAlias.Alias+"\n%",
		)
	}
	err = query.Scan(&result).Error
	return result, err
}
