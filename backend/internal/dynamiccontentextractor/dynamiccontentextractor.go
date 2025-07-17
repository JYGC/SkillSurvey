package dynamiccontentextractor

import (
	"context"
	"time"

	"github.com/JYGC/SkillSurvey/internal/entities"
	"github.com/chromedp/chromedp"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"
const webdriverProperty = `Object.defineProperty(navigator, 'webdriver', {get: () => undefined});`
const pluginsProperty = `Object.defineProperty(navigator, 'plugins', {get: () => [1, 2, 3, 4, 5]});`

type DynamicContentExtractor struct {
	chromedpOptions []chromedp.ExecAllocatorOption
}

func NewDynamicContentExtractor() DynamicContentExtractor {
	chromedpOptions := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(userAgent),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),                                 // Headless mode; set to false for headful if needed
		chromedp.Flag("disable-blink-features", "AutomationControlled"), // Hide automation signals
	}
	return DynamicContentExtractor{
		chromedpOptions,
	}
}

func (d DynamicContentExtractor) GetInboundJobPost(
	url string,
	matchSiteSelectorToProperties func(
		newInboundJobPost *entities.InboundJobPost,
	) map[string]*string,
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
	chromedpActions := []chromedp.Action{
		chromedp.Evaluate(webdriverProperty, nil),
		chromedp.Evaluate(pluginsProperty, nil),
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQueryAll),
	}
	matchedSiteSelectors := matchSiteSelectorToProperties(&newInboundJobPost)
	for selector, propertyPointer := range matchedSiteSelectors {
		chromedpActions = append(
			chromedpActions,
			chromedp.Text(selector, propertyPointer),
		)
	}
	err = chromedp.Run(timeoutCtx, chromedpActions...)

	if err != nil {
		return entities.InboundJobPost{}, err
	}

	return newInboundJobPost, nil
}
