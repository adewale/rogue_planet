# Missing Abstractions Analysis

**Date**: 2025-11-02 (Original analysis)
**Updated**: 2025-11-04 (Decisions finalized)
**Context**: Code review identifying opportunities to improve testability and modularity

## Executive Summary

**Status: Architecture Work Complete ✅**

The Rogue Planet codebase has successfully implemented context propagation throughout all layers (commit 65d22f3). After comprehensive analysis of proposed abstractions, we determined that:

- ✅ **Context Propagation** - COMPLETED (22 methods updated across 4 packages)
- ❌ **FetchService Interface** - NOT PURSUING (current 2-layer architecture is sufficient)
- ❌ **RateLimiter Interface** - NOT PURSUING (removed from plan)
- ❌ **FileSystem Interface** - NOT PURSUING (removed from plan)
- ❌ **OPMLService Interface** - NOT PURSUING (removed from plan)
- ❌ **Configuration Injection** - NOT PURSUING (removed from plan)

The existing abstractions (Fetcher, Crawler, Normalizer, Repository interfaces) combined with context propagation provide sufficient testability and modularity. Future extractions should be driven by actual need, not speculation.

This document is preserved as historical reference showing what was considered and why decisions were made.

---

## 🔴 High Priority Missing Abstractions (Historical)

### 1. ❌ FetchService Interface - NOT PURSUING

**Decision**: After comprehensive analysis, we determined that extracting a FetchService would add complexity without clear value. The current 2-layer architecture (cmd_helpers orchestration → pkg/fetcher business logic) is appropriate for the codebase's scale and needs.

**Rationale**:
- Business logic already well-separated in pkg/fetcher (88.5% test coverage)
- Only 2 call sites with identical needs (cmd_fetch, cmd_update)
- No second consumer exists or is planned
- Would add ~450 lines of code for marginal testability improvement
- Progress reporting and signal handling are correctly placed in CLI layer

See detailed tradeoff analysis conducted 2025-11-04.

---

**Original Analysis** (preserved for reference):

**Current Problem** (`cmd/rp/cmd_helpers.go:75-225`):
- 150 lines of hard-coded orchestration logic in command helper
- Creates all dependencies directly (crawler, normalizer, rate limiter)
- Complex concurrency management embedded in helper function
- Cannot unit test commands without real database/HTTP

**Current Code Pattern**:
```go
func fetchFeeds(cfg *config.Config, logger logging.Logger) error {
    // Hard-coded dependency creation
    c := crawler.NewWithConfig(...)
    n := normalizer.New()
    rateLimiter := ratelimit.New(...)

    // 76 lines of goroutine orchestration
    for i, feed := range feeds {
        wg.Add(1)
        go func(index int, f repository.Feed) {
            // Concurrency, rate limiting, progress tracking...
        }(i, feed)
    }
}
```

**Proposed Solution**:
```go
// pkg/fetcher/service.go
type FetchService interface {
    FetchAllFeeds(ctx context.Context) (FetchResults, error)
}

type FetchResults struct {
    SuccessCount int
    ErrorCount   int
    StoredEntries int
    Errors       []FeedError
}

type FeedError struct {
    FeedURL string
    Error   error
}

// Production implementation
type Service struct {
    fetcher      *Fetcher
    repo         repository.FeedRepository
    rateLimiter  RateLimiter
    logger       logging.Logger
    concurrency  int
}

func NewService(
    fetcher *Fetcher,
    repo repository.FeedRepository,
    rateLimiter RateLimiter,
    logger logging.Logger,
    concurrency int,
) *Service {
    return &Service{
        fetcher:     fetcher,
        repo:        repo,
        rateLimiter: rateLimiter,
        logger:      logger,
        concurrency: concurrency,
    }
}

func (s *Service) FetchAllFeeds(ctx context.Context) (FetchResults, error) {
    // Move orchestration logic from cmd_helpers.go here
    // - Get feeds from repository
    // - Launch workers with concurrency limit
    // - Rate limiting per domain
    // - Progress tracking
    // - Error aggregation
}
```

**Impact**:
- Commands become 10 lines instead of 150
- Full unit test coverage of fetch orchestration
- Reusable across CLI, potential API, scheduled jobs
- Easy to test error scenarios (partial failures, rate limiting, cancellation)

**Effort**: 2-3 hours
**Files to Create**: `pkg/fetcher/service.go`, `pkg/fetcher/service_test.go`
**Files to Modify**: `cmd/rp/cmd_helpers.go`, `cmd/rp/cmd_update.go`, `cmd/rp/cmd_fetch.go`

---

### 2. ❌ RateLimiter Interface - NOT PURSUING

**Decision**: Removed from roadmap. Current concrete `Manager` type in pkg/ratelimit is sufficient.

