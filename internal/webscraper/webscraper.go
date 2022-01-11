package webscraper

import (
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/JYGC/SkillSurvey/internal/siteadapters"
	"github.com/gocolly/colly/v2"
)

type WebScraper struct {
	siteAdapter   siteadapters.ISiteAdapter
	scraperEngine colly.Collector
}

func NewWebScraper(siteadapter siteadapters.ISiteAdapter, userAgent string) *WebScraper {
	newWebScraper := new(WebScraper)
	newWebScraper.siteAdapter = siteadapter
	newWebScraper.scraperEngine = *colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains(
			newWebScraper.siteAdapter.GetConfigSettings().AllowedDomains...,
		),
	)
	return newWebScraper
}

func (w WebScraper) Scrape() []entities.InboundJobPost {
	return w.getJobPosts(w.getJobPostLinks())
}

func (w *WebScraper) getJobPostLinks() (jobPostLinks []string) {
	w.scraperEngine.OnHTML(
		w.siteAdapter.GetConfigSettings().SiteSelectors.JobPostLink,
		func(e *colly.HTMLElement) {
			link := e.Attr("href")
			jobPostLinks = append(jobPostLinks, link)
		},
	)
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
	return jobPostLinks
}

func (w WebScraper) getJobPosts(jobPostLinksSlice []string) (newInboundJobPostSlice []entities.InboundJobPost) {
	w.scraperEngine.OnHTML("html", func(doc *colly.HTMLElement) {
		newInboundJobPost := new(entities.InboundJobPost)
		defer exception.ReportErrorIfPanic(map[string]interface{}{
			"Url":               doc.Request.URL.String(),
			"newInboundJobPost": newInboundJobPost,
		})
		newInboundJobPost.SiteName = w.siteAdapter.GetConfigSettings().SiteSelectors.SiteName
		newInboundJobPost.JobSiteNumber = w.siteAdapter.GetJobSiteNumber(doc)
		newInboundJobPost.Title = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.TitleSelector)
		newInboundJobPost.Body = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.BodySelector)
		newInboundJobPost.PostedDate = w.siteAdapter.GetPostedDate(doc)
		newInboundJobPost.City = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.CitySelector)
		newInboundJobPost.Country = w.siteAdapter.GetConfigSettings().SiteSelectors.Country
		newInboundJobPost.Suburb = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.SuburbSelector)
		newInboundJobPostSlice = append(newInboundJobPostSlice, *newInboundJobPost)
	})
	if len(jobPostLinksSlice) == 0 {
		exception.ReportError(map[string]interface{}{
			"Message": "No job post links found. Possible site selector error",
		})
	}
	for _, jobPostLink := range jobPostLinksSlice {
		link := w.siteAdapter.GetConfigSettings().BaseUrl + jobPostLink
		w.scraperEngine.Visit(link)
	}
	return newInboundJobPostSlice
}
