package pagepool

import (
	"context"
	"testing"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// requireChrome skips the test when no real Chrome is available to
// launch (e.g. this sandbox). It still runs wherever Chrome is
// installed, including the project's CI.
func requireChrome(t *testing.T) *scraper.Browser {
	t.Helper()

	b, err := scraper.New(context.Background())
	if err != nil {
		t.Skipf("skipping: no Chrome available to launch: %v", err)
	}
	return b
}

func TestPoolReusesTheSameTabAcrossPages(t *testing.T) {
	browser := requireChrome(t)
	defer browser.Close()

	ctx := context.Background()
	pool, err := New(ctx, browser, 1)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	first, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	title, err := first.Title(ctx)
	if err != nil {
		t.Fatalf("Title returned an error: %v", err)
	}
	if title == "" {
		t.Fatal("expected a non-empty title after navigating to example.com")
	}
	pool.Release(first)

	second, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(second)

	if first != second {
		t.Fatal("Get should hand back the same tab instead of opening a new one")
	}
}

func TestPoolReplacesACrashedPage(t *testing.T) {
	browser := requireChrome(t)
	defer browser.Close()

	ctx := context.Background()
	pool, err := New(ctx, browser, 1)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	page, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	// Simulate a crash: the tab dies while checked out, then the
	// caller (not knowing that) gives it back to the pool.
	page.Close()
	pool.Release(page)

	replacement, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(replacement)

	if replacement == page {
		t.Fatal("Get should not hand out a crashed page")
	}
	if !replacement.Healthy(ctx) {
		t.Fatal("Get should replace a crashed page with a healthy one")
	}
}
