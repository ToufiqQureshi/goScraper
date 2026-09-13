// Package browserpool keeps a fixed number of browsers ready to use,
// so scraping code can reuse them instead of starting a new Chrome
// process for every page.
package browserpool

import (
	"context"

	pool "github.com/ToufiqQureshi/Scraper/internal/pool"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// Stats is a snapshot of pool activity, useful for production
// dashboards and alerts.
type Stats pool.Stats

// Pool hands out browsers and takes them back when you're done.
// It reuses a fixed number of browsers and replaces any that crash
// or go stale.
type Pool struct {
	core *pool.Pool[*scraper.Browser]
}

// New starts `size` browsers, in parallel, and returns a pool holding
// them.
func New(size int) (*Pool, error) {
	core, err := pool.New(
		size,
		(*scraper.Browser).Healthy,
		scraper.New,
		func(b *scraper.Browser) { b.Close() },
	)
	if err != nil {
		return nil, &Error{Op: "new", Err: err}
	}

	return &Pool{core: core}, nil
}

// Get takes a free browser out of the pool. A browser that crashed or
// has sat idle too long is replaced automatically. It blocks until a
// browser is free, ctx is done, or the pool closes.
func (p *Pool) Get(ctx context.Context) (*scraper.Browser, error) {
	b, err := p.core.Get(ctx)
	if err != nil {
		return nil, &Error{Op: "get", Err: err}
	}
	return b, nil
}

// Release gives a browser back to the pool so someone else can use it.
// If the pool has already been closed, the browser is closed instead
// of being kept around.
func (p *Pool) Release(b *scraper.Browser) {
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
