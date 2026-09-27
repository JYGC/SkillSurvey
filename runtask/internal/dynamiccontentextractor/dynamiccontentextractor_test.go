package dynamiccontentextractor_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"keybook/runtask/internal/dynamiccontentextractor"
)

// newSlowResponseStubServer's handler sleeps via select on stopSleepingEarlyForTestTeardown rather
// than a bare time.Sleep. A bare time.Sleep ignores request cancellation, so when the client (Chrome,
// orchestrated by chromedp) abandons the request after the configured extraction timeout elapses, the
// handler goroutine keeps sleeping for the full configured duration regardless — and httptest.Server's
// Close, called during test cleanup, blocks until that goroutine returns. Measured before this fix:
// TestExtractDynamicContentAbortsWhenConfiguredTimeoutElapses reported 31.5s+ wall time (a 30s sleep
// minus the 2s configured timeout, spent entirely blocked inside the deferred Close), even though the
// test's own elapsed-time assertion — measuring only the ExtractDynamicContent call — correctly showed
// the abort happening within 2s. Closing stopSleepingEarlyForTestTeardown before the server during
// t.Cleanup lets a still-sleeping handler return immediately once the test itself no longer needs it.
func newSlowResponseStubServer(t *testing.T, sleepBeforeResponding time.Duration) *httptest.Server {
	t.Helper()
	stopSleepingEarlyForTestTeardown := make(chan struct{})
	stubServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(sleepBeforeResponding):
		case <-stopSleepingEarlyForTestTeardown:
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body><h1>Hello</h1></body></html>`)
	}))
	t.Cleanup(func() {
		close(stopSleepingEarlyForTestTeardown)
		stubServer.Close()
	})
	return stubServer
}

func TestExtractDynamicContentAbortsWhenConfiguredTimeoutElapses(t *testing.T) {
	stubServer := newSlowResponseStubServer(t, 30*time.Second)

	extractor := dynamiccontentextractor.NewDynamicContentExtractor(2 * time.Second)

	start := time.Now()
	err := extractor.ExtractDynamicContent(stubServer.URL, func(ctx context.Context) error {
		return nil
	})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error when the configured timeout elapses, got nil")
	}
	if elapsed >= 20*time.Second {
		t.Fatalf("expected abort well under the stub's 30s sleep, took %v", elapsed)
	}
}

func TestExtractDynamicContentSucceedsWithinGenerousConfiguredTimeout(t *testing.T) {
	stubServer := newSlowResponseStubServer(t, 0)

	extractor := dynamiccontentextractor.NewDynamicContentExtractor(30 * time.Second)

	var heading string
	err := extractor.ExtractDynamicContent(stubServer.URL, func(ctx context.Context) error {
		return dynamiccontentextractor.GetTextBySelector("h1", &heading, ctx)
	})

	if err != nil {
		t.Fatalf("expected extraction to succeed within a generous timeout, got: %v", err)
	}
	if heading != "Hello" {
		t.Fatalf("expected extractFunction to observe page content %q, got %q", "Hello", heading)
	}
}
