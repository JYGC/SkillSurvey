package dynamiccontentextractor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JYGC/SkillSurvey/internal/config"
	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/chromedp/chromedp"
)

type DynamicContentExtractor struct {
	configSettings  config.SiteAdapterConfig
	chromedpOptions []chromedp.ExecAllocatorOption
}

func NewDynamicContentExtractor(
	configSettingsInterface interface{},
) DynamicContentExtractor {
	var configSettings config.SiteAdapterConfig
	configSettingsBytes, configSettingsBytesErr := json.Marshal(configSettingsInterface)
	if configSettingsBytesErr != nil {
		configSettings = config.SiteAdapterConfig{}
	}
	json.Unmarshal(configSettingsBytes, &configSettings)
	chromedpOptions := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),                                 // Headless mode; set to false for headful if needed
		chromedp.Flag("disable-blink-features", "AutomationControlled"), // Hide automation signals
	}
	return DynamicContentExtractor{
		configSettings,
		chromedpOptions,
	}
}

func (d DynamicContentExtractor) GetInboundJobPost(
	url string,
) (
	newInboundJobPost entities.InboundJobPost,
	err error,
) {
	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(
		context.Background(),
		d.chromedpOptions...,
	)
	defer allocatorCancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()

	newInboundJobPost = entities.InboundJobPost{}
	err = chromedp.Run(
		timeoutCtx,
		chromedp.Evaluate(
			`Object.defineProperty(navigator, 'webdriver', {get: () => undefined});`,
			nil,
		),
		chromedp.Evaluate(
			`Object.defineProperty(navigator, 'plugins', {get: () => [1, 2, 3, 4, 5]});`,
			nil,
		),
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQueryAll),
		chromedp.Text(
			d.configSettings.SiteSelectors.TitleSelector,
			&(newInboundJobPost.Title),
		),
		chromedp.Text(
			d.configSettings.SiteSelectors.BodySelector,
			&newInboundJobPost.Body,
		),
	)

	if err != nil {
		return entities.InboundJobPost{}, err
	}

	return newInboundJobPost, nil
}
