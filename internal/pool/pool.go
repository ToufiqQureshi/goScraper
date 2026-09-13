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

// DefaultRecycleAfter bounds how long an item can sit idle in the
// pool before Get replaces it, even if it still looks healthy.
const DefaultRecycleAfter = 10 * time.Minute

// DefaultSkipHealthCheckWithin avoids paying for a health check on an
// item that was handed back this recently. Only the expensive check is
// skipped: the free one (alive) runs on every Get, so an item that has
// already died is still never handed out.
const DefaultSkipHealthCheckWithin = 2 * time.Second

// Options tunes recycling behavior. A zero Options uses the defaults
// above; either field can be set independently.
type Options struct {
	RecycleAfter          time.Duration
	SkipHealthCheckWithin time.Duration
}

func (o Options) withDefaults() Options {
	if o.RecycleAfter <= 0 {
		o.RecycleAfter = DefaultRecycleAfter
	}
	if o.SkipHealthCheckWithin <= 0 {
		o.SkipHealthCheckWithin = DefaultSkipHealthCheckWithin
	}
	return o
}

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

	recycleAfter          time.Duration
	skipHealthCheckWithin time.Duration

	// alive is the free check: no I/O, so Get runs it every time.
	// healthy is the expensive one and is skipped for items used very
	// recently. Both healthy and spawn take the caller's context so a
	// Get(ctx) that ends up recycling still respects ctx's deadline
	// instead of blocking on a check or a fresh launch indefinitely.
	alive   func(T) bool
	healthy func(context.Context, T) bool
	spawn   func(context.Context) (T, error)
	discard func(T)
}

// New creates a pool of `size` items, launching them in parallel with
// spawn. alive is a free check for an item that has already died;
// healthy is the expensive one; discard releases an item's own
// resources (e.g. closing a browser). ctx bounds the initial launch
// only. opts tunes recycling; its zero value uses sensible defaults.
func New[T any](ctx context.Context, size int, opts Options, alive func(T) bool, healthy func(context.Context, T) bool, spawn func(context.Context) (T, error), discard func(T)) (*Pool[T], error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}
	opts = opts.withDefaults()

	p := &Pool[T]{
		items:                 make(chan *entry[T], size),
		recycleAfter:          opts.RecycleAfter,
		skipHealthCheckWithin: opts.SkipHealthCheckWithin,
		alive:                 alive,
		healthy:               healthy,
		spawn:                 spawn,
		discard:               discard,
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
			v, err := spawn(ctx)
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

		// Replace the item if it has sat idle too long, if it has
		// already died (free to check, so always checked - an item
		// that crashed while checked out must never be handed on),
		// or if the deeper check says it is no longer healthy.
		idle := time.Since(e.lastUsed)
		if idle > p.recycleAfter ||
			!p.alive(e.value) ||
			(idle > p.skipHealthCheckWithin && !p.healthy(ctx, e.value)) {
			p.discard(e.value)
			p.recycled.Add(1)
			v, err := p.spawn(ctx)
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
