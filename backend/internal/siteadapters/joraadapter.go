package siteadapters

import (
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/environment"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/webscraper"
	"github.com/gocolly/colly/v2"
)

const joraConfigFileName = "jora.json"

type JoraAdapter struct {
	ConfigSettings JoraAdapterConfig
	webScraper     *webscraper.WebScraper
}

func NewJoraAdapter() *JoraAdapter {
	jora := new(JoraAdapter)
	config.JsonToConfig(
		&jora.ConfigSettings,
		environment.AttachToExecutableDir(joraConfigFileName),
	)
	jora.webScraper = webscraper.NewWebScraper(
		jora.ConfigSettings.BaseUrl,
		jora.ConfigSettings.SiteSelectors.SiteName,
		jora.ConfigSettings.AllowedDomains,
	)
	return jora
}

func (j JoraAdapter) RunSurvey() (
	[]entities.InboundJobPost,
	error,
) {
	searchUrls := []string{}
	for _, searchCriteria := range j.ConfigSettings.SearchCriterias {
		searchUrls = append(searchUrls, searchCriteria.Url)
	}
	return j.webScraper.Scrape(
		j.ConfigSettings.SiteSelectors.JobPostLink,
		j.ConfigSettings.Pages,
		j.ConfigSettings.PageFlag,
		j.ConfigSettings.SecondsBetweenJobPosts,
		searchUrls,
		func(doc *colly.HTMLElement) entities.InboundJobPost {
			newInboundJobPost := entities.InboundJobPost{}
			newInboundJobPost.SiteName = j.ConfigSettings.SiteSelectors.SiteName
			newInboundJobPost.JobSiteNumber = j.getJobSiteNumber(doc)
			newInboundJobPost.PostedDate = j.getPostedDate(doc)
			newInboundJobPost.Title = doc.ChildText(j.ConfigSettings.SiteSelectors.TitleSelector)
			newInboundJobPost.Body = doc.ChildText(j.ConfigSettings.SiteSelectors.BodySelector)
			newInboundJobPost.City = doc.ChildText(j.ConfigSettings.SiteSelectors.CitySelector)
			newInboundJobPost.Country = j.ConfigSettings.SiteSelectors.Country
			newInboundJobPost.Suburb = doc.ChildText(j.ConfigSettings.SiteSelectors.SuburbSelector)
			return newInboundJobPost
		},
	)
}

// Advertisement's post date can calculated by subtracting how old the advert is in days from
// the current date.
func (j JoraAdapter) getPostedDate(doc *colly.HTMLElement) time.Time {
	variableRef := make(map[string]any)
	defer exception.ReportErrorIfPanic(map[string]any{
		"Url":       doc.Request.URL.String(),
		"Variables": variableRef,
	})
	ageString := doc.ChildText(j.ConfigSettings.SiteSelectors.PostedDateSelector) // .date contains either "N days" or "today"
	variableRef["ageString"] = ageString
	daysOld := 0
	daysIndex := strings.Index(ageString, "days")
	variableRef["daysIndex"] = daysIndex
	currentDate := time.Now()
	postedDate := time.Now()
	// TODO: support hours and minutes ago
	if daysIndex != -1 {
		// If "today", the advert is 0 days old, leave daysOld as 0
		var err error
		daysOld, err = strconv.Atoi(ageString[0 : daysIndex-1])
		if err != nil {
			panic(err)
		}
	}
	// set to next month if argument is more then number of days in
	// current month
	postedDate = currentDate.AddDate(0, 0, -daysOld)
	return postedDate
}

func (j JoraAdapter) getJobSiteNumber(doc *colly.HTMLElement) string {
	url := doc.Request.URL.String()
	return url[strings.Index(url, "/job/")+5 : strings.LastIndex(url, "?")]
}