---

**Original Analysis** (preserved for reference):

**Current Problem** (`pkg/ratelimit/ratelimit.go:16-35`):
- `Manager` is a concrete type with no interface
- Cannot mock rate limiting in tests
- Cannot test concurrent fetch behavior with controlled delays
- Cannot simulate backoff scenarios

**Current Code**:
```go
type Manager struct {  // Concrete type, no abstraction
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex
    limit    rate.Limit
    burst    int
}
```

**Proposed Solution**:
```go
// pkg/ratelimit/interface.go
package ratelimit

import "context"

// RateLimiter controls the rate of requests per feed URL.
type RateLimiter interface {
    // Wait blocks until the rate limiter allows the request for the given feed URL.
    // Returns an error if the context is cancelled.
    Wait(ctx context.Context, feedURL string) error

    // Allow reports whether a request for the given feed URL may happen now.
    Allow(feedURL string) bool

    // Stats returns statistics for a specific feed URL (for observability).
    Stats(feedURL string) RateLimitStats
}

type RateLimitStats struct {
    URL         string
    Limit       float64
    Burst       int
    Available   int
}

// Ensure Manager implements RateLimiter
var _ RateLimiter = (*Manager)(nil)
```

**Mock Implementation for Tests**:
```go
// pkg/ratelimit/mock.go or in test files
type MockRateLimiter struct {
    WaitDelay time.Duration
    WaitError error
    AllowResult bool
}

func (m *MockRateLimiter) Wait(ctx context.Context, feedURL string) error {
    if m.WaitDelay > 0 {
        time.Sleep(m.WaitDelay)
    }
    return m.WaitError
}

func (m *MockRateLimiter) Allow(feedURL string) bool {
    return m.AllowResult
}
```

**Impact**:
- Can test concurrent scenarios with mock delays
- Can simulate rate limiting errors
- Better concurrent fetch test coverage (complements the tests we just added)
- Can verify rate limiting logic without waiting

**Effort**: 30 minutes
**Files to Create**: `pkg/ratelimit/interface.go`
**Files to Modify**: `pkg/ratelimit/ratelimit.go` (add interface check), test files

---

### 3. ❌ FileSystem Interface - NOT PURSUING

**Decision**: Removed from roadmap. Tests using t.TempDir() work well enough.

---

**Original Analysis** (preserved for reference):

**Current Problem** (scattered across codebase):
- Direct `os` package calls throughout
- Tests must use actual filesystem
- Cannot test error conditions (permissions, disk full, etc.)
- Tests are slower and leave artifacts
- Cannot test without filesystem side effects

**Current Pattern**:
```go
// pkg/config/config.go:145
file, err := os.Open(path)

// pkg/generator/generator.go:146
os.MkdirAll(outputDir, 0755)
os.Create(filepath.Join(outputDir, "index.html"))

// pkg/config/config.go:363
data, err := os.ReadFile(path)
```

**Proposed Solution**:
```go
// pkg/filesystem/filesystem.go
package filesystem

import (
    "io"
    "io/fs"
    "os"
)

// FileSystem provides an abstraction over filesystem operations.
// This enables testing without actual filesystem access.
type FileSystem interface {
    // ReadFile reads the entire file at path.
    ReadFile(path string) ([]byte, error)

    // WriteFile writes data to the file at path.
    WriteFile(path string, data []byte, perm os.FileMode) error

    // MkdirAll creates a directory path with all parent directories.
    MkdirAll(path string, perm os.FileMode) error

    // Stat returns file information.
    Stat(path string) (os.FileInfo, error)

    // Open opens a file for reading.
    Open(path string) (io.ReadCloser, error)

    // Create creates or truncates a file.
    Create(path string) (io.WriteCloser, error)

    // ReadDir reads the directory at path.
    ReadDir(path string) ([]fs.DirEntry, error)

    // Remove removes a file or empty directory.
    Remove(path string) error

    // Exists checks if a path exists.
    Exists(path string) bool
}

// OSFileSystem implements FileSystem using the real OS filesystem.
type OSFileSystem struct{}

func NewOSFileSystem() *OSFileSystem {
    return &OSFileSystem{}
}

func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}

func (fs *OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
    return os.WriteFile(path, data, perm)
}

// ... implement all methods ...

// MemFileSystem implements FileSystem using an in-memory map (for testing).
type MemFileSystem struct {
    files map[string][]byte
    dirs  map[string]bool
    mu    sync.RWMutex
}

func NewMemFileSystem() *MemFileSystem {
    return &MemFileSystem{
        files: make(map[string][]byte),
        dirs:  make(map[string]bool),
    }
}

func (fs *MemFileSystem) ReadFile(path string) ([]byte, error) {
    fs.mu.RLock()
    defer fs.mu.RUnlock()

    data, exists := fs.files[path]
    if !exists {
        return nil, os.ErrNotExist
    }
    return data, nil
}

// ... implement all methods with in-memory storage ...
```

