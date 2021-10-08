package database

import (
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/JYGC/SkillSurvey/pkg/entities"
)

type JobPostCollection struct {
	CollectionBase
}

func NewJobPostCollection(database *db.DB) *JobPostCollection {
	collection := new(JobPostCollection)
	collection.database = database
	collection.CollectionName = "JobPosts"
	collection.collection = collection.database.ForceUse(collection.CollectionName)
	return collection
}

func (j JobPostCollection) AddMany(inboundJobPosts []entities.JobPost) {
	inboundJobPostMap := make(map[string]entities.JobPost)
	var inboundJobPostSiteNumbers []string
	for _, jobPost := range inboundJobPosts {
		inboundJobPostSiteNumbers = append(inboundJobPostSiteNumbers, jobPost.JobSiteNumber)
		inboundJobPostMap[jobPost.JobSiteNumber] = jobPost
	}
	// Get existing JobPost SiteNumbers
	//var query interface{}
	//json.Unmarshal([]byte(`{"eq":}`), &query)
}
