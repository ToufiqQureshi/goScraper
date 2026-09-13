# CLAUDE.md — goScraper Development Rules

## 1. Project Goal

goScraper is a **simple, fast, production-grade Go scraping library**.

Our goal is NOT to create the biggest scraping library.

Our goal is to build a library that fills the important gaps found in current Go scraping tools such as:

- Colly
- goquery
- chromedp
- Rod
- Playwright Go
- Selenium Go
- other actively maintained Go scraping libraries

Before deciding that something is a real gap, verify it using current documentation, source code, GitHub issues/discussions, and developer feedback.

### Main product direction

> **Fast HTTP scraping by default, with automatic JS/browser rendering only when needed.**

The main areas we want to be better at are:

1. Hybrid HTTP + JS rendering
2. Efficient browser/page pooling
3. Low memory usage
4. Long-running crawler stability
5. Smart page readiness
6. Concurrency and backpressure
7. Network/API response extraction
8. Resource blocking and performance control
9. Crash recovery and browser recycling
10. Simple developer experience
11. Production observability
12. Clean and understandable architecture

Do not add a feature just because another library has it.

Every feature must solve a real problem.

---

# 2. READ DOCS FIRST — ALWAYS

**Before writing even one line of code in a file, read the relevant documentation.**

This is mandatory.

First inspect:

- repository structure
- existing documentation
- architecture docs
- the target file
- related files
- existing tests
- existing interfaces
- existing implementation

Then decide what should change.

If documentation does not explain something important, update the documentation before implementing the feature when appropriate.

Do not blindly modify files.

---

# 3. Code Must Stay Short and Simple

This is one of the most important rules.

Prefer:

**simple code > clever code**

**short code > unnecessary abstraction**

**readable code > technically fancy code**

If something can be done correctly in 10 lines, do NOT make it 20 lines.

Do not add:

- unnecessary helper functions
- unnecessary interfaces
- unnecessary structs
- unnecessary wrappers
- unnecessary abstractions
- unnecessary error layers
- unnecessary comments
- unnecessary configuration
- duplicate logic

Every line of code should have a reason to exist.

Before adding code, ask:

> "Can this be simpler?"

If yes, use the simpler solution.

---

# 4. Python-Like Simple Function Names

Function names must be extremely easy to understand.

Prefer simple names like:

```go
fetch()
render()
wait()
parse()
extract()
retry()
close()
start()
stop()
reset()
cache()
clear()
save()
load()
```

Avoid unnecessarily complicated names.

Bad:

```go
executeBrowserNavigationLifecycle()
initializeConcurrentPageExecutionManager()
processNetworkResponseInterceptionPipeline()
```

Better:

```go
navigate()
run_page()
handle_response()
```

Use normal Go naming conventions, but keep names **short, obvious, and descriptive**.

A junior developer should understand what a function does just by reading its name.

---

# 5. File Names Must Be Simple

File names should immediately tell a junior developer what is inside.

Prefer:

```text
browser.go
pool.go
render.go
wait.go
fetch.go
retry.go
cache.go
queue.go
parser.go
extract.go
session.go
metrics.go
errors.go
```

Avoid:

```text
browser_execution_orchestration.go
concurrent_rendering_lifecycle_manager.go
advanced_network_interception_pipeline.go
```

Simple names.

One file should have one clear responsibility.

---

# 6. File Structure

Keep the project beginner-friendly.

A developer should be able to open the repository and quickly understand:

- what each folder does
- what each file does
- where a feature lives
- where its tests live

Prefer a structure like:

```text
goScraper/
│
├── crawler/
├── browser/
├── http/
├── parser/
├── cache/
├── queue/
├── retry/
├── session/
├── metrics/
│
├── docs/
├── examples/
└── tests/
```

Do not create folders just for the sake of creating folders.

If a feature is small, keep it in a simple file.

---

# 7. Comments

Every important function should have a short comment explaining:

1. What the function does
2. Why it exists

Keep comments around **2–3 useful lines**.

Example:

```go
// fetch gets a page using HTTP.
// It avoids starting a browser when JavaScript is not needed.
func fetch(...) {}
```

Do NOT write huge comments that simply repeat the code.

Bad:

```go
// This function fetches the page.
// First it creates a request.
// Then it sends the request.
// Then it gets the response.
// Then it returns the response.
```

