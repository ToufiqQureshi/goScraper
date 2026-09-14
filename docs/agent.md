# Scraper — Agent Guide

> **Superseded — kept for history only.**
>
> This was the v0.1 plan. The current rules live in `CLAUDE.md` and the
> current plan lives in `docs/ROADMAP.md`. Parts of this file now
> contradict the shipped code (it describes a *visible* browser and
> says not to add pooling yet — goScraper is headless by default and
> ships both a browser pool and a page pool). Read `CLAUDE.md` instead;
> don't follow this file.

## Why are we building this?

We want a **Go-native browser automation/scraping library** that gives developers a tiny, production-friendly API while hiding browser/CDP complexity.

The long-term direction is inspired by the simplicity of tools like Playwright/Patchright/nodriver, but we are **not trying to clone them**.

## Current v0.1 goal

Keep the first version intentionally tiny:

```go
browser, _ := scraper.New()
defer browser.Close()

page, _ := browser.Open("https://example.com")
defer page.Close()

text, _ := page.Text("h1")
fmt.Println(text)
```

### v0.1 supports

- Launch visible Chromium
- Open a URL
- Query an element with a CSS selector
- Extract its text
- Cleanly close the page/browser

## Development rule

**One feature at a time.**

Do not add proxies, stealth, concurrency, browser pools, scraping pipelines, CAPTCHA systems, or complicated abstractions until the core API is stable and tested.

## Long-term roadmap

1. Stable browser/page lifecycle
2. CSS selector + HTML/attribute extraction
3. Click, type, wait and JavaScript evaluation
4. Direct CDP transport instead of depending on a high-level automation framework
5. Multiple pages/tabs and context management
6. Concurrency and browser pooling
7. Memory/performance benchmarks
8. Production-grade errors, timeouts and cancellation
9. HTTP-first + browser-fallback extraction
10. Observability, logging and metrics
11. Documentation, examples and CI
12. Security, responsible automation and compatibility testing

## Engineering principles

- Go-first API
- Small public surface
- Context-aware cancellation
- Explicit errors
- Low memory overhead
- Benchmark before making performance claims
- Prefer simple internals over premature abstractions
- Keep browser-specific implementation behind the public API

## Important

The project must **not promise "undetectable" browsing**. Anti-bot systems change continuously. The goal is reliable browser automation and extraction, with responsible use and compatibility testing.

## Definition of success

A developer should be able to install the package and go from:

**Go code → Chromium → website → selector → data**

with only a few lines of code.
