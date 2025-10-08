package siteadapters

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/dynamiccontentextractor"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/JYGC/SkillSurvey/internal/environment"
	"github.com/JYGC/SkillSurvey/internal/webscraper"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

const joraConfigFileName = "jora.json"

type JoraAdapter struct {
	ConfigSettings          JoraAdapterConfig
	webScraper              *webscraper.WebScraper
	dynamicContentExtractor *dynamiccontentextractor.DynamicContentExtractor
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
	jora.dynamicContentExtractor = dynamiccontentextractor.NewDynamicContentExtractor()
	return jora
}

func (j JoraAdapter) RunSurvey() (
	newInboundJobPosts []entities.InboundJobPost,
	err error,
) {
	var errParts []error
	var pageErrors []error
	jobPostLinks := []string{}
	for _, searchCriteria := range j.ConfigSettings.SearchCriterias {
		for searchPage := 1; searchPage <= j.ConfigSettings.Pages; searchPage++ {
			pageUrl := strings.ReplaceAll(
				searchCriteria.Url,
				j.ConfigSettings.PageFlag,
				strconv.Itoa(searchPage),
			)
			pageError := j.dynamicContentExtractor.ExtractDynamicContent(
				pageUrl,
				func(ctx context.Context) (err error) {
					var nodes []*cdp.Node
					if err := chromedp.Nodes(".job-link", &nodes, chromedp.ByQueryAll).Do(ctx); err != nil {
						return err
					}

					for _, node := range nodes {
						if jobPostPath, ok := node.Attribute("href"); ok {
							jobPostLink := j.ConfigSettings.BaseUrl + jobPostPath
							jobPostLinks = append(jobPostLinks, jobPostLink)
						}
					}
					return nil
				},
			)
			if pageError != nil {
				pageErrors = append(pageErrors, pageError)
			}
		}
	}
	if len(pageErrors) > 0 {
		errParts = append(errParts, fmt.Errorf("pageErrors: %v", pageErrors))
	}

	var jobPostErrors []error
	for _, jobPostLink := range jobPostLinks {
		jobPostError := j.dynamicContentExtractor.ExtractDynamicContent(
			jobPostLink,
			func(ctx context.Context) (jobPostErr error) {
				var jobPostErrParts []error
				newInboundJobPost := entities.InboundJobPost{}
				newInboundJobPost.SiteName = j.ConfigSettings.SiteSelectors.SiteName
				newInboundJobPost.JobSiteNumber = j.getJobSiteNumber(jobPostLink)
				postedDate, getPostedDateErr := j.getPostedDate(ctx)
				if getPostedDateErr != nil {
					jobPostErrParts = append(jobPostErrParts, fmt.Errorf("getPostedDateErr: %v", getPostedDateErr))
				}
				newInboundJobPost.PostedDate = postedDate
				if getTitleErr := dynamiccontentextractor.GetTextBySelector(j.ConfigSettings.SiteSelectors.TitleSelector, &newInboundJobPost.Title, ctx); getTitleErr != nil {
					jobPostErrParts = append(jobPostErrParts, fmt.Errorf("getTitleErr: %v", getTitleErr))
				}
				if getBodyErr := dynamiccontentextractor.GetTextBySelector(j.ConfigSettings.SiteSelectors.BodySelector, &newInboundJobPost.Body, ctx); getBodyErr != nil {
					jobPostErrParts = append(jobPostErrParts, fmt.Errorf("getBodyErr: %v", getBodyErr))
				}
				if getCityErr := dynamiccontentextractor.GetTextBySelector(j.ConfigSettings.SiteSelectors.CitySelector, &newInboundJobPost.City, ctx); getCityErr != nil {
					jobPostErrParts = append(jobPostErrParts, fmt.Errorf("getCityErr: %v", getCityErr))
				}
				newInboundJobPost.Country = j.ConfigSettings.SiteSelectors.Country
				if getSuburbErr := dynamiccontentextractor.GetTextBySelector(j.ConfigSettings.SiteSelectors.SuburbSelector, &newInboundJobPost.Suburb, ctx); getSuburbErr != nil {
					jobPostErrParts = append(jobPostErrParts, fmt.Errorf("getSuburbErr: %v", getSuburbErr))
				}
				if len(jobPostErrParts) > 0 {
					jobPostErr = fmt.Errorf("jobPostLink: %v %v", jobPostLink, jobPostErrParts)
					fmt.Printf("jobPostErr: %v\n", jobPostErr)
				} else {
					fmt.Printf("jobPostLink: %v\n", jobPostLink)
					newInboundJobPosts = append(newInboundJobPosts, newInboundJobPost)
				}
				return jobPostErr
			},
		)
		if jobPostError != nil {
			jobPostErrors = append(jobPostErrors, jobPostError)
		}
	}
	if len(jobPostErrors) > 0 {
		errParts = append(errParts, fmt.Errorf("jobPostErrors: %v", jobPostErrors))
	}

	if len(errParts) > 0 {
		err = fmt.Errorf("%v", errParts)
	}

	return newInboundJobPosts, err
}

// Advertisement's post date can calculated by subtracting how old the advert is in days from
// the current date.
func (j JoraAdapter) getPostedDate(ctx context.Context) (postedDate time.Time, err error) {
	var ageString string
	if getAgeStringErr := dynamiccontentextractor.GetTextBySelector(j.ConfigSettings.SiteSelectors.PostedDateSelector, &ageString, ctx); getAgeStringErr != nil {
		return postedDate, getAgeStringErr
	}

	currentDate := time.Now()

	if daysIndex := strings.Index(ageString, "d ago"); daysIndex > 0 {
		var daysOld int
		if daysOld, err = strconv.Atoi(ageString[0:daysIndex]); err != nil {
			return postedDate, err
		}
		postedDate = currentDate.AddDate(0, 0, -daysOld)
		return postedDate, nil
	}

	if hoursIndex := strings.Index(ageString, "h ago"); hoursIndex > 0 {

		var hoursOld int
		if hoursOld, err = strconv.Atoi(ageString[0:hoursIndex]); err != nil {
			return postedDate, err
		}
		postedDate = currentDate.Add(time.Duration(-hoursOld) * time.Hour)
		return postedDate, nil
	}

	if monthsIndex := strings.Index(ageString, "mo age"); monthsIndex > 0 {
		var monthsOld int
		if monthsOld, err = strconv.Atoi(ageString[0:monthsIndex]); err != nil {
			return postedDate, err
		}
		postedDate = currentDate.AddDate(0, -monthsOld, 0)
		return postedDate, nil
	}

	return currentDate, errors.New("cannot determine posted date. using current date")
}

func (j JoraAdapter) getJobSiteNumber(url string) string {
	return url[strings.Index(url, "/job/")+5 : strings.LastIndex(url, "?")]
}
