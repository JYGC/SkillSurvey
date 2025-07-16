package webscraper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	secondsBetweenJobPosts int,
	searchUrls []string,
	createNewInboundJobPost func(doc *colly.HTMLElement) entities.InboundJobPost,
) (jobPosts []entities.InboundJobPost, err error) {
	jobPostLinks, jobPostLinksErr := w.getJobPostLinks(
		jobPostLinkSelector,
		numberOfPages,
		pageFlag,
		searchUrls,
	)
	if jobPostLinksErr != nil {
		err = fmt.Errorf("jobPostLinksErr: %v", jobPostLinksErr)
	}
	if len(jobPostLinks) > 0 {
		var jobPostErr error
		jobPosts, jobPostErr = w.getJobPosts(
			secondsBetweenJobPosts,
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
		exception.ErrorLogger.Println(err)
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
				exception.LogErrorWithLabel("pageError", pageError)
			}
		}
	}
	if len(pageErrors) > 0 {
		err = fmt.Errorf("Page errors: %v", pageErrors)
	}
	return jobPostLinks, err
}

func (w WebScraper) getJobPosts(
	secondsBetweenJobPosts int,
	jobPostLinksSlice []string,
	createNewInboundJobPost func(doc *colly.HTMLElement) entities.InboundJobPost,
) (
	newInboundJobPosts []entities.InboundJobPost,
	err error,
) {
	w.scraperEngine.OnHTML("html", func(doc *colly.HTMLElement) {
		var newInboundJobPost entities.InboundJobPost
		defer (func() {
			extraVariables := fmt.Sprintf(
				"(Url: %s)",
				doc.Request.URL.String(),
			)
			if errInterface := recover(); errInterface != nil {
				switch x := errInterface.(type) {
				case string:
					err = fmt.Errorf("%v scraperErr: %s %s", err, x, extraVariables)
				default:
					err = fmt.Errorf("%v scraperErr: %v %s", err, x, extraVariables)
				}
			}
		})()
		newInboundJobPost = createNewInboundJobPost(doc)
		newInboundJobPosts = append(newInboundJobPosts, newInboundJobPost)
	})
	var jobPostLinkErrs []error
	for _, jobPostLink := range jobPostLinksSlice {
		link := w.baseUrl + jobPostLink
		jobPostLinkErr := w.scraperEngine.Visit(link)
		if jobPostLinkErr != nil {
			jobPostLinkErrs = append(jobPostLinkErrs, jobPostLinkErr)
			exception.LogErrorWithLabel("jobPostLinkErr", jobPostLinkErr)
		}
		time.Sleep(time.Duration(secondsBetweenJobPosts) * time.Second)
	}
	if len(jobPostLinkErrs) > 0 {
		err = fmt.Errorf("%v\njobPostLinkErrs: %v", err, jobPostLinkErrs)
	}
	return newInboundJobPosts, err
}
