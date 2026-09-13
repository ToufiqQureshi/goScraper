// Package chrometest is test-only support for tests that need a real
// Chrome. It is internal: nothing outside this module uses it.
package chrometest

import (
	"context"
	"os"

	"github.com/ToufiqQureshi/Scraper/browser"
)

// RequireEnv, when set to "1", turns a failed Chrome launch into a
// test failure instead of a skip. CI sets it so the suite can never
// go green while every browser test quietly skips itself.
const RequireEnv = "GOSCRAPER_REQUIRE_CHROME"

// TB is the part of *testing.T this package needs, so that importing
// it doesn't drag the testing package into non-test builds.
type TB interface {
	Helper()
	Skipf(format string, args ...any)
	Fatalf(format string, args ...any)
}

// Require returns a running browser, or skips the test when Chrome
// isn't available here. Where GOSCRAPER_REQUIRE_CHROME=1 it fails
// instead, so a broken browser environment can't hide behind a skip.
func Require(t TB) *browser.Browser {
	t.Helper()

	b, err := browser.New(context.Background())
	if err != nil {
		if os.Getenv(RequireEnv) == "1" {
			t.Fatalf("Chrome is required here (%s=1) but failed to launch: %v", RequireEnv, err)
		}
		t.Skipf("skipping: no Chrome available to launch: %v", err)
	}

	return b
}
