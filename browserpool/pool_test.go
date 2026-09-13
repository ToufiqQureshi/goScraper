package browserpool

import (
	"testing"
	"time"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// newTestPool builds a pool without launching real Chrome, so these
// tests can run anywhere (no browser binary needed).
func newTestPool(size int) *Pool {
	pool := &Pool{
		browsers: make(chan *scraper.Browser, size),
	}
	for i := 0; i < size; i++ {
		pool.browsers <- &scraper.Browser{}
	}
	return pool
}

func TestNewRejectsInvalidSize(t *testing.T) {
	if _, err := New(0); err == nil {
		t.Fatal("New(0) should return an error")
	}

	if _, err := New(-1); err == nil {
		t.Fatal("New(-1) should return an error")
	}
}

func TestGetReturnsABrowser(t *testing.T) {
	pool := newTestPool(2)
	defer pool.Close()

	b, err := pool.Get()
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if b == nil {
		t.Fatal("Get returned a nil browser")
	}
}

func TestGetBlocksWhenPoolIsEmpty(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	// Take the only browser, so the pool is now empty.
	b, err := pool.Get()
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	done := make(chan struct{})
	go func() {
		pool.Get() // should block because the pool is empty
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("Get should block when no browser is free")
	case <-time.After(100 * time.Millisecond):
		// expected: still blocked
	}

	// Give the browser back, the blocked Get should now succeed.
	pool.Release(b)

	select {
	case <-done:
		// expected: unblocked after Release
	case <-time.After(time.Second):
		t.Fatal("Get should unblock after Release")
	}
}

func TestReleasePutsBrowserBackForReuse(t *testing.T) {
	pool := newTestPool(1)
	defer pool.Close()

	first, err := pool.Get()
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	pool.Release(first)

	second, err := pool.Get()
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	if first != second {
		t.Fatal("Release should make the same browser available again")
	}
}

func TestGetAfterCloseReturnsError(t *testing.T) {
	pool := newTestPool(1)
	pool.Close()

	if _, err := pool.Get(); err == nil {
		t.Fatal("Get should return an error once the pool is closed")
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	pool := newTestPool(1)
	pool.Close()
	pool.Close() // should not panic
}

func TestReleaseAfterCloseDoesNotPanic(t *testing.T) {
	pool := newTestPool(1)
	b, _ := pool.Get()
	pool.Close()
	pool.Release(b) // should not panic or block
}
