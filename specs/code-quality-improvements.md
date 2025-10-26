# Rogue Planet - Code Quality Improvement Plan

**Status**: Recommendations
**Priority**: Post v1.0.0 (technical debt reduction)

This document outlines code quality improvements to increase maintainability, testability, and robustness. These are **not bugs** but opportunities to improve the codebase architecture.

---

## Executive Summary

**Current State**: B+ quality code
- ✅ Good: Error wrapping, resource cleanup, no library logging, security-first
- ⚠️  Needs Work: Test coverage (26.6% cmd/rp), no interfaces, large functions, global state

**Target State**: A quality code
- Testable architecture with interfaces
- >60% coverage on cmd/rp
- Separated concerns (business logic vs orchestration)
- Context propagation
- Structured logging

---

## Priority 1: Architecture & Testability 🏗️

### 1.1 Introduce Interfaces for Core Dependencies

**Problem**: All code depends on concrete types (crawler.Crawler, repository.Repository, etc.). This makes testing harder and couples components tightly.

**Current**:
```go
func fetchFeeds(cfg *config.Config) error {
    repo, err := repository.New(cfg.Database.Path)  // Concrete type
    c := crawler.NewWithUserAgent(cfg.Planet.UserAgent)  // Concrete type
    n := normalizer.New()  // Concrete type
    // ...
}
```

**Proposed**: Define interfaces in pkg packages:

```go
// pkg/repository/interface.go
type FeedRepository interface {
    GetFeeds(activeOnly bool) ([]Feed, error)
    AddFeed(url, title string) (int64, error)
    UpdateFeedCache(id int64, etag, lastModified string, lastFetched time.Time) error
    UpsertEntry(entry *Entry) error
    // ... other methods
}

// pkg/crawler/interface.go
type FeedCrawler interface {
    Fetch(ctx context.Context, feedURL string, cache FeedCache) (*FeedResponse, error)
}

// pkg/normalizer/interface.go
type FeedNormalizer interface {
    Parse(feedData []byte, feedURL string, fetchTime time.Time) (*FeedMetadata, []Entry, error)
}
```

**Benefits**:
- Easy to mock for testing
- Dependency injection possible
- Can swap implementations (e.g., test database)
- Clear contracts between components

**Estimate**: 2 days
**Files**: Create `interface.go` in each pkg/

---

### 1.2 Extract Business Logic from fetchFeeds()

**Problem**: `fetchFeeds()` in commands.go is 143 lines mixing orchestration, concurrency, and business logic. Hard to test individual pieces.

**Current Structure** (817-919):
```go
func fetchFeeds(cfg *config.Config) error {
    // Setup (20 lines)
    // Concurrency control (10 lines)
    // Goroutine with fetch logic (90 lines)
    // Wait and cleanup (5 lines)
}
```

**Proposed**: Extract to pkg/fetcher with interfaces

```go
// pkg/fetcher/fetcher.go
type Fetcher struct {
    repo       repository.FeedRepository
    crawler    crawler.FeedCrawler
    normalizer normalizer.FeedNormalizer
    logger     Logger
}

func (f *Fetcher) FetchFeed(ctx context.Context, feed repository.Feed) error {
    // Single feed fetch logic (50 lines)
    // No concurrency here - that's orchestration
}

func (f *Fetcher) FetchAllFeeds(ctx context.Context, concurrency int) error {
    // Orchestrate concurrent fetches (40 lines)
    // Uses semaphore pattern
}
```

**Benefits**:
- Business logic testable without goroutines
- Clear separation: fetcher package = logic, commands = CLI
- Can test single feed fetch independently
- Context properly propagated

**Estimate**: 3 days
**Files**: Create `pkg/fetcher/`, refactor `cmd/rp/commands.go`

---

### 1.3 Improve Test Coverage for cmd/rp

**Problem**: Only 26.6% coverage vs 80-96% in library packages.

**Current Coverage**:
```
cmd/rp:          26.6%  ⚠️
pkg/crawler:     96.6%  ✅
pkg/config:      93.8%  ✅
pkg/opml:        91.8%  ✅
```

