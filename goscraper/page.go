package goscraper

import (
	"context"

	"github.com/ToufiqQureshi/Scraper/browser"
)

// Page is a browser.Page bound to one request's context, so callbacks
// passed to Scraper.Get can call page.Text("h1") without having to
// thread ctx through every call themselves. Use Raw() to reach the
// context-aware methods directly (e.g. to pass a shorter per-call
// timeout than the request's own).
type Page struct {
	page *browser.Page
	ctx  context.Context
}

// Raw returns the underlying page and the context Get was called
// with, for callers who need explicit control over context per call.
func (p *Page) Raw() (*browser.Page, context.Context) {
	return p.page, p.ctx
}

// Text returns the visible text of the first element matching selector.
func (p *Page) Text(selector string) (string, error) {
	return p.page.Text(p.ctx, selector)
}

// TextAll returns the visible text of every element matching selector.
func (p *Page) TextAll(selector string) ([]string, error) {
	return p.page.TextAll(p.ctx, selector)
}

// Title returns the current page title.
func (p *Page) Title() (string, error) {
	return p.page.Title(p.ctx)
}

// HTML returns the outer HTML of the first element matching selector.
func (p *Page) HTML(selector string) (string, error) {
	return p.page.HTML(p.ctx, selector)
}

// Attr returns the named attribute value and whether it was found.
func (p *Page) Attr(selector string, attribute string) (string, bool, error) {
	return p.page.Attr(p.ctx, selector, attribute)
}

// Click triggers a mouse click on the first element matching selector.
func (p *Page) Click(selector string) error {
	return p.page.Click(p.ctx, selector)
}

// Type sends keyboard input to the first element matching selector.
func (p *Page) Type(selector string, value string) error {
	return p.page.Type(p.ctx, selector, value)
}

// Wait waits until the first element matching selector is visible.
func (p *Page) Wait(selector string) error {
	return p.page.Wait(p.ctx, selector)
}

// Eval runs a JavaScript expression in the page context and stores the result.
func (p *Page) Eval(expression string, result any) error {
	return p.page.Eval(p.ctx, expression, result)
}
