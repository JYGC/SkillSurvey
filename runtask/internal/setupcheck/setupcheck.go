package setupcheck

import (
	"context"
	"time"

	"github.com/chromedp/chromedp"

	"keybook/runtask/internal/config"
	"keybook/runtask/internal/dynamiccontentextractor"
	"keybook/runtask/internal/pbclient"
	"keybook/runtask/internal/siteadapters"
)

type CheckOutcome struct {
	CheckName string
	Failure   error
}

const blankPageThatNeedsNoNetwork = "data:text/html,<html><body></body></html>"

var browserOptionsAddedToChromiumCheckOnlyByTests []chromedp.ExecAllocatorOption

// RunAllChecks runs every check even after one fails, so a single run shows everything that is wrong. It signs in
// to PocketBase and reads, and writes nothing.
func RunAllChecks(cfg config.Config, pb *pbclient.Client) []CheckOutcome {
	return []CheckOutcome{
		{CheckName: "seek settings", Failure: checkSeekSettingsFileLoads(cfg)},
		{CheckName: "jora settings", Failure: checkJoraSettingsFileLoads(cfg)},
		{CheckName: "pocketbase sign-in", Failure: checkServiceAccountCanSignInAndReadSites(pb)},
		{CheckName: "chromium", Failure: checkChromiumStartsAndLoadsBlankPage(cfg.DynamicContentExtractionTimeout())},
	}
}

func checkSeekSettingsFileLoads(cfg config.Config) error {
	_, err := siteadapters.NewSeekAdapter(cfg.SeekConfigFile, cfg.DynamicContentExtractionTimeout())
	return err
}

func checkJoraSettingsFileLoads(cfg config.Config) error {
	_, err := siteadapters.NewJoraAdapter(cfg.JoraConfigFile, cfg.DynamicContentExtractionTimeout())
	return err
}

func checkServiceAccountCanSignInAndReadSites(pb *pbclient.Client) error {
	_, err := pb.GetSites()
	return err
}

// Starting the browser goes further than finding it, because a wrong home folder, login class or filesystem flag
// would all let the program be found and then fail in the first scrape.
func checkChromiumStartsAndLoadsBlankPage(dynamicContentExtractionTimeout time.Duration) error {
	extractor := dynamiccontentextractor.NewDynamicContentExtractor(
		dynamicContentExtractionTimeout,
		browserOptionsAddedToChromiumCheckOnlyByTests...,
	)
	return extractor.ExtractDynamicContent(blankPageThatNeedsNoNetwork, extractNothing)
}

func extractNothing(context.Context) error {
	return nil
}
