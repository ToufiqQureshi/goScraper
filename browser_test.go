package scraper

import "testing"

func TestBrowserHealthyOnZeroValueIsFalse(t *testing.T) {
	b := &Browser{}
	if b.Healthy() {
		t.Fatal("a browser with no running Chrome should not report healthy")
	}
}
