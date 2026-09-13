package pagepool

import (
	"context"
	"errors"
	"testing"
	"time"

	pool "github.com/ToufiqQureshi/Scraper/internal/pool"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// newTestPool builds a pool without launching real Chrome, so these
// tests can run anywhere. The hard concurrency/recycling logic lives
// in internal/pool and is tested exhaustively there; these tests only
// check that this package wires it up correctly for *scraper.Page,
// including the navigate-then-serve behavior specific to pages.
func newTestPool(t *testing.T, size int) *Pool {
	t.Helper()

	core, err := pool.New(
		size,
		func(*scraper.Page) bool { return true },
		func() (*scraper.Page, error) { return &scraper.Page{}, nil },
		func(*scraper.Page) {},
	)
	if err != nil {
		t.Fatalf("pool.New returned an error: %v", err)
	}

	return &Pool{
		core:     core,
		open:     func(string) (*scraper.Page, error) { return &scraper.Page{}, nil },
		navigate: func(*scraper.Page, string) error { return nil },
	}
}

func TestNewRejectsNilBrowser(t *testing.T) {
	if _, err := New(nil, 1); err == nil {
		t.Fatal("New(nil, 1) should return an error")
	}
}

func TestGetReturnsAPage(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	page, err := p.Get(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if page == nil {
		t.Fatal("Get returned a nil page")
	}
}

func TestGetNavigatesTheReusedPage(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	var gotURL string
	p.navigate = func(_ *scraper.Page, url string) error {
		gotURL = url
		return nil
	}

	if _, err := p.Get(context.Background(), "https://example.com/products"); err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	if gotURL != "https://example.com/products" {
		t.Fatalf("expected Navigate to be called with the requested url, got %q", gotURL)
	}
}

func TestGetReplacesPageWhenNavigateFails(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	navigateErr := errors.New("tab crashed mid-navigation")
	p.navigate = func(*scraper.Page, string) error { return navigateErr }

	opened := &scraper.Page{}
	var openedURL string
	p.open = func(url string) (*scraper.Page, error) {
		openedURL = url
		return opened, nil
	}

	page, err := p.Get(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Get should recover by opening a fresh page, got error: %v", err)
	}
	if page != opened {
		t.Fatal("Get should return the freshly opened page when Navigate fails")
	}
	if openedURL != "https://example.com" {
		t.Fatalf("expected the fallback open to use the requested url, got %q", openedURL)
	}
}

func TestGetPropagatesFallbackOpenError(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	p.navigate = func(*scraper.Page, string) error { return errors.New("crashed") }
	openErr := errors.New("chrome is gone")
	p.open = func(string) (*scraper.Page, error) { return nil, openErr }

	if _, err := p.Get(context.Background(), "https://example.com"); !errors.Is(err, openErr) {
		t.Fatalf("expected the fallback open error to propagate, got: %v", err)
	}
}

func TestGetReturnsErrorWhenContextCancelled(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	if _, err := p.Get(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := p.Get(ctx, "https://example.com"); err == nil {
		t.Fatal("Get should return an error when its context is cancelled")
	}
}

func TestReleasePutsPageBackForReuse(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	first, err := p.Get(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	p.Release(first)

	second, err := p.Get(context.Background(), "https://example.com/other")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	if first != second {
		t.Fatal("Release should make the same page available again")
	}
}

func TestGetAfterCloseReturnsError(t *testing.T) {
	p := newTestPool(t, 1)
	p.Close()

	if _, err := p.Get(context.Background(), "https://example.com"); !errors.Is(err, ErrClosed) {
		t.Fatal("Get should return ErrClosed once the pool is closed")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	p := newTestPool(t, 1)
	p.Close()
	p.Close() // should not panic
}

func TestReleaseNilDoesNothing(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()
	p.Release(nil) // should not panic
}

func TestStatsReflectsActivity(t *testing.T) {
	p := newTestPool(t, 2)
	defer p.Close()

	page, err := p.Get(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	p.Release(page)

	stats := p.Stats()
	if stats.Created != 2 {
		t.Fatalf("expected Created=2, got %d", stats.Created)
	}
	if stats.Gets != 1 {
		t.Fatalf("expected Gets=1, got %d", stats.Gets)
	}
	if stats.Releases != 1 {
		t.Fatalf("expected Releases=1, got %d", stats.Releases)
	}
}