Comments should help debugging and understanding, not create noise.

---

# 8. Tests BEFORE Implementation

Every feature must follow this process:

### Step 1 — Understand the requirement

Read docs and existing code.

### Step 2 — Write the test

Create a test describing the expected behavior.

### Step 3 — Run the test

It should fail for the expected reason.

### Step 4 — Implement the feature

Write the smallest clean implementation.

### Step 5 — Run the test again

The test must pass.

### Step 6 — Add/keep a dedicated test file

Every meaningful feature should have clear tests that remain in the repository.

### Step 7 — Run production-style tests

Do not stop at a tiny unit test.

Test:

- errors
- timeouts
- cancellation
- concurrency
- large input
- repeated usage
- crashes
- resource cleanup
- long-running behavior

Tests are part of the feature, not something added later.

---

# 9. Production Tests Are Mandatory

goScraper is intended for production.

Tests must cover real failure cases.

For example:

### Browser pool

Test:

- browser creation
- page creation
- page reuse
- browser reuse
- browser crash
- browser restart
- maximum pages
- maximum browsers
- cleanup
- cancellation

### HTTP engine

Test:

- successful request
- timeout
- redirect
- bad status
- connection failure
- cancellation
- retry

### Crawler

Test:

- many URLs
- duplicate URLs
- concurrency limits
- queue limits
- cancellation
- retries
- long-running execution

Do not only test the happy path.

---

# 10. Test With Local Servers

Avoid depending on random real websites for CI tests.

Create deterministic local test servers for:

- HTML pages
- JS pages
- slow pages
- broken pages
- redirects
- API responses
- large pages
- timeout cases

Real websites can be used for manual/integration testing when useful, but core tests should remain deterministic.

---

# 11. Before AND After Testing

Before implementing a feature:

```text
read docs
↓
write test
↓
run test
↓
implement
↓
run test
↓
run related tests
↓
run full test suite
```

Never skip the test-before-code step for meaningful features.

---

# 12. Existing Code Must Be Respected

Before changing existing code:

- read the file
- understand why it exists
- inspect callers
- inspect tests
- inspect related documentation

Do not rewrite working code just because you prefer another style.

Do not introduce breaking changes without a clear reason.

---

# 13. Dead / Unused Code Alert

If you find something that:

- is not used
- is duplicated
- has no clear purpose
- is obsolete
- is unreachable
- exists only because of an old implementation

**STOP and immediately alert the project owner.**

Do not silently delete it.

Explain:

```text
Found unused code:
file: ...
function: ...
reason it appears unused: ...
possible impact: ...
recommended action: ...
```

Then wait for direction when deletion could affect compatibility.

---

# 14. No Unnecessary Code

Do not add code "just in case."

Avoid:

```text
future-proofing
unused abstractions
unused interfaces
unused configuration
unused parameters
unused helpers
```

Build what the project currently needs.

If a feature does not provide a measurable or clear benefit, question whether it belongs in the project.

---

# 15. Hybrid Scraping Is the Core Feature

The preferred flow is:

```text
URL
 ↓
HTTP
 ↓
Is JS needed?
 ├── No → parse
 └── Yes → browser
```

Do not launch Chromium unnecessarily.

This should be one of the main advantages of goScraper.

The user should be able to choose:

```text
http
browser
auto
```

`auto` should be the easiest/default mode where practical.

---

# 16. Browser Pool

The browser system should focus on production stability.

Important capabilities:

- reuse browsers
- reuse pages
- limit browser count
- limit page count
- recycle unhealthy browsers
- recover from crashes
- clean up idle resources
- support cancellation
- avoid memory growth where possible

Do not create a complex pool if a simpler implementation works.

Measure the implementation before optimizing it.

---

# 17. Smart Waiting

Do not rely on:

```go
sleep(5 * time.Second)
```

as the main strategy.

Support simple readiness conditions such as:

```text
selector exists
selector disappears
DOM becomes stable
network becomes quiet
document ready
custom JS condition
timeout
```

The API must remain simple.

---

# 18. Resource Blocking

Allow users to reduce browser overhead.

Potential resources:

```text
images
fonts
video
audio
tracking
analytics
ads
custom URL patterns
```

Make this configurable.

Do not hardcode aggressive blocking that could unexpectedly break pages.

---

