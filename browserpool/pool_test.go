package browserpool

import (
	"context"
	"errors"
	"testing"
	"time"

	pool "github.com/ToufiqQureshi/Scraper/internal/pool"

	"github.com/ToufiqQureshi/Scraper/browser"
)

// newTestPool builds a pool without launching real Chrome, so these
// tests can run anywhere. The hard concurrency/recycling logic lives
// in internal/pool and is tested exhaustively there; these tests only
// check that this package wires it up correctly for *browser.Browser.
func newTestPool(t *testing.T, size int) *Pool {
	t.Helper()

	core, err := pool.New(
		context.Background(),
		size,
		pool.Options{},
		func(context.Context, *browser.Browser) bool { return true },
		func(context.Context) (*browser.Browser, error) { return &browser.Browser{}, nil },
		func(*browser.Browser) {},
	)
	if err != nil {
		t.Fatalf("pool.New returned an error: %v", err)
	}
	return &Pool{core: core}
}

func TestNewRejectsInvalidSize(t *testing.T) {
	if _, err := New(context.Background(), 0); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("New(0) should return ErrInvalidSize, got: %v", err)
	}
	if _, err := New(context.Background(), -1); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("New(-1) should return ErrInvalidSize, got: %v", err)
	}
}

func TestGetReturnsABrowser(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()

	b, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	if b == nil {
		t.Fatal("Get returned a nil browser")
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

func TestReleasePutsBrowserBackForReuse(t *testing.T) {
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
		t.Fatal("Release should make the same browser available again")
	}
}

func TestGetAfterCloseReturnsError(t *testing.T) {
	p := newTestPool(t, 1)
	p.Close()

	_, err := p.Get(context.Background())
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("expected errors.Is(err, ErrClosed) to be true, got: %v", err)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	p := newTestPool(t, 1)
	p.Close()
	p.Close() // should not panic
}

func TestReleaseNilDoesNothing(t *testing.T) {
	p := newTestPool(t, 1)
	defer p.Close()
	p.Release(nil) // should not panic
}

func TestReleaseAfterCloseDoesNotPanic(t *testing.T) {
	p := newTestPool(t, 1)
	b, _ := p.Get(context.Background())
	p.Close()
	p.Release(b) // should not panic or block
}

func TestStatsReflectsActivity(t *testing.T) {
	p := newTestPool(t, 2)
	defer p.Close()

	b, err := p.Get(context.Background())
	if err != nil {
		t.Fatalf("Get returned an error: %v", err)
	}
	p.Release(b)

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
