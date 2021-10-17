package webscraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/siteadapters"
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
	w.scraperEngine.OnHTML(w.siteAdapter.GetConfigSettings().SiteSelectors.JobPostLink, func(e *colly.HTMLElement) {
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
	var newInboundJobPostSlice []entities.InboundJobPost
	w.scraperEngine.OnHTML("html", func(doc *colly.HTMLElement) {
		newInboundJobPost := new(entities.InboundJobPost)
		newInboundJobPost.SiteName = w.siteAdapter.GetConfigSettings().SiteSelectors.SiteName
		newInboundJobPost.JobSiteNumber = w.siteAdapter.GetJobSiteNumber(doc)
		newInboundJobPost.Body = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.BodySelector)
		newInboundJobPost.PostedDate = w.siteAdapter.GetPostedDate(doc)
		newInboundJobPost.City = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.CitySelector)
		newInboundJobPost.Country = w.siteAdapter.GetConfigSettings().SiteSelectors.Country
		newInboundJobPost.Suburb = doc.ChildText(w.siteAdapter.GetConfigSettings().SiteSelectors.SuburbSelector)
		newInboundJobPostSlice = append(newInboundJobPostSlice, *newInboundJobPost)
	})
	for _, jobPostLink := range w.jobPostLinks {
		link := w.siteAdapter.GetConfigSettings().BaseUrl + jobPostLink
		w.scraperEngine.Visit(link)
	}
	fmt.Println(len(w.jobPostLinks))
	fmt.Println(newInboundJobPostSlice)
}
