package scrape

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/exception"
	"keybook/runtask/internal/pbclient"
	"keybook/runtask/internal/siteadapters"
)

// Run fetches job listings from all configured sites and upserts them into PocketBase.
func Run(cfg config.Config, pb *pbclient.Client) error {
	sites, err := pb.GetSites()
	if err != nil {
		return fmt.Errorf("get sites: %w", err)
	}

	for _, site := range sites {
		adapter, err := adapterForSite(site.Name, cfg)
		if err != nil {
			exception.LogErrorWithLabel("adapterForSite", err)
			continue
		}

		posts, err := adapter.RunSurvey()
		if err != nil {
			exception.LogErrorWithLabel("RunSurvey", err)
			// Continue — errors are per-page/post, we may have partial results.
		}

		for _, post := range posts {
			if err := pb.UpsertJobPost(pbclient.JobPost{
				JobSiteNumber: post.JobSiteNumber,
				SiteId:        site.Id,
				Content: pbclient.JobPostContent{
					Title: post.Title,
					Body:  post.Body,
				},
				Location: pbclient.JobPostLocation{
					City:    post.City,
					Country: post.Country,
					Suburb:  post.Suburb,
				},
				PostedDate: post.PostedDate,
			}); err != nil {
				exception.LogErrorWithLabel("UpsertJobPost", err)
			}
		}
	}
	return nil
}

// adapterForSite selects the right adapter based on the site name matching a config file name.
func adapterForSite(siteName string, cfg config.Config) (siteadapters.ISiteAdapter, error) {
	seekName := strings.TrimSuffix(filepath.Base(cfg.SeekConfigFile), filepath.Ext(cfg.SeekConfigFile))
	joraName := strings.TrimSuffix(filepath.Base(cfg.JoraConfigFile), filepath.Ext(cfg.JoraConfigFile))

	switch {
	case strings.EqualFold(siteName, seekName) || strings.EqualFold(siteName, "seek"):
		return siteadapters.NewSeekAdapter(cfg.SeekConfigFile)
	case strings.EqualFold(siteName, joraName) || strings.EqualFold(siteName, "jora"):
		return siteadapters.NewJoraAdapter(cfg.JoraConfigFile)
	default:
		return nil, fmt.Errorf("no adapter configured for site %q", siteName)
	}
}

// ensure time is used (postedDate field)
var _ = time.Now
