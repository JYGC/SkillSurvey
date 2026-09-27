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

func newSlowResponseStubServer(sleepBeforeResponding time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleepBeforeResponding)
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!DOCTYPE html><html><body><h1>Hello</h1></body></html>`)
	}))
}

func TestExtractDynamicContentAbortsWhenConfiguredTimeoutElapses(t *testing.T) {
	stubServer := newSlowResponseStubServer(30 * time.Second)
	defer stubServer.Close()

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
	stubServer := newSlowResponseStubServer(0)
	defer stubServer.Close()

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
