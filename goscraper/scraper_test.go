package goscraper

import (
	"context"
	"testing"
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