**Impact**:
- Fast in-memory tests (microseconds vs milliseconds)
- Test error conditions (permissions, disk full, concurrent access)
- No test artifacts left on disk
- Deterministic tests (no filesystem race conditions)
- Easier CI/CD (no filesystem setup required)

**Effort**: 1-2 hours
**Files to Create**:
- `pkg/filesystem/filesystem.go` (interface + OS implementation)
- `pkg/filesystem/memory.go` (in-memory implementation)
- `pkg/filesystem/filesystem_test.go`

**Files to Modify**:
- `pkg/config/config.go` (inject FileSystem)
- `pkg/generator/generator.go` (inject FileSystem)
- `pkg/opml/opml.go` (inject FileSystem for Write())
- Command tests (use MemFileSystem)

---

## 🟡 Medium Priority Missing Abstractions (Historical)

### 4. ❌ OPMLService Interface - NOT PURSUING

**Decision**: Removed from roadmap. OPML functions are simple enough without service layer.

---

**Original Analysis** (preserved for reference):

**Current Problem**:
- Package exports only functions, no interface
- Commands directly call package functions
- Cannot mock OPML parsing/generation for command tests

**Current Pattern**:
```go
// pkg/opml/opml.go
func Parse(r io.Reader) (*OPML, error)
func Generate(feeds []Feed) (*OPML, error)
```

**Proposed Solution**:
```go
// pkg/opml/interface.go
type OPMLService interface {
    Parse(r io.Reader) (*OPML, error)
    Generate(feeds []Feed) (*OPML, error)
}

// pkg/opml/service.go
type Service struct{}

func NewService() *Service {
    return &Service{}
}

func (s *Service) Parse(r io.Reader) (*OPML, error) {
    return Parse(r) // Delegate to existing function
}

func (s *Service) Generate(feeds []Feed) (*OPML, error) {
    return Generate(feeds)
}
```

**Impact**: Command tests can mock OPML operations

**Effort**: 30 minutes

---

### 5. ✅ Context Propagation - COMPLETED

**Status**: Implemented in commit 65d22f3 (2025-11-04)

**What Was Completed**:
- ✅ Repository interface: All 15 methods now accept context.Context
- ✅ Repository implementation: Converted to context-aware SQL (QueryContext, ExecContext)
- ✅ Generator: All 5 methods (Generate, GenerateToFile, CopyStaticAssets, copyDir, copyFile)
- ✅ Normalizer: Parse() method accepts context
- ✅ OPML: ParseFile() and Write() accept context
- ✅ Command layer: Signal handling in main.go with signal.NotifyContext
- ✅ Test files: All 38 test files updated to pass context.Background()

**Impact**:
- Users can press Ctrl+C to gracefully cancel long-running operations
- Database queries can be cancelled via context
- All operations support timeout enforcement
- Maintained 70.4% test coverage across all packages

See CHANGELOG.md for full details.

---

**Original Analysis** (preserved for reference):

**Current Good Examples**:
- ✅ `pkg/fetcher/fetcher.go:77` - `FetchFeed(ctx context.Context, ...)` - Excellent!
- ✅ `pkg/crawler/crawler.go:264` - `Fetch(ctx context.Context, ...)` - Excellent!
- ✅ `pkg/crawler/crawler.go:437` - `FetchWithRetry(ctx context.Context, ...)` - Excellent!

These packages demonstrate the correct pattern: context flows from command layer through fetcher to crawler, enabling cancellation and timeout control.

**Missing Context - Repository Methods (CRITICAL - 14 methods)** [NOW COMPLETED]

`pkg/repository/repository.go` - All database operations lack context:

**Feed Operations**:
- Line 316: `AddFeed(url, title string)`
- Line 330: `UpdateFeed(id int64, title, link string, updated time.Time)`
- Line 345: `UpdateFeedCache(id int64, etag, lastModified string, lastFetched time.Time)`
- Line 360: `UpdateFeedError(id int64, errorMsg string)`
- Line 377: `UpdateFeedURL(id int64, newURL string)`
- Line 391: `GetFeeds(activeOnly bool)`
- Line 408: `GetFeedByURL(url string)`
- Line 428: `RemoveFeed(id int64)`

**Entry Operations**:
- Line 441: `UpsertEntry(entry *Entry)` - Can be slow with large entries
- Line 467: `GetRecentEntries(days int)` - Large queries can block
- Line 524: `GetRecentEntriesWithOptions(days int, filterByFirstSeen bool, sortBy string)`
- Line 584: `CountEntries()`
- Line 594: `CountRecentEntries(days int)`
- Line 611: `GetEntryCountForFeed(feedID int64)`
- Line 627: `PruneOldEntries(days int)` - Potentially long-running DELETE

