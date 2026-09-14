package browser_test

import (
	"context"
	"testing"
	"time"

	"github.com/ToufiqQureshi/Scraper/internal/chrometest"
)

func TestPageActionRespectsContextTimeout(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	page, err := b.Open(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Open returned an error: %v", err)
	}
	defer page.Close()

	// An expression that never returns must not hang the caller
	// forever: it should be aborted the moment ctx's deadline passes.
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	start := time.Now()
	var result int
	err = page.Eval(ctx, "new Promise(() => {})", &result)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Eval should return an error when it never resolves before the deadline")
	}
	if elapsed > 2*time.Second {
		t.Fatalf("Eval should abort close to the 200ms deadline, took %v", elapsed)
	}
}

func TestPageActionSucceedsWithinDeadline(t *testing.T) {
	b := chrometest.Require(t)
	defer b.Close()

	page, err := b.Open(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Open returned an error: %v", err)
	}
	defer page.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := page.Title(ctx); err != nil {
		t.Fatalf("Title should succeed comfortably within a 10s deadline: %v", err)
	}
}
