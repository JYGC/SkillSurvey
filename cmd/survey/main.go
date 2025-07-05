package main

import (
	"fmt"

	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"

	//"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/siteadapters"
)

func main() {
	variableRef := make(map[string]any)
	//defer exception.ReportErrorIfPanic(map[string]interface{}{"Variables": variableRef})
	// get jobposts from websites
	var newInboundJobPosts []entities.InboundJobPost
	for _, webScraperSite := range []siteadapters.ISiteAdapter{
		*siteadapters.NewJoraAdapter(),
		*siteadapters.NewSeekAdapter(),
	} {
		newInboundJobPosts = append(newInboundJobPosts, webScraperSite.RunSurvey()...)
	}
	existingSites, err := database.DbAdapter.Site.GetAll()
	if err != nil {
		panic(err)
	}
	// insert jobposts to database
	siteMap := entities.MakeSiteMap(existingSites)
	newJobPosts := entities.CreateJobPosts(siteMap, newInboundJobPosts)
	if err := database.DbAdapter.JobPost.BulkUpdateOrInsert(newJobPosts); err != nil {
		variableRef["existingSites"] = existingSites
		variableRef["siteMap"] = siteMap
		variableRef["newInboundJobPostSlice"] = newInboundJobPosts
		variableRef["newJobPostSlice"] = newJobPosts
		panic(err)
	}
}
