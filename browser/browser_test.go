package browser

import (
	"context"
	"testing"
)

func TestBrowserHealthyOnZeroValueIsFalse(t *testing.T) {
	b := &Browser{}
	if b.Healthy(context.Background()) {
		t.Fatal("a browser with no running Chrome should not report healthy")
	}
}
