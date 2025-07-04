package main

import (
	"github.com/JYGC/SkillSurvey/internal/database"
	"github.com/JYGC/SkillSurvey/internal/entities"

	//"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/siteadapters"
)

func main() {
	variableRef := make(map[string]interface{})
	//defer exception.ReportErrorIfPanic(map[string]interface{}{"Variables": variableRef})
	// get jobposts from websites
	var newInboundJobPostSlice []entities.InboundJobPost
	for _, webScraperSite := range []siteadapters.ISiteAdapter{
		//*siteadapters.NewJoraAdapter(),
		*siteadapters.NewSeekAdapter(),
	} {
		newInboundJobPostSlice = append(newInboundJobPostSlice, webScraperSite.RunSurvey()...)
	}
	existingSites, err := database.DbAdapter.Site.GetAll()
	if err != nil {
		panic(err)
	}
	// insert jobposts to database
	siteMap := entities.MakeSiteMap(existingSites)
	newJobPosts := entities.CreateJobPosts(siteMap, newInboundJobPostSlice)
	if err := database.DbAdapter.JobPost.BulkUpdateOrInsert(newJobPosts); err != nil {
		variableRef["existingSites"] = existingSites
		variableRef["siteMap"] = siteMap
		variableRef["newInboundJobPostSlice"] = newInboundJobPostSlice
		variableRef["newJobPostSlice"] = newJobPosts
		panic(err)
	}
}
