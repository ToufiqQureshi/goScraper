// Package goscraper is goScraper's simple, one-stop API: a pool of
// browsers, each with its own pool of reusable tabs, behind a single
// Get call.
//
// Known gap: if a whole Chrome process dies (not just one tab),
// Scraper does not yet replace that browser. Its page pool will
// return errors for that one slot (roughly 1/Browsers of requests)
// until the process is restarted. Per-tab crashes, the far more
// common case, are recovered automatically. See docs/ROADMAP.md.
package goscraper

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	scraper "github.com/ToufiqQureshi/Scraper"
	"github.com/ToufiqQureshi/Scraper/pagepool"
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
	browsers []*scraper.Browser
	pages    []*pagepool.Pool
	next     atomic.Uint64
}

// New launches browsers and their tab pools in parallel, using cfg
// (or its defaults). ctx bounds only this startup.
func New(ctx context.Context, cfg Config) (*Scraper, error) {
	cfg = cfg.withDefaults()

	type slot struct {
		browser *scraper.Browser
		pages   *pagepool.Pool
		err     error
	}

	slots := make(chan slot, cfg.Browsers)
	var wg sync.WaitGroup
	for i := 0; i < cfg.Browsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			b, err := scraper.New(ctx)
			if err != nil {
				slots <- slot{err: err}
				return
			}

			pp, err := pagepool.New(ctx, b, cfg.PagesPerBrowser, pagepool.Options{
				RecycleAfter:          cfg.RecycleAfter,
				SkipHealthCheckWithin: cfg.SkipHealthCheckWithin,
			})
			if err != nil {
				b.Close()
				slots <- slot{err: err}
				return
			}

			slots <- slot{browser: b, pages: pp}
		}()
	}
	wg.Wait()
	close(slots)

	s := &Scraper{}
	var firstErr error
	for sl := range slots {
		if sl.err != nil {
			if firstErr == nil {
				firstErr = sl.err
			}
			continue
		}
		s.browsers = append(s.browsers, sl.browser)
		s.pages = append(s.pages, sl.pages)
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
// errors.
func (s *Scraper) Get(ctx context.Context, url string, fn func(*Page) error) error {
	if len(s.pages) == 0 {
		return errors.New("goscraper: scraper is closed or was never started")
	}

	idx := int(s.next.Add(1)-1) % len(s.pages)
	pool := s.pages[idx]

	page, err := pool.Get(ctx, url)
	if err != nil {
		return err
	}
	defer pool.Release(page)

	return fn(&Page{page: page, ctx: ctx})
}

// Close shuts down every browser and tab. Call it when you're done
// scraping. Safe to call more than once.
func (s *Scraper) Close() {
	for _, p := range s.pages {
		p.Close()
	}
	for _, b := range s.browsers {
		b.Close()
	}
	s.pages = nil
	s.browsers = nil
}