**Why This Is Critical**:
- Database operations are synchronous and can block indefinitely
- No way to cancel long-running queries
- No timeout enforcement
- Graceful shutdown impossible during bulk operations
- Users cannot Ctrl+C during slow operations

**Required Changes**:
```go
// pkg/repository/interface.go - Update interface
type FeedRepository interface {
    AddFeed(ctx context.Context, url, title string) (int64, error)
    UpdateFeed(ctx context.Context, id int64, title, link string, updated time.Time) error
    GetRecentEntries(ctx context.Context, days int) ([]Entry, error)
    PruneOldEntries(ctx context.Context, days int) (int64, error)
    // ... update all 14 methods
}

// pkg/repository/repository.go - Implementation
func (r *Repository) GetRecentEntries(ctx context.Context, days int) ([]Entry, error) {
    // Use context-aware database methods
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    // Query can be cancelled via context
    for rows.Next() {
        // Parse rows...
    }
    return entries, nil
}
```

---

**Missing Context - Generator Methods (HIGH - 5 methods)**

`pkg/generator/generator.go` - File I/O operations lack context:

- Line 119: `Generate(w io.Writer, data TemplateData)` - Template rendering
- Line 144: `GenerateToFile(outputPath string, data TemplateData)` - File creation + dir creation
- Line 178: `CopyStaticAssets(outputDir string)` - Recursive directory copy
- Line 209: `copyDir(src, dst string)` - Recursive file operations
- Line 249: `copyFile(src, dst string)` - File I/O

**Why This Matters**:
- Site generation can take significant time with many entries
- User should be able to cancel during file operations
- Recursive directory copies can be interrupted

**Required Changes**:
```go
func (g *Generator) GenerateToFile(ctx context.Context, outputPath string, data TemplateData) error {
    // Check context before expensive operations
    if err := ctx.Err(); err != nil {
        return err
    }

    if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
        return err
    }

    // Check context again before file creation
    if err := ctx.Err(); err != nil {
        return err
    }

    file, err := os.Create(outputPath)
    // ...
}

func (g *Generator) copyDir(ctx context.Context, src, dst string) error {
    entries, err := os.ReadDir(src)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        // Check context in loop - enables cancellation during recursive copy
        if err := ctx.Err(); err != nil {
            return err
        }

        // Process entry...
    }
    return nil
}
```

---

**Missing Context - Normalizer Methods (MEDIUM - 1 method)**

`pkg/normalizer/normalizer.go`:
- Line 73: `Parse(feedData []byte, feedURL string, fetchTime time.Time)`

**Why This Matters**:
- Feed parsing can be CPU-intensive for large feeds (10MB limit)
- Context enables cancellation during parsing loops

**Required Changes**:
```go
// pkg/normalizer/interface.go - Update interface
type FeedNormalizer interface {
    Parse(ctx context.Context, feedData []byte, feedURL string, fetchTime time.Time) (*FeedMetadata, []Entry, error)
}

// pkg/normalizer/normalizer.go - Implementation
func (n *Normalizer) Parse(ctx context.Context, feedData []byte, feedURL string, fetchTime time.Time) (*FeedMetadata, []Entry, error) {
    // Check context before expensive parsing
    if err := ctx.Err(); err != nil {
        return nil, nil, err
    }

    feed, err := n.parser.ParseString(string(feedData))
    // ...
}
```

---

**Missing Context - OPML Methods (LOW - 2 methods)**

`pkg/opml/opml.go`:
- Line 75: `ParseFile(path string)` - Uses os.ReadFile
- Line 176: `Write(path string)` - Uses os.WriteFile

**Why This Matters**:
- OPML files are typically small, but context provides consistency
- Follows Go best practices

**Required Changes**:
```go
func ParseFile(ctx context.Context, path string) (*OPML, error) {
    // Check context before I/O
    if err := ctx.Err(); err != nil {
        return nil, err
    }
    data, err := os.ReadFile(path)
    // ...
}
```

---

**Implementation Priority Order**:

1. **Repository (1-2 days)** - HIGHEST IMPACT
   - Update interface at `pkg/repository/interface.go` (14 method signatures)
   - Update implementation to use `db.QueryContext()`, `db.ExecContext()`
   - Update all callers in pkg/fetcher and cmd/rp

2. **Command Layer Signal Handling (1 day)**
   - Add context with signal handling in cmd/rp/main.go
   - Propagate context through fetchFeeds() and generateSite() in cmd_helpers.go
   - Test cancellation with Ctrl+C during operations