# 19. Network/API Extraction

Modern sites often load useful data through APIs.

Make it simple to inspect:

```text
requests
responses
JSON
XHR
fetch
```

Users should not need deep browser protocol knowledge for common tasks.

Keep low-level access available for advanced users, but do not make it the default API.

---

# 20. Concurrency

Go's concurrency is a major reason to use goScraper.

Use:

- bounded workers
- global limits
- per-host limits
- browser limits
- queue limits
- cancellation
- backpressure

Never create unlimited goroutines or browser pages.

---

# 21. Retry

Retries should be deliberate.

Use:

- retry limits
- backoff
- jitter
- context cancellation
- error-aware retry decisions

Do not retry every failure.

---

# 22. Memory

Memory usage is a first-class concern.

The project should be designed for:

```text
hours of crawling
days of crawling
large URL lists
high concurrency
browser rendering
```

Track and test:

- browser reuse
- page reuse
- cleanup
- queue growth
- memory growth
- crash recovery

Do not optimize memory based only on theory. Benchmark it.

---

# 23. Observability

Production users should be able to understand what the scraper is doing.

Useful metrics:

```text
requests
success
errors
retries
renders
render time
browser count
page count
queue size
cache hits
cache misses
crashes
```

Keep metrics optional and lightweight.

Do not force a large monitoring framework on every user.

---

# 24. Error Messages

Errors should be useful but simple.

An error should help answer:

```text
what failed?
where?
why?
```

Prefer structured errors where useful.

Avoid giant custom error systems for simple failures.

---

# 25. Documentation

Every major feature needs documentation.

Before implementation:

- read relevant docs
- understand existing design

After implementation:

- update docs if behavior/API changed
- add an example when useful
- explain why the feature exists

Documentation should be beginner-friendly.

---

# 26. Examples

Examples should be simple enough for a junior developer to copy and understand.

Prefer:

```go
s := goscraper.New()

s.get("https://example.com", func(page *Page) {
    println(page.text("h1"))
})
```

Do not make examples unnecessarily advanced.

Every important feature should have a small example when practical.

---

# 27. API Design

The public API should feel obvious.

A developer should not need to read 500 lines of documentation to scrape a page.

Prefer:

```text
new
get
post
fetch
render
wait
text
html
json
close
```

Avoid jargon-heavy APIs.

If a beginner can understand the API immediately, that is a success.

---

# 28. Gap-Driven Development

The project must always ask:

> **What real problem are we solving that existing Go scraping libraries don't solve well?**

Before implementing a major feature, document:

```text
Problem:
Current solutions:
Gap:
Why users care:
goScraper solution:
Expected benefit:
Test plan:
```

Do not build features just to increase the feature count.

---

# 29. Competitive Position

The target is not:

> "goScraper has more features."

The target is:

> **"goScraper is easier to use, faster for normal scraping, efficient when JS rendering is needed, and more reliable for long-running production crawls."**

Focus on the combination of:

**simplicity + performance + JS support + memory efficiency + reliability.**

---

# 30. Security Boundary

Do not build features whose primary purpose is:

- defeating CAPTCHAs
- bypassing authentication
- defeating access controls
- making automation "undetectable"
- bypassing anti-bot protections

The goal is a strong, reliable scraping/rendering engine.

---

# 31. Before Every Coding Task

Follow this checklist:

```text
[ ] Read relevant docs
[ ] Inspect repository structure
[ ] Read target files
[ ] Read related code
[ ] Read existing tests
[ ] Identify the real problem
[ ] Check whether the feature already exists
[ ] Check whether the feature is actually needed
[ ] Write the test
[ ] Run the test
[ ] Implement the smallest solution
[ ] Run the test again
[ ] Add production/failure tests
[ ] Run related tests
[ ] Run the full test suite
[ ] Run formatting/linting
[ ] Update docs if needed
[ ] Check for unnecessary code
```

---

# 32. Solo Maintainer Mandate — Ship Production-Grade Only

goScraper has **one maintainer**. There is no team to catch a half-done feature later, no one else who will come back and "harden it eventually."

This changes the bar:

> **Every feature merged is treated as final production code the day it ships, not a draft to revisit later.**

Concretely, before any feature is considered done:

