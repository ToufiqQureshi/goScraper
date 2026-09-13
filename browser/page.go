package browser

import (
	"context"
	"errors"
	"fmt"

	"github.com/chromedp/chromedp"
)

// Page represents a browser tab.
type Page struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// run executes actions against the tab's own long-lived context,
// bounded by ctx so a hung command (e.g. a page that never finishes
// loading) can't block its caller forever.
func (p *Page) run(ctx context.Context, actions ...chromedp.Action) error {
	if p.ctx == nil {
		return errors.New("page is not open")
	}

	runCtx, cancel := context.WithCancel(p.ctx)
	defer cancel()

	stop := context.AfterFunc(ctx, cancel)
	defer stop()

	return chromedp.Run(runCtx, actions...)
}

// Text returns the visible text of the first element matching selector.
func (p *Page) Text(ctx context.Context, selector string) (string, error) {
	if selector == "" {
		return "", errors.New("selector cannot be empty")
	}

	var text string
	err := p.run(ctx, chromedp.Text(selector, &text, chromedp.ByQuery))
	if err != nil {
		return "", err
	}

	return text, nil
}

// TextAll returns the visible text of every element matching selector.
func (p *Page) TextAll(ctx context.Context, selector string) ([]string, error) {
	if selector == "" {
		return nil, errors.New("selector cannot be empty")
	}

	var texts []string
	expression := fmt.Sprintf(
		"Array.from(document.querySelectorAll(%q)).map((el) => (el.innerText || '').trim())",
		selector,
	)

	if err := p.run(ctx, chromedp.EvaluateAsDevTools(expression, &texts)); err != nil {
		return nil, err
	}

	return texts, nil
}

// Title returns the current page title.
func (p *Page) Title(ctx context.Context) (string, error) {
	var title string
	if err := p.run(ctx, chromedp.Title(&title)); err != nil {
		return "", err
	}

	return title, nil
}

// HTML returns the outer HTML of the first element matching selector.
func (p *Page) HTML(ctx context.Context, selector string) (string, error) {
	if selector == "" {
		return "", errors.New("selector cannot be empty")
	}

	var html string
	if err := p.run(ctx, chromedp.OuterHTML(selector, &html, chromedp.ByQuery)); err != nil {
		return "", err
	}

	return html, nil
}

// Attr returns the named attribute value and whether it was found.
func (p *Page) Attr(ctx context.Context, selector string, attribute string) (string, bool, error) {
	if selector == "" {
		return "", false, errors.New("selector cannot be empty")
	}
	if attribute == "" {
		return "", false, errors.New("attribute cannot be empty")
	}

	var value string
	var ok bool
	if err := p.run(ctx, chromedp.AttributeValue(selector, attribute, &value, &ok, chromedp.ByQuery)); err != nil {
		return "", false, err
	}

	return value, ok, nil
}

// Click triggers a mouse click on the first element matching selector.
func (p *Page) Click(ctx context.Context, selector string) error {
	if selector == "" {
		return errors.New("selector cannot be empty")
	}

	return p.run(ctx, chromedp.Click(selector, chromedp.ByQuery))
}

// Type sends keyboard input to the first element matching selector.
func (p *Page) Type(ctx context.Context, selector string, value string) error {
	if selector == "" {
		return errors.New("selector cannot be empty")
	}

	return p.run(ctx, chromedp.SendKeys(selector, value, chromedp.ByQuery))
}

// Wait waits until the first element matching selector is visible.
func (p *Page) Wait(ctx context.Context, selector string) error {
	if selector == "" {
		return errors.New("selector cannot be empty")
	}

	return p.run(ctx, chromedp.WaitVisible(selector, chromedp.ByQuery))
}

// Eval runs a JavaScript expression in the page context and stores the result.
func (p *Page) Eval(ctx context.Context, expression string, result any) error {
	if expression == "" {
		return errors.New("expression cannot be empty")
	}
	if result == nil {
		return errors.New("result cannot be nil")
	}

	return p.run(ctx, chromedp.Evaluate(expression, result))
}

// Navigate loads a new URL in this same tab. This is what makes page
// reuse possible: a pool can hand the same tab out again and again by
// calling Navigate instead of opening a new one each time.
func (p *Page) Navigate(ctx context.Context, url string) error {
	if url == "" {
		return errors.New("url cannot be empty")
	}

	return p.run(ctx, chromedp.Navigate(url))
}

// Healthy reports whether the tab can still respond to commands,
// bounded by ctx. A crashed browser or a killed tab fails this check,
// so callers (like a pool) know to discard it instead of handing out
// a dead page.
func (p *Page) Healthy(ctx context.Context) bool {
	if p.ctx == nil || p.ctx.Err() != nil {
		return false
	}

	var result int
	return p.run(ctx, chromedp.Evaluate("1", &result)) == nil
}

// Close closes this browser tab.
func (p *Page) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