3. **Generator (1 day)**
   - Add context to 5 methods
   - Add context checks in file I/O loops
   - Test cancellation during site generation

4. **Normalizer (0.5 days)**
   - Update interface and implementation
   - Update callers in pkg/fetcher

5. **OPML (0.5 days)**
   - Add context for consistency
   - Update callers in cmd/rp

**Total Effort**: 3-5 days

**Impact**:
- ✅ Graceful shutdown during long operations
- ✅ User can Ctrl+C to cancel operations
- ✅ Timeout enforcement for all I/O and database operations
- ✅ Better resource management
- ✅ Follows Go best practices
- ✅ Enables future work (distributed tracing, request-scoped logging)

---

### 6. ❌ Configuration Injection - NOT PURSUING

**Decision**: After comprehensive analysis (2025-11-04), determined that configuration injection would add complexity without solving a real pain point.

**Rationale**:
- Current pattern works well (tests use t.TempDir() successfully)
- Config loading is fast (~0.8ms, not a bottleneck)
- Would require 120 lines of changes across 17 files (breaking changes to all command signatures)
- Main.go would grow by ~50 lines
- Integration tests would still need real config files for E2E validation
- Benefit score: 25/40, Cost score: 22/40, Net value: marginal

See detailed consequence analysis conducted 2025-11-04.

---

**Original Analysis** (preserved for reference):

**Current Problem**:
```go
// Every command loads config independently
func cmdUpdate(opts UpdateOptions) error {
    cfg, err := loadConfig(opts.ConfigPath)  // Hard-coded
    // ...
}
```

**Better Pattern**:
```go
// Commands accept config, caller handles loading
func cmdUpdate(opts UpdateOptions, cfg *config.Config) error {
    // Focus on business logic, not infrastructure
}

// main.go does the loading
func main() {
    cfg, err := loadConfig(configPath)
    if err != nil {
        log.Fatal(err)
    }

    // Pass to commands
    cmdUpdate(opts, cfg)
}
```

**Impact**:
- Easier command testing with mock configs
- Clearer separation of concerns
- Config loading errors separated from business logic errors

**Effort**: 30 minutes (but lower priority than context propagation)

---

## ✅ Excellent Existing Abstractions (Don't Change)

### 1. Fetcher Package (`pkg/fetcher/fetcher.go:37-53`)

**Exemplary Dependency Injection**:
```go
func New(
    c crawler.FeedCrawler,        // Interface ✓
    n normalizer.FeedNormalizer,  // Interface ✓
    r repository.FeedRepository,  // Interface ✓
    repoMutex sync.Locker,
    logger logging.Logger,        // Interface ✓
    maxRetries int,
) *Fetcher
```

This is **textbook dependency injection** - all dependencies are interfaces, constructor makes dependencies explicit, fully testable with mocks.

### 2. Core Package Interfaces

All essential interfaces exist and are well-designed:
- `crawler.FeedCrawler` (pkg/crawler/interface.go:10-14)
- `normalizer.FeedNormalizer` (pkg/normalizer/interface.go:10-14)
- `repository.FeedRepository` (pkg/repository/interface.go:8-54)
- `logging.Logger` (pkg/logging/logging.go:8-13)
- `timeprovider.TimeProvider` (pkg/timeprovider/timeprovider.go:16-23)

### 3. Command Options Pattern

**Good Design** (`cmd/rp/cmd_options.go`):
- All commands accept Options structs
- Options include `io.Writer` for testable output
- Clean separation of parsing and execution

---

## Implementation Priority

| Abstraction | Effort | Impact | Test Improvement | Priority |
|-------------|--------|--------|------------------|----------|
| **Context Propagation** | **Medium (3-5 days)** | **Very High** | **Cancellation, timeouts, graceful shutdown** | **1. Do First** |
| RateLimiter Interface | Low (30min) | High | Better concurrency test coverage | **2. Do Second** |
| FetchService | Medium (2-3h) | Very High | Commands become unit testable | **3. Do Third** |
| FileSystem Interface | Medium (1-2h) | High | Fast in-memory tests | 4. Later |
| OPML Interface | Low (30min) | Medium | Command test coverage | 5. Later |
| Config Injection | Low (30min) | Low | Minor test improvement | 6. Deferred |

---

## Context Propagation Implementation Details

### Repository Context Pattern

**Before (No Context)**:
```go
func (r *Repository) GetRecentEntries(days int) ([]Entry, error) {
    query := `SELECT ... WHERE published >= ?`
    rows, err := r.db.Query(query, cutoff)  // Cannot be cancelled
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entries []Entry
    for rows.Next() {  // Blocks until all rows processed
        // Parse entry...
    }
    return entries, nil
}
```