**Missing Tests**:
- [ ] Command execution paths (init, add-feed, etc.)
- [ ] Error handling in commands
- [ ] Config validation flows
- [ ] Output formatting

**Proposed Tests**:
```go
// cmd/rp/commands_test.go
func TestCmdInit_Success(t *testing.T) { ... }
func TestCmdInit_ConfigExists(t *testing.T) { ... }
func TestCmdAddFeed_InvalidURL(t *testing.T) { ... }
func TestCmdAddFeed_DatabaseError(t *testing.T) { ... }
func TestCmdListFeeds_EmptyDatabase(t *testing.T) { ... }
func TestCmdStatus_OutputFormat(t *testing.T) { ... }
// ... 20-30 more tests
```

**Target**: >60% coverage (currently 26.6%)

**Estimate**: 2 days
**Files**: Expand `cmd/rp/*_test.go`

---

## Priority 2: Context & Concurrency 🔄

### 2.1 Proper Context Propagation

**Problem**: Creating `context.Background()` inside goroutines instead of passing parent context from command.

**Current** (commands.go:838):
```go
go func(index int, f repository.Feed) {
    defer wg.Done()
    // ...
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    resp, err := c.Fetch(ctx, f.URL, cache)
    cancel()
    // ...
}(i, feed)
```

**Issue**: Can't cancel all goroutines if user hits Ctrl+C

**Proposed**:
```go
func fetchFeeds(ctx context.Context, cfg *config.Config) error {
    // ...
    for i, feed := range feeds {
        wg.Add(1)
        go func(index int, f repository.Feed) {
            defer wg.Done()

            // Create child context from parent
            fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
            defer cancel()

            resp, err := c.Fetch(fetchCtx, f.URL, cache)
            // ...
        }(i, feed)
    }

    // Add context cancellation on interrupt
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()
    }
}
```

**Benefits**:
- Graceful shutdown on Ctrl+C
- Proper cancellation propagation
- Better timeout handling
- Go best practices

**Estimate**: 1 day
**Files**: `cmd/rp/commands.go`, update all cmd* functions

---

### 2.2 Goroutine Leak Prevention

**Problem**: If a goroutine panics or context is cancelled, other goroutines may leak (they'll complete but resources aren't cleaned up properly).

**Proposed**:
```go
// Add goroutine leak detection in tests
func TestFetchFeeds_NoGoroutineLeaks(t *testing.T) {
    before := runtime.NumGoroutine()

    // Run fetch
    err := fetchFeeds(context.Background(), cfg)
    require.NoError(t, err)

    // Wait for cleanup
    time.Sleep(100 * time.Millisecond)

    after := runtime.NumGoroutine()
    assert.Equal(t, before, after, "goroutine leak detected")
}
```

**Add Context Cancellation**:
```go
// In fetchFeeds, handle context cancellation properly
for _, feed := range feeds {
    select {
    case <-ctx.Done():
        return ctx.Err()  // Stop spawning new goroutines
    default:
    }

    wg.Add(1)
    go func(f repository.Feed) {
        defer wg.Done()
        // ...
    }(feed)
}
```

**Estimate**: 1 day
**Files**: `cmd/rp/commands.go`, add leak tests

---

## Priority 3: Logging & Observability 📊

### 3.1 Replace Global Logger with Structured Logger

**Problem**: `globalLogger` is a global variable with no context, not thread-safe for level changes.

**Current** (commands.go:34):
```go
var globalLogger = &Logger{level: LogLevelInfo}

func (l *Logger) SetLevel(level string) {
    // Not thread-safe!
    l.level = LogLevelInfo
}
```

**Proposed**: Use structured logging (slog in Go 1.21+):

