<p align="center"><img src="https://i.ibb.co/sdCWkC8P/file-0000000072248207bc2389779c2a0175.png" alt="goScraper Logo" width="180"></p>

# goScraper

A production-focused browser automation and web scraping library for Go.

> 🚧 Early development — API and internals may change.

## Features

- Headless Chromium by default (works on servers with no display)
- Every blocking call takes a `context.Context`, so a stuck page
  can't hang your program
- Browser pool with crash recovery, idle recycling, and stats
  (`browserpool`)
- Page (tab) pool that reuses tabs instead of opening a new one per
  page, with the same crash recovery (`pagepool`)
- Simple one-stop API combining both (`goscraper`)
- Extract text, HTML, and attributes with CSS selectors
- Click, type, wait for elements, execute JavaScript
- Designed for long-running crawls, not just one-off scrapes

## Installation

```bash
go get github.com/ToufiqQureshi/Scraper
```

## Quick Start

The simple API launches a small pool of browsers and tabs for you and
reuses them automatically:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/ToufiqQureshi/Scraper/goscraper"
)

func main() {
    ctx := context.Background()

    s, err := goscraper.New(ctx, goscraper.Config{})
    if err != nil {
        log.Fatal(err)
    }
    defer s.Close()

    err = s.Get(ctx, "https://example.com", func(page *goscraper.Page) error {
        text, err := page.Text("h1")
        if err != nil {
            return err
        }
        fmt.Println(text)
        return nil
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

Output:

```text
Example Domain
```

`goscraper.Config{}` uses sensible defaults (2 browsers, 5 tabs each).
Set `Browsers` and `PagesPerBrowser` to size the pool for your workload.

## Low-level API

For direct control over one browser and its tabs:

```go
import "github.com/ToufiqQureshi/Scraper/browser"

ctx := context.Background()

b, _ := browser.New(ctx)
defer b.Close()

page, _ := b.Open(ctx, "https://example.com")
defer page.Close()

text, _ := page.Text(ctx, "h1")
html, _ := page.HTML(ctx, ".content")
href, _, _ := page.Attr(ctx, "a", "href")

_ = page.Click(ctx, "#button")
_ = page.Type(ctx, "#search", "hello")
_ = page.Wait(ctx, ".result")
```

For pooling a fleet of browsers or tabs directly, see the
`browserpool` and `pagepool` packages — `goscraper` is a thin
convenience layer built on top of both.

## Architecture

```text
Go Application
      │
      ▼
   goscraper (simple API)
      │
      ▼
browserpool + pagepool (reuse, crash recovery, stats)
      │
      ▼
   browser (headless Chromium via chromedp)
      │
      ▼
     Website
```

An interactive version of this picture — with the request path, trust
boundaries, and links into the real source files — is at
[docs/diagrams/goscraper-architecture.html](docs/diagrams/goscraper-architecture.html)
(open it in a browser; its source spec sits beside it).

Where everything lives:

```text
goScraper/
├── browser/        Chromium itself: Browser and Page
├── browserpool/    a reusable pool of browsers
├── pagepool/       a reusable pool of tabs on one browser
├── goscraper/      the simple one-stop API (start here)
├── internal/pool/  the generic reuse pool both pools share
└── docs/           roadmap and project notes
```

`browserpool` and `pagepool` share their concurrency-critical logic
(race-safe close, idle recycling, crash detection) through that
internal generic pool, so it only needs to be correct in one place.

## Running in Docker or on Ubuntu 23.10+

Chrome's sandbox needs unprivileged user namespaces, which most
containers and Ubuntu 23.10+ block. Where they're blocked, Chrome exits
at startup with `No usable sandbox!`.

Two ways out, best first:

1. **Allow the sandbox.** On a host you control:
   `sudo sysctl -w kernel.apparmor_restrict_unprivileged_userns=0`.
   In Docker, run with `--security-opt seccomp=unconfined` or grant
   `SYS_ADMIN`. This keeps the sandbox — a real security boundary
   between a malicious page and your machine.

2. **Turn the sandbox off**, if you can't change the environment and
   you trust the pages you visit:

```go
s, err := goscraper.New(ctx, goscraper.Config{
    Browser: browser.Options{NoSandbox: true},
})
```

`browserpool.Options` and `browser.New` take the same option.

## Crash handling

Both levels recover on their own:

- A crashed **tab** is detected on checkout and replaced with a fresh
  one.
- A crashed **Chrome process** is detected by asking the browser
  itself whether it is still alive (so a single bad URL is never
  mistaken for a dead browser), then that browser and its whole tab
  pool are rebuilt and the request is retried once.

`Browser.PID()` exposes the underlying Chrome process ID for
per-process monitoring, and `Close` is verified by test to leave no
running or zombie process behind.

## Roadmap

See [docs/ROADMAP.md](docs/ROADMAP.md) for what's built and what's next —
including the HTTP-first hybrid engine that's the main thing still
missing before goScraper matches its stated goal.

## Philosophy

goScraper is being developed incrementally, one feature at a time, and
every feature is meant to be production-usable the day it ships — see
`CLAUDE.md` for the full set of development rules this project follows.

**Simple API → Correctness → Reliability → Performance → Scale**

## Responsible Use

goScraper is intended for legitimate browser automation, testing, research, and data extraction.

It does not guarantee or promise "undetectable" browsing or bypassing anti-bot or security systems.

## License

MIT
