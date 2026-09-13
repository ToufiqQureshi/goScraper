package goscraper

import (
	"context"
	"errors"
	"testing"

	scraper "github.com/ToufiqQureshi/Scraper"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}.withDefaults()
	if cfg.Browsers != 2 {
		t.Fatalf("expected default Browsers=2, got %d", cfg.Browsers)
	}
	if cfg.PagesPerBrowser != 5 {
		t.Fatalf("expected default PagesPerBrowser=5, got %d", cfg.PagesPerBrowser)
	}
}

func TestConfigKeepsExplicitValues(t *testing.T) {
	cfg := Config{Browsers: 3, PagesPerBrowser: 7}.withDefaults()
	if cfg.Browsers != 3 || cfg.PagesPerBrowser != 7 {
		t.Fatalf("withDefaults should not override explicit values, got %+v", cfg)
	}
}

func TestGetOnUnstartedScraperReturnsError(t *testing.T) {
	s := &Scraper{}
	err := s.Get(context.Background(), "https://example.com", func(*Page) error { return nil })
	if err == nil {
		t.Fatal("Get on a zero-value Scraper should return an error, not panic or hang")
	}
}

func TestCloseOnZeroValueScraperDoesNotPanic(t *testing.T) {
	s := &Scraper{}
	s.Close()
}

func TestSlotCloseIsIdempotent(t *testing.T) {
	sl := &slot{}
	sl.close()
	sl.close() // should not panic
}

func TestSlotRefusesWorkOnceClosed(t *testing.T) {
	sl := &slot{}
	sl.close()

	if _, _, err := sl.current(); !errors.Is(err, errClosed) {
		t.Fatalf("current on a closed slot should return errClosed, got: %v", err)
	}

	if _, _, err := sl.get(context.Background(), "https://example.com"); !errors.Is(err, errClosed) {
		t.Fatalf("get on a closed slot should return errClosed, got: %v", err)
	}

	// A rebuild must never resurrect browsers after Close, or they'd
	// leak with nothing left holding a reference to them.
	if err := sl.rebuild(context.Background(), nil); !errors.Is(err, errClosed) {
		t.Fatalf("rebuild on a closed slot should return errClosed, got: %v", err)
	}
}

func TestSlotRebuildSkipsWhenBrowserAlreadyReplaced(t *testing.T) {
	// Two callers hitting the same crash must not each launch a
	// replacement: the second sees a different browser and no-ops.
	replaced := &scraper.Browser{}
	sl := &slot{browser: replaced}

	stale := &scraper.Browser{}
	if err := sl.rebuild(context.Background(), stale); err != nil {
		t.Fatalf("rebuild should no-op when the browser was already replaced, got: %v", err)
	}

	if sl.browser != replaced {
		t.Fatal("rebuild should have left the already-replaced browser alone")
	}
}
