package scraper

import "testing"

func TestPageValidationErrors(t *testing.T) {
	p := &Page{}

	if _, err := p.Text(""); err == nil {
		t.Fatal("Text should reject empty selector")
	}

	if _, err := p.TextAll(""); err == nil {
		t.Fatal("TextAll should reject empty selector")
	}

	if _, err := p.HTML(""); err == nil {
		t.Fatal("HTML should reject empty selector")
	}

	if _, _, err := p.Attr("", "href"); err == nil {
		t.Fatal("Attr should reject empty selector")
	}

	if _, _, err := p.Attr("a", ""); err == nil {
		t.Fatal("Attr should reject empty attribute")
	}

	if err := p.Click(""); err == nil {
		t.Fatal("Click should reject empty selector")
	}

	if err := p.Type("", "hello"); err == nil {
		t.Fatal("Type should reject empty selector")
	}

	if err := p.Wait(""); err == nil {
		t.Fatal("Wait should reject empty selector")
	}

	if err := p.Eval("", nil); err == nil {
		t.Fatal("Eval should reject empty expression")
	}

	if err := p.Eval("1 + 1", nil); err == nil {
		t.Fatal("Eval should reject nil result")
	}
}

func TestPageCloseDoesNotPanic(t *testing.T) {
	p := &Page{}
	p.Close()
}

func TestPageNavigateRejectsEmptyURL(t *testing.T) {
	p := &Page{}
	if err := p.Navigate(""); err == nil {
		t.Fatal("Navigate should reject an empty url")
	}
}

func TestPageNavigateOnUnopenedPageReturnsError(t *testing.T) {
	p := &Page{}
	if err := p.Navigate("https://example.com"); err == nil {
		t.Fatal("Navigate should fail on a page that was never opened")
	}
}

func TestPageHealthyOnZeroValueIsFalse(t *testing.T) {
	p := &Page{}
	if p.Healthy() {
		t.Fatal("a page with no running tab should not report healthy")
	}
}
