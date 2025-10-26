# Go Auditing Heuristics - Additional Checks from Real-World Audits

**Purpose:** Lessons learned from auditing real Go projects that aren't in generic guidelines
**Source:** Rogue Planet audit (2025-10-26)
**Status:** Proposed additions to GO_AUDITING_HEURISTICS.md

---

## Table of Contents

1. [Documentation vs Reality](#documentation-vs-reality)
2. [Build System Integrity](#build-system-integrity)
3. [Unused Infrastructure](#unused-infrastructure)
4. [Undocumented Features](#undocumented-features)
5. [Architecture Patterns](#architecture-patterns)
6. [Context Propagation](#context-propagation)
7. [Performance Low-Hanging Fruit](#performance-low-hanging-fruit)

---

## Documentation vs Reality

**Problem Found:** Documentation claimed features were complete when they weren't implemented.

### Check 1: TODO/Checklist Accuracy

**Heuristic:** Verify all ✅ checkmarks are actually true

**What to check:**
```bash
# Find all claimed completions
grep -rn "✅\|✓\|\[x\]" TODO.md ROADMAP.md CHECKLIST.md

# For each claimed item:
# 1. Does the file exist?
# 2. Is the dependency in go.mod?
# 3. Is the feature actually implemented?
```

**Real-world example (Rogue Planet):**
```
TODO.md claimed:
✅ QUICKSTART.md - 5-minute setup guide     → FILE DOESN'T EXIST
✅ THEMES.md - Complete theme guide         → FILE DOESN'T EXIST
✅ golang.org/x/time/rate - Rate limiting   → NOT IN go.mod, NEVER IMPORTED
```

**Prevention:**
- Automate checklist verification in CI
- Link checklist items to actual files/commits
- Use `grep -q "expected content" file.md || exit 1` in scripts

---

### Check 2: Dependency Documentation Accuracy

**Heuristic:** Compare documented dependencies vs actual imports

**Commands:**
```bash
# List documented dependencies
grep -E "^- .*:" README.md CLAUDE.md | grep -oE "`[^`]+`"

# List actual dependencies
go list -m all | cut -d' ' -f1

# Find mismatches
comm -23 <(sort documented.txt) <(sort actual.txt)
```

**What to verify:**
- Every documented dependency is in go.mod
- Every go.mod dependency has a purpose
- No "future features" listed as "Key Dependencies"

**Real-world example:**
```
CLAUDE.md claimed:
"Rate limiting: golang.org/x/time/rate"

Reality:
$ grep "time/rate" go.mod
(nothing)
```

**Fix:** Move to "Planned Dependencies" or remove entirely

---

### Check 3: Example Files Referenced vs Present

**Heuristic:** Find all file references in documentation, verify existence

**Command:**
```bash
# Find all file references in markdown
grep -rh -oE '[a-zA-Z0-9_.-]+\.(go|md|sh|ini|yaml|json)' *.md | sort -u > referenced.txt

# Check existence
while read file; do
    [ ! -f "$file" ] && echo "MISSING: $file"
done < referenced.txt
```

**Real-world example (Rogue Planet):**
```
examples/README.md references:
- examples/config.ini     → DOESN'T EXIST (breaks make examples)
- QUICKSTART.md           → DOESN'T EXIST (broken link)
- THEMES.md               → DOESN'T EXIST (broken link)

Makefile references:
- setup-example-planet.sh → DOESN'T EXIST (make target fails)
```

**Fix:** Create missing files OR update documentation

---

## Build System Integrity

**Problem Found:** Build targets and scripts broke on different platforms or referenced non-existent files.

### Check 4: Platform-Specific Commands in Makefiles

**Heuristic:** Find commands that differ between macOS and Linux

**Commands to check:**
```bash
# Find macOS-specific sed
grep -n "sed -i ''" Makefile

# Find other platform-specific commands
grep -n "greadlink\|gsed\|ggrep" Makefile
```

**Real-world example (Rogue Planet):**
```makefile
# BREAKS ON LINUX
sed -i '' 's/foo/bar/' file.txt

# PORTABLE
sed -i.bak 's/foo/bar/' file.txt && rm file.txt.bak

# OR USE PLATFORM DETECTION
ifeq ($(shell uname),Darwin)
    SED := sed -i ''
else
    SED := sed -i
endif
```

**What to verify:**
- sed, awk, grep commands are portable
- No assumptions about GNU vs BSD tools
- Use Go code instead of shell where possible

---

### Check 5: Makefile Target Dependencies

**Heuristic:** Test every make target on clean checkout

**Process:**
```bash
# List all targets
grep "^[a-zA-Z0-9_-]*:" Makefile | cut -d: -f1 > targets.txt

# Test each one
for target in $(cat targets.txt); do
    echo "Testing: make $target"
    make $target || echo "FAILED: $target"
done
```

**What breaks commonly:**
- Targets depending on files that don't exist
- Targets assuming previous targets ran
- Scripts referenced but not committed
- Paths hardcoded for one developer's machine

**Real-world example:**
```makefile
examples:
    @cp examples/config.ini tmp/config.ini  # FAILS - file doesn't exist

setup-example: build
    @./setup-example-planet.sh  # FAILS - script doesn't exist
```

**Prevention:**
- Test make targets in CI
- Use `.PHONY` for targets that don't create files
- Check file existence before referencing

---

## Unused Infrastructure

**Problem Found:** Database fields, struct fields, and functions existed but were never used.

### Check 6: Database Schema vs Queries

**Heuristic:** Find database columns that are written but never read

**Process:**
```bash
# 1. List all database columns from schema
grep -E "^\s+[a-z_]+ (INTEGER|TEXT|BLOB)" schema.sql | awk '{print $1}'

# 2. For each column, search if it's queried
for col in $(cat columns.txt); do
    grep -r "SELECT.*$col" . || echo "NEVER READ: $col"
done
```

**Real-world example (Rogue Planet):**
```sql
-- Schema defines:
CREATE TABLE feeds (
    next_fetch TEXT,           -- Written on every update
    fetch_interval INTEGER     -- Written on every update
);

-- Code does:
UPDATE feeds SET next_fetch = ?, fetch_interval = ? ...

-- But queries are:
SELECT * FROM feeds WHERE active = 1  -- Ignores next_fetch!

-- Should be:
SELECT * FROM feeds WHERE active = 1 AND next_fetch <= ?
```

**What this reveals:**
- Infrastructure for unimplemented features
- Fields that should be removed
- Missed optimization opportunities

---

### Check 7: Unused Struct Fields

**Heuristic:** Find struct fields that are set but never read

**Detection (manual):**
```bash
# 1. Find all struct definitions
grep -n "type.*struct {" *.go

# 2. For each field, search for reads
# Look for: variable.FieldName or struct.FieldName
# Exclude: variable.FieldName = (this is a write)
```

**Real-world example (Rogue Planet):**
```go
type FeedData struct {
    Title       string
    Link        string
    Subscribers int  // NEVER SET (always zero)
}

// Set everywhere:
feed := FeedData{
    Title: f.Title,
    Link:  f.Link,
    // Subscribers not set!
}

// Template uses it:
{{if gt .Subscribers 0}}...{{end}}  // Never true!
```

**Fix:** Remove field or implement feature

---

### Check 8: Functions Tested But Unused in Production

**Heuristic:** Find functions only called in *_test.go files

**Process:**
```bash
# For each exported function
for func in $(grep "^func.*{" pkg/*.go | grep -v "_test.go" | awk '{print $2}' | cut -d'(' -f1); do
    # Check if used in production code
    grep -r "\\.$func\|$func(" --include="*.go" --exclude="*_test.go" | grep -v "^pkg/.*:func $func" || echo "ONLY IN TESTS: $func"
done
```

**Real-world example (Rogue Planet):**
```go
// pkg/crawler/crawler.go
func (c *Crawler) FetchWithRetry(...) error {
    // Well-implemented, 100% tested
}

// But cmd/rp/commands.go uses:
resp, err := c.Fetch(ctx, url, cache)  // Not FetchWithRetry!

// FetchWithRetry only called in:
// - crawler_test.go
// - crawler_comprehensive_test.go
```

**When this is OK:**
- Public API for library users
- Future feature infrastructure

**When this is a problem:**
- Code duplication (retry logic implemented twice)
- Missed quality improvements

---

### Check 9: Stub Implementations

**Heuristic:** Find functions with --dry-run or similar flags that don't actually work

**Pattern to search:**
```bash
# Find functions with dry-run parameters
grep -rn "DryRun.*bool\|dryRun.*bool" --include="*.go"

# Check if they actually preview actions
grep -A 10 "if.*DryRun" *.go
```

**Real-world example (Rogue Planet):**
```go
func cmdPrune(opts PruneOptions) error {
    if opts.DryRun {
        fmt.Printf("Dry run: would delete entries older than %d days\n", opts.Days)
        // In a real implementation, we'd query and show what would be deleted
        return nil
    }
    // ... actual deletion
}
```

**What's wrong:**
- User thinks they're seeing a preview
- Actually just getting a generic message
- No actual query to show what WOULD be deleted

**Fix:**
```go
if opts.DryRun {
    count, err := repo.CountEntriesOlderThan(cutoff)
    if err != nil {
        return err
    }
    fmt.Printf("Would delete %d entries older than %s\n", count, cutoff)
    return nil
}
```

---

## Undocumented Features

**Problem Found:** Features implemented but not in user documentation.

### Check 10: Config Options vs Documentation

**Heuristic:** Find all config fields vs documented options

**Process:**
```bash
# 1. Find all config struct fields
grep -A 50 "type.*Config struct" pkg/config/*.go | grep -E "^\s+[A-Z]" | awk '{print $1}'

# 2. For each field, check if documented
for field in $(cat config_fields.txt); do
    grep -q "$field" README.md || echo "UNDOCUMENTED: $field"
done
```

**Real-world example (Rogue Planet):**
```go
// pkg/config/config.go
type PlanetConfig struct {
    Days             int
    FilterByFirstSeen bool  // IMPLEMENTED
    SortBy           string  // IMPLEMENTED
}

// README.md only documents:
// - days = 7
// (missing FilterByFirstSeen and SortBy!)
```

**Impact:**
- Users don't know powerful features exist
- Features stay unused
- Support burden ("how do I...?")

**Fix:** Document ALL config options in README

---

## Architecture Patterns

**Problem Found:** No interfaces, making testing difficult.

### Check 11: Concrete Types vs Interfaces

**Heuristic:** Check if main business logic depends on concrete types

**What to look for:**
```go
// HARD TO TEST - concrete dependencies
func processFeeds(repo *repository.Repository, crawler *crawler.Crawler) {
    feeds, _ := repo.GetFeeds()  // Can't mock!
    for _, feed := range feeds {
        data, _ := crawler.Fetch(feed.URL)  // Can't mock!
    }
}

// EASY TO TEST - interface dependencies
func processFeeds(repo FeedRepository, crawler FeedCrawler) {
    // Can inject mocks!
}
```

**Detection:**
```bash
# Find interface definitions
grep -rn "type.*interface" pkg/*.go | grep -v "_test.go"

# Should find interfaces for:
# - Database operations
# - HTTP fetching
# - External services
```

**If count is zero:** Architecture problem!

**Benefits of interfaces:**
- Unit tests don't need real database
- Can test error paths easily
- Faster tests (no I/O)
- Dependency injection

---

### Check 12: Global State in Packages

**Heuristic:** Find package-level variables that hold state

**Pattern:**
```bash
# Find global variables
grep -rn "^var [a-z]" pkg/*.go | grep -v "_test.go" | grep -v "^var Err"
```

**Real-world example (Rogue Planet):**
```go
// cmd/rp/commands.go
var globalLogger = &Logger{level: LogLevelInfo}

func (l *Logger) SetLevel(level string) {
    l.level = LogLevelInfo  // NOT THREAD-SAFE!
}
```

**Problems:**
- Not thread-safe
- Makes testing harder (global state leaks between tests)
- Can't have multiple instances with different config

**Better pattern:**
```go
// Pass logger as parameter
func cmdFetch(opts FetchOptions, logger *Logger) error {
    // ...
}
```

---

## Context Propagation

**Problem Found:** Creating context.Background() in goroutines instead of passing parent context.

### Check 13: Context Creation in Goroutines

**Heuristic:** Find context.Background() inside go func()

**Pattern to search:**
```bash
# Find goroutines
grep -B 2 -A 10 "go func" *.go

# Look for context.Background() inside
grep -B 2 -A 10 "go func" *.go | grep "context.Background()"
```

**Real-world example (Rogue Planet):**
```go
// BAD - can't cancel from parent
for _, feed := range feeds {
    go func(f Feed) {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        crawler.Fetch(ctx, f.URL, cache)
    }(feed)
}
```

**What's wrong:**
- User hits Ctrl+C → parent context cancelled
- But goroutines keep running (using Background context!)
- No graceful shutdown

**Fix:**
```go
// GOOD - context from parent
func fetchFeeds(ctx context.Context, feeds []Feed) error {
    for _, feed := range feeds {
        go func(f Feed) {
            fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
            defer cancel()
            crawler.Fetch(fetchCtx, f.URL, cache)
        }(feed)
    }

    // Wait with cancellation support
    select {
    case <-done:
        return nil
    case <-ctx.Done():
        return ctx.Err()  // Cancelled!
    }
}
```

**Detection checklist:**
- [ ] All functions that spawn goroutines take context.Context
- [ ] No context.Background() inside goroutines
- [ ] Context passed to all long-running operations
- [ ] Main function sets up signal handling

---

## Performance Low-Hanging Fruit

**Problem Found:** Simple configuration changes that provide 10-20% speedup.

### Check 14: HTTP Client Configuration

**Heuristic:** Check if http.Client uses connection pooling

**Default transport settings:**
```go
// What Go does by default
http.DefaultTransport = &http.Transport{
    MaxIdleConns:       100,  // OK
    MaxIdleConnsPerHost: 2,   // TOO LOW!
}
```

**Check in code:**
```bash
# Find http.Client creation
grep -rn "http.Client{" --include="*.go"

# Check if Transport is configured
```

**Real-world example (Rogue Planet):**
```go
// SLOW - default transport
client := &http.Client{
    Timeout: 30 * time.Second,
    // Uses default transport (only 2 conns per host!)
}

// FAST - configured transport
transport := &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,  // 5x more connections!
    IdleConnTimeout:     90 * time.Second,
}

client := &http.Client{
    Timeout:   30 * time.Second,
    Transport: transport,
}
```

**Speedup:** 10-20% for concurrent HTTP requests

---

### Check 15: Database Prepared Statements

**Heuristic:** Check if queries are prepared once or every time

**Bad pattern:**
```go
func (r *Repo) GetUser(id int) (*User, error) {
    // Prepared EVERY call!
    row := r.db.QueryRow("SELECT * FROM users WHERE id = ?", id)
}
```

**Good pattern:**
```go
type Repo struct {
    db   *sql.DB
    stmt struct {
        getUser *sql.Stmt
    }
}

func (r *Repo) init() error {
    var err error
    r.stmt.getUser, err = r.db.Prepare("SELECT * FROM users WHERE id = ?")
    return err
}

func (r *Repo) GetUser(id int) (*User, error) {
    // Reuses prepared statement!
    row := r.stmt.getUser.QueryRow(id)
}
```

**Speedup:** 10-20% for database-heavy operations

---

### Check 16: Batch Operations

**Heuristic:** Look for INSERT/UPDATE in loops

**Anti-pattern:**
```bash
# Find loops with database operations
grep -B 5 -A 5 "for.*range" *.go | grep "db.Exec\|stmt.Exec"
```

**Real-world example (Rogue Planet):**
```go
// SLOW - one transaction per entry
for _, entry := range entries {
    repo.UpsertEntry(&entry)  // Each is a transaction!
}

// FAST - batch in single transaction
repo.UpsertEntriesBatch(entries)  // 5-10x faster!
```

---

## Coverage Reporting Tricks

**Problem Found:** Excluding packages from coverage to inflate averages.

### Check 17: Coverage Exclusions

**Heuristic:** Check what's excluded from coverage reports

**Commands:**
```bash
# Check coverage report
go test -coverprofile=coverage.out ./...

# See what's included
go tool cover -func=coverage.out | grep -v "100.0%"

# Compare documented coverage vs actual
```

**Real-world example (Rogue Planet):**
```
TODO.md claims: "88.4% average coverage (Excellent)"

Reality:
- pkg/crawler:  96.6%
- pkg/config:   93.8%
- cmd/rp:       26.6%  ← EXCLUDED from "average"!

Actual average: (96.6 + 93.8 + 26.6) / 3 = 72.3%
```

**What to verify:**
- Coverage includes all packages (including cmd/)
- Exclusions are documented
- Don't cherry-pick packages to report

---

## Summary: New Checks to Add

These are the checks **NOT** in the existing GO_AUDITING_HEURISTICS.md:

1. ✅ **Documentation Accuracy** - TODO checklists, dependency claims
2. ✅ **Build System Integrity** - Platform-specific commands, missing scripts
3. ✅ **Unused Infrastructure** - DB fields, struct fields, tested-but-unused functions
4. ✅ **Stub Implementations** - Flags that don't actually work
5. ✅ **Undocumented Features** - Config options not in README
6. ✅ **Missing Interfaces** - All concrete types
7. ✅ **Global State** - Package-level variables
8. ✅ **Context in Goroutines** - context.Background() instead of parent
9. ✅ **HTTP Connection Pooling** - Default transport settings
10. ✅ **Coverage Reporting Tricks** - Excluding packages

---

## Recommended Integration

Add these as new sections to GO_AUDITING_HEURISTICS.md:

1. New section: **Documentation Integrity** (before or after Code Quality)
2. New section: **Build & Infrastructure** (after Automated Tools)
3. Expand **Performance** section with HTTP pooling and prepared statements
4. Add **Context Propagation** to Concurrency Patterns
5. Add **Unused Code Detection** to Code Quality

---

**Total new checks:** 17
**Estimated time to add:** 2-3 hours
**Value:** High - catches issues that automated tools miss
