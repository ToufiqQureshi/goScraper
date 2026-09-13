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

// Options tunes how Chrome is launched. The zero value is the safe
// default: headless, with Chrome's own sandbox left switched on.
type Options struct {
	// NoSandbox launches Chrome with --no-sandbox.
	//
	// You need this in most containers, and on Ubuntu 23.10+, which
	// block the unprivileged user namespaces Chrome's sandbox relies
	// on; without it Chrome exits at startup with "No usable sandbox!".
	//
	// It is off by default because that sandbox is a real security
	// boundary: it is what stops a malicious page that exploits a
	// browser bug from reaching the rest of the machine. Prefer fixing
	// the environment (give the container the right permissions) and
	// use this only when you can't, or when you already trust every
	// page you visit.
	NoSandbox bool
}

// New launches a headless Chromium browser. Headless is the correct
// default for production: servers have no display to show a real
// window on. ctx bounds only the launch itself; cancelling it after
// New returns has no effect (use Close for that). opts is optional;
// pass nothing for the defaults.
func New(ctx context.Context, opts ...Options) (*Browser, error) {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	flags := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		// Containers usually give /dev/shm only 64MB, which makes
		// Chrome crash on heavier pages. Writing shared memory to
		// /tmp instead costs a little speed and avoids that entirely.
		chromedp.Flag("disable-dev-shm-usage", true),
	}
	if o.NoSandbox {
		flags = append(flags, chromedp.Flag("no-sandbox", true))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), flags...)

	browserCtx, cancel := chromedp.NewContext(allocCtx)

	b := &Browser{
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		ctx:         browserCtx,
		cancel:      cancel,
	}

	// Start Chrome now so New can report launch errors. This first Run
	// must be handed the browser's own context: chromedp ties the
	// Chrome process to whatever context starts it, so running it on a
	// cancellable child would kill the browser the moment New returned.
	// The caller's ctx therefore bounds the launch from outside, tearing
	// the browser down if it never comes up.
	if err := waitForRun(ctx, func() error { return chromedp.Run(b.ctx) }); err != nil {
		b.Close()
		return nil, err
	}

	return b, nil
}

// waitForRun runs fn, giving up as soon as ctx is done. It exists for
// the two calls that establish a lifetime - launching the browser and
// opening a tab - where the work cannot simply be handed a cancellable
// context without destroying the thing it just created. fn keeps
// running in the background after a timeout; its caller tears down the
// half-built browser or tab, which is what stops it.
func waitForRun(ctx context.Context, fn func() error) error {
	done := make(chan error, 1)
	go func() { done <- fn() }()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// run executes actions against the browser's own long-lived context,
// bounded by ctx so a hung command can't block its caller forever.
// Safe to cancel: by the time anything calls this, New has already
// created the browser and its first target, so cancelling only aborts
// the command in flight.
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
// ctx. The tab stays open after ctx ends; only this first navigation
// is given up on if ctx runs out first, and then the tab is closed.
func (b *Browser) Open(ctx context.Context, url string) (*Page, error) {
	if b.ctx == nil {
		return nil, errors.New("browser is not open")
	}

	pageCtx, cancel := chromedp.NewContext(b.ctx)
	page := &Page{ctx: pageCtx, cancel: cancel}

	// Like the launch in New, this first Run creates the tab and ties
	// it to the context it is given, so it gets the page's own context
	// rather than a cancellable child - otherwise the tab would close
	// as soon as this navigation finished.
	if err := waitForRun(ctx, func() error {
		return chromedp.Run(pageCtx, chromedp.Navigate(url))
	}); err != nil {
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
