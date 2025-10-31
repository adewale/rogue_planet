# Lessons Learned - Rogue Planet Development

**Project**: Rogue Planet v0.1.0 → v0.2.0
**Period**: October 2025
**Purpose**: Document key lessons from development to guide future work

---

## Table of Contents

1. [Architecture & Design](#architecture--design)
2. [Testing & Quality](#testing--quality)
3. [Feed Handling & HTTP](#feed-handling--http)
4. [Security](#security)
5. [Go Best Practices](#go-best-practices)
6. [Documentation & Organization](#documentation--organization)
7. [Project Management](#project-management)
8. [Historical Lessons from Planet/Venus](#historical-lessons-from-planetvenus)

---

## Architecture & Design

### 1. Static Output is King 👑

**Lesson**: Generate static HTML files, not dynamic pages
**Why**: Survives traffic spikes (HN, Reddit), no database queries on reads, can be served by any web server
**Evidence**: 20+ years of Planet/Venus success

**Implementation**:
- Generate complete HTML files that work without backend
- Use CDN/nginx/Apache for serving
- No database required for reads

### 2. Separate Concerns Cleanly 🔧

**Lesson**: Keep fetch, normalize, store, and generate as distinct phases
**Why**: Easier to debug, test, and maintain; can run phases independently

**Our Pipeline**:
```
Config → Crawler → Normaliser → Repository → Generator → HTML
```

Each component:
- Has single responsibility
- Operates independently
- Can be tested in isolation
- Communicates through database

### 3. Simple is Sustainable 🌱

**Lesson**: Fewer features = fewer bugs = longer project life
**Why**: Moonmoon's "stupidly simple" approach outlasted complex alternatives

**Our Choices**:
- No comments system
- No voting system
- No user accounts
- Just aggregation
- Single binary deployment

**Result**: 6 packages, ~6,000 lines of code, 88/100 Go standards compliance

---

## Testing & Quality

### 4. Boundary Conditions Require Explicit Testing 📏

**Discovered**: Off-by-one error in size limit check (10MB boundary)
**Issue**: Test passed for 9.9MB and 10.1MB but not exactly 10MB

**Always Test**:
- Under limit ✓
- **Exactly at limit** ← Often missed!
- Over limit ✓

**Example**:
```go
tests := []struct {
    size     int
    expected bool
}{
    {10*1024*1024 - 1, true},   // Under
    {10*1024*1024,     true},   // Exactly at (CRITICAL)
    {10*1024*1024 + 1, false},  // Over
}
```

### 5. Time-Dependent Tests Are Fragile ⏰

**Discovered**: Daring Fireball test failed as entries aged out
**Issue**: Expected 10+ recent entries, got 4 (time-based assertion)

**Solution**: Make tests time-invariant
- Test **parsing** rather than time windows
- Use smart fallbacks in implementation
- Check for "some results" not specific counts
- Save feed snapshots in testdata/

**Our Fix**:
```go
// Bad (time-dependent)
if len(recentEntries) < 10 {
    t.Error("expected 10+ recent entries")
}

// Good (time-invariant)
if len(allEntries) == 0 {
    t.Error("parsing failed")
}
```

### 6. Test Pragmatism vs Purity ⚖️

**Discovered**: Obfuscated script test was too strict
**Issue**: Demanded removal of harmless text `alert(1)` when only `<script>` tags needed removal

**Lesson**: Tests should verify **actual security properties**, not implementation details

**Fixed Test**:
```go
// Verify <script> tags removed (security)
if strings.Contains(sanitized, "<script>") {
    t.Error("script tag not removed")
}
// Don't worry about harmless text content
```

### 7. Error Type Consistency Matters 🎯

**Discovered**: URL validation failed because `url.Parse()` succeeds for invalid inputs
**Issue**: `url.Parse("not a url")` returns no error but invalid URL

**Always Validate Thoroughly**:
- Check for empty inputs explicitly
- Verify required fields exist (scheme, host)
- Don't rely solely on library behavior

**Our Validation**:
```go
if rawURL == "" {
    return ErrInvalidURL
}
parsed, err := url.Parse(rawURL)
if err != nil {
    return fmt.Errorf("%w: %v", ErrInvalidURL, err)
}
if parsed.Scheme == "" {  // CRITICAL CHECK
    return ErrInvalidURL
}
```

### 8. Test Coverage Goals Work 📊

**Target**: >75% coverage on all packages
**Achieved**: 78% average (config: 96%, crawler: 97%, normalizer: 80%, repository: 75%, generator: 54%)

**Discovery**: Generator package below target revealed missing error path tests
**Action**: Now targeting 85% for v0.2.0 with better template error tests

---

## Feed Handling & HTTP

### 9. HTTP Conditional Requests Are MANDATORY 🚨

**Lesson**: Proper ETag/Last-Modified support is non-negotiable
**Why**: Feed publishers will block/rate-limit aggregators that waste bandwidth
**Evidence**: rachelbythebay.com documented extensive problems with poorly-behaved readers

**Critical Rules**:
1. Store ETag and Last-Modified **EXACTLY** as received (don't modify!)
2. ETag values often include quotes - **quotes are part of the value**
3. Send BOTH If-None-Match AND If-Modified-Since if you have both
4. NEVER make up values
5. Update stored values on EVERY response
6. Don't hash body content - use server headers

**Our Implementation**:
```go
// Store EXACTLY as received
ETag:         resp.Header.Get("ETag"),         // Includes quotes!
LastModified: resp.Header.Get("Last-Modified"), // Exact value

// Send in next request
req.Header.Set("If-None-Match", cache.ETag)
req.Header.Set("If-Modified-Since", cache.LastModified)
```

### 10. Feeds Are Messy 🌪️

**Reality**: Real-world feeds violate specs, have encoding issues, broken HTML
**Why**: Feed publishers use varied tools and don't always validate

**Handle Gracefully**:
- Malformed XML
- Incorrect content-types
- Missing dates
- Future dates
- Broken HTML
- Character encoding lies (claims UTF-8 but uses Windows-1252)

**Our Approach**: Robust parsing + graceful degradation + error logging

### 11. Entry IDs Are Often Missing or Unstable 🆔

**Problem**: Not all feeds provide stable GUIDs/IDs
**Why**: Spec violations, CMS bugs, feed regeneration

**Fallback Strategy**:
1. Use provided GUID/ID if present and stable
2. Generate from permalink if available
3. Fall back to content hash (but warn about duplicates on edits)
4. **Never** use publish date as ID (not unique)

**Implementation**:
```go
if entry.GUID != "" {
    return entry.GUID
}
if entry.Link != "" {
    return hashLink(entry.Link)
}
return hashContent(entry.Title, entry.Content)
```

---

## Security

### 12. CVE-2009-2937: Never Trust Feed Content 🛡️

**Historical Lesson**: Planet Venus suffered XSS via malicious feed content
**Attack Vector**: `<img src="javascript:alert(1)">`

**Defense Strategy**:
1. **Sanitize on input** (when normalizing feed)
2. **Sanitize on output** (use `html/template` with auto-escape)
3. **CSP headers** in generated HTML
4. **Only allow http/https URLs** in src/href
5. **Remove ALL event handlers** (onclick, onerror, etc.)
6. **Remove dangerous tags** (`<script>`, `<object>`, `<embed>`, `<iframe>`, `<base>`)

**Our Implementation**:
```go
// bluemonday.UGCPolicy() + strict URL validation
policy := bluemonday.UGCPolicy()
policy.AllowURLSchemes("http", "https")
sanitized := policy.Sanitize(content)
```

### 13. SSRF Prevention Is Critical 🔒

**Threat**: Feed URLs pointing to localhost/internal networks
**Attack**: `http://localhost:6379/` (Redis), `http://169.254.169.254/` (AWS metadata)

**Our Defense**:
```go
// Block localhost and private IPs
if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
    return ErrPrivateIP
}
// Only allow http/https
if scheme != "http" && scheme != "https" {
    return ErrInvalidScheme
}
```

### 14. SQL Injection Prevention: Always Use Prepared Statements 💉

**Rule**: ALWAYS use parameterized queries
**NEVER** concatenate user input into SQL

**Good**:
```go
db.Exec("INSERT INTO feeds (url, title) VALUES (?, ?)", url, title)
```

**Bad** (NEVER DO THIS):
```go
db.Exec(fmt.Sprintf("INSERT INTO feeds (url, title) VALUES ('%s', '%s')", url, title))
```

**Result**: 100% of our queries use prepared statements

---

## Go Best Practices

### 15. Package Documentation Is Critical for godoc 📚

**Discovered**: Go Standards Audit revealed missing package docs
**Impact**: godoc.org won't properly document packages without them
**Priority**: CRITICAL for v0.2.0

**Required Format**:
```go
// Package crawler provides HTTP fetching with conditional request support
// for feed aggregation. It implements SSRF prevention, proper caching,
// and rate limiting for well-behaved feed readers.
package crawler
```

**Action**: Add to all pkg/ packages before v0.2.0

### 16. Error Messages Should Be Lowercase ✍️

**Go Convention**: Error messages start with lowercase, no trailing punctuation
**Our Violations**: Some capitalized errors (`log.Fatalf("Error: %v", err)`)

**Correct**:
```go
return fmt.Errorf("failed to fetch feed: %w", err)  // lowercase
```

**Incorrect**:
```go
return fmt.Errorf("Failed to fetch feed: %w", err)  // capitalized
```

### 17. io.Writer Pattern Enables Testability 🧪

**Pattern**: Accept `io.Writer` for output instead of hardcoding `os.Stdout`

**Benefits**:
- Commands testable without capturing stdout
- Can write to buffers, files, or stdout
- Clean separation of logic and output

**Our Usage**:
```go
type InitOptions struct {
    Output io.Writer  // Can be os.Stdout or bytes.Buffer for tests
}

func cmdInit(opts InitOptions) error {
    fmt.Fprintln(opts.Output, "Success!")
    return nil
}
```

### 18. Table-Driven Tests Scale Well 📋

**Pattern**: Use table-driven tests for multiple test cases

**Benefits**:
- Easy to add new cases
- Clear test coverage
- Consistent structure

**Our Usage**: 30+ table-driven test functions (165 test cases in crawler alone)

```go
tests := []struct {
    name    string
    input   string
    want    bool
}{
    {"valid http", "http://example.com", true},
    {"localhost", "http://localhost", false},
    // ... 23 more cases
}
```

### 19. Context for Cancellation 🚫

**Lesson**: Use `context.Context` for cancellable operations
**Our Usage**: All HTTP fetches accept context

**Good Pattern**:
```go
func (c *Crawler) Fetch(ctx context.Context, url string, cache FeedCache) (*FeedResponse, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    // ...
}
```

**Missing**: Goroutine cancellation in concurrent fetch (TODO for v0.2.0)

### 20. Defer for Resource Cleanup 🧹

**Pattern**: Always defer resource cleanup
**Our Usage**: 58 defer statements across codebase

**Common Uses**:
```go
defer resp.Body.Close()
defer db.Close()
defer f.Close()
defer wg.Done()
```

---

## Documentation & Organization

### 21. Documentation Sprawl Hurts Discoverability 📑

**Problem**: Started with 14 root markdown files, some overlapping
**Impact**: New contributors confused about which doc to read

**Solution**: Consolidated to 10 focused files
- Created `TESTING.md` (merged 6 test docs)
- Archived historical docs to `docs/archive/`
- Archived development notes to `docs/archive/development-notes/`

**Result**: 36% reduction in root files, clearer structure

### 22. Test Documentation Needs Consolidation 🗂️

**Problem**: 6 overlapping test documents (TEST_PLAN, TEST_FAILURES, TEST_IMPLEMENTATION_SUMMARY, etc.)
**Solution**: Created single `TESTING.md` with:
- Quick reference (how to run tests)
- Current status (375 tests, 100% pass)
- Test structure overview
- Archived historical analysis

**Lesson**: One canonical document > multiple overlapping docs

### 23. Development Notes Should Be Archived 📦

**Realization**: Internal dev notes (CONSISTENCY_REVIEW, CLEANUP_SUMMARY, etc.) clutter root
**Action**: Moved to `docs/archive/development-notes/` with explanatory README

**Benefit**: Root directory focused on user-facing docs

---

## Project Management

### 24. Version Consistency Checks Prevent Errors ✅

**Tool**: Created CONSISTENCY_CHECK_REPORT to verify version numbers
**Checks**: go.mod, main.go, Makefile, User-Agent, CHANGELOG

**Discovered**: Prevented v1.0.0 references in v0.1.0 release
**Process**: Run before every release

### 25. Research Agents Save Time 🤖

**Usage**: Deployed sub-agents for OPML spec research and Go standards audit

**OPML Research Agent**:
- Analyzed official specs (OPML 1.0/2.0)
- Identified 52 test cases
- Discovered `text` vs `title` attribute problem
- Provided real-world examples

**Go Standards Audit Agent**:
- Comprehensive codebase review
- 88/100 compliance score
- Identified critical documentation gaps
- Prioritized fixes for v0.2.0

**Lesson**: Specialized research agents provide thorough, structured analysis

### 26. Iterative Planning Works Better 🔄

**Approach**: Review and adjust plan before implementation
**Process**:
1. Create initial plan
2. Review with user (Option C)
3. Adjust based on feedback
4. Lock in scope
5. Begin implementation

**Benefit**: Catches scope creep early, aligns expectations

---

## Historical Lessons from Planet/Venus

### 27. Be a Good Netizen 🤝

**Principles**:
- Proper caching headers (ETag/Last-Modified)
- Clear User-Agent: `RoguePlanet/0.1 (+https://github.com/user/repo)`
- Respect Cache-Control and Retry-After
- Default: 1 hour between fetches (minimum 15 minutes)
- Rate limiting per domain

### 28. Concurrent Fetching with Limits ⚡

**Pattern**: Worker pool with semaphore
**Our Implementation**: 5-20 workers (configurable)

```go
sem := make(chan struct{}, concurrency)
var wg sync.WaitGroup

for i, feed := range feeds {
    wg.Add(1)
    go func(f Feed) {
        defer wg.Done()
        sem <- struct{}{}        // Acquire
        defer func() { <-sem }() // Release
        // ... fetch feed ...
    }(feed)
}
wg.Wait()
```

**Lesson**: Fast updates without overwhelming servers

### 29. Database Indexes Matter 📈

**Critical Indexes**:
- `idx_entries_published ON entries(published DESC)` - Most important!
- `idx_entries_feed_id ON entries(feed_id)`
- `idx_feeds_next_fetch ON feeds(next_fetch)`

**Impact**: Site generation queries 100x faster

### 30. SQLite is Perfect for This Use Case 💎

**Why SQLite**:
- Simple, portable, fast enough
- Single-file database
- WAL mode for better concurrency
- No server setup required

**Not Needed**: PostgreSQL/MySQL overkill for feed aggregator

### 31. Group by Date, Not by Feed 📅

**UX Lesson**: "River of news" format (chronological) more engaging
**Why**: Users want latest content, not per-feed organization

**Implementation**:
- Single chronological stream
- Newest first
- Optional grouping by day ("Today", "Yesterday")

### 32. Link to Original Post ↗️

**Attribution**: Always link to original article on source site
**Why**: Attribution, full content, comments, respect for publisher

**Implementation**: Prominent link on entry title + "read more" link

---

## Concurrency & Performance

### 33. Mutex Placement is Critical for Performance 🔒⚡

**Lesson**: ONLY protect what needs protection, not entire operations

**Problem Discovered (v0.4.0)**: Placed mutex around entire FetchFeed() call, serializing HTTP fetching, parsing, AND database writes. Result: 10x performance regression.

**What Happened**:
```go
// WRONG - Serializes everything (10x slower!)
mu.Lock()
result := feedFetcher.FetchFeed(ctx, f)  // HTTP + Parse + DB all locked
mu.Unlock()
```

**Root Cause Analysis**:
- HTTP fetching: CAN run concurrently (no shared state)
- Feed parsing: CAN run concurrently (no shared state)
- Database writes: MUST be serialized (shared resource)

**Correct Approach**:
```go
// HTTP fetch - NO LOCK (concurrent)
resp, err := f.crawler.FetchWithRetry(ctx, url, cache, maxRetries)

// Database write - WITH LOCK
f.lock()
f.repo.UpdateFeedError(feedID, err.Error())
f.unlock()

// Parse - NO LOCK (concurrent)
metadata, entries, err := f.normalizer.Parse(resp.Body, url, fetchTime)

// Database writes - WITH LOCK
f.lock()
for _, entry := range entries {
    f.repo.UpsertEntry(entry)
}
f.unlock()
```

**Impact**:
- With concurrency=10 and 50 feeds @ 2s each:
  - Expected: ~10 seconds (5 batches of 10)
  - With bug: ~100 seconds (sequential)
  - **10x slower!**

**How to Think About Mutex Placement**:
1. Identify shared resources (database, files, network connections with limits)
2. Identify operations that CAN run in parallel (HTTP, CPU work, parsing)
3. Lock ONLY around shared resource access
4. Keep critical sections as short as possible

**Testing**: See lesson #34

---

### 34. Test Concurrency with Timing Assertions ⏱️

**Lesson**: Write tests that verify concurrent execution, not just correctness

**Problem**: Unit tests can pass even when code is accidentally serialized

**What Unit Tests Don't Catch**:
```go
// These could be serialized or concurrent - test would pass either way:
✓ FetchFeed returns correct data
✓ Database is updated correctly
✓ Errors are handled properly

// Missing: Are they actually running concurrently?
```

**Solution: Timing-Based Concurrency Tests**:
```go
func TestFetchFeed_Concurrency(t *testing.T) {
    // Setup: 6 feeds, each takes 100ms
    // Expected with concurrency=3: ~200ms (2 batches)
    // If serialized: 600ms (all sequential)

    start := time.Now()
    // ... process 6 feeds with concurrency=3
    elapsed := time.Since(start)

    if elapsed > 250*time.Millisecond {
        t.Errorf("Took %v (expected ~200ms). Feeds processing serially!", elapsed)
    }

    if maxConcurrent < 2 {
        t.Errorf("Max concurrent was %d. Feeds appear to be serialized!", maxConcurrent)
    }
}
```

**Test Verifies**:
1. **Timing**: Completes in time consistent with parallelism
2. **Concurrency tracking**: Actually achieved concurrent execution
3. **Clear failure messages**: Explains WHAT is wrong (serialized) and WHY

**Our Results**:
- Test output: `✓ Processed 6 feeds in 202ms with max concurrency of 3`
- Would have failed immediately with broken code: `✗ Took 602ms, feeds processing serially`

**Pattern for Testing Mutex Protection**:
```go
func TestFetchFeed_MutexProtectsDatabase(t *testing.T) {
    // Track concurrent operations separately:
    // - HTTP fetches (should be concurrent)
    // - Database writes (should be serialized)

    // Verify:
    if maxConcurrentFetches < 2 {
        t.Error("HTTP fetching is serialized!") // Bug caught!
    }
    if maxConcurrentDBWrites > 1 {
        t.Error("Database writes not protected!") // Race condition!
    }
}
```

**When to Use This Pattern**:
- Any concurrent processing (worker pools, goroutines)
- Performance-critical code
- After refactoring that changes concurrency model
- When extracting business logic from concurrent code

---

### 35. Measure Performance, Don't Assume 📊

**Lesson**: Always benchmark after refactoring, even if "logic is the same"

**What We Assumed**:
- "Extracting business logic won't change performance"
- "Simplified code is good enough"
- "Tests pass, so it works"

**What We Missed**:
- Mutex placement completely changed concurrency model
- 10x performance regression
- Would have been caught by measurement

**Measurement Tools**:

**1. Timing Tests** (quick, in test suite):
```go
func TestFeaturePerformance(t *testing.T) {
    start := time.Now()
    processNItems(100)
    elapsed := time.Since(start)

    if elapsed > 500*time.Millisecond {
        t.Errorf("Too slow: %v (expected <500ms)", elapsed)
    }
}
```

**2. Benchmark Tests** (detailed):
```go
func BenchmarkFetchFeed(b *testing.B) {
    for i := 0; i < b.N; i++ {
        fetcher.FetchFeed(ctx, feed)
    }
}
// Run: go test -bench=. -benchmem
```

**3. Before/After Comparison**:
```bash
# Before refactoring
go test -bench=. -benchmem > before.txt

# After refactoring
go test -bench=. -benchmem > after.txt

# Compare
benchcmp before.txt after.txt
```

**What to Measure**:
- Execution time (wall clock)
- CPU time
- Memory allocations
- Concurrent operations achieved
- Database query counts

**Our Benchmarks**:
```
BenchmarkFetchFeed_Sequential    5194248    211 ns/op
BenchmarkFetchFeed_Concurrent    1887361    629 ns/op
```

3x overhead per operation acceptable for thread safety, but real-world gains from parallelism are massive.

**Lesson**: If you're changing concurrency code, MEASURE IT.

---

### 36. Race Detector is Mandatory for Concurrent Code 🏁

**Lesson**: Always run `go test -race` for code with goroutines

**What It Catches**:
- Data races (concurrent reads/writes without synchronization)
- Missing mutex protection
- Unsafe shared state access

**Our Experience**:
```bash
go test -race ./pkg/fetcher -v
```

**Found**: Mock logger not thread-safe (slices accessed without mutex)

**Fixed**:
```go
type mockLogger struct {
    mu sync.Mutex  // Added
    debugCalls []string
}

func (m *mockLogger) Debug(format string, args ...interface{}) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.debugCalls = append(m.debugCalls, format)
}
```

**When to Run**:
- Every test run in CI
- Before committing concurrent code changes
- When debugging mysterious test failures
- After refactoring with goroutines/mutexes

**Cost**: 2-10x slower test execution, but **critical** for finding bugs

**False Positives**: Very rare with Go's race detector

**Real Bugs It Finds**:
- Forgotten mutex locks
- Shared slices/maps without protection
- Variables accessed from multiple goroutines
- Channel operations on closed channels

**Integration with CI**:
```bash
# .github/workflows/test.yml
- name: Test with race detector
  run: go test -race ./...
```

---

### 37. Write Tests That Would Catch Your Specific Bug 🎯

**Lesson**: Don't just test that code works - test that specific failure modes fail

**Anti-Pattern** (what we had before):
```go
// Generic test: "Does it work?"
func TestFetchFeed_Success(t *testing.T) {
    result := fetcher.FetchFeed(ctx, feed)
    if result.Error != nil {
        t.Error("expected success")
    }
}
// This passes whether concurrent or serialized!
```

**Better** (what we added):
```go
// Specific test: "Does it run concurrently?"
func TestFetchFeed_Concurrency(t *testing.T) {
    // Would fail with serialization bug
    if elapsed > 250*time.Millisecond {
        t.Error("Feeds processing serially instead of concurrently")
    }
    if maxConcurrent < 2 {
        t.Error("Feeds appear to be serialized!")
    }
}

// Specific test: "Are database writes protected?"
func TestFetchFeed_MutexProtectsDatabase(t *testing.T) {
    // Would fail if mutex missing
    if maxConcurrentDBWrites > 1 {
        t.Error("Database writes not serialized - race condition!")
    }
}
```

**Questions Your Tests Should Answer**:
1. **Correctness**: Does it produce right result?
2. **Performance**: Does it run fast enough?
3. **Concurrency**: Does it actually run in parallel?
4. **Safety**: Are shared resources protected?
5. **Error handling**: Do error paths work?

**Test Design Process**:
1. Identify the bug you're fixing or feature you're adding
2. Write test that FAILS with the bug
3. Verify test fails
4. Fix the bug
5. Verify test passes
6. Keep the test (regression prevention)

**Our Example**:
- **Bug**: Mutex around entire operation (serialization)
- **Test**: Timing assertion + concurrency tracking
- **Result**: Test would have immediately failed with clear message
- **Value**: Prevents this bug from happening again

**Test Naming Convention**:
```go
TestFeatureName_SpecificBehavior
TestFetchFeed_Concurrency           // Tests concurrent execution
TestFetchFeed_MutexProtectsDatabase // Tests database protection
TestFetchFeed_301Redirect           // Tests redirect handling
```

Name clearly states WHAT is being tested, makes it obvious if test is missing.

---

### 38. Refactoring Checklist: Concurrency Edition ✅

**Lesson**: Refactoring concurrent code requires extra verification steps

**Standard Refactoring Checklist**:
- ✅ Tests pass
- ✅ Code compiles
- ✅ No new warnings

**Concurrent Code ALSO Requires**:
- ✅ `go test -race` passes
- ✅ Timing tests verify concurrency maintained
- ✅ Benchmark shows no regression
- ✅ Mutex placement analyzed (what needs protection?)
- ✅ Load test with multiple goroutines
- ✅ Manual verification of concurrent behavior

**Our Checklist for v0.4.0 Fix**:
1. ✅ Identified shared resources (database)
2. ✅ Identified concurrent operations (HTTP, parsing)
3. ✅ Moved mutex to protect ONLY database
4. ✅ All tests pass (528 tests)
5. ✅ Race detector clean
6. ✅ Concurrency test verifies parallel execution (202ms for 6 feeds)
7. ✅ Mutex protection test verifies database serialized
8. ✅ Benchmark shows acceptable overhead (3x per-op for thread safety)
9. ✅ Coverage improved (88.5%)

**Red Flags During Refactoring**:
- 🚩 Wrapping large code blocks in mutex "to be safe"
- 🚩 Not measuring performance before/after
- 🚩 Skipping race detector "because tests pass"
- 🚩 No timing assertions for concurrent code
- 🚩 Assuming "simpler code = same performance"

**Safe Refactoring Pattern**:
1. Write test that verifies current behavior (including performance)
2. Make change
3. Run ALL checks (tests, race detector, benchmarks)
4. If any fail, understand why before proceeding
5. Commit with clear description of what changed

**Time Investment**:
- Writing concurrency tests: 1 hour
- Running with race detector: +5 minutes per test run
- Benchmarking: 10 minutes

**ROI**:
- Caught critical 10x performance regression
- Would have been very hard to debug in production
- Tests now prevent this class of bugs permanently

---

## Summary: Key Principles

### Architecture
1. **Static output** - Fast, scalable, survivable
2. **Separate concerns** - Fetch → Normalize → Store → Generate
3. **Simple is sustainable** - Fewer features, fewer bugs

### Security
4. **Security first** - Sanitize everything, assume feeds are hostile
5. **Defense in depth** - Multiple layers of protection
6. **Never trust input** - Validate and sanitize all feed content

### Quality
7. **Test boundaries** - Especially exact limits
8. **Time-invariant tests** - Don't rely on current date
9. **Pragmatic testing** - Test properties, not implementation

### HTTP
10. **Be a good netizen** - Proper caching, identification, respect
11. **Conditional requests** - ETag/Last-Modified are mandatory
12. **Graceful degradation** - One bad feed doesn't break everything

### Go
13. **Package docs required** - For godoc
14. **Error wrapping** - Always use `%w`
15. **io.Writer pattern** - For testability
16. **Table-driven tests** - Scale well

### Process
17. **Document everything** - Future you will thank you
18. **Iterate on plans** - Review before implementing
19. **Use research agents** - For deep technical analysis
20. **Archive dev notes** - Keep root directory clean

---

## Metrics & Achievements

### v0.1.0 Success Metrics
- **Test Coverage**: 78% (target: 75%+)
- **Tests Passing**: 375/375 (100%)
- **Test Coverage by Package**:
  - config: 96.4%
  - crawler: 96.6%
  - generator: 54.3% (needs improvement)
  - normalizer: 79.8%
  - repository: 75.3%
- **Go Standards Compliance**: 88/100 (B+)
- **Security**: XSS, SSRF, SQL injection all tested and prevented
- **Documentation**: 10 focused docs (down from 16)
- **Code Quality**: gofmt 100%, go vet clean

### Lessons Applied Count
- **Architectural decisions**: 32 lessons from 20+ years of Planet history
- **Security practices**: 10 defensive layers implemented
- **Testing improvements**: 7 patterns discovered and fixed
- **Go best practices**: 15+ patterns adopted
- **Documentation improvements**: 8 organizational changes

---

## Actions for v0.2.0

Based on lessons learned:

### Critical (Must Fix)
1. ✅ Add package documentation (all pkg/ packages)
2. ✅ Increase generator test coverage (54% → 85%+)
3. ✅ Fix capitalized error messages

### High Priority (Should Add)
4. ✅ Add context cancellation to goroutines
5. ✅ Consider interfaces for major types (testability)
6. ✅ OPML import/export (apply lessons from research)

### Nice to Have
7. ✅ Refactor module name (`rogue_planet` → `rogue-planet`)
8. ✅ Add godoc examples
9. ✅ Benchmark critical paths

---

## Final Thoughts

The Rogue Planet project has been a masterclass in:
- Learning from history (20+ years of Planet/Venus)
- Applying modern Go practices
- Security-first development
- Iterative improvement
- Thorough testing

**Key Insight**: Standing on the shoulders of giants (Planet/Venus history) while applying modern best practices (Go standards, security, testing) creates robust, sustainable software.

**Most Valuable Lesson**: Simplicity and good netizen behavior trump features. A feed aggregator that respects servers and handles edge cases gracefully will outlast one with more bells and whistles.

---

*Document Created*: 2025-10-12
*For*: Rogue Planet v0.2.0 Planning
*Based On*: Development of v0.1.0 and planning for v0.2.0
*Sources*: specs/rogue-planet-spec.md, Go Standards Audit, OPML Research, Test Failure Analysis, Consistency Reviews
