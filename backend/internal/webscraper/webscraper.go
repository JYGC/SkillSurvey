package webscraper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JYGC/SkillSurvey/internal/entities"
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
	createNewInboundJobPost func(doc *colly.HTMLElement) entities.InboundJobPost,
) (jobPosts []entities.InboundJobPost, err error) {
	jobPostLinks, jobPostLinkErr := w.getJobPostLinks(
		jobPostLinkSelector,
		numberOfPages,
		pageFlag,
		searchUrls,
	)
	if jobPostLinkErr != nil {
		err = fmt.Errorf("jobPostLinkErr: %v", jobPostLinkErr)
	}
	if len(jobPostLinks) > 0 {
		// exception.ReportError(map[string]any{
		// 	"Message":  "No job post links found. Possible site selector error",
		// 	"SiteName": w.siteName,
		// })
		var jobPostErr error
		jobPosts, jobPostErr = w.getJobPosts(
			jobPostLinks,
			createNewInboundJobPost,
		)
		if jobPostErr != nil {
			err = fmt.Errorf("%v\njobPostErr: %v", err, jobPostErr)
		}
	} else {
		err = fmt.Errorf(
			"%v\nNo job post links found. Possible site selector error. SiteName: %s",
			err,
			w.siteName,
		)
	}

	return jobPosts, err
}

func (w *WebScraper) getJobPostLinks(
	jobPostLinkSelector string,
	numberOfPages int,
	pageFlag string,
	searchUrls []string,
) (jobPostLinks []string, err error) {
	w.scraperEngine.OnHTML(
		jobPostLinkSelector,
		func(e *colly.HTMLElement) {
			link := e.Attr("href")
			jobPostLinks = append(jobPostLinks, link)
		},
	)
	var pageErrors []error
	for searchPage := 1; searchPage <= numberOfPages; searchPage++ {
		for _, searchUrl := range searchUrls {
			fullUrl := strings.ReplaceAll(
				searchUrl,
				pageFlag,
				strconv.Itoa(searchPage),
			)
			pageError := w.scraperEngine.Visit(fullUrl)
			if pageError != nil {
				pageErrors = append(pageErrors, pageError)
			}
		}
	}
	if len(pageErrors) > 0 {
		err = fmt.Errorf("Page errors: %v", pageErrors)
	}
	return jobPostLinks, err
}

func (w WebScraper) getJobPosts(
	jobPostLinksSlice []string,
	createNewInboundJobPost func(doc *colly.HTMLElement) entities.InboundJobPost,
) (
	newInboundJobPostSlice []entities.InboundJobPost,
	err error,
) {
	w.scraperEngine.OnHTML("html", func(doc *colly.HTMLElement) {
		var newInboundJobPost entities.InboundJobPost
		defer (func() {
			extraVariables := fmt.Sprintf(
				"(Url: %s)",
				doc.Request.URL.String(),
			)
			errInterface := recover()
			switch x := errInterface.(type) {
			case string:
				err = fmt.Errorf("%v scraperErr: %s %s", err, x, extraVariables)
			default:
				err = fmt.Errorf("%v scraperErr: %v %s", err, x, extraVariables)
			}
		})()
		newInboundJobPost = createNewInboundJobPost(doc)
		newInboundJobPostSlice = append(newInboundJobPostSlice, newInboundJobPost)
	})
	var jobPostLinkErrs []error
	for _, jobPostLink := range jobPostLinksSlice {
		link := w.baseUrl + jobPostLink
		jobPostLinkErr := w.scraperEngine.Visit(link)
		if jobPostLinkErr != nil {
			jobPostLinkErrs = append(jobPostLinkErrs, jobPostLinkErr)
		}
	}
	if len(jobPostLinkErrs) > 0 {
		err = fmt.Errorf("%v\njobPostLinkErrs: %v", err, jobPostLinkErrs)
	}
	return newInboundJobPostSlice, err
}
