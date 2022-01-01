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
		defer exception.ReportErrorIfPanic(map[string]string{
			"Url": doc.Request.URL.String(),
		})
		newInboundJobPost := new(entities.InboundJobPost)
		newInboundJobPost.SiteName = w.siteAdapter.GetConfigSettings().SiteSelectors.SiteName
		newInboundJobPost.JobSiteNumber = w.siteAdapter.GetJobSiteNumber(doc)
		newInboundJobPost.Title = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.TitleSelector)
		newInboundJobPost.Body = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.BodySelector)
		newInboundJobPost.PostedDate = w.siteAdapter.GetPostedDate(doc)
		newInboundJobPost.City = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.CitySelector)
		newInboundJobPost.Country = w.siteAdapter.GetConfigSettings().SiteSelectors.Country
		newInboundJobPost.Suburb = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.SuburbSelector)
		w.reportIfEmptyStr(doc, "JobSiteNumber", newInboundJobPost.JobSiteNumber)
		w.reportIfEmptyStr(doc, "Title", newInboundJobPost.Title)
		w.reportIfEmptyStr(doc, "Body", newInboundJobPost.Body)
		w.reportIfEmptyStr(doc, "City", newInboundJobPost.City)
		w.reportIfEmptyStr(doc, "Country", newInboundJobPost.Country)
		w.reportIfEmptyStr(doc, "Suburb", newInboundJobPost.Suburb)
		newInboundJobPostSlice = append(newInboundJobPostSlice, *newInboundJobPost)
	})
	for _, jobPostLink := range jobPostLinksSlice {
		link := w.siteAdapter.GetConfigSettings().BaseUrl + jobPostLink
		w.scraperEngine.Visit(link)
	}
	return newInboundJobPostSlice
}

func (w WebScraper) reportIfEmptyStr(doc *colly.HTMLElement, fieldName string, fieldValue string) {
	if len(strings.TrimSpace(fieldValue)) == 0 {
		exception.ReportError(map[string]string{
			"Url":           doc.Request.URL.String(),
			"Missing Field": fieldName,
		})
	}
}
