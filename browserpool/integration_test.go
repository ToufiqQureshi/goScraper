package browserpool

import (
	"context"
	"testing"

	scraper "github.com/ToufiqQureshi/Scraper"
)

// requireChrome skips the test when no real Chrome is available to
// launch (e.g. this sandbox). It still runs wherever Chrome is
// installed, including the project's CI.
func requireChrome(t *testing.T) *scraper.Browser {
	t.Helper()

	b, err := scraper.New(context.Background())
	if err != nil {
		t.Skipf("skipping: no Chrome available to launch: %v", err)
	}
	return b
}

func TestPoolReplacesACrashedBrowser(t *testing.T) {
	probe := requireChrome(t)
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
