package goscraper

import (
	"context"
	"sync"
	"testing"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// requireChrome skips the test when no real Chrome is available to
// launch (e.g. this sandbox). It still runs wherever Chrome is
// installed, including the project's CI.
func requireChrome(t *testing.T) {
	t.Helper()

	b, err := scraper.New(context.Background())
	if err != nil {
		t.Skipf("skipping: no Chrome available to launch: %v", err)
	}
	b.Close()
}

func TestScraperGetFetchesAPage(t *testing.T) {
	requireChrome(t)

	ctx := context.Background()
	s, err := New(ctx, Config{Browsers: 1, PagesPerBrowser: 1})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer s.Close()

	var title string
	err = s.Get(ctx, "https://example.com", func(p *Page) error {
		var err error
		title, err = p.Title()
		return err
	})
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if title == "" {
		t.Fatal("expected a non-empty title after fetching example.com")
	}
}

func TestScraperGetReturnsCallbackError(t *testing.T) {
	requireChrome(t)

	ctx := context.Background()
	s, err := New(ctx, Config{Browsers: 1, PagesPerBrowser: 1})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer s.Close()

	sentinel := errorString("callback failed")
	err = s.Get(ctx, "https://example.com", func(*Page) error { return sentinel })
	if err != sentinel {
		t.Fatalf("expected the callback's own error to propagate, got: %v", err)
	}
}

type errorString string

func (e errorString) Error() string { return string(e) }

func TestScraperHandlesConcurrentGetsAcrossBrowsers(t *testing.T) {
	requireChrome(t)

	ctx := context.Background()
	s, err := New(ctx, Config{Browsers: 2, PagesPerBrowser: 2})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer s.Close()

	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- s.Get(ctx, "https://example.com", func(p *Page) error {
				_, err := p.Title()
				return err
			})
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Get returned an error: %v", err)
		}
	}
}
