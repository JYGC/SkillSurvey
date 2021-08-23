package webscraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/pkg/entities"
	"github.com/JYGC/SkillSurvey/pkg/siteadapters"
	"github.com/gocolly/colly/v2"
)

type WebScraper struct {
	siteAdapter   siteadapters.ISiteAdapter
	scraperEngine colly.Collector
	jobPostLinks  []string
}

func NewWebScraper(siteadapter siteadapters.ISiteAdapter, userAgent string) *WebScraper {
	newWebScraper := new(WebScraper)
	newWebScraper.siteAdapter = siteadapter
	newWebScraper.scraperEngine = *colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains(newWebScraper.siteAdapter.GetConfigSettings().AllowedDomains...),
	)
	return newWebScraper
}

func (w WebScraper) Start() {
	w.getJobPostLinks()
	w.getJobPosts()
}

func (w *WebScraper) getJobPostLinks() {
	w.scraperEngine.OnHTML(w.siteAdapter.GetJobPostLink(), func(e *colly.HTMLElement) {
		link := e.Attr("href")
		w.jobPostLinks = append(w.jobPostLinks, link)
	})
	for _, searchCriteria := range w.siteAdapter.GetConfigSettings().SearchCriterias {
		for searchPage := 1; searchPage <= w.siteAdapter.GetConfigSettings().Pages; searchPage++ {
			fullUrl := strings.ReplaceAll(
				searchCriteria.Url,
				w.siteAdapter.GetConfigSettings().PageFlag,
				strconv.Itoa(searchPage),
			)
			w.scraperEngine.Visit(fullUrl)
		}
	}
}

func (w WebScraper) getJobPosts() {
	var newJobPostSlice []entities.JobPost
	w.scraperEngine.OnHTML("html", func(doc *colly.HTMLElement) {
		newJobPost := new(entities.JobPost)
		newJobPost.SiteName = w.siteAdapter.GetSiteName()
		newJobPost.JobSiteNumber = w.siteAdapter.GetJobSiteNumber(doc)
		newJobPost.Body = doc.ChildText(w.siteAdapter.GetBodySelector())
		newJobPost.PostedDate = w.siteAdapter.GetPostedDate(doc)
		newJobPost.City = doc.ChildText(w.siteAdapter.GetCitySelector())
		newJobPost.Country = w.siteAdapter.GetCountry()
		newJobPost.Suburb = doc.ChildText(w.siteAdapter.GetSuburbSelector())
		newJobPostSlice = append(newJobPostSlice, *newJobPost)
	})
	for _, jobPostLink := range w.jobPostLinks {
		link := w.siteAdapter.GetConfigSettings().BaseUrl + jobPostLink
		//fmt.Println(link)
		w.scraperEngine.Visit(link)
	}
	fmt.Println(newJobPostSlice)
}
