package setupcheck

import (
	"testing"

	"github.com/chromedp/chromedp"
)

func UseBrowserExecutableForChromiumCheck(t *testing.T, browserExecutablePath string) {
	browserOptionsBeforeTest := browserOptionsAddedToChromiumCheckOnlyByTests
	browserOptionsAddedToChromiumCheckOnlyByTests = []chromedp.ExecAllocatorOption{chromedp.ExecPath(browserExecutablePath)}
	t.Cleanup(func() { browserOptionsAddedToChromiumCheckOnlyByTests = browserOptionsBeforeTest })
}
