package browser_test

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ToufiqQureshi/Scraper/browser"
)

// processState reads a Linux process's state letter from /proc. ok is
// false once the process is gone from the table entirely, which is
// what a fully cleaned-up browser looks like.
func processState(pid int) (state string, ok bool) {
	raw, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return "", false
	}

	// The line is "<pid> (<comm>) <state> ...", and comm can itself
	// contain spaces and parens, so read the state after the last ')'.
	line := string(raw)
	end := strings.LastIndex(line, ")")
	if end < 0 {
		return "", true
	}

	fields := strings.Fields(line[end+1:])
	if len(fields) == 0 {
		return "", true
	}

	return fields[0], true
}

// TestCloseLeavesNoChromeProcessBehind is what a long-running crawler
// actually depends on: every Close must reap its Chrome process, not
// leave it running and not leave a zombie. A leak here shows up as
// thousands of dead processes after a day of crawling.
func TestCloseLeavesNoChromeProcessBehind(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping: process check reads /proc, which is Linux-only")
	}

	b := requireChrome(t)

	pid := b.PID()
	if pid == 0 {
		t.Fatal("expected a real Chrome PID for a running browser")
	}
	if _, alive := processState(pid); !alive {
		t.Fatalf("Chrome process %d should be running before Close", pid)
	}

	b.Close()

	// chromedp reaps the process asynchronously, so allow a moment for
	// it, but never accept a lingering or zombie process.
	deadline := time.Now().Add(10 * time.Second)
	for {
		state, alive := processState(pid)
		if !alive {
			return // reaped, nothing left behind
		}

		if time.Now().After(deadline) {
			if state == "Z" {
				t.Fatalf("Chrome process %d is a zombie after Close: killed but never reaped", pid)
			}
			t.Fatalf("Chrome process %d is still alive (state %q) 10s after Close", pid, state)
		}

		time.Sleep(50 * time.Millisecond)
	}
}

// TestPIDIsZeroOnceClosed guards the accessor itself: a closed browser
// must not hand out a PID that may since have been reused by an
// unrelated process.
func TestPIDIsZeroOnceClosed(t *testing.T) {
	b := requireChrome(t)
	b.Close()

	if pid := b.PID(); pid != 0 {
		t.Fatalf("expected PID 0 after Close, got %d", pid)
	}
}

func TestPIDIsZeroOnZeroValueBrowser(t *testing.T) {
	var b browser.Browser
	if pid := b.PID(); pid != 0 {
		t.Fatalf("expected PID 0 for a browser that was never started, got %d", pid)
	}
}
