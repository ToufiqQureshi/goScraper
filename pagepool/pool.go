// Package pagepool keeps a fixed number of browser tabs ready per
// browser, so scraping code can reuse tabs instead of opening and
// closing a new one for every page.
package pagepool

import (
	"context"
	"errors"

	pool "github.com/ToufiqQureshi/Scraper/internal/pool"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// blank is what a pooled tab sits on between uses. Get moves it to
// the real URL with Navigate, which is what makes reuse possible: no
// new tab is created on the common path.
const blank = "about:blank"

// Stats is a snapshot of pool activity, useful for production
// dashboards and alerts.
type Stats pool.Stats

// Pool hands out pages (browser tabs) and takes them back when you're
// done. It reuses a fixed number of tabs on one browser and replaces
// any that crash or go stale.
type Pool struct {
	core *pool.Pool[*scraper.Page]

	// open and navigate exist so tests can fake tab creation and
	// navigation without a real browser. Production code always uses
	// browser.Open and (*scraper.Page).Navigate.
	open     func(string) (*scraper.Page, error)
	navigate func(*scraper.Page, string) error
}

// New opens `size` tabs on browser, in parallel, and returns a pool
// holding them.
func New(browser *scraper.Browser, size int) (*Pool, error) {
	if browser == nil {
		return nil, &Error{Op: "new", Err: errors.New("browser cannot be nil")}
	}

	open := browser.Open
	core, err := pool.New(
		size,
		(*scraper.Page).Healthy,
		func() (*scraper.Page, error) { return open(blank) },
		func(p *scraper.Page) { p.Close() },
	)
	if err != nil {
		return nil, &Error{Op: "new", Err: err}
	}

	return &Pool{core: core, open: open, navigate: (*scraper.Page).Navigate}, nil
}

// Get takes a free page out of the pool and navigates it to url,
// reusing the same tab. A page that crashed or has sat idle too long
// is replaced automatically. It blocks until a page is free, ctx is
// done, or the pool closes.
func (p *Pool) Get(ctx context.Context, url string) (*scraper.Page, error) {
	page, err := p.core.Get(ctx)
	if err != nil {
		return nil, &Error{Op: "get", Err: err}
	}

	if err := p.navigate(page, url); err != nil {
		// The tab passed its last health check but failed to navigate
		// anyway (e.g. it crashed in between). Replace it once instead
		// of leaking this slot out of the pool or handing back a
		// broken page. This bypasses the pool's own Stats counters,
		// since it's a fallback outside the normal reuse path.
		page.Close()
		fresh, openErr := p.open(url)
		if openErr != nil {
			return nil, &Error{Op: "get", Err: openErr}
		}
		return fresh, nil
	}

	return page, nil
}

// Release gives a page back to the pool so someone else can use it.
// If the pool has already been closed, the page is closed instead of
// being kept around.
func (p *Pool) Release(page *scraper.Page) {
	if page == nil {
		return
	}
	p.core.Release(page)
}

// Close closes every page in the pool. Call it when you're done with
// this browser's tabs. Safe to call more than once.
func (p *Pool) Close() {
	p.core.Close()
}

// Stats returns a snapshot of pool activity so far.
func (p *Pool) Stats() Stats {
	return Stats(p.core.Stats())
}
