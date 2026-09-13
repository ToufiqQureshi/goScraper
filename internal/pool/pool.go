// Package pool implements a small generic reuse pool. It is internal,
// not part of goScraper's public API: it exists so the hard part of
// pooling (race-safe close, idle recycling, crash detection) is
// correct in exactly one place, shared by browserpool and pagepool
// instead of duplicated and drifting apart.
package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrClosed is returned by Get once the pool has been closed.
var ErrClosed = errors.New("pool is closed")

// ErrInvalidSize is returned by New when size is not positive.
var ErrInvalidSize = errors.New("pool size must be greater than 0")

// RecycleAfter bounds how long an item can sit idle in the pool
// before Get replaces it, even if it still looks healthy.
const RecycleAfter = 10 * time.Minute

// SkipHealthCheckWithin avoids paying for a health check on an item
// that was handed back this recently; a real failure still surfaces
// immediately through the caller's own use of the item.
const SkipHealthCheckWithin = 2 * time.Second

// Stats is a snapshot of pool activity.
type Stats struct {
	Created  int64 // items launched in total, including replacements
	Recycled int64 // items replaced for being unhealthy or idle too long
	Gets     int64
	Releases int64
}

type entry[T any] struct {
	value    T
	lastUsed time.Time
}

// Pool reuses a fixed number of items of type T, replacing any that
// go unhealthy or sit idle too long.
type Pool[T any] struct {
	items  chan *entry[T]
	mu     sync.Mutex
	closed bool

	created  atomic.Int64
	recycled atomic.Int64
	gets     atomic.Int64
	releases atomic.Int64

	healthy func(T) bool
	spawn   func() (T, error)
	discard func(T)
}

// New creates a pool of `size` items, launching them in parallel with
// spawn. healthy checks whether an item is still usable, and discard
// releases an item's own resources (e.g. closing a browser).
func New[T any](size int, healthy func(T) bool, spawn func() (T, error), discard func(T)) (*Pool[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	p := &Pool[T]{
		items:   make(chan *entry[T], size),
		healthy: healthy,
		spawn:   spawn,
		discard: discard,
	}

	type launch struct {
		value T
		err   error
	}

	launches := make(chan launch, size)
	var wg sync.WaitGroup
	for i := 0; i < size; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, err := spawn()
			launches <- launch{value: v, err: err}
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
		p.created.Add(1)
		p.items <- &entry[T]{value: l.value, lastUsed: time.Now()}
	}

	if firstErr != nil {
		p.Close()
		return nil, firstErr
	}

	return p, nil
}

// Get takes a free item out of the pool. An item that has gone
// unhealthy or sat idle past RecycleAfter is replaced automatically.
// It blocks until an item is free, ctx is done, or the pool closes.
func (p *Pool[T]) Get(ctx context.Context) (T, error) {
	var zero T

	select {
	case e, ok := <-p.items:
		if !ok {
			return zero, ErrClosed
		}

		idle := time.Since(e.lastUsed)
		if idle > RecycleAfter || (idle > SkipHealthCheckWithin && !p.healthy(e.value)) {
			p.discard(e.value)
			p.recycled.Add(1)
			v, err := p.spawn()
			if err != nil {
				return zero, err
			}
			p.created.Add(1)
			p.gets.Add(1)
			return v, nil
		}

		p.gets.Add(1)
		return e.value, nil
	case <-ctx.Done():
		return zero, ctx.Err()
	}
}

// Release gives an item back to the pool. If the pool has already
// been closed, the item is discarded instead of being kept around.
func (p *Pool[T]) Release(v T) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.discard(v)
		return
	}

	p.releases.Add(1)
	p.items <- &entry[T]{value: v, lastUsed: time.Now()}
}

// Close discards every item in the pool. Safe to call more than once.
func (p *Pool[T]) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}
	p.closed = true

	close(p.items)
	for e := range p.items {
		p.discard(e.value)
	}
}

// Stats returns a snapshot of pool activity so far.
func (p *Pool[T]) Stats() Stats {
	return Stats{
		Created:  p.created.Load(),
		Recycled: p.recycled.Load(),
		Gets:     p.gets.Load(),
		Releases: p.releases.Load(),
	}
}
