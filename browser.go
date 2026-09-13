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

// New launches a visible Chromium browser.
func New() (*Browser, error) {
	allocCtx, allocCancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", false),
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

// Close shuts down Chromium and releases all resources.
func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}
	if b.allocCancel != nil {
		b.allocCancel()
	}
}
