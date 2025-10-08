package dynamiccontentextractor

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36"
const webdriverProperty = `Object.defineProperty(navigator, 'webdriver', {get: () => undefined});`
const pluginsProperty = `Object.defineProperty(navigator, 'plugins', {get: () => [1, 2, 3, 4, 5]});`

type DynamicContentExtractor struct {
	chromedpOptions []chromedp.ExecAllocatorOption
}

func NewDynamicContentExtractor() *DynamicContentExtractor {
	chromedpOptions := []chromedp.ExecAllocatorOption{
		chromedp.UserAgent(userAgent),
		chromedp.WindowSize(1920, 1080),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("headless", true),                                 // Headless mode; set to false for headful if needed
		chromedp.Flag("disable-blink-features", "AutomationControlled"), // Hide automation signals
	}
	return &DynamicContentExtractor{
		chromedpOptions,
	}
}

func (d DynamicContentExtractor) ExtractDynamicContent(
	url string,
	extractFunction func(context.Context) error,
) (err error) {
	allocatorCtx, allocatorCancel := chromedp.NewExecAllocator(
		context.Background(),
		d.chromedpOptions...,
	)
	defer allocatorCancel()

	ctx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, 60*time.Second)
	defer timeoutCancel()

	chromedpActions := []chromedp.Action{
		chromedp.Evaluate(webdriverProperty, nil),
		chromedp.Evaluate(pluginsProperty, nil),
		chromedp.Navigate(url),
		chromedp.WaitVisible("html", chromedp.ByQueryAll),
		chromedp.ActionFunc(extractFunction),
	}

	err = chromedp.Run(timeoutCtx, chromedpActions...)

	if err != nil {
		return err
	}

	return nil
}

func getContentFromContext(selector string, getContentFunc func([]*cdp.Node) error, ctx context.Context) (err error) {
	var errParts []error
	var nodes []*cdp.Node
	if getNodesErr := chromedp.Nodes(selector, &nodes, chromedp.ByQuery).Do(ctx); getNodesErr != nil {
		errParts = append(errParts, fmt.Errorf("getNodesErr: %v", getNodesErr))
	}
	if len(nodes) > 0 {
		if getContentFuncErr := getContentFunc(nodes); getContentFuncErr != nil {
			errParts = append(errParts, getContentFuncErr)
		}
	}
	if len(errParts) > 0 {
		err = fmt.Errorf("%v", errParts)
	}
	return err
}

func GetTextBySelector(selector string, text *string, ctx context.Context) (err error) {
	return getContentFromContext(selector, func(nodes []*cdp.Node) error {
		if getTextErr := chromedp.Text(nodes[0].FullXPath(), text).Do(ctx); getTextErr != nil {
			return fmt.Errorf("getTextErr: %v", getTextErr)
		}
		return nil
	}, ctx)
}

func GetAttributeValue(selector string, attributeName string, value *string, ctx context.Context) (err error) {
	return getContentFromContext(selector, func(nodes []*cdp.Node) error {
		attributeExists := false
		*value, attributeExists = nodes[0].Attribute(attributeName)
		if !attributeExists {
			return fmt.Errorf("attribute not found")
		}
		return nil
	}, ctx)
}
