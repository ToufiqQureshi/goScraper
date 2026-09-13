package browser

import (
	"context"
	"errors"

	"github.com/chromedp/chromedp"
)

// Browser controls a Chromium instance.
type Browser struct {
	allocCtx    context.Context
	allocCancel context.CancelFunc
	ctx         context.Context
	cancel      context.CancelFunc
}

// New launches a headless Chromium browser. Headless is the correct
// default for production: servers have no display to show a real
// window on. ctx bounds only the launch itself; cancelling it after
// New returns has no effect (use Close for that).
func New(ctx context.Context) (*Browser, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)

	browserCtx, cancel := chromedp.NewContext(allocCtx)

	b := &Browser{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		ctx:         browserCtx,
		cancel:      cancel,
	}

	// Force browser startup now so New can report launch errors,
	// bounded by the caller's ctx instead of blocking forever if
	// Chrome never comes up.
	if err := b.run(ctx); err != nil {
		b.Close()
		return nil, err
	}

	return b, nil
}

// run executes actions against the browser's own long-lived context,
// bounded by ctx so a hung command can't block its caller forever.
func (b *Browser) run(ctx context.Context, actions ...chromedp.Action) error {
	if b.ctx == nil {
		return errors.New("browser is not open")
	}

	runCtx, cancel := context.WithCancel(b.ctx)
	defer cancel()

	stop := context.AfterFunc(ctx, cancel)
	defer stop()

	return chromedp.Run(runCtx, actions...)
}

// Open creates a new browser tab and navigates it to url, bounded by
// ctx. The tab itself stays open after ctx ends; only this initial
// navigation is cancelled if ctx runs out first.
func (b *Browser) Open(ctx context.Context, url string) (*Page, error) {
	pageCtx, cancel := chromedp.NewContext(b.ctx)

	page := &Page{ctx: pageCtx, cancel: cancel}
	if err := page.run(ctx, chromedp.Navigate(url)); err != nil {
		cancel()
		return nil, err
	}

	return page, nil
}

// Healthy reports whether the browser can still respond to commands,
// bounded by ctx. A crashed or killed Chrome process fails this
// check, so callers (like a pool) know to discard it instead of
// handing out a dead browser.
func (b *Browser) Healthy(ctx context.Context) bool {
	if b.ctx == nil || b.ctx.Err() != nil {
		return false
	}

	var result int
	return b.run(ctx, chromedp.Evaluate("1", &result)) == nil
}

// PID returns the operating-system process ID of this browser's
// Chrome process, or 0 if it isn't known. Useful for collecting
// per-process memory/CPU metrics in production, and for verifying
// that Close really did clean the process up.
func (b *Browser) PID() int {
	if b.ctx == nil || b.ctx.Err() != nil {
		return 0
	}

	c := chromedp.FromContext(b.ctx)
	if c == nil || c.Browser == nil {
		return 0
	}

	process := c.Browser.Process()
	if process == nil {
		return 0
	}

	return process.Pid
}

// Close shuts down Chromium and releases all resources.
func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}
	if b.allocCancel != nil {
		b.allocCancel()
	}
}