```go
// pkg/logging/logger.go
type Logger struct {
    slog *slog.Logger
}

func New(level string, output io.Writer) *Logger {
    var logLevel slog.Level
    switch level {
    case "debug":
        logLevel = slog.LevelDebug
    case "info":
        logLevel = slog.LevelInfo
    // ...
    }

    handler := slog.NewTextHandler(output, &slog.HandlerOptions{
        Level: logLevel,
    })

    return &Logger{
        slog: slog.New(handler),
    }
}

func (l *Logger) Info(msg string, args ...any) {
    l.slog.Info(msg, args...)
}

// Usage:
logger.Info("fetching feed",
    "url", feedURL,
    "feed_id", feedID,
    "attempt", retryCount)
```

**Benefits**:
- Structured logs (JSON-parseable)
- Context tracking (request IDs, feed IDs)
- Thread-safe
- Standard library (Go 1.21+)

**Migration Path**:
1. Keep current Logger for backward compatibility
2. Add slog wrapper
3. Migrate gradually
4. Remove old Logger in v2.0

**Estimate**: 2 days
**Files**: Create `pkg/logging/`, update commands.go

---

### 3.2 Add Metrics Collection

**Problem**: No visibility into performance (fetch times, error rates, cache hit rates).

**Proposed**: Add simple metrics:

```go
// pkg/metrics/metrics.go
type Metrics struct {
    FetchTotal        int64
    FetchErrors       int64
    FetchDuration     time.Duration
    CacheHits         int64
    EntriesStored     int64
    BytesFetched      int64
}

// Collect during operation
metrics := &Metrics{}
start := time.Now()
resp, err := c.Fetch(ctx, feedURL, cache)
metrics.FetchDuration += time.Since(start)
if err != nil {
    atomic.AddInt64(&metrics.FetchErrors, 1)
}
if resp.NotModified {
    atomic.AddInt64(&metrics.CacheHits, 1)
}

// Report at end
logger.Info("fetch complete",
    "total_feeds", len(feeds),
    "errors", metrics.FetchErrors,
    "cache_hits", metrics.CacheHits,
    "avg_duration_ms", metrics.FetchDuration.Milliseconds() / int64(len(feeds)))
```

**Estimate**: 1 day
**Files**: Create `pkg/metrics/`, integrate in commands.go

---

## Priority 4: Error Handling 🚨

### 4.1 Consistent Sentinel Errors

**Problem**: Some packages export sentinel errors, some don't. Inconsistent wrapping.

**Current**:
```go
// repository.go exports them
var ErrFeedNotFound = errors.New("feed not found")

// crawler.go exports them
var ErrInvalidURL = errors.New("invalid URL")

// normalizer.go doesn't export any
// generator.go doesn't export any
```

**Proposed**: Define package-level errors consistently:

```go
// pkg/normalizer/errors.go
var (
    ErrInvalidFeed = errors.New("invalid feed data")
    ErrNoEntries   = errors.New("feed contains no entries")
    ErrEmptyFeed   = errors.New("feed has no content")
)

// pkg/generator/errors.go
var (
    ErrInvalidTemplate = errors.New("invalid template")
    ErrMissingData     = errors.New("required template data missing")
)
```

**Usage**:
```go
// Caller can check specific errors
if errors.Is(err, normalizer.ErrInvalidFeed) {
    // Handle invalid feed specifically
}
```

**Estimate**: 4 hours
**Files**: Add `errors.go` to each pkg/

---

### 4.2 Error Context Enhancement

**Problem**: Some errors lack context about what operation failed.

**Current**:
```go
if err != nil {
    return fmt.Errorf("update feed: %w", err)
}
```

**Better**:
```go
if err != nil {
    return fmt.Errorf("update feed %s (id=%d): %w", feedURL, feedID, err)
}
```

**Review Checklist**:
- [ ] All database errors include feed/entry identifiers
- [ ] All HTTP errors include URL
- [ ] All parsing errors include source
- [ ] All file operations include path

**Estimate**: 1 day
**Files**: Review all error returns in pkg/

---

## Priority 5: Performance 🚀

### 5.1 Connection Pooling

**Problem**: HTTP client creates new connections for each fetch.

