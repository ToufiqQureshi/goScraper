// Package browserpool keeps a fixed number of browsers ready to use,
// so scraping code can reuse them instead of starting a new Chrome
// process for every page.
package browserpool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// recycleAfter bounds how long a browser can sit idle in the pool
// before Get replaces it, even if it still looks healthy. Chrome's own
// memory use tends to creep upward over time, so this keeps a
// long-running pool's memory bounded.
const recycleAfter = 10 * time.Minute

// skipHealthCheckWithin avoids paying for a health check (a real
// round trip to Chrome) on every single Get. A browser used this
// recently is extremely unlikely to have crashed in between; a real
// page-level error will still surface immediately if it did.
const skipHealthCheckWithin = 2 * time.Second

// pooledBrowser tracks when a browser was last handed back, so Get can
// decide whether to recycle it or skip its health check.
type pooledBrowser struct {
	browser  *scraper.Browser
	lastUsed time.Time
}

// Stats is a snapshot of pool activity, useful for production
// dashboards and alerts.
type Stats struct {
	Created  int64 // browsers launched in total, including replacements
	Recycled int64 // browsers replaced for crashing or sitting idle too long
	Gets     int64
	Releases int64
}

// Pool hands out browsers and takes them back when you're done.
// It reuses a fixed number of browsers and replaces any that crash
// or go stale.
type Pool struct {
	browsers chan *pooledBrowser
	mu       sync.Mutex
	closed   bool

	created  atomic.Int64
	recycled atomic.Int64
	gets     atomic.Int64
	releases atomic.Int64

	// healthy and spawn exist so tests can fake a browser's health and
	// creation without launching real Chrome. Production code always
	// uses the defaults set in New.
	healthy func(*scraper.Browser) bool
	spawn   func() (*scraper.Browser, error)
}

// New starts `size` browsers, in parallel, and returns a pool holding
// them.
func New(size int) (*Pool, error) {
	if size <= 0 {
		return nil, &Error{Op: "new", Err: ErrInvalidSize}
	}

	pool := &Pool{
		browsers: make(chan *pooledBrowser, size),
		healthy:  (*scraper.Browser).Healthy,
		spawn:    scraper.New,
	}

	type launch struct {
		browser *scraper.Browser
		err     error
	}

	launches := make(chan launch, size)
	var wg sync.WaitGroup
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b, err := pool.spawn()
			launches <- launch{browser: b, err: err}
		}()
	}
	wg.Wait()
	close(launches)

	var firstErr error
	for l := range launches {
		if l.err != nil {
			if firstErr == nil {
				firstErr = l.err
			}
			continue
		}
		pool.created.Add(1)
		pool.browsers <- &pooledBrowser{browser: l.browser, lastUsed: time.Now()}
	}

	if firstErr != nil {
		pool.Close()
		return nil, &Error{Op: "new", Err: firstErr}
	}

	return pool, nil
}

// Get takes a free browser out of the pool. A browser that crashed or
// has sat idle too long is replaced automatically. It blocks until a
// browser is free, ctx is done, or the pool closes.
func (p *Pool) Get(ctx context.Context) (*scraper.Browser, error) {
	select {
	case entry, ok := <-p.browsers:
		if !ok {
			return nil, &Error{Op: "get", Err: ErrClosed}
		}

		idle := time.Since(entry.lastUsed)

		if idle > recycleAfter || (idle > skipHealthCheckWithin && !p.healthy(entry.browser)) {
			entry.browser.Close()
			p.recycled.Add(1)
			b, err := p.spawn()
			if err != nil {
				return nil, &Error{Op: "get", Err: err}
			}
			p.created.Add(1)
			p.gets.Add(1)
			return b, nil
		}

		p.gets.Add(1)
		return entry.browser, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Release gives a browser back to the pool so someone else can use it.
// If the pool has already been closed, the browser is closed instead
// of being kept around.
func (p *Pool) Release(b *scraper.Browser) {
	if b == nil {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		b.Close()
		return
	}

	p.releases.Add(1)
	p.browsers <- &pooledBrowser{browser: b, lastUsed: time.Now()}
}

// Close shuts down every browser in the pool. Call it when you're
// done scraping. Safe to call more than once.
func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true

	close(p.browsers)
	for entry := range p.browsers {
		entry.browser.Close()
	}
}

// Stats returns a snapshot of pool activity so far.
func (p *Pool) Stats() Stats {
	return Stats{
		Created:  p.created.Load(),
		Recycled: p.recycled.Load(),
		Gets:     p.gets.Load(),
		Releases: p.releases.Load(),
	}
}