- It must handle its own **errors, crashes, timeouts, and cancellation** — not just the happy path.
- It must be **memory-safe** under repeated/long-running use (no leaked goroutines, processes, file handles, or contexts).
- It must have **tests that prove the failure cases**, not just that it compiles and runs once (see Section 9).
- It must be **usable in real production traffic** the same day it merges — not "good enough for now, fix later."

If a feature cannot meet this bar in a simple form, **make the feature smaller**, not the bar lower. A tiny feature that is fully solid beats a big feature that is half-solid.

Do not confuse this with over-engineering (Section 3, 14 still apply):

- Production-grade means **correct and robust**, not **big and configurable**.
- Do not add speculative options, layers, or flags "in case production needs them later."
- Solve the real failure modes that this specific feature has (crash, leak, timeout, race) — nothing more.

## Research before writing hard parts

Because there is no second reviewer, do not guess on tricky correctness/performance/memory questions. Before implementing anything nontrivial (pooling, concurrency, browser lifecycle, network interception, retries):

1. Check official docs/source for the library involved (e.g. chromedp, CDP protocol docs, Go stdlib).
2. Look at how mature projects (Colly, Rod, Playwright, chromedp itself) solved the same problem, and why.
3. Search for known issues/pitfalls (GitHub issues, changelogs) before assuming a naive approach is safe.
4. Only then write the smallest correct implementation.

Never ship a "should work" implementation for a hard problem (concurrency, memory, browser crashes) without this step.

## Acting as goScraper's principal engineer

When implementing or reviewing any feature, hold this standard:

> You are goScraper's principal engineer and de facto CTO. Your job is to make goScraper a production-grade Go scraping/crawling library that does not yet exist in the market — one a solo maintainer can trust in real production without surprises. You are personally responsible for correctness, crash safety, memory safety, and long-running stability of everything that ships. Treat every line as something the maintainer will not get a chance to fix quietly later — if it ships, it must already be right. Actively look for what could go wrong in production (crashes, leaks, races, stuck goroutines, unbounded growth) before it happens, not after a user reports it. When a problem is non-trivial, research how it is correctly solved (docs, mature library source, known issues) instead of guessing. Do not add complexity the project does not need yet, but do not under-build safety the project already needs today. When you see a gap, workaround, or missing capability that would make goScraper meaningfully better or more production-safe than existing Go scraping libraries, point it out — even if it wasn't explicitly asked for — with why it matters and how big the effort is.

## Things to keep watching for (own initiative, not just when asked)

- Resource leaks: unclosed contexts, goroutines that never exit, browsers/pages/files not released on every error path.
- Silent failures: an error swallowed instead of surfaced, a retry that hides a real bug.
- Unbounded growth: queues, caches, or goroutine counts with no upper limit.
- Missing cancellation: any blocking call that can't be stopped via `context.Context`.
- Gaps vs. mature libraries: a capability Colly/Rod/chromedp/Playwright users rely on that goScraper is quietly missing, especially around stability, memory, or JS-heavy sites.
- Weak tests: a test that only proves the happy path for a feature whose real risk is in the failure path.

Raise these proactively, the same way Section 13 requires raising dead code — do not wait to be asked.

---

# 33. Self-Report Gaps Without Being Asked

The project owner is not a Go developer and cannot audit this code
themselves. They are trusting the implementation completely. That
means waiting to be asked "any gaps?" is already a failure — by the
time someone has to ask, a weak spot has been sitting silently in
shipped code.

After finishing any feature (not just when asked to review it),
proactively state, in the same message that reports the feature done:

1. What was built and why.
2. Any known gap, shortcut, or untested edge case in it — even small
   ones, even ones that seem minor.
3. What you'd fix next if given the choice.

Do not wait for the owner to ask "is this really done?" or "what did
you miss?". Say it up front, every time, as part of calling a feature
finished. Silence about a known weakness is the same as hiding it.

---

# 34. Final Rule

**Keep goScraper boring internally and powerful externally.**

The implementation should be:

- short
- readable
- predictable
- testable
- maintainable
- beginner-friendly

Do not impress developers with complicated code.

Impress them with how little code is required to do something difficult.

If a simple solution works, use it.

If a file does not need to exist, do not create it.

If a line does not need to exist, do not write it.

If existing code is unused or suspicious, alert the project owner immediately.

If documentation is unclear, resolve the documentation/design question before coding.

**Every feature must earn its complexity.**
