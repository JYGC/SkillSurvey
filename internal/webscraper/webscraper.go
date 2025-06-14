package webscraper

import (
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/gocolly/colly/v2"
)

const userAgent = "node-spider"

type WebScraper struct {
	configSettings   config.SiteAdapterConfig
	scraperEngine    colly.Collector
	getJobSiteNumber func(doc *colly.HTMLElement) string
	getPostedDate    func(doc *colly.HTMLElement) time.Time
}

func NewWebScraper(
	configSettings config.SiteAdapterConfig,
	getJobSiteNumber func(doc *colly.HTMLElement) string,
	getPostedDate func(doc *colly.HTMLElement) time.Time,
) *WebScraper {
	newWebScraper := new(WebScraper)
	newWebScraper.configSettings = configSettings
	newWebScraper.scraperEngine = *colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains(
			newWebScraper.configSettings.AllowedDomains...,
		),
	)
	newWebScraper.getJobSiteNumber = getJobSiteNumber
	newWebScraper.getPostedDate = getPostedDate
	return newWebScraper
}

func (w WebScraper) Scrape() []entities.InboundJobPost {
	return w.getJobPosts(w.getJobPostLinks())
}

func (w *WebScraper) getJobPostLinks() (jobPostLinks []string) {
	w.scraperEngine.OnHTML(
		w.configSettings.SiteSelectors.JobPostLink,
		func(e *colly.HTMLElement) {
			link := e.Attr("href")
			jobPostLinks = append(jobPostLinks, link)
		},
	)
	for _, searchCriteria := range w.configSettings.SearchCriterias {
		for searchPage := 1; searchPage <= w.configSettings.Pages; searchPage++ {
			fullUrl := strings.ReplaceAll(
				searchCriteria.Url,
				w.configSettings.PageFlag,
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
		newInboundJobPost.SiteName = w.configSettings.SiteSelectors.SiteName
		newInboundJobPost.JobSiteNumber = w.getJobSiteNumber(doc)
		newInboundJobPost.Title = doc.ChildText(w.configSettings.SiteSelectors.TitleSelector)
		newInboundJobPost.Body = doc.ChildText(w.configSettings.SiteSelectors.BodySelector)
		newInboundJobPost.PostedDate = w.getPostedDate(doc)
		newInboundJobPost.City = doc.ChildText(w.configSettings.SiteSelectors.CitySelector)
		newInboundJobPost.Country = w.configSettings.SiteSelectors.Country
		newInboundJobPost.Suburb = doc.ChildText(w.configSettings.SiteSelectors.SuburbSelector)
		newInboundJobPostSlice = append(newInboundJobPostSlice, *newInboundJobPost)
	})
	if len(jobPostLinksSlice) == 0 {
		exception.ReportError(map[string]interface{}{
			"Message":  "No job post links found. Possible site selector error",
			"SiteName": w.configSettings.SiteSelectors.SiteName,
		})
	}
	for _, jobPostLink := range jobPostLinksSlice {
		link := w.configSettings.BaseUrl + jobPostLink
		w.scraperEngine.Visit(link)
	}
	return newInboundJobPostSlice
}
