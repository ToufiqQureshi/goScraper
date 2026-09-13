// Package goscraper is goScraper's simple, one-stop API: a pool of
// browsers, each with its own pool of reusable tabs, behind a single
// Get call. A crashed tab — and a crashed browser process — are both
// replaced automatically.
package goscraper

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// Config controls pool sizes and recycling. Zero values fall back to
// sensible defaults.
type Config struct {
	Browsers        int // how many Chrome browsers to run (default 2)
	PagesPerBrowser int // how many tabs to keep ready per browser (default 5)

	// RecycleAfter bounds how long a tab can sit idle before it's
	// replaced, even if it still looks healthy (default 10 minutes).
	RecycleAfter time.Duration
	// SkipHealthCheckWithin avoids a health check for a tab used this
	// recently (default 2 seconds).
	SkipHealthCheckWithin time.Duration
}

func (c Config) withDefaults() Config {
	if c.Browsers <= 0 {
		c.Browsers = 2
	}
	if c.PagesPerBrowser <= 0 {
		c.PagesPerBrowser = 5
	}
	return c
}

// Scraper is a ready-to-use pool of browsers and tabs.
type Scraper struct {
	slots []*slot
	next  atomic.Uint64
}

// New launches browsers and their tab pools in parallel, using cfg
// (or its defaults). ctx bounds only this startup.
func New(ctx context.Context, cfg Config) (*Scraper, error) {
	cfg = cfg.withDefaults()

	type launch struct {
		slot *slot
		err  error
	}

	launches := make(chan launch, cfg.Browsers)
	var wg sync.WaitGroup
	for i := 0; i < cfg.Browsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sl, err := newSlot(ctx, cfg)
			launches <- launch{slot: sl, err: err}
		}()
	}
	wg.Wait()
	close(launches)

	s := &Scraper{}
	var firstErr error
	for l := range launches {
		if l.err != nil {
			if firstErr == nil {
				firstErr = l.err
			}
			continue
		}
		s.slots = append(s.slots, l.slot)
	}

	if firstErr != nil {
		s.Close()
		return nil, firstErr
	}

	return s, nil
}

// Get fetches url using one of the pooled browsers, reusing one of
// its tabs, and calls fn with the resulting page. The tab is returned
// to its pool automatically once fn returns, whether or not fn
// errors. If the browser handling this request has died, it is
// replaced and the request retried once.
func (s *Scraper) Get(ctx context.Context, url string, fn func(*Page) error) error {
	if len(s.slots) == 0 {
		return errClosed
	}

	idx := int(s.next.Add(1)-1) % len(s.slots)

	page, pool, err := s.slots[idx].get(ctx, url)
	if err != nil {
		return err
	}
	defer pool.Release(page)

	return fn(&Page{page: page, ctx: ctx})
}

// Close shuts down every browser and tab. Call it when you're done
// scraping. Safe to call more than once.
func (s *Scraper) Close() {
	for _, sl := range s.slots {
		sl.close()
	}
	s.slots = nil
}