**After (With Context)**:
```go
func (r *Repository) GetRecentEntries(ctx context.Context, days int) ([]Entry, error) {
    query := `SELECT ... WHERE published >= ?`
    rows, err := r.db.QueryContext(ctx, query, cutoff)  // Respects context cancellation
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var entries []Entry
    for rows.Next() {  // Query can be cancelled via context
        // Parse entry...
    }
    return entries, nil
}
```

**Key Changes**:
- `db.Query()` → `db.QueryContext(ctx, ...)`
- `db.Exec()` → `db.ExecContext(ctx, ...)`
- Database driver handles context cancellation automatically
- No additional error checking needed in most cases

---

### Command Layer Signal Handling Pattern

**Implementation in main.go**:
```go
func run() error {
    // Create context that cancels on SIGINT or SIGTERM
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    switch command {
    case "update":
        return runUpdate(ctx)  // Pass context to command
    case "fetch":
        return runFetch(ctx)
    case "generate":
        return runGenerate(ctx)
    default:
        // Commands that don't need cancellation can use context.Background()
        return runOtherCommand()
    }
}

func runUpdate(ctx context.Context) error {
    opts, err := parseUpdateFlags(os.Args[2:])
    if err != nil {
        return err
    }
    opts.Output = os.Stdout
    return cmdUpdate(ctx, opts)  // Pass context
}
```

**Updated Command Signature**:
```go
// Before
func cmdUpdate(opts UpdateOptions) error

// After
func cmdUpdate(ctx context.Context, opts UpdateOptions) error {
    cfg, err := loadConfig(opts.ConfigPath)
    if err != nil {
        return err
    }

    repo, err := repository.New(cfg.Database.Path)
    if err != nil {
        return err
    }
    defer repo.Close()

    // Pass context to helper functions
    return fetchFeeds(ctx, cfg, repo, opts.Logger)
}
```

---

### Generator Context Checking Pattern

**Loop-Based Cancellation**:
```go
func (g *Generator) copyDir(ctx context.Context, src, dst string) error {
    entries, err := os.ReadDir(src)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        // Check context at the start of each iteration
        if err := ctx.Err(); err != nil {
            return fmt.Errorf("operation cancelled: %w", err)
        }

        srcPath := filepath.Join(src, entry.Name())
        dstPath := filepath.Join(dst, entry.Name())

        if entry.IsDir() {
            if err := g.copyDir(ctx, srcPath, dstPath); err != nil {
                return err
            }
        } else {
            if err := g.copyFile(ctx, srcPath, dstPath); err != nil {
                return err
            }
        }
    }
    return nil
}
```

**Why This Works**:
- Check `ctx.Err()` at loop boundaries (not every line)
- Returns immediately when context is cancelled
- User gets fast response to Ctrl+C
- No resource leaks (deferred cleanup still runs)

---

### Testing Context Cancellation

**Test Pattern 1: Timeout**:
```go
func TestRepository_GetRecentEntries_Timeout(t *testing.T) {
    repo := setupTestRepository(t)

    // Create context with very short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    time.Sleep(5 * time.Millisecond)  // Ensure timeout expires

    _, err := repo.GetRecentEntries(ctx, 7)

    // Should return context.DeadlineExceeded
    if !errors.Is(err, context.DeadlineExceeded) {
        t.Errorf("Expected DeadlineExceeded, got: %v", err)
    }
}
```

**Test Pattern 2: Cancellation**:
```go
func TestGenerator_CopyAssets_Cancellation(t *testing.T) {
    gen := generator.New()

    ctx, cancel := context.WithCancel(context.Background())

    // Cancel context during operation
    go func() {
        time.Sleep(10 * time.Millisecond)
        cancel()
    }()

    err := gen.CopyStaticAssets(ctx, testOutputDir)

    // Should return context.Canceled
    if !errors.Is(err, context.Canceled) {
        t.Errorf("Expected Canceled, got: %v", err)
    }
}
```

**Test Pattern 3: Success with Context**:
```go
func TestRepository_AddFeed_WithContext(t *testing.T) {
    repo := setupTestRepository(t)

    // Normal context that doesn't cancel
    ctx := context.Background()

    id, err := repo.AddFeed(ctx, "http://example.com/feed.xml", "Example Feed")

    if err != nil {
        t.Fatalf("AddFeed failed: %v", err)
    }

    if id <= 0 {
        t.Error("Expected valid feed ID")
    }
}
```

---

### Context Propagation Best Practices

