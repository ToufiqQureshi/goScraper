package scraper

import (
	"context"

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
// window on.
func New() (*Browser, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
	)

	ctx, cancel := chromedp.NewContext(allocCtx)

	b := &Browser{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		ctx:         ctx,
		cancel:      cancel,
	}

	// Force browser startup now so New() can report launch errors.
	if err := chromedp.Run(ctx); err != nil {
		b.Close()
		return nil, err
	}

	return b, nil
}

// Open creates a new browser page and navigates to url.
func (b *Browser) Open(url string) (*Page, error) {
	ctx, cancel := chromedp.NewContext(b.ctx)

	if err := chromedp.Run(ctx, chromedp.Navigate(url)); err != nil {
		cancel()
		return nil, err
	}

	return &Page{
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

// Healthy reports whether the browser can still respond to commands.
// A crashed or killed Chrome process fails this check, so callers
// (like a pool) know to discard it instead of handing out a dead browser.
func (b *Browser) Healthy() bool {
	if b.ctx == nil || b.ctx.Err() != nil {
		return false
	}

	var result int
	return chromedp.Run(b.ctx, chromedp.Evaluate("1", &result)) == nil
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