**Current**:
```go
// crawler.go:68-81
func New() *Crawler {
    return &Crawler{
        client: &http.Client{
            Timeout: DefaultTimeout,
            // Uses default transport (no pooling config)
        },
        // ...
    }
}
```

**Proposed**: Configure connection pooling:

```go
func New() *Crawler {
    transport := &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
        DisableKeepAlives:   false,
    }

    return &Crawler{
        client: &http.Client{
            Timeout:   DefaultTimeout,
            Transport: transport,
        },
        // ...
    }
}
```

**Benefit**: Reuse connections, faster fetches

**Estimate**: 2 hours
**Files**: `pkg/crawler/crawler.go`

---

### 5.2 Database Prepared Statement Reuse

**Problem**: Prepared statements created on each call, not cached.

**Current**: Repository methods create statements each time:
```go
func (r *Repository) AddFeed(url, title string) (int64, error) {
    result, err := r.db.Exec(`INSERT INTO feeds ...`, url, title)
    // Statement prepared and discarded each call
}
```

**Proposed**: Cache prepared statements:

```go
type Repository struct {
    db   *sql.DB
    stmt struct {
        addFeed      *sql.Stmt
        updateFeed   *sql.Stmt
        getFeedByURL *sql.Stmt
        // ...
    }
}

func (r *Repository) prepareStatements() error {
    var err error
    r.stmt.addFeed, err = r.db.Prepare(`INSERT INTO feeds ...`)
    if err != nil {
        return err
    }
    // ... prepare others
    return nil
}

func (r *Repository) AddFeed(url, title string) (int64, error) {
    result, err := r.stmt.addFeed.Exec(url, title)
    // Reuses prepared statement
}
```

**Benefit**: 10-20% faster database operations

**Estimate**: 1 day
**Files**: `pkg/repository/repository.go`

---

### 5.3 Batch Entry Inserts

**Problem**: Entries inserted one-by-one in loop (commands.go:889-910).

**Current**:
```go
for _, entry := range entries {
    repoEntry := &repository.Entry{...}
    if err := repo.UpsertEntry(repoEntry); err != nil {
        // Each call is a separate transaction
    }
}
```

**Proposed**: Batch insert in transaction:

```go
// pkg/repository/repository.go
func (r *Repository) UpsertEntriesBatch(entries []*Entry) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare(`INSERT INTO entries ...`)
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, entry := range entries {
        if _, err := stmt.Exec(...); err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

**Benefit**: 5-10x faster for feeds with many entries

**Estimate**: 4 hours
**Files**: `pkg/repository/repository.go`

---

## Priority 6: Code Organization 📁

### 6.1 Split Large Files

**Files over 500 lines**:
- `pkg/generator/generator.go`: 648 lines (includes 300-line embedded template)
- `pkg/repository/repository.go`: 593 lines

**Proposed**:
```
pkg/generator/
  generator.go          (100 lines - core logic)
  template.go           (200 lines - template functions)
  template_default.go   (300 lines - embedded template)
  formats.go            (50 lines - date/time formatting)
```

```
pkg/repository/
  repository.go         (200 lines - core interface)
  feeds.go              (150 lines - feed operations)
  entries.go            (150 lines - entry operations)
  schema.go             (100 lines - schema & migration)
```

**Estimate**: 1 day
**Files**: Split large files

---

### 6.2 Extract Helper Functions

**Problem**: fetchFeeds() has nested helper logic that could be functions.

**Extract**:
```go
// fetchFeeds() currently has inline:
// - Feed cache preparation (5 lines)
// - Entry conversion (15 lines)
// - Error handling patterns (repeated 3x)

