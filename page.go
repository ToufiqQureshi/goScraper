package scraper

import (
	"context"
	"errors"

	"github.com/chromedp/chromedp"
)

// Page represents a browser tab.
type Page struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// Text returns the visible text of the first element matching selector.
func (p *Page) Text(selector string) (string, error) {
	if selector == "" {
		return "", errors.New("selector cannot be empty")
	}

	var text string
	err := chromedp.Run(
		p.ctx,
		chromedp.Text(selector, &text, chromedp.ByQuery),
	)
	if err != nil {
		return "", err
	}

	return text, nil
}

// Close closes this browser tab.
func (p *Page) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