1. **Always accept context as first parameter**: `func Foo(ctx context.Context, other params)`
2. **Propagate context through call chain**: Don't create new contexts unless you need timeouts
3. **Check context in loops**: Use `ctx.Err()` at loop boundaries for long-running operations
4. **Use context-aware methods**: `db.QueryContext()`, not `db.Query()`
5. **Don't ignore context errors**: Return them immediately
6. **Signal handling**: Use `signal.NotifyContext()` in main.go
7. **Testing**: Test both success path and cancellation/timeout paths

**Common Mistakes to Avoid**:
- ❌ Creating `context.Background()` in library code (accept it as parameter instead)
- ❌ Not checking context in long loops
- ❌ Ignoring context.Err() return value
- ❌ Using non-context database methods when context is available

---

## Testing Impact Examples

### Before (Current State)
```go
func TestCmdUpdate(t *testing.T) {
    // Must create real database file
    dir := t.TempDir()
    dbPath := filepath.Join(dir, "test.db")

    // Must write actual config file
    configPath := filepath.Join(dir, "config.ini")
    os.WriteFile(configPath, configData, 0644)

    // Must set up HTTP test server for feeds
    server := httptest.NewServer(...)

    // Test is slow (100-500ms)
    // Hard to test error conditions
    // Leaves filesystem artifacts
}
```

### After (With Abstractions)
```go
func TestCmdUpdate(t *testing.T) {
    // Mock service with controlled behavior
    mockService := &mockFetchService{
        fetchResult: FetchResults{
            SuccessCount: 5,
            ErrorCount: 0,
        },
    }

    // Mock config (no file I/O)
    cfg := &config.Config{
        Planet: config.Planet{Name: "Test"},
    }

    // Fast (<1ms), deterministic, easy error testing
    err := cmdUpdate(opts, cfg, mockService)
    assert.NoError(t, err)
}
```

---

### Context Propagation Testing Impact

**Before (No Context - Cannot Test Cancellation)**:
```go
func TestPruneOldEntries(t *testing.T) {
    repo := setupTestRepository(t)

    // Add 1000 old entries
    for i := 0; i < 1000; i++ {
        repo.UpsertEntry(&repository.Entry{...})
    }

    // This operation runs to completion, no way to cancel
    deleted, err := repo.PruneOldEntries(30)

    // Cannot test:
    // - Timeout scenarios
    // - User cancellation (Ctrl+C)
    // - Graceful shutdown during long operations
}
```

**After (With Context - Full Cancellation Testing)**:
```go
func TestPruneOldEntries_Success(t *testing.T) {
    repo := setupTestRepository(t)

    // Add 1000 old entries
    for i := 0; i < 1000; i++ {
        repo.UpsertEntry(context.Background(), &repository.Entry{...})
    }

    ctx := context.Background()
    deleted, err := repo.PruneOldEntries(ctx, 30)

    assert.NoError(t, err)
    assert.Equal(t, int64(1000), deleted)
}

func TestPruneOldEntries_ContextCancelled(t *testing.T) {
    repo := setupTestRepository(t)

    // Add many entries to make operation slow
    for i := 0; i < 10000; i++ {
        repo.UpsertEntry(context.Background(), &repository.Entry{...})
    }

    // Cancel context immediately
    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    deleted, err := repo.PruneOldEntries(ctx, 30)

    // Operation should fail with context error
    assert.Error(t, err)
    assert.True(t, errors.Is(err, context.Canceled))
    // May have deleted some entries before cancellation, but not all
}

func TestPruneOldEntries_Timeout(t *testing.T) {
    repo := setupTestRepository(t)

    // Add many entries
    for i := 0; i < 10000; i++ {
        repo.UpsertEntry(context.Background(), &repository.Entry{...})
    }

    // Very short timeout
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()

    time.Sleep(5 * time.Millisecond)  // Ensure timeout

    deleted, err := repo.PruneOldEntries(ctx, 30)

    // Operation should timeout
    assert.Error(t, err)
    assert.True(t, errors.Is(err, context.DeadlineExceeded))
}

func TestCmdUpdate_UserCancellation(t *testing.T) {
    // Test full command with cancellation
    ctx, cancel := context.WithCancel(context.Background())

    // Cancel after 100ms to simulate Ctrl+C
    go func() {
        time.Sleep(100 * time.Millisecond)
        cancel()
    }()

    err := cmdUpdate(ctx, opts)

    // Command should handle cancellation gracefully
    assert.Error(t, err)
    assert.True(t, errors.Is(err, context.Canceled))
    // No corrupted database state
    // No partial files written
    // Resources cleaned up via defer
}
```

**Benefits of Context in Tests**:
- ✅ Test timeout scenarios
- ✅ Test user cancellation (Ctrl+C simulation)
- ✅ Test graceful shutdown during long operations
- ✅ Verify resource cleanup on cancellation
- ✅ Test partial completion scenarios
- ✅ Fast tests (cancel after short duration instead of waiting for completion)

---

