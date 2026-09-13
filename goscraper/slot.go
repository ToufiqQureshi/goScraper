package goscraper

import (
	"context"
	"errors"
	"sync"

	"github.com/ToufiqQureshi/Scraper/browser"
	"github.com/ToufiqQureshi/Scraper/pagepool"
)

// errClosed is returned once the slot (or its Scraper) is shut down.
var errClosed = errors.New("goscraper: scraper is closed")

// slot is one browser plus the tab pool living on it. It rebuilds
// itself if that whole Chrome process dies, so a crashed browser
// doesn't permanently take a share of traffic down with it.
type slot struct {
	cfg Config

	mu      sync.Mutex
	browser *browser.Browser
	pages   *pagepool.Pool
	closed  bool
}

// newSlot launches one browser and opens its pool of tabs.
func newSlot(ctx context.Context, cfg Config) (*slot, error) {
	b, pages, err := startPair(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &slot{cfg: cfg, browser: b, pages: pages}, nil
}

// startPair launches a browser and its tab pool together, closing the
// browser if the pool can't be opened on it.
func startPair(ctx context.Context, cfg Config) (*browser.Browser, *pagepool.Pool, error) {
	b, err := browser.New(ctx, cfg.Browser)
	if err != nil {
		return nil, nil, err
	}

	pages, err := pagepool.New(ctx, b, cfg.PagesPerBrowser, pagepool.Options{
		RecycleAfter:          cfg.RecycleAfter,
		SkipHealthCheckWithin: cfg.SkipHealthCheckWithin,
	})
	if err != nil {
		b.Close()
		return nil, nil, err
	}

	return b, pages, nil
}

// get returns a page for url plus the pool it must be released to. If
// the request failed because this slot's whole browser died, the slot
// is rebuilt once and the request retried on the fresh browser.
func (s *slot) get(ctx context.Context, url string) (*browser.Page, *pagepool.Pool, error) {
	b, pages, err := s.current()
	if err != nil {
		return nil, nil, err
	}

	page, err := pages.Get(ctx, url)
	if err == nil {
		return page, pages, nil
	}

	// Don't blame the browser for the caller giving up, and don't
	// blame it for one bad URL either: ask the browser itself whether
	// it is still alive before deciding to replace it.
	if ctx.Err() != nil || b.Healthy(ctx) {
		return nil, nil, err
	}

	if rebuildErr := s.rebuild(ctx, b); rebuildErr != nil {
		return nil, nil, rebuildErr
	}

	_, pages, err = s.current()
	if err != nil {
		return nil, nil, err
	}

	page, err = pages.Get(ctx, url)
	if err != nil {
		return nil, nil, err
	}

	return page, pages, nil
}

// current reads this slot's browser and pool under lock.
func (s *slot) current() (*browser.Browser, *pagepool.Pool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, nil, errClosed
	}

	return s.browser, s.pages, nil
}

// rebuild replaces a dead browser and its tabs. dead is the browser
// the caller saw fail, so concurrent callers that hit the same crash
// don't each launch a replacement.
func (s *slot) rebuild(ctx context.Context, dead *browser.Browser) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return errClosed
	}
	if s.browser != dead {
		return nil // someone else already replaced it
	}

	b, pages, err := startPair(ctx, s.cfg)
	if err != nil {
		return err
	}

	s.pages.Close()
	s.browser.Close()
	s.browser, s.pages = b, pages

	return nil
}

// close shuts down this slot's tabs and browser. Safe to call twice.
func (s *slot) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}
	s.closed = true

	if s.pages != nil {
		s.pages.Close()
	}
	if s.browser != nil {
		s.browser.Close()
	}
}