// Proposed helpers:
func prepareFeedCache(feed repository.Feed) crawler.FeedCache { ... }
func convertToRepoEntry(entry normalizer.Entry, feedID int64) *repository.Entry { ... }
func logFetchError(logger *Logger, feedURL string, err error) { ... }
```

**Estimate**: 4 hours
**Files**: `cmd/rp/commands.go`

---

## Priority 7: Configuration 🔧

### 7.1 Config Validation on Load

**Problem**: Config validation happens late (on use), not on load.

**Current**:
```go
cfg, err := config.LoadFromFile(path)
// cfg.Validate() called separately, sometimes not at all
```

**Proposed**:
```go
func LoadFromFile(path string) (*Config, error) {
    cfg := Default()
    // ... parse file ...

    // Validate before returning
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("invalid config: %w", err)
    }

    return cfg, nil
}
```

**Benefit**: Fail fast with clear error messages

**Estimate**: 2 hours
**Files**: `pkg/config/config.go`

---

### 7.2 Typed Config Errors

**Problem**: Generic error messages for config problems.

**Current**:
```go
return fmt.Errorf("invalid days value: %s", value)
```

**Proposed**:
```go
type ConfigError struct {
    Field   string
    Value   string
    Message string
}

func (e *ConfigError) Error() string {
    return fmt.Sprintf("config.%s: %s (got: %s)", e.Field, e.Message, e.Value)
}

// Usage:
return &ConfigError{
    Field:   "planet.days",
    Value:   value,
    Message: "must be >= 1",
}
```

**Benefit**: Better error messages for users

**Estimate**: 4 hours
**Files**: `pkg/config/config.go`

---

## Implementation Roadmap

### Phase 1: Foundation (1-2 weeks) - Post v1.0.0
1. Introduce interfaces (P1.1) - 2 days
2. Extract business logic (P1.2) - 3 days
3. Context propagation (P2.1) - 1 day
4. Goroutine leak prevention (P2.2) - 1 day

**Benefits**: Testable architecture, better concurrency

### Phase 2: Testing (1 week)
1. Improve cmd/rp coverage (P1.3) - 2 days
2. Add integration tests with new interfaces - 2 days
3. Add goroutine leak tests - 1 day

**Benefits**: >60% cmd/rp coverage, confidence in refactoring

### Phase 3: Observability (1 week)
1. Structured logging (P3.1) - 2 days
2. Metrics collection (P3.2) - 1 day
3. Error handling improvements (P4.1, P4.2) - 2 days

**Benefits**: Better debugging, performance visibility

### Phase 4: Performance (3-4 days)
1. Connection pooling (P5.1) - 2 hours
2. Prepared statement reuse (P5.2) - 1 day
3. Batch inserts (P5.3) - 4 hours
4. Benchmark tests - 1 day

**Benefits**: 2-3x faster fetching

### Phase 5: Organization (3-4 days)
1. Split large files (P6.1) - 1 day
2. Extract helpers (P6.2) - 4 hours
3. Config improvements (P7.1, P7.2) - 6 hours
4. Documentation updates - 1 day

**Benefits**: Easier maintenance

**Total**: 4-5 weeks of focused work

---

## Success Metrics

- [ ] Test coverage >60% on cmd/rp (from 26.6%)
- [ ] All packages have interfaces defined
- [ ] Zero goroutine leaks in tests
- [ ] Structured logging with context
- [ ] No file >400 lines
- [ ] 2-3x faster feed fetching (benchmarks)
- [ ] Proper context cancellation (Ctrl+C works)
- [ ] All sentinel errors exported and documented

---

## Out of Scope

These are NOT recommended:
- ❌ Rewrite in another language
- ❌ Add ORM (SQLite is fine as-is)
- ❌ Microservices architecture
- ❌ GraphQL API
- ❌ Web UI (against project philosophy)

---

## Quick Wins (Can Do Today)

These are trivial improvements with immediate benefit:

1. **Connection Pooling** (2 hours) - P5.1
2. **Config Validation on Load** (2 hours) - P7.1
3. **Export Sentinel Errors** (4 hours) - P4.1
4. **Add SQL Safety Comments** (30 min) - Already in v0.4.0 plan

**Total**: 1 day of work, measurable improvement

---

**Status**: Ready for post-v1.0.0 implementation
**Priority**: Technical debt reduction, not urgent
**Goal**: Move from B+ to A quality codebase
