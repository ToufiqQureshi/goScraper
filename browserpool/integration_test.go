package browserpool

import (
	"context"
	"testing"
	"time"

	"github.com/ToufiqQureshi/Scraper/internal/chrometest"
)

func TestOptionsRecycleAfterIsRespected(t *testing.T) {
	probe := chrometest.Require(t)
	probe.Close()

	// A short RecycleAfter must actually reach the pool: a browser
	// idle past it is replaced even though it is perfectly healthy.
	pool, err := New(context.Background(), 1, Options{RecycleAfter: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	first, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	pool.Release(first)

	time.Sleep(100 * time.Millisecond)

	second, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(second)

	if second == first {
		t.Fatal("a browser idle past RecycleAfter should have been replaced")
	}
	if got := pool.Stats().Recycled; got != 1 {
		t.Fatalf("expected 1 recycled browser, got %d", got)
	}
}

func TestDefaultOptionsDoNotRecycleImmediately(t *testing.T) {
	probe := chrometest.Require(t)
	probe.Close()

	// With defaults, a browser handed straight back must be reused,
	// not thrown away - the zero Options must not mean "recycle now".
	pool, err := New(context.Background(), 1)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	first, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	pool.Release(first)

	second, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(second)

	if second != first {
		t.Fatal("with default options the same browser should be reused")
	}
}

func TestPoolReplacesACrashedBrowser(t *testing.T) {
	probe := chrometest.Require(t)
	probe.Close()

	pool, err := New(context.Background(), 1)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	b, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	// Simulate a crash: the browser dies while it's checked out, then
	// the caller (not knowing that) gives it back to the pool.
	b.Close()
	pool.Release(b)

	replacement, err := pool.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(replacement)

	if replacement == b {
		t.Fatal("Get should not hand out a crashed browser")
	}
	if !replacement.Healthy(context.Background()) {
		t.Fatal("Get should replace a crashed browser with a healthy one")
	}
}
