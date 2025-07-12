package main

import (
	"fmt"

	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"

	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/siteadapters"
)

func main() {
	defer exception.ReportErrorIfPanic(map[string]any{})
	// get jobposts from websites
	var err error
	var newInboundJobPosts []entities.InboundJobPost
	for _, webScraperSite := range []siteadapters.ISiteAdapter{
		*siteadapters.NewJoraAdapter(),
		*siteadapters.NewSeekAdapter(),
	} {
		newInboundJobPost, newInboundJobPostErr := webScraperSite.RunSurvey()
		if newInboundJobPostErr != nil {
			err = fmt.Errorf("%v\nnewInboundJobPostErr: %v", err, newInboundJobPostErr)
		}
		newInboundJobPosts = append(newInboundJobPosts, newInboundJobPost...)
	}
	existingSites, getAllErr := database.DbAdapter.Site.GetAll()
	if getAllErr != nil {
		err = fmt.Errorf("%v\ngetAllErr: %v", err, getAllErr)
	}
	// insert jobposts to database
	siteMap := entities.MakeSiteMap(existingSites)
	newJobPosts := entities.CreateJobPosts(siteMap, newInboundJobPosts)
	if updateOrInsertErr := database.DbAdapter.JobPost.BulkUpdateOrInsert(newJobPosts); updateOrInsertErr != nil {
		err = fmt.Errorf("%v\nupdateOrInsertErr: %v", err, updateOrInsertErr)
	}
	if err != nil {
		panic(err)
	}
}
