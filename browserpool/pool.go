// Package browserpool keeps a fixed number of browsers ready to use,
// so scraping code can reuse them instead of starting a new Chrome
// process for every page.
package browserpool

import (
	"context"
	"errors"
	"sync"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// Pool hands out browsers and takes them back when you're done.
// It reuses a fixed number of browsers and replaces any that crash.
type Pool struct {
	browsers chan *scraper.Browser
	mu       sync.Mutex
	closed   bool

	// healthy and spawn exist so tests can fake a browser's health and
	// creation without launching real Chrome. Production code always
	// uses the defaults set in New.
	healthy func(*scraper.Browser) bool
	spawn   func() (*scraper.Browser, error)
}

// New starts `size` browsers and returns a pool holding them.
func New(size int) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("pool size must be greater than 0")
	}

	pool := &Pool{
		browsers: make(chan *scraper.Browser, size),
		healthy:  (*scraper.Browser).Healthy,
		spawn:    scraper.New,
	}

	for i := 0; i < size; i++ {
		b, err := scraper.New()
		if err != nil {
			pool.Close()
			return nil, err
		}
		pool.browsers <- b
	}

	return pool, nil
}

// Get takes a free browser out of the pool. If that browser crashed
// while it was idle, Get replaces it with a fresh one automatically.
// It blocks until a browser is free, ctx is done, or the pool closes.
func (p *Pool) Get(ctx context.Context) (*scraper.Browser, error) {
	select {
	case b, ok := <-p.browsers:
		if !ok {
			return nil, errors.New("pool is closed")
		}
		if p.healthy(b) {
			return b, nil
		}
		b.Close()
		return p.spawn()
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

	p.browsers <- b
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
	for b := range p.browsers {
		b.Close()
	}
}
