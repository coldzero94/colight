package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	// headlessTimeout is the max time for a headless browser render
	headlessTimeout = 30 * time.Second

	// headlessWaitAfterLoad is extra wait after page load for JS rendering
	headlessWaitAfterLoad = 3 * time.Second

	// minRenderedHTMLLen is the minimum HTML length to consider rendering successful
	minRenderedHTMLLen = 500
)

// HeadlessFetcher renders JavaScript-heavy SPA pages using a headless Chrome browser.
// Falls back gracefully if Chrome is not available.
type HeadlessFetcher struct{}

// NewHeadlessFetcher creates a new HeadlessFetcher.
func NewHeadlessFetcher() *HeadlessFetcher {
	return &HeadlessFetcher{}
}

// FetchRenderedHTML navigates to a URL with headless Chrome, waits for JS to render,
// and returns the fully-rendered DOM HTML.
func (f *HeadlessFetcher) FetchRenderedHTML(ctx context.Context, targetURL string) (string, error) {
	// Create a timeout context for the entire operation
	ctx, cancel := context.WithTimeout(ctx, headlessTimeout)
	defer cancel()

	// Create headless Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-background-networking", true),
		chromedp.Flag("disable-default-apps", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-translate", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var renderedHTML string

	err := chromedp.Run(browserCtx,
		chromedp.Navigate(targetURL),
		// Wait for the page to finish loading
		chromedp.WaitReady("body"),
		// Extra wait for JS frameworks to render content
		chromedp.Sleep(headlessWaitAfterLoad),
		// Extract the full rendered HTML
		chromedp.OuterHTML("html", &renderedHTML),
	)
	if err != nil {
		return "", fmt.Errorf("headless render failed: %w", err)
	}

	if len(renderedHTML) < minRenderedHTMLLen {
		return "", fmt.Errorf("headless render returned too little content (%d bytes)", len(renderedHTML))
	}

	slog.Info("headless_render_success", "url", targetURL, "html_len", len(renderedHTML))
	return renderedHTML, nil
}

// FetchRenderedHTMLWithWait is like FetchRenderedHTML but waits for a specific CSS selector
// to appear in the DOM before extracting HTML. Useful for sites where content loads lazily.
func (f *HeadlessFetcher) FetchRenderedHTMLWithWait(ctx context.Context, targetURL string, waitSelector string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, headlessTimeout)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	var renderedHTML string

	err := chromedp.Run(browserCtx,
		chromedp.Navigate(targetURL),
		chromedp.WaitVisible(waitSelector, chromedp.ByQuery),
		// Extra wait for any remaining JS rendering
		chromedp.Sleep(1*time.Second),
		chromedp.OuterHTML("html", &renderedHTML),
	)
	if err != nil {
		return "", fmt.Errorf("headless render with wait failed: %w", err)
	}

	if len(renderedHTML) < minRenderedHTMLLen {
		return "", fmt.Errorf("headless render returned too little content (%d bytes)", len(renderedHTML))
	}

	slog.Info("headless_render_with_wait_success", "url", targetURL, "selector", waitSelector, "html_len", len(renderedHTML))
	return renderedHTML, nil
}
