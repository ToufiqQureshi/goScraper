// Package pagepool keeps a fixed number of browser tabs ready per
// browser, so scraping code can reuse tabs instead of opening and
// closing a new one for every page.
package pagepool

import (
	"context"
	"errors"
	"time"

	pool "github.com/ToufiqQureshi/Scraper/internal/pool"

	"github.com/ToufiqQureshi/Scraper/browser"
)

// blank is what a pooled tab sits on between uses. Get moves it to
// the real URL with Navigate, which is what makes reuse possible: no
// new tab is created on the common path.
const blank = "about:blank"

// Stats is a snapshot of pool activity, useful for production
// dashboards and alerts.
type Stats pool.Stats

// Options tunes recycling behavior. A zero Options uses sensible
// defaults (10 minutes / 2 seconds).
type Options struct {
	// RecycleAfter bounds how long a page can sit idle in the pool
	// before Get replaces it, even if it still looks healthy.
	RecycleAfter time.Duration
	// SkipHealthCheckWithin avoids a health check for a page used
	// this recently.
	SkipHealthCheckWithin time.Duration
}

// Pool hands out pages (browser tabs) and takes them back when you're
// done. It reuses a fixed number of tabs on one browser and replaces
// any that crash or go stale.
type Pool struct {
	core *pool.Pool[*browser.Page]

	// open and navigate exist so tests can fake tab creation and
	// navigation without a real browser. Production code always uses
	// browser.Open and (*browser.Page).Navigate.
	open     func(context.Context, string) (*browser.Page, error)
	navigate func(context.Context, *browser.Page, string) error
}

// New opens `size` tabs on browser, in parallel, and returns a pool
// holding them. ctx bounds only this initial launch. opts is
// optional; pass nothing for the defaults.
func New(ctx context.Context, b *browser.Browser, size int, opts ...Options) (*Pool, error) {
	if b == nil {
		return nil, &Error{Op: "new", Err: errors.New("browser cannot be nil")}
	}

	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	open := b.Open
	core, err := pool.New(
		ctx,
		size,
		pool.Options(o),
		func(ctx context.Context, p *browser.Page) bool { return p.Healthy(ctx) },
		func(ctx context.Context) (*browser.Page, error) { return open(ctx, blank) },
		func(p *browser.Page) { p.Close() },
	)
	if err != nil {
		return nil, &Error{Op: "new", Err: err}
	}

	navigate := func(ctx context.Context, p *browser.Page, url string) error { return p.Navigate(ctx, url) }
	return &Pool{core: core, open: open, navigate: navigate}, nil
}

// Get takes a free page out of the pool and navigates it to url,
// reusing the same tab. A page that crashed or has sat idle too long
// is replaced automatically. It blocks until a page is free, ctx is
// done, or the pool closes.
func (p *Pool) Get(ctx context.Context, url string) (*browser.Page, error) {
	page, err := p.core.Get(ctx)
	if err != nil {
		return nil, &Error{Op: "get", Err: err}
	}

	if err := p.navigate(ctx, page, url); err != nil {
		// The tab passed its last health check but failed to navigate
		// anyway (e.g. it crashed in between). Replace it once instead
		// of leaking this slot out of the pool or handing back a
		// broken page. This bypasses the pool's own Stats counters,
		// since it's a fallback outside the normal reuse path.
		page.Close()
		fresh, openErr := p.open(ctx, url)
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
func (p *Pool) Release(page *browser.Page) {
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
