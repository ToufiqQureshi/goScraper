// Package browserpool keeps a fixed number of browsers ready to use,
// so scraping code can reuse them instead of starting a new Chrome
// process for every page.
package browserpool

import (
	"errors"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// Pool hands out browsers and takes them back when you're done.
type Pool struct {
	browsers chan *scraper.Browser
	closed   bool
}

// New starts `size` browsers and returns a pool holding them.
func New(size int) (*Pool, error) {
	if size <= 0 {
		return nil, errors.New("pool size must be greater than 0")
	}

	pool := &Pool{
		browsers: make(chan *scraper.Browser, size),
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

// Get takes a free browser out of the pool. It blocks if every
// browser is currently in use.
func (p *Pool) Get() (*scraper.Browser, error) {
	if p.closed {
		return nil, errors.New("pool is closed")
	}

	b, ok := <-p.browsers
	if !ok {
		return nil, errors.New("pool is closed")
	}

	return b, nil
}

// Release gives a browser back to the pool so someone else can use it.
func (p *Pool) Release(b *scraper.Browser) {
	if p.closed || b == nil {
		return
	}

	p.browsers <- b
}

// Close shuts down every browser in the pool. Call it when you're
// done scraping.
func (p *Pool) Close() {
	if p.closed {
		return
	}
	p.closed = true

	close(p.browsers)
	for b := range p.browsers {
		b.Close()
	}
}
