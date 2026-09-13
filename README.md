# goScraper

A lightweight browser automation and web scraping library for Go.

> 🚧 Early development — API and internals may change.

## Features

- Launch Chromium
- Navigate to URLs
- Extract text using CSS selectors
- Extract HTML
- Read element attributes
- Click elements
- Type into inputs
- Wait for elements
- Execute JavaScript
- Context-aware browser lifecycle
- Designed with performance and low resource usage in mind

## Installation

```bash
go get github.com/ToufiqQureshi/goScraper
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    scraper "github.com/ToufiqQureshi/goScraper"
)

func main() {
    browser, err := scraper.New()
    if err != nil {
        log.Fatal(err)
    }
    defer browser.Close()

    page, err := browser.Open("https://example.com")
    if err != nil {
        log.Fatal(err)
    }
    defer page.Close()

    text, err := page.Text("h1")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(text)
}
```

Output:

```text
Example Domain
```

## Basic API

```go
browser, _ := scraper.New()
defer browser.Close()

page, _ := browser.Open("https://example.com")
defer page.Close()

text, _ := page.Text("h1")

html, _ := page.HTML(".content")

href, _ := page.Attr("a", "href")

_ = page.Click("#button")

_ = page.Type("#search", "hello")

_ = page.Wait(".result")
```

## Architecture

```text
Go Application
      │
      ▼
   goScraper
      │
      ▼
 Chromium / Browser
      │
      ▼
   Website
```

goScraper provides a simple Go API while hiding the browser-control implementation from the developer.

## Roadmap

- [x] Chromium launch
- [x] Page navigation
- [x] CSS selector text extraction
- [x] HTML extraction
- [x] Attribute extraction
- [x] Basic interactions
- [ ] Improve reliability
- [ ] Improve error handling
- [ ] Performance benchmarks
- [ ] Memory benchmarks
- [ ] Better browser/page lifecycle
- [ ] Direct CDP-based implementation
- [ ] Concurrency and browser pooling
- [ ] Production-ready API

## Philosophy

goScraper is being developed incrementally.

**Simple API → Correctness → Reliability → Performance → Scale**

We focus on making existing functionality stable and well-tested before adding new features.

## Responsible Use

goScraper is intended for legitimate browser automation, testing, research, and data extraction.

It does not guarantee or promise "undetectable" browsing or bypassing anti-bot or security systems.

## License

MIT
