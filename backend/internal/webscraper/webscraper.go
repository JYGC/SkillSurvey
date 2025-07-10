package webscraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/exception"
	"github.com/gocolly/colly/v2"
)

const userAgent = "node-spider"

type WebScraper struct {
	baseUrl       string
	siteName      string
	scraperEngine colly.Collector
}

func NewWebScraper(
	baseUrl string,
	siteName string,
	allowedDomains []string,
) *WebScraper {
	newWebScraper := new(WebScraper)
	newWebScraper.baseUrl = baseUrl
	newWebScraper.siteName = siteName
	newWebScraper.scraperEngine = *colly.NewCollector(
		colly.UserAgent(userAgent),
		colly.AllowedDomains(allowedDomains...),
	)
	return newWebScraper
}

func (w WebScraper) Scrape(
	jobPostLinkSelector string,
	numberOfPages int,
	pageFlag string,
	searchUrls []string,
	createnewInboundJobPost func(doc *colly.HTMLElement) entities.InboundJobPost,
) []entities.InboundJobPost {
	return w.getJobPosts(
		w.getJobPostLinks(
			jobPostLinkSelector,
			numberOfPages,
			pageFlag,
			searchUrls,
		),
		createnewInboundJobPost,
	)
}

func (w *WebScraper) getJobPostLinks(
	jobPostLinkSelector string,
	numberOfPages int,
	pageFlag string,
	searchUrls []string,
) (jobPostLinks []string) {
	w.scraperEngine.OnHTML(
		jobPostLinkSelector,
		func(e *colly.HTMLElement) {
			link := e.Attr("href")
			jobPostLinks = append(jobPostLinks, link)
		},
	)
	for searchPage := 1; searchPage <= numberOfPages; searchPage++ {
		for _, searchUrl := range searchUrls {
			fullUrl := strings.ReplaceAll(
				searchUrl,
				pageFlag,
				strconv.Itoa(searchPage),
			)
			w.scraperEngine.Visit(fullUrl)
		}
	}
	return jobPostLinks
}

func (w WebScraper) getJobPosts(
	jobPostLinksSlice []string,
	createnewInboundJobPost func(doc *colly.HTMLElement) entities.InboundJobPost,
) (newInboundJobPostSlice []entities.InboundJobPost) {
	w.scraperEngine.OnHTML("html", func(doc *colly.HTMLElement) {
		var newInboundJobPost entities.InboundJobPost
		defer exception.ReportErrorIfPanic(map[string]any{
			"Url":               doc.Request.URL.String(),
			"newInboundJobPost": newInboundJobPost,
		})
		newInboundJobPost = createnewInboundJobPost(doc)
		newInboundJobPostSlice = append(newInboundJobPostSlice, newInboundJobPost)
	})
	if len(jobPostLinksSlice) == 0 {
		exception.ReportError(map[string]any{
			"Message":  "No job post links found. Possible site selector error",
			"SiteName": w.siteName,
		})
	}
	for _, jobPostLink := range jobPostLinksSlice {
		link := w.baseUrl + jobPostLink
		w.scraperEngine.Visit(link)
	}
	return newInboundJobPostSlice
}
