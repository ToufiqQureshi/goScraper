# goScraper Roadmap

This is a plain-English list of what's built, what's next, and what would
make goScraper the best production Go scraping library on the market.

Not a wishlist. Every item here solves a real problem existing Go
scraping libraries (Colly, chromedp, Rod, Playwright Go, Selenium Go)
handle badly or not at all. See `CLAUDE.md` for the rules every item
must follow before it counts as done.

**Out of scope, on purpose:** anything whose main job is defeating
CAPTCHAs, bypassing logins, or making automation "undetectable." That's
not what this project is for.

---

## Done

- [x] **Browser pool** — reuse a fixed number of Chrome instances instead
      of launching one per scrape.
- [x] **Crash recovery** — a dead browser is detected and replaced
      automatically, never handed to a caller.
- [x] **Idle recycling** — a browser sitting unused too long is replaced,
      so Chrome's own memory creep can't build up forever.
- [x] **Cancellation** — every blocking call takes a `context.Context`.
- [x] **Structured errors** — errors say what failed and why, and work
      with `errors.Is`.
- [x] **Basic stats** — counts of browsers created, recycled, and
      checked in/out.
- [x] **Headless by default** — works on a real server with no display.
- [x] **Page pooling** — reuse tabs inside each browser instead of
      opening/closing a new one per page (`pagepool`). Shares its
      core reuse/crash-recovery logic with `browserpool` via a small
      internal generic pool (`internal/pool`), so that hard
      concurrency code is correct in one place instead of two.
- [x] **Context on every blocking call** — `Page`/`Browser` methods
      (Text, Click, Navigate, Eval, ...) all take a `context.Context`
      and are aborted the moment it's done, instead of being able to
      hang forever on a stuck page.
- [x] **Configurable recycling** — `browserpool.Options` /
      `pagepool.Options` let a caller tune idle-recycle time and the
      health-check skip window for their own workload, instead of
      fixed internal constants.
- [x] **Unified simple API (`goscraper`)** — wires `browserpool` +
      `pagepool` together behind one `Get(ctx, url, func(page))` call,
      matching the "few lines of code" goal.
- [x] **Whole-browser crash recovery** — if a Chrome *process* dies,
      that slot's browser and its entire tab pool are rebuilt and the
      request is retried once. A failure is only blamed on the browser
      after asking the browser itself whether it is still alive, so a
      single bad URL never triggers a needless restart.
- [x] **Process cleanup verified** — `Browser.PID()` exposes the
      Chrome process ID, and a test asserts `Close` leaves no running
      or zombie process behind (the failure mode that turns a day-long
      crawl into thousands of dead processes).

---

## Next up

- [ ] **HTTP fetcher** — see P0 item 2 below. This starts the actual
      hybrid engine: HTTP by default, browser only when needed.

---

## P0 — the core differentiator (do these next)

- [ ] **2. HTTP fetcher** — a plain `net/http`-based fetch path with no
      browser involved. Most pages don't need JavaScript; paying for a
      full Chrome tab on every request is the single biggest reason
      Go scrapers waste memory today.
- [ ] **3. JS-required detection** — a simple, honest heuristic (empty
      body, missing expected content, known SPA shell) that decides
      whether a page needs the browser. Never claim it's perfect —
      let the user force either mode.
- [ ] **4. One switch: http / browser / auto** — the user picks the
      mode; `auto` tries HTTP first and only escalates to the browser
      pool when the heuristic says so. This combination (fast HTTP +
      pooled browser fallback) is the actual product pitch.
- [ ] **5. Resource blocking** — let users block images/fonts/video/ads
      before they download, so a rendered page costs far less bandwidth,
      memory, and time.
- [ ] **6. Retry system** — retries with backoff, jitter, a max count,
      and rules for which errors deserve a retry. No blind retry-everything.
- [ ] **7. Concurrency limits** — a global cap and a per-domain cap, so
      one bad URL list can't spawn unlimited goroutines or browser pages.

## P1 — makes goScraper meaningfully better

- [ ] **8. Smart waiting** — wait for "selector appears," "network goes
      quiet," or a custom JS condition, instead of `sleep(5 * time.Second)`.
- [ ] **9. Network/API capture** — a simple callback for reading a
      page's own XHR/fetch responses (many sites load their real data
      through an API, not the HTML).
- [ ] **10. Crawl queue** — a proper URL queue with priority, depth, and
      a bounded size, instead of a plain loop over a slice.
- [ ] **11. URL dedupe** — skip a URL (and its equivalent variants)
      that's already been crawled, with a pluggable storage backend.
- [ ] **12. Caching layer** — cache an HTTP response or rendered page for
      a configurable time, so re-crawls don't repeat work.
- [ ] **13. Session management** — cookies and headers that persist
      across requests as one logical session, without touching browser
      internals directly.
- [ ] **14. Per-domain settings** — different concurrency, delay, and
      retry rules per domain, e.g. `example.com = 10 workers`.
- [ ] **15. Real metrics** — expand today's basic counters into
      request/render rates, latency, cache hit rate, queue depth —
      still optional and dependency-free.
- [ ] **16. Long-running soak tests** — an actual multi-hour local test
      that proves memory stays flat, not just a fast proxy test.
- [ ] **17. Benchmarks** — HTTP-only vs. browser-only vs. hybrid, with
      real numbers for memory and latency, published in the repo.

## P2 — nice to have once the core is solid

- [ ] **18. Debug CLI** — `goscraper fetch <url>` / `render <url>` /
      `inspect <url>` for quick manual checks, kept as a separate binary
      from the core library.
- [ ] **19. Pluggable storage** — interfaces for where crawl results,
      cache entries, and the queue live, so the core stays dependency-free.
- [ ] **20. Examples + docs site** — a docs folder with one short,
      copy-pasteable example per feature, written for a beginner.

---

## How to read this list

Work top to bottom. A P0 item is only "done" when it meets the bar in
`CLAUDE.md` Section 32 — tested against real failure cases, not just
the happy path — before moving to the next one.