## Migration Strategy

### Phase 1: Context Propagation (Week 1) - HIGHEST PRIORITY

**Day 1-2: Repository Context Migration**
1. Update `pkg/repository/interface.go` - Add `context.Context` as first parameter to all 14 methods
2. Update `pkg/repository/repository.go` - Replace `db.Query()` with `db.QueryContext()`, `db.Exec()` with `db.ExecContext()`
3. Add context error checking in long-running operations (PruneOldEntries, GetRecentEntries)
4. Run full test suite to verify no breaking changes

**Day 3: Command Layer Signal Handling**
1. Add signal handling in `cmd/rp/main.go` - Create context that cancels on SIGINT/SIGTERM
2. Update `cmd_helpers.go:fetchFeeds()` - Accept and propagate context
3. Update `cmd_helpers.go:generateSite()` - Accept and propagate context
4. Update all command functions to create and pass context
5. Test cancellation with Ctrl+C during fetch and generate operations

**Day 4: Generator Context Migration**
1. Add context to `Generator.Generate()`, `Generator.GenerateToFile()`, `Generator.CopyStaticAssets()`
2. Add context checks in `copyDir()` and `copyFile()` loops
3. Update callers in cmd_helpers.go
4. Test cancellation during site generation with many files

**Day 5: Normalizer & OPML Context**
1. Update `pkg/normalizer/interface.go` - Add context to Parse()
2. Update `pkg/normalizer/normalizer.go` - Implement context support
3. Update `pkg/fetcher/fetcher.go` - Pass context to normalizer
4. Add context to `pkg/opml/opml.go` ParseFile() and Write()
5. Update callers in command layer
6. Run full test suite

**Deliverable**: Complete context propagation throughout codebase, graceful cancellation support, timeout enforcement

---

### Phase 2: Core Abstractions (Week 2)

**Day 1: RateLimiter Interface**
1. Create `pkg/ratelimit/interface.go` with RateLimiter interface (30 min)
2. Ensure Manager implements interface (10 min)
3. Create mock implementation for tests (20 min)
4. Update fetcher to use interface type (10 min)

**Day 2-3: FetchService**
1. Create `pkg/fetcher/service.go` with FetchService interface (1 hour)
2. Move orchestration logic from cmd_helpers.go to Service.FetchAllFeeds() (2 hours)
3. Update cmd_fetch and cmd_update to use FetchService (1 hour)
4. Add comprehensive tests for FetchService (2 hours)

**Deliverable**: Commands use reusable service layer, better testability

---

### Phase 3: Infrastructure (Week 3)

**Day 1-2: FileSystem Interface**
1. Create `pkg/filesystem/filesystem.go` interface + OS implementation (2 hours)
2. Create `pkg/filesystem/memory.go` in-memory implementation (1 hour)
3. Update pkg/config to accept FileSystem (1 hour)
4. Update pkg/generator to accept FileSystem (1 hour)
5. Convert tests to use MemFileSystem (2 hours)

**Day 3: Polish**
1. Add OPMLService interface (30 min)
2. Documentation updates in CLAUDE.md (1 hour)
3. Add examples and migration guides (1 hour)

**Deliverable**: Fast in-memory tests, comprehensive abstraction layer

**Config Injection**: Deferred to future work (lower priority than context propagation)

---

## Open Questions

1. **FetchService Placement**: Should it be `pkg/fetcher/service.go` or a new package `pkg/service`?
   - Recommendation: `pkg/fetcher/service.go` (keeps related code together)

2. **FileSystem in Config**: Should Config accept FileSystem in constructor or load method?
   - Recommendation: Constructor (makes dependency explicit)

3. **Breaking Changes**: These changes would break existing code. Acceptable?
   - Recommendation: Make changes in a feature branch, update incrementally

4. **Mock Placement**: Should mocks be in test files or separate `mock` package?
   - Recommendation: Small mocks in test files, complex mocks in `pkg/*/mock.go`

---

## References

- Current Architecture: `pkg/fetcher/fetcher.go` (good example of DI)
- Problem Areas: `cmd/rp/cmd_helpers.go:75-225` (hard-coded orchestration)
- Testing Patterns: `pkg/fetcher/fetcher_test.go` (excellent mock implementations)
- Interface Examples: `pkg/*/interface.go` files (well-designed)

---

## Next Steps

1. **Discuss with team**: Agree on priority and approach
2. **Create feature branch**: `feature/abstractions` for incremental changes
3. **Start with FetchService**: Highest impact, proves the pattern
4. **Iterate**: Add abstractions incrementally, update tests
5. **Document**: Update CLAUDE.md with new patterns

---

**Last Updated**: 2025-11-02
**Review Date**: 2025-12-01 (reassess progress and priorities)
