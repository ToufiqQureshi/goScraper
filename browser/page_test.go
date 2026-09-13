package browser

import (
	"context"
	"testing"
)

func TestPageValidationErrors(t *testing.T) {
	p := &Page{}
	ctx := context.Background()

	if _, err := p.Text(ctx, ""); err == nil {
		t.Fatal("Text should reject empty selector")
	}

	if _, err := p.TextAll(ctx, ""); err == nil {
		t.Fatal("TextAll should reject empty selector")
	}

	if _, err := p.HTML(ctx, ""); err == nil {
		t.Fatal("HTML should reject empty selector")
	}

	if _, _, err := p.Attr(ctx, "", "href"); err == nil {
		t.Fatal("Attr should reject empty selector")
	}

	if _, _, err := p.Attr(ctx, "a", ""); err == nil {
		t.Fatal("Attr should reject empty attribute")
	}

	if err := p.Click(ctx, ""); err == nil {
		t.Fatal("Click should reject empty selector")
	}

	if err := p.Type(ctx, "", "hello"); err == nil {
		t.Fatal("Type should reject empty selector")
	}

	if err := p.Wait(ctx, ""); err == nil {
		t.Fatal("Wait should reject empty selector")
	}

	if err := p.Eval(ctx, "", nil); err == nil {
		t.Fatal("Eval should reject empty expression")
	}

	if err := p.Eval(ctx, "1 + 1", nil); err == nil {
		t.Fatal("Eval should reject nil result")
	}
}

func TestPageCloseDoesNotPanic(t *testing.T) {
	p := &Page{}
	p.Close()
}

func TestPageNavigateRejectsEmptyURL(t *testing.T) {
	p := &Page{}
	if err := p.Navigate(context.Background(), ""); err == nil {
		t.Fatal("Navigate should reject an empty url")
	}
}

func TestPageNavigateOnUnopenedPageReturnsError(t *testing.T) {
	p := &Page{}
	if err := p.Navigate(context.Background(), "https://example.com"); err == nil {
		t.Fatal("Navigate should fail on a page that was never opened")
	}
}

func TestPageHealthyOnZeroValueIsFalse(t *testing.T) {
	p := &Page{}
	if p.Healthy(context.Background()) {
		t.Fatal("a page with no running tab should not report healthy")
	}
}
