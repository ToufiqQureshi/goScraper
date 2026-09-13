package goscraper

import (
	"context"
	"sync"
	"testing"

	"github.com/ToufiqQureshi/Scraper/internal/chrometest"
)

func TestScraperGetFetchesAPage(t *testing.T) {
	chrometest.Require(t)

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
	chrometest.Require(t)

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

func TestScraperReplacesACrashedBrowser(t *testing.T) {
	chrometest.Require(t)

	ctx := context.Background()
	s, err := New(ctx, Config{Browsers: 1, PagesPerBrowser: 1})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer s.Close()

	if err := s.Get(ctx, "https://example.com", func(*Page) error { return nil }); err != nil {
		t.Fatalf("first Get returned an error: %v", err)
	}

	// Kill the whole Chrome process behind this slot, not just a tab.
	dead, _, err := s.slots[0].current()
	if err != nil {
		t.Fatalf("current returned an error: %v", err)
	}
	dead.Close()

	// The next request must still succeed: the slot rebuilds itself.
	var title string
	err = s.Get(ctx, "https://example.com", func(p *Page) error {
		var err error
		title, err = p.Title()
		return err
	})
	if err != nil {
		t.Fatalf("Get after a browser crash should recover, got: %v", err)
	}
	if title == "" {
		t.Fatal("expected a working page after the browser was replaced")
	}

	alive, _, err := s.slots[0].current()
	if err != nil {
		t.Fatalf("current returned an error: %v", err)
	}
	if alive == dead {
		t.Fatal("the crashed browser should have been replaced, not reused")
	}
	if !alive.Healthy(ctx) {
		t.Fatal("the replacement browser should be healthy")
	}
}

func TestScraperKeepsWorkingAfterRepeatedBrowserCrashes(t *testing.T) {
	chrometest.Require(t)

	ctx := context.Background()
	s, err := New(ctx, Config{Browsers: 1, PagesPerBrowser: 1})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer s.Close()

	// A crash loop must not degrade into a permanently broken slot.
	for i := 0; i < 3; i++ {
		dead, _, err := s.slots[0].current()
		if err != nil {
			t.Fatalf("current returned an error: %v", err)
		}
		dead.Close()

		if err := s.Get(ctx, "https://example.com", func(*Page) error { return nil }); err != nil {
			t.Fatalf("Get after crash %d returned an error: %v", i+1, err)
		}
	}
}

func TestScraperHandlesConcurrentGetsAcrossBrowsers(t *testing.T) {
	chrometest.Require(t)

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
