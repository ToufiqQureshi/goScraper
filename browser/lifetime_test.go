package browser_test

import (
	"context"
	"testing"

	"github.com/ToufiqQureshi/Scraper/internal/chrometest"
)

// TestBrowserStaysAliveAfterNew is a regression test. New used to run
// its launch on a cancellable child context, and chromedp ties the
// Chrome process to whatever context starts it - so Chrome died the
// instant New returned, and every later call failed with "context
// canceled".
func TestBrowserStaysAliveAfterNew(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	if !b.Healthy(context.Background()) {
		t.Fatal("the browser should still be running after New returns")
	}
}

// TestPageStaysUsableAfterOpen is the same regression one level down:
// the first Run on a page context creates the tab, so running it on a
// cancellable child closed the tab as soon as Open finished.
func TestPageStaysUsableAfterOpen(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	ctx := context.Background()

	page, err := b.Open(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Open returned an error: %v", err)
	}
	defer page.Close()

	if !page.Healthy(ctx) {
		t.Fatal("the tab should still be open after Open returns")
	}

	// A second command on the same tab is the part that actually broke.
	title, err := page.Title(ctx)
	if err != nil {
		t.Fatalf("using the page after Open returned an error: %v", err)
	}
	if title == "" {
		t.Fatal("expected a title from the page opened by Open")
	}
}

// TestPageIsReusableAfterNavigate covers the behaviour the page pool
// depends on: the same tab can be sent to another URL and still work.
func TestPageIsReusableAfterNavigate(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	ctx := context.Background()

	page, err := b.Open(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Open returned an error: %v", err)
	}
	defer page.Close()

	if err := page.Navigate(ctx, "https://example.org"); err != nil {
		t.Fatalf("Navigate returned an error: %v", err)
	}

	if _, err := page.Title(ctx); err != nil {
		t.Fatalf("the tab should still work after being navigated: %v", err)
	}
}
