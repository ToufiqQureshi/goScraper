package pagepool

import (
	"context"
	"testing"
	"time"

	"github.com/ToufiqQureshi/Scraper/internal/chrometest"
)

func TestPoolReusesTheSameTabAcrossPages(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	ctx := context.Background()
	pool, err := New(ctx, b, 1)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	first, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	title, err := first.Title(ctx)
	if err != nil {
		t.Fatalf("Title returned an error: %v", err)
	}
	if title == "" {
		t.Fatal("expected a non-empty title after navigating to example.com")
	}
	pool.Release(first)

	second, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(second)

	if first != second {
		t.Fatal("Get should hand back the same tab instead of opening a new one")
	}
}

func TestOptionsRecycleAfterIsRespected(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	ctx := context.Background()

	// A short RecycleAfter must actually reach the pool: a tab idle
	// past it is replaced even though it is perfectly healthy.
	pool, err := New(ctx, b, 1, Options{RecycleAfter: 50 * time.Millisecond})
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	first, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	pool.Release(first)

	time.Sleep(100 * time.Millisecond)

	second, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(second)

	if second == first {
		t.Fatal("a tab idle past RecycleAfter should have been replaced")
	}
	if got := pool.Stats().Recycled; got != 1 {
		t.Fatalf("expected 1 recycled page, got %d", got)
	}
}

func TestPoolReplacesACrashedPage(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	ctx := context.Background()
	pool, err := New(ctx, b, 1)
	if err != nil {
		t.Fatalf("New returned an error: %v", err)
	}
	defer pool.Close()

	page, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}

	// Simulate a crash: the tab dies while checked out, then the
	// caller (not knowing that) gives it back to the pool.
	page.Close()
	pool.Release(page)

	replacement, err := pool.Get(ctx, "https://example.com")
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	defer pool.Release(replacement)

	if replacement == page {
		t.Fatal("Get should not hand out a crashed page")
	}
	if !replacement.Healthy(ctx) {
		t.Fatal("Get should replace a crashed page with a healthy one")
	}
}
