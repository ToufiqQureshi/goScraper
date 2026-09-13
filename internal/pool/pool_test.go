package pool

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// item is a fake pooled value used only in tests, so these tests need
// no real browser or network resource.
type item struct {
	id     int
	closed bool
}

func newTestPool(t *testing.T, size int) *Pool[*item] {
	t.Helper()

	var next atomic.Int64
	p, err := New(
		size,
		func(*item) bool { return true },
		func() (*item, error) {
			return &item{id: int(next.Add(1))}, nil
		},
		func(i *item) { i.closed = true },
	)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	return p
}

func TestNewRejectsInvalidSize(t *testing.T) {
	spawn := func() (*item, error) { return &item{}, nil }
	healthy := func(*item) bool { return true }
	discard := func(*item) {}

	if _, err := New(0, healthy, spawn, discard); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("New(0) should return ErrInvalidSize, got: %v", err)
	}
	if _, err := New(-1, healthy, spawn, discard); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("New(-1) should return ErrInvalidSize, got: %v", err)
	}
}

func TestNewReturnsSpawnErrorAndCleansUp(t *testing.T) {
	var created, discarded atomic.Int64
	spawnErr := errors.New("boom")

	_, err := New(3,
		func(*item) bool { return true },
		func() (*item, error) {
			n := created.Add(1)
			if n == 2 {
				return nil, spawnErr
			}
			return &item{id: int(n)}, nil
		},
		func(*item) { discarded.Add(1) },
	)

	if !errors.Is(err, spawnErr) {
		t.Fatalf("expected spawn error to propagate, got: %v", err)
	}
	// The items that did launch successfully must still be discarded,
	// not leaked.
	if discarded.Load() == 0 {
		t.Fatal("New should discard successfully-launched items when one launch fails")
	}
}

func TestGetReturnsAnItem(t *testing.T) {
	p := newTestPool(t, 2)
	defer p.Close()

	v, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if v == nil {
		t.Fatal("Get returned a nil item")
	}
}

func TestGetBlocksWhenPoolIsEmpty(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	v, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		p.Get(context.Background())
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Get should block when no item is free")
	case <-time.After(100 * time.Millisecond):
	}

	p.Release(v)

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Get should unblock after Release")
	}
}

func TestGetReturnsErrorWhenContextCancelled(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	if _, err := p.Get(context.Background()); err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	if _, err := p.Get(ctx); err == nil {
		t.Fatal("Get should return an error when its context is cancelled")
	}
}

func TestReleasePutsItemBackForReuse(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	first, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	p.Release(first)

	second, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	if first != second {
		t.Fatal("Release should make the same item available again")
	}
}

func TestGetAfterCloseReturnsErrClosed(t *testing.T) {
	p := newTestPool(t, 1)
	p.Close()

	if _, err := p.Get(context.Background()); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got: %v", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	p := newTestPool(t, 1)
	p.Close()
	p.Close()
}

func TestCloseDiscardsEveryItem(t *testing.T) {
	p := newTestPool(t, 3)
	p.Close()
	// nothing left to inspect directly, but Close must not panic or
	// hang, and a subsequent Get must fail cleanly.
	if _, err := p.Get(context.Background()); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed after Close, got: %v", err)
	}
}

func TestReleaseAfterCloseDiscardsItem(t *testing.T) {
	p := newTestPool(t, 1)
	v, _ := p.Get(context.Background())
	p.Close()

	p.Release(v) // should not panic or block

	if !v.closed {
		t.Fatal("Release after Close should discard the item")
	}
}

func TestConcurrentReleaseAndCloseDoNotRace(t *testing.T) {
	p := newTestPool(t, 4)

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		v, err := p.Get(context.Background())
		if err != nil {
			t.Fatalf("Get returned an error: %v", err)
		}
		wg.Add(1)
		go func(v *item) {
			defer wg.Done()
			p.Release(v)
		}(v)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		p.Close()
	}()

	wg.Wait()
}

func TestGetRecyclesUnhealthyItem(t *testing.T) {
	stale := &item{id: 1}
	spawned := &item{id: 2}

	p := &Pool[*item]{
		items:   make(chan *entry[*item], 1),
		healthy: func(*item) bool { return false },
		spawn:   func() (*item, error) { return spawned, nil },
		discard: func(i *item) { i.closed = true },
	}
	p.items <- &entry[*item]{value: stale, lastUsed: time.Now().Add(-SkipHealthCheckWithin - time.Second)}

	v, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer p.Close()

	if v != spawned {
		t.Fatal("Get should replace an unhealthy item")
	}
	if !stale.closed {
		t.Fatal("Get should discard the unhealthy item it replaced")
	}
	if p.Stats().Recycled != 1 {
		t.Fatalf("expected 1 recycled item, got %d", p.Stats().Recycled)
	}
}

func TestGetRecyclesItemIdleTooLong(t *testing.T) {
	stale := &item{id: 1}
	spawned := &item{id: 2}

	p := &Pool[*item]{
		items:   make(chan *entry[*item], 1),
		healthy: func(*item) bool { return true }, // healthy, but too old
		spawn:   func() (*item, error) { return spawned, nil },
		discard: func(i *item) { i.closed = true },
	}
	p.items <- &entry[*item]{value: stale, lastUsed: time.Now().Add(-RecycleAfter - time.Minute)}

	v, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer p.Close()

	if v != spawned {
		t.Fatal("Get should replace an item idle past RecycleAfter")
	}
}

func TestGetSkipsHealthCheckForRecentlyUsedItem(t *testing.T) {
	recent := &item{id: 1}

	p := &Pool[*item]{
		items:   make(chan *entry[*item], 1),
		healthy: func(*item) bool { return false }, // would fail if checked
		spawn:   func() (*item, error) { return &item{id: 2}, nil },
		discard: func(i *item) { i.closed = true },
	}
	p.items <- &entry[*item]{value: recent, lastUsed: time.Now()}

	v, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer p.Close()

	if v != recent {
		t.Fatal("Get should skip the health check for an item used moments ago")
	}
}

func TestStatsTracksActivity(t *testing.T) {
	p := newTestPool(t, 2)
	defer p.Close()

	v, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	p.Release(v)

	stats := p.Stats()
	if stats.Created != 2 {
		t.Fatalf("expected Created=2, got %d", stats.Created)
	}
	if stats.Gets != 1 {
		t.Fatalf("expected Gets=1, got %d", stats.Gets)
	}
	if stats.Releases != 1 {
		t.Fatalf("expected Releases=1, got %d", stats.Releases)
	}
}

// TestManyGetReleaseCyclesDoNotLeakGoroutines is a fast stand-in for a
// long-running crawl: it hammers Get/Release and checks the goroutine
// count doesn't creep up.
func TestManyGetReleaseCyclesDoNotLeakGoroutines(t *testing.T) {
	p := newTestPool(t, 4)
	defer p.Close()

	for i := 0; i < 500; i++ {
		v, err := p.Get(context.Background())
		if err != nil {
			t.Fatalf("Get returned an error on iteration %d: %v", i, err)
		}
		p.Release(v)
	}

	runtime.GC()
	before := runtime.NumGoroutine()

	for i := 0; i < 500; i++ {
		v, err := p.Get(context.Background())
		if err != nil {
			t.Fatalf("Get returned an error on iteration %d: %v", i, err)
		}
		p.Release(v)
	}

	runtime.GC()
	after := runtime.NumGoroutine()

	if after > before+2 {
		t.Fatalf("goroutine count grew from %d to %d after 500 Get/Release cycles", before, after)
	}
}
