// Package browserpool keeps a fixed number of browsers ready to use,
// so scraping code can reuse them instead of starting a new Chrome
// process for every page.
package browserpool

import (
	"context"
	"time"

	pool "github.com/ToufiqQureshi/Scraper/internal/pool"

	"github.com/ToufiqQureshi/Scraper/browser"
)

// Stats is a snapshot of pool activity, useful for production
// dashboards and alerts.
type Stats pool.Stats

// Options tunes recycling behavior. A zero Options uses sensible
// defaults (10 minutes / 2 seconds).
type Options struct {
	// RecycleAfter bounds how long a browser can sit idle in the pool
	// before Get replaces it, even if it still looks healthy.
	RecycleAfter time.Duration
	// SkipHealthCheckWithin avoids a health check for a browser used
	// this recently.
	SkipHealthCheckWithin time.Duration
}

// Pool hands out browsers and takes them back when you're done.
// It reuses a fixed number of browsers and replaces any that crash
// or go stale.
type Pool struct {
	core *pool.Pool[*browser.Browser]
}

// New starts `size` browsers, in parallel, and returns a pool holding
// them. ctx bounds only this initial launch. opts is optional; pass
// nothing for the defaults.
func New(ctx context.Context, size int, opts ...Options) (*Pool, error) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	core, err := pool.New(
		ctx,
		size,
		pool.Options(o),
		func(ctx context.Context, b *browser.Browser) bool { return b.Healthy(ctx) },
		browser.New,
		func(b *browser.Browser) { b.Close() },
	)
	if err != nil {
		return nil, &Error{Op: "new", Err: err}
	}

	return &Pool{core: core}, nil
}

// Get takes a free browser out of the pool. A browser that crashed or
// has sat idle too long is replaced automatically. It blocks until a
// browser is free, ctx is done, or the pool closes.
func (p *Pool) Get(ctx context.Context) (*browser.Browser, error) {
	b, err := p.core.Get(ctx)
	if err != nil {
		return nil, &Error{Op: "get", Err: err}
	}
	return b, nil
}

// Release gives a browser back to the pool so someone else can use it.
// If the pool has already been closed, the browser is closed instead
// of being kept around.
func (p *Pool) Release(b *browser.Browser) {
	if b == nil {
		return
	}
	p.core.Release(b)
}

// Close shuts down every browser in the pool. Call it when you're
// done scraping. Safe to call more than once.
func (p *Pool) Close() {
	p.core.Close()
}

// Stats returns a snapshot of pool activity so far.
func (p *Pool) Stats() Stats {
	return Stats(p.core.Stats())
}
