# Go Code Auditing Heuristics

**Purpose:** Comprehensive checklist for auditing Go codebases
**Audience:** Developers, code reviewers, AI agents, security auditors
**Version:** 2.0
**Last Updated:** 2025-10-26

---

## Table of Contents

1. [Quick Start Checklist](#quick-start-checklist)
2. [Documentation Integrity](#documentation-integrity)
3. [Security Auditing](#security-auditing)
4. [Resource Management](#resource-management)
5. [Concurrency Patterns](#concurrency-patterns)
6. [Error Handling](#error-handling)
7. [Code Quality](#code-quality)
8. [Architecture & Design](#architecture--design)
9. [Build & Infrastructure](#build--infrastructure)
10. [Test Quality](#test-quality)
11. [Performance](#performance)
12. [Automated Tools](#automated-tools)
13. [Priority Matrix](#priority-matrix)

---

## Quick Start Checklist

Use this for rapid assessment of any Go codebase:

### Critical (Must Check)
- [ ] No SQL injection (parameterized queries only)
- [ ] No XSS vulnerabilities (use `html/template`, sanitize user content)
- [ ] No SSRF vulnerabilities (validate URLs, block private IPs)
- [ ] All resources properly closed (files, HTTP connections, database)
- [ ] No race conditions (`go test -race` passes)
- [ ] Errors are not ignored (search for `_ = `)
- [ ] No panics in library code
- [ ] TODO/checklist items are actually complete (not aspirational)
- [ ] All referenced files exist (documentation, Makefiles, examples)
- [ ] Version numbers consistent (code, CHANGELOG, README)
- [ ] Implemented features documented (README, CHANGELOG, examples/)

### High Priority (Should Check)
- [ ] Errors wrapped with context (`%w` verb)
- [ ] Test coverage >80% for critical packages
- [ ] No code duplication >10 lines
- [ ] Test assertions are effective (not just `t.Log()`)
- [ ] No functions >100 lines

### Medium Priority (Good to Check)
- [ ] Function complexity reasonable (cyclomatic <15)
- [ ] No flaky tests (run with `-count=10`)
- [ ] Proper mutex usage for shared state
- [ ] Channel operations follow best practices
- [ ] No dead code (unused functions/variables)
- [ ] Build targets work on clean checkout
- [ ] Config options documented in README
- [ ] HTTP client uses connection pooling
- [ ] Context passed to goroutines (not context.Background())

---

## Documentation Integrity

**Problem:** Documentation claims features are complete when they're not implemented, or references files that don't exist.

### TODO/Checklist Accuracy

**Heuristic:** Verify all ✅ checkmarks are actually true

**Commands:**
```bash
# Find all claimed completions
grep -rn "✅\|✓\|\[x\]" TODO.md ROADMAP.md CHECKLIST.md specs/

# For each claimed item, verify:
# 1. Does the file exist?
# 2. Is the dependency in go.mod?
# 3. Is the feature actually implemented?
```

**Real-world example:**
```
TODO.md claimed:
✅ QUICKSTART.md - 5-minute setup guide     → FILE DOESN'T EXIST
✅ THEMES.md - Complete theme guide         → FILE DOESN'T EXIST
✅ golang.org/x/time/rate - Rate limiting   → NOT IN go.mod, NEVER IMPORTED
```

**What to verify:**
- Every checked item is actually complete
- Files claimed to exist are present
- Dependencies claimed to exist are in go.mod
- Features claimed to work are implemented

**Prevention:**
- Automate checklist verification in CI
- Link checklist items to actual files/commits
- Use `test -f file.md || exit 1` in validation scripts

---

### Dependency Documentation Accuracy

**Heuristic:** Compare documented dependencies vs actual imports

**Commands:**
```bash
# List documented dependencies
grep -E "^- .*:" README.md CLAUDE.md docs/ | grep -oE "`[^`]+`" | sort -u

# List actual dependencies
go list -m all | cut -d' ' -f1 | sort -u

# Find documented but not used
comm -23 <(grep -oE "`[a-z]+\.[a-z]+/[^`]+" README.md | tr -d '`' | sort -u) \
         <(go list -m all | cut -d' ' -f1 | sort -u)
```

**What to verify:**
- Every documented dependency is in go.mod
- Every go.mod dependency has a purpose (or remove it)
- No "planned features" listed as "Key Dependencies"
- Dependency versions match if specified

**Real-world example:**
```
CLAUDE.md claimed:
"Rate limiting: golang.org/x/time/rate"

Reality:
$ grep "time/rate" go.mod
(nothing - dependency doesn't exist)
```

**Fix:** Move to "Planned Dependencies" section or remove

---

### Referenced Files Must Exist

**Heuristic:** Find all file references in documentation, verify they exist

**Command:**
```bash
# Find all file references in markdown
grep -rh -oE '[a-zA-Z0-9_/-]+\.(go|md|sh|ini|yaml|json|txt)' *.md docs/ specs/ | sort -u > /tmp/referenced.txt

# Check existence
while read file; do
    [ ! -f "$file" ] && echo "MISSING: $file (referenced in docs)"
done < /tmp/referenced.txt
```

**What breaks commonly:**
- Example files referenced but not created
- Scripts in Makefile that don't exist
- Documentation cross-references to files not yet written
- Old references to files that were renamed/deleted

**Real-world example:**
```
examples/README.md references:
- examples/config.ini     → DOESN'T EXIST (breaks `make examples`)
- QUICKSTART.md           → DOESN'T EXIST (404 link)
- THEMES.md               → DOESN'T EXIST (404 link)

Makefile line 172:
./setup-example-planet.sh → DOESN'T EXIST (make target fails)
```

**Fix:** Create missing files OR update documentation to remove references

---

### Version Number Consistency

**Heuristic:** Version constants in code must match documentation and changelog

**Problem:** Code declares one version, but docs claim another version

**Detection:**
```bash
# Find version constants in code
grep -rn "const version\|Version.*=.*\"[0-9]" --include="*.go" cmd/ main.go

# Compare to CHANGELOG.md
head -50 CHANGELOG.md | grep "^\#\# \["

# Compare to README.md
grep -i "version\|release\|v[0-9]\.[0-9]" README.md | head -5
```

**Real-world example:**
```
cmd/rp/main.go:10:    const version = "0.3.0"
CHANGELOG.md:27:      ## [0.4.0] - 2025-10-30
README.md:5:          Development Release v0.4.0

Three different version numbers!
```

**What to verify:**
- Version constant in code matches CHANGELOG latest release
- README development status matches current version
- Git tags match release versions
- Binary output (`./app version`) shows correct number

**Impact:**
- Users confused about what version they're running
- Bug reports reference wrong version
- Deployment confusion (deploying wrong version)
- Release notes don't match artifacts

**Prevention:**
- Automated check in CI: extract version from code, verify matches CHANGELOG
- Single source of truth (generate CHANGELOG from code or vice versa)
- Pre-release checklist includes version number verification

---

### Documentation Synchronization After Features

**Heuristic:** Implementing a feature requires updating multiple documentation files

**Problem:** Feature is implemented but documentation in various locations is stale

**Checklist when implementing new features:**
```bash
# 1. Check if feature mentioned in specs/plans
grep -rn "new_feature\|NewFeature" specs/*.md

# 2. Update main README.md (features list, configuration, examples)
grep -rn "new_feature" README.md || echo "Not in README!"

# 3. Update CHANGELOG.md (added/changed/fixed sections)
head -50 CHANGELOG.md | grep "new_feature" || echo "Not in CHANGELOG!"

# 4. Update CLAUDE.md if it affects dev workflow
grep -rn "new_feature" CLAUDE.md || echo "Not in CLAUDE!"

# 5. Update TODO.md status (planned → completed)
grep -rn "new_feature" specs/TODO.md

# 6. Update relevant spec documents
grep -rn "new_feature" specs/*-plan.md

# 7. Update examples/config.ini if config option added
grep -rn "new_feature\|new_config_option" examples/*.ini || echo "Not in examples!"
```

**Real-world example from v0.4.0 rate limiting:**
```
Files that needed updates:
✓ pkg/ratelimit/ratelimit.go (new package)
✓ pkg/config/config.go (config fields)
✓ cmd/rp/commands.go (integration)
✓ examples/config.ini (documented options)
✓ README.md (features list)
✓ CLAUDE.md (implementation notes)
✓ CHANGELOG.md (v0.4.0 section)
✗ specs/TODO.md (still said "not implemented") ← MISSED
✗ specs/v0.4.0-plan.md (still said "not implemented") ← MISSED
✗ specs/NETWORKING_FEATURES_STATUS.md (outdated) ← MISSED
```

**Minimum documentation checklist:**
1. **CHANGELOG.md** - Required for all changes
2. **README.md** - Required if user-visible
3. **examples/config.ini** - Required if config option added
4. **specs/TODO.md** - Update status if feature was planned
5. **CLAUDE.md** - Update if dev workflow changes
6. **Spec documents** - Mark complete or add completion notes

**Prevention:**
- Documentation update checklist in PR template
- Automated check: feature branch name matches CHANGELOG entry
- Final review step: "Have you updated all docs?"

---

### Spec Document Lifecycle Management

**Heuristic:** Planning documents should indicate their status (planned/in-progress/complete)

**Problem:** Specification documents become stale after features are implemented

**Detection:**
```bash
# Find spec/plan documents
find specs/ -name "*-plan.md" -o -name "*-spec.md" -o -name "TODO.md"

# Check if they have status indicators
for file in specs/*-plan.md; do
    echo "=== $file ==="
    head -10 "$file" | grep -i "status\|complete\|done\|finished" || echo "NO STATUS"
done
```

**Good practice - status header:**
```markdown
# Feature X Implementation Plan

**Status**: ✅ COMPLETED (2025-10-30)
**Target**: v0.4.0
**Dependencies**: None

## What Was Implemented

- [x] Component A (completed 2025-10-28)
- [x] Component B (completed 2025-10-29)
- [x] Component C (completed 2025-10-30)

## What Changed From Plan

- Decided to use library X instead of Y
- Added integration test coverage
- Deferred feature D to v1.0

---

[Original plan content below]
```

**Real-world example:**
```markdown
specs/v0.4.0-plan.md currently says:

**Status**: Planning          ← WRONG (it's complete!)
**Target**: Address critical issues

Should say:

**Status**: ✅ COMPLETED (2025-10-30)
**Actual Release**: v0.4.0 on 2025-10-30

## Implementation Summary
- P2.3 (Rate limiting): Implemented pkg/ratelimit/
- P2.4 (301 redirects): Implemented UpdateFeedURL()
- P3.2 (301 auto-update): Integrated into fetchFeeds
```

**Status types:**
- `📋 PLANNED` - Not started, spec complete
- `🔧 IN PROGRESS` - Currently being implemented
- `✅ COMPLETED (date)` - Finished, note completion date
- `❌ CANCELLED` - Not doing, explain why
- `⏸️ DEFERRED TO v#.#` - Postponed, note which version

**Impact of not managing:**
- False expectations (users think it's still planned)
- Duplicate work (someone implements it again)
- Lost context (why was it done this way?)
- Documentation debt accumulates

**Prevention:**
- Update spec status in same PR as feature
- Quarterly spec review: mark old plans as complete/cancelled
- Template for spec documents includes status field

---

## Security Auditing

### SQL Injection Prevention

**Heuristic:** Search for SQL query construction patterns

**Bad Patterns:**
```go
// VULNERABLE - Never do this
query := fmt.Sprintf("SELECT * FROM users WHERE id = %d", userID)
query := "INSERT INTO items VALUES ('" + name + "')"
```

**Good Pattern:**
```go
// SAFE - Always use parameterized queries
query := "SELECT * FROM users WHERE id = ?"
rows, err := db.Query(query, userID)

// Or with named parameters
query := "INSERT INTO items (name, value) VALUES (?, ?)"
result, err := db.Exec(query, name, value)
```

**Detection:**
```bash
# Find potential SQL injection risks
grep -rn "fmt.Sprintf.*SELECT\|INSERT\|UPDATE\|DELETE" --include="*.go"
grep -rn '"SELECT.*" +' --include="*.go"
```

**What to verify:**
- All queries use `?` placeholders or `$1`, `$2` style
- No string concatenation in SQL construction
- User input never directly interpolated into SQL

---

### XSS Prevention

**Heuristic:** Check template usage and HTML sanitization

**Bad Patterns:**
```go
// VULNERABLE - text/template doesn't escape
import "text/template"
tmpl.Execute(w, userContent)

// VULNERABLE - no sanitization
html := "<div>" + userInput + "</div>"
```

**Good Patterns:**
```go
// SAFE - html/template auto-escapes
import "html/template"
tmpl.Execute(w, userContent)

// SAFE - sanitize user HTML
import "github.com/microcosm-cc/bluemonday"
policy := bluemonday.UGCPolicy()
clean := policy.Sanitize(userHTML)
```

**Detection:**
```bash
# Check template imports
grep -rn '"text/template"' --include="*.go"

# Find HTML construction without sanitization
grep -rn '= "<[^"]*" +' --include="*.go"
```

**What to verify:**
- Using `html/template` for user-provided content
- HTML from users is sanitized (bluemonday or similar)
- Content Security Policy headers present
- JavaScript: only allows http/https URLs in src/href

---

### SSRF Prevention

**Heuristic:** Validate all URLs before fetching

**Bad Pattern:**
```go
// VULNERABLE - fetches any URL
resp, err := http.Get(userProvidedURL)
```

**Good Pattern:**
```go
// SAFE - validate before fetching
func ValidateURL(urlStr string) error {
    u, err := url.Parse(urlStr)
    if err != nil {
        return err
    }

    // Only allow http/https
    if u.Scheme != "http" && u.Scheme != "https" {
        return fmt.Errorf("invalid scheme: %s", u.Scheme)
    }

    // Resolve hostname to IPs
    ips, err := net.LookupIP(u.Hostname())
    if err != nil {
        return err
    }

    // Block private/internal IPs
    for _, ip := range ips {
        if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
            return fmt.Errorf("private IP not allowed")
        }
    }

    return nil
}
```

**Detection:**
```bash
# Find HTTP fetches
grep -rn "http.Get\|http.Post\|http.Do" --include="*.go"
```

**What to verify:**
- All user-provided URLs validated before use
- Blocks localhost (127.0.0.1, ::1)
- Blocks private IP ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
- Blocks link-local (169.254.0.0/16, fe80::/10)
- Only allows http/https schemes

---

## Resource Management

### File Handles

**Heuristic:** Every `os.Open()` must have `defer f.Close()`

**Bad Pattern:**
```go
// LEAK - file never closed
f, err := os.Open(path)
if err != nil {
    return err
}
data, _ := io.ReadAll(f)
return nil  // f still open!
```

**Good Pattern:**
```go
// SAFE - deferred close with error check
func readFile(path string) (data []byte, err error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer func() {
        if closeErr := f.Close(); closeErr != nil && err == nil {
            err = fmt.Errorf("close: %w", closeErr)
        }
    }()

    data, err = io.ReadAll(f)
    return data, err
}
```

**Detection:**
```bash
# Find file opens
grep -n "os.Open\|os.Create\|os.OpenFile" *.go

# Then manually verify each has corresponding defer
```

**What to verify:**
- Every file open has `defer f.Close()`
- Defer checks for nil before closing
- Close errors are captured and returned
- Uses named return values when capturing close errors

---

### HTTP Response Bodies

**Heuristic:** Every `http.Do()` must have `defer resp.Body.Close()`

**Bad Pattern:**
```go
// LEAK - body never closed
resp, err := http.Get(url)
if err != nil {
    return err
}
body, _ := io.ReadAll(resp.Body)
```

**Good Pattern:**
```go
// SAFE - always close body
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()

body, err := io.ReadAll(resp.Body)
```

**Detection:**
```bash
# Find HTTP calls
grep -n "http.Get\|http.Post\|http.Do\|client.Do" *.go

# Verify each has defer resp.Body.Close()
```

**What to verify:**
- Every HTTP request closes response body
- Body closed even when status code indicates error
- Defer happens immediately after error check

---

### Database Resources

**Heuristic:** Check `rows.Close()`, `stmt.Close()`, connection cleanup

**Bad Pattern:**
```go
// LEAK - rows never closed
rows, err := db.Query("SELECT * FROM users")
for rows.Next() {
    // process
}
// rows still open!
```

**Good Pattern:**
```go
// SAFE - always close rows
rows, err := db.Query("SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    // process
}

// Always check for errors from iteration
if err := rows.Err(); err != nil {
    return err
}
```

**Detection:**
```bash
# Find database queries
grep -n "db.Query\|db.QueryRow" *.go

# Verify defer rows.Close() and rows.Err() check
```

**What to verify:**
- `rows.Close()` called with defer
- `rows.Err()` checked after iteration
- Prepared statements closed when no longer needed
- Database connections properly managed (connection pool)

---

## Concurrency Patterns

### Race Condition Detection

**Heuristic:** Run `go test -race` on all packages

**Command:**
```bash
go test -race ./...
```

**What to look for:**
- Test output should show no "WARNING: DATA RACE" messages
- Pay special attention to packages with goroutines

**Manual checks:**
- Shared state access protected by mutex
- WaitGroup usage for goroutine synchronization
- Proper channel usage (no sending on closed channels)

---

### Mutex Usage

**Heuristic:** Shared state must be protected

**Bad Pattern:**
```go
// RACE - unprotected shared state
var counter int

func increment() {
    counter++  // Multiple goroutines = race!
}
```

**Good Pattern:**
```go
// SAFE - mutex protects shared state
var (
    mu      sync.Mutex
    counter int
)

func increment() {
    mu.Lock()
    counter++
    mu.Unlock()
}

// Or use defer for safety
func increment() {
    mu.Lock()
    defer mu.Unlock()
    counter++
}
```

**What to verify:**
- Shared variables accessed by multiple goroutines have mutex protection
- Mutex is locked before access, unlocked after
- Consider using `defer mu.Unlock()` to prevent forgetting
- Watch for long critical sections (mutex held too long)

---

### Goroutine Lifecycle

**Heuristic:** All goroutines should have clear termination

**Bad Pattern:**
```go
// LEAK - goroutine runs forever
func startWorker() {
    go func() {
        for {
            // No way to stop!
            doWork()
        }
    }()
}
```

**Good Pattern:**
```go
// SAFE - context-based cancellation
func startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                doWork()
            }
        }
    }()
}
```

**What to verify:**
- Goroutines have termination condition (context, done channel)
- WaitGroup used to ensure completion
- No goroutine leaks (use pprof to detect)

---

### Context Propagation in Goroutines

**Heuristic:** Goroutines should receive context from parent, not create context.Background()

**Bad Pattern:**
```go
// PROBLEM - can't cancel from parent
func fetchFeeds(feeds []Feed) {
    for _, feed := range feeds {
        go func(f Feed) {
            // Creates independent context!
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()
            crawler.Fetch(ctx, f.URL)
        }(feed)
    }
}
```

**Good Pattern:**
```go
// CORRECT - context from parent
func fetchFeeds(ctx context.Context, feeds []Feed) error {
    for _, feed := range feeds {
        go func(f Feed) {
            // Derives from parent context
            fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
            defer cancel()
            crawler.Fetch(fetchCtx, f.URL)
        }(feed)
    }

    // Can be cancelled from parent
    <-ctx.Done()
    return ctx.Err()
}
```

**Detection:**
```bash
# Find goroutines with context.Background()
grep -B 2 -A 10 "go func" --include="*.go" -r . | grep "context.Background()"

# Should find very few (only in main() or tests)
```

**What's wrong with Background() in goroutines:**
- User hits Ctrl+C → main context cancelled
- Goroutines with Background() keep running
- No graceful shutdown
- Can't enforce timeouts from parent

**What to verify:**
- Functions spawning goroutines accept context.Context parameter
- No context.Background() inside goroutine functions
- Context passed to all long-running operations
- Main function sets up signal handling with context

---

## Error Handling

### Ignored Errors

**Heuristic:** Search for `_ = ` pattern

**Command:**
```bash
# Find ignored errors
grep -rn "_ = " --include="*.go"
```

**Bad Pattern:**
```go
// IGNORED - error silently discarded
_ = file.Close()
_ = json.Unmarshal(data, &result)
_ = rows.Close()
```

**Acceptable Pattern:**
```go
// OK - explicitly documented why ignored
_ = conn.Close() // Already have error from read, this is cleanup
```

**Better Pattern:**
```go
// BEST - check and handle
if err := file.Close(); err != nil {
    log.Printf("Warning: failed to close file: %v", err)
}
```

**What to verify:**
- Every `_ = ` is intentional and documented
- Critical operations (parse, unmarshal, close) have errors checked
- Defer close operations capture errors if possible

---

### Error Wrapping

**Heuristic:** Errors should preserve context with `%w`

**Bad Pattern:**
```go
// LOSES CONTEXT
if err != nil {
    return fmt.Errorf("operation failed")
}

// CAN'T UNWRAP
if err != nil {
    return fmt.Errorf("operation failed: %v", err)
}
```

**Good Pattern:**
```go
// PRESERVES CONTEXT
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// Can later use errors.Is() or errors.As()
```

**Detection:**
```bash
# Find error formatting
grep -rn "fmt.Errorf" --include="*.go"

# Check they use %w not %v for wrapped errors
```

**What to verify:**
- Errors wrapped with `%w` to preserve chain
- Error messages include context (what operation failed)
- Sentinel errors defined as variables, not inline strings

---

### Panic Usage

**Heuristic:** Library code should never panic

**Detection:**
```bash
# Find panics
grep -rn "panic(" --include="*.go"
```

**What to verify:**
- `panic()` only in main package or init functions
- Library packages return errors, don't panic
- Use `recover()` only in appropriate places (HTTP handlers, worker pools)

---

## Code Quality

### Code Duplication

**Heuristic:** Look for repeated patterns >10 lines

**Manual detection:**
1. Compare similar function implementations
2. Look for copy-pasted code with minor variations
3. Check for repeated initialization patterns
4. Review test setup code

**Common patterns to extract:**
- Config/database initialization
- Progress reporting loops
- Validation logic
- Error handling boilerplate

**Example refactoring:**
```go
// BEFORE - duplicated 10 times
cfg, err := loadConfig(path)
if err != nil {
    return err
}
db, err := openDB(cfg.DBPath)
if err != nil {
    return err
}
defer db.Close()

// AFTER - extracted helper
db, cleanup, err := openDatabase(path)
if err != nil {
    return err
}
defer cleanup()
```

**Tool-based detection:**
```bash
# Use duplication detection tools
dupl -t 50 .           # Find blocks >50 tokens
gocyclo -over 10 .     # Find complex functions
```

---

### Function Complexity

**Heuristic:** Measure lines of code and cyclomatic complexity

**Thresholds:**
- Lines of code: <50 good, 50-100 warning, >100 critical
- Cyclomatic complexity: <10 good, 10-15 warning, >15 critical
- Nesting depth: <3 good, 3-4 warning, >4 critical

**Manual measurement:**
```bash
# Count lines in functions (approximate)
grep -n "^func " file.go   # Find function starts

# Use tools for precise measurement
gocyclo -over 10 .
gocognit -over 15 .
```

**What to look for:**
- Functions doing multiple things (split responsibilities)
- Deep nesting (extract helper functions)
- Long parameter lists (>5 parameters = consider struct)
- Functions with multiple return values (>3 = reconsider design)

**Refactoring approach:**
- Extract helper functions
- Use early returns to reduce nesting
- Split large functions by responsibility

---

### Unused Database Columns

**Heuristic:** Find database columns that are written but never queried

**Process:**
```bash
# 1. Extract column names from schema
grep -E "^\s+[a-z_]+ (INTEGER|TEXT|BLOB|REAL)" schema.sql | awk '{print $1}' > /tmp/columns.txt

# 2. For each column, search if it's queried
while read col; do
    grep -r "SELECT.*\b$col\b" . --include="*.go" || echo "NEVER READ: $col"
done < /tmp/columns.txt

# 3. Check WHERE clauses specifically
while read col; do
    grep -r "WHERE.*\b$col\b" . --include="*.go" || echo "NEVER FILTERED ON: $col"
done < /tmp/columns.txt
```

**Real-world example:**
```sql
-- Schema defines:
CREATE TABLE feeds (
    next_fetch TEXT,           -- Written on every update
    fetch_interval INTEGER     -- Written on every update
);

-- Code writes:
UPDATE feeds SET next_fetch = ?, fetch_interval = ? ...

-- But queries ignore it:
SELECT * FROM feeds WHERE active = 1  -- Ignores next_fetch!

-- Should be:
SELECT * FROM feeds WHERE active = 1 AND next_fetch <= ?
```

**What this reveals:**
- Infrastructure for unimplemented features
- Fields that should be removed
- Missed optimization opportunities
- Planned features never finished

**Fix:** Implement the feature OR remove unused columns

---

### Undocumented Future/Reserved Fields

**Heuristic:** Database columns or struct fields reserved for future use should have TODO comments

**Problem:** Fields exist and are populated but never used - no explanation why

**Detection:**
```bash
# 1. Find database columns from schema
grep -E "^\s+[a-z_]+ (INTEGER|TEXT)" pkg/repository/*.go | grep "CREATE TABLE" -A 20

# 2. Search for each column in WHERE/SELECT clauses
# If a column is written but never queried, it needs a comment

# 3. Check struct fields for TODO comments
grep -B 2 -A 2 "NextFetch\|FetchInterval\|Scheduled" pkg/repository/*.go
```

**Bad example - no explanation:**
```go
type Feed struct {
    ID              int64
    URL             string
    NextFetch       time.Time    // Why is this here?
    FetchInterval   int          // Never used anywhere!
}
```

**Good example - documented intent:**
```go
type Feed struct {
    ID              int64
    URL             string
    NextFetch       time.Time    // TODO(v1.0): Used for intelligent scheduling (not yet implemented)
    FetchInterval   int          // seconds - TODO(v1.0): Used for adaptive polling (not yet implemented)
}
```

**What the TODO comment should explain:**
- Why the field exists if not currently used
- What version/feature will use it
- Whether it's safe to populate it now (infrastructure for future)
- Link to spec/plan document if applicable

**Real-world example:**
```go
// Database schema has:
CREATE TABLE feeds (
    next_fetch TEXT,           -- Populated on every fetch
    fetch_interval INTEGER     -- Set to 3600 (1 hour)
);

// But application code does:
SELECT * FROM feeds WHERE active = 1  -- Ignores next_fetch!

// Should have:
NextFetch time.Time  // TODO(v1.0): Intelligent scheduling - see specs/v1.0.0-plan.md Phase 2
```

**Impact of not documenting:**
- Developers don't know if field is abandoned or planned
- May delete field thinking it's dead code
- May implement feature differently than planned
- QA doesn't know if it's a bug or intended

**Prevention:**
- Code review checklist: unused fields have TODO comments
- Grep for struct fields that match unimplemented features
- Link TODO comments to spec documents for context

---

### Unused Struct Fields

**Heuristic:** Find struct fields that are set but never read

**Process:**
```bash
# 1. Find all struct definitions
grep -n "type.*struct {" --include="*.go" -r .

# 2. For each field, search for reads (manual review)
# Look for: variable.FieldName
# Exclude: variable.FieldName = (this is a write)
```

**Real-world example:**
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
{{if gt .Subscribers 0}}
    Subscribers: {{.Subscribers}}
{{end}}  // Never displays (always 0)
```

**What to verify:**
- Field is set somewhere (not always default value)
- Field is read somewhere (not just written)
- If unused, consider removing or documenting as TODO

**Tools:**
```bash
# Find struct fields that might be unused
structcheck ./...
ineffassign ./...
```

---

### Functions Tested But Unused in Production

**Heuristic:** Find exported functions only called in test files

**Process:**
```bash
# Find exported functions in package
grep "^func.*[A-Z].*(" pkg/**/*.go | grep -v "_test.go" | cut -d' ' -f2 | cut -d'(' -f1 > /tmp/funcs.txt

# For each function, check if used outside tests
while read func; do
    # Search in non-test files
    grep -r "\.$func\|$func(" --include="*.go" --exclude="*_test.go" -q || \
        echo "ONLY IN TESTS: $func"
done < /tmp/funcs.txt
```

**Real-world example:**
```go
// pkg/crawler/crawler.go
func (c *Crawler) FetchWithRetry(ctx context.Context, url string, maxRetries int) error {
    // Well-implemented, thoroughly tested
}

// But cmd/rp/commands.go uses:
resp, err := c.Fetch(ctx, url, cache)  // Not FetchWithRetry!

// FetchWithRetry only called in:
// - crawler_test.go (100% coverage)
// - Never in production code
```

**When this is OK:**
- Public API for library users
- Future feature being built incrementally
- Explicitly documented as "available but not used internally"

**When this is a problem:**
- Code duplication (retry logic implemented twice)
- Missed quality improvements
- Dead code that should be removed

---

### Stub Implementations (Dry-Run Flags That Don't Work)

**Heuristic:** Find functions with `--dry-run` flags that don't actually preview

**Pattern:**
```bash
# Find dry-run parameters
grep -rn "DryRun.*bool\|dryRun.*bool\|dry-run" --include="*.go"

# Check if they query before reporting
grep -B 5 -A 15 "if.*DryRun" --include="*.go" -r . | grep "Count\|Query\|List"
```

**Bad implementation:**
```go
func cmdPrune(opts PruneOptions) error {
    if opts.DryRun {
        fmt.Printf("Dry run: would delete entries older than %d days\n", opts.Days)
        return nil  // Doesn't actually check what would be deleted!
    }
    // ... actual deletion
}
```

**Good implementation:**
```go
func cmdPrune(opts PruneOptions) error {
    cutoff := time.Now().AddDate(0, 0, -opts.Days)

    if opts.DryRun {
        // Actually query what would be deleted
        count, err := repo.CountEntriesOlderThan(cutoff)
        if err != nil {
            return err
        }
        fmt.Printf("Would delete %d entries older than %s\n", count, cutoff)

        // Optionally show sample entries
        preview, _ := repo.GetEntriesOlderThan(cutoff, 5)
        for _, entry := range preview {
            fmt.Printf("  - %s (%s)\n", entry.Title, entry.Published)
        }
        return nil
    }

    // Actual deletion
    return repo.DeleteEntriesOlderThan(cutoff)
}
```

**What to verify:**
- Dry-run actually queries database/filesystem
- Shows specific items that would be affected
- Same logic as real operation (just doesn't commit)

---

### Undocumented Configuration Options

**Heuristic:** Find all config fields vs documented options

**Process:**
```bash
# 1. Find all config struct fields
grep -A 50 "type.*Config struct" pkg/config/*.go | \
    grep -E "^\s+[A-Z]" | awk '{print $1}' | sort -u > /tmp/config_fields.txt

# 2. For each field, check if documented in README
while read field; do
    grep -q "$field\|$(echo $field | sed 's/\([A-Z]\)/_\L\1/g' | sed 's/^_//')" README.md || \
        echo "UNDOCUMENTED: $field"
done < /tmp/config_fields.txt
```

**Real-world example:**
```go
// pkg/config/config.go
type PlanetConfig struct {
    Days             int     // Documented ✓
    FilterByFirstSeen bool   // NOT in README ✗
    SortBy           string  // NOT in README ✗
}
```

**Impact:**
- Users don't know powerful features exist
- Features stay unused
- Support burden ("how do I sort by date?")
- Wasted development effort

**Fix:** Document ALL config options in README with:
- Option name and type
- Default value
- Description
- Example usage

---

## Architecture & Design

**Problem:** Hard-to-test code, global state, concrete dependencies

### Interfaces vs Concrete Types

**Heuristic:** Check if business logic depends on concrete types or interfaces

**Detection:**
```bash
# Find interface definitions in non-test files
grep -rn "type.*interface" pkg/ --include="*.go" | grep -v "_test.go" | wc -l

# If count is 0 or very low: architecture problem!
```

**Bad pattern (hard to test):**
```go
// Concrete dependencies
func ProcessFeeds(repo *repository.Repository, crawler *crawler.Crawler) error {
    feeds, err := repo.GetFeeds()  // Can't mock for testing!
    if err != nil {
        return err
    }

    for _, feed := range feeds {
        data, err := crawler.Fetch(feed.URL)  // Can't mock HTTP!
        if err != nil {
            continue
        }
        // process...
    }
    return nil
}
```

**Good pattern (easy to test):**
```go
// Interface dependencies
type FeedRepository interface {
    GetFeeds() ([]Feed, error)
    SaveEntry(entry *Entry) error
}

type FeedCrawler interface {
    Fetch(url string) (*FeedData, error)
}

func ProcessFeeds(repo FeedRepository, crawler FeedCrawler) error {
    // Same code, but can inject mocks in tests!
}
```

**Benefits:**
- Unit tests don't need real database
- Can test error paths easily (inject failing mock)
- Faster tests (no I/O)
- Dependency injection
- Easier to swap implementations

**Where to add interfaces:**
- Database operations (Repository)
- HTTP clients (Crawler, API clients)
- File I/O operations
- External services
- Time (for testing time-dependent code)

---

### Global State in Packages

**Heuristic:** Find package-level variables that hold mutable state

**Detection:**
```bash
# Find global variables (exclude error sentinels)
grep -rn "^var [a-z]" pkg/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "^var Err"
```

**Bad pattern:**
```go
// cmd/rp/commands.go
var globalLogger = &Logger{level: LogLevelInfo}

func (l *Logger) SetLevel(level string) {
    l.level = parseLevel(level)  // NOT THREAD-SAFE!
}

func cmdFetch(opts FetchOptions) error {
    globalLogger.SetLevel(opts.LogLevel)  // Mutates global state!
    // ...
}
```

**Problems:**
- Not thread-safe (data races)
- Makes testing harder (state leaks between tests)
- Can't have multiple instances with different config
- Harder to reason about program behavior

**Good pattern:**
```go
// Pass dependencies explicitly
type App struct {
    logger *Logger
    config *Config
}

func NewApp(cfg *Config) *App {
    return &App{
        logger: NewLogger(cfg.LogLevel),
        config: cfg,
    }
}

func (a *App) Fetch(opts FetchOptions) error {
    a.logger.Info("fetching feeds")
    // ...
}
```

**What to verify:**
- Package-level vars are truly constant (or sentinel errors)
- Mutable state is instance variables, not global
- Dependencies passed as parameters or struct fields

**Acceptable global state:**
- Sentinel errors: `var ErrNotFound = errors.New(...)`
- True constants (even if declared as var)
- sync.Once for one-time initialization
- Package-level loggers if explicitly documented as global

---

## Build & Infrastructure

**Problem:** Build scripts break on different platforms or reference missing files

### Platform-Specific Commands in Makefiles

**Heuristic:** Find commands that differ between macOS and Linux

**Detection:**
```bash
# Find macOS-specific sed
grep -n "sed -i ''" Makefile

# Find other platform-specific commands
grep -n "greadlink\|gsed\|ggrep\|gawk" Makefile

# Find Darwin/Linux detection
grep -n "uname.*Darwin" Makefile
```

**Bad pattern:**
```makefile
# BREAKS ON LINUX
examples:
	sed -i '' 's/foo/bar/' file.txt
```

**Portable alternatives:**
```makefile
# Option 1: Use .bak extension (works everywhere)
examples:
	sed -i.bak 's/foo/bar/' file.txt && rm file.txt.bak

# Option 2: Platform detection
ifeq ($(shell uname),Darwin)
    SED := sed -i ''
else
    SED := sed -i
endif

examples:
	$(SED) 's/foo/bar/' file.txt

# Option 3: Use Go instead of shell
examples:
	go run scripts/transform.go file.txt
```

**What to verify:**
- `sed`, `awk`, `grep` commands are portable
- No assumptions about GNU vs BSD tools
- Test on both macOS and Linux
- Consider using Go for complex scripts

**Common gotchas:**
- `sed -i ''` (macOS) vs `sed -i` (Linux)
- `readlink -f` (Linux only, use `greadlink` on macOS)
- `date` command syntax differences
- `find -printf` (GNU only)

---

### Makefile Target Dependencies and File Existence

**Heuristic:** Test every make target on clean checkout

**Process:**
```bash
# List all targets
make -qp | grep "^[a-zA-Z0-9_-]*:" | cut -d: -f1 | sort -u > /tmp/targets.txt

# Test each one in clean directory
for target in $(cat /tmp/targets.txt); do
    echo "Testing: make $target"
    make clean
    make $target || echo "❌ FAILED: $target"
done
```

**What breaks commonly:**
- Targets depending on files that don't exist
- Targets assuming previous targets ran
- Scripts referenced but not committed to git
- Hardcoded paths specific to one developer

**Real-world example:**
```makefile
# BREAKS - file doesn't exist
examples:
	@cp examples/config.ini /tmp/test.ini

# BREAKS - script not in repo
setup-example: build
	@./setup-example-planet.sh

# WORKS - checks existence first
examples:
	@test -f examples/config.ini || (echo "examples/config.ini missing" && exit 1)
	@cp examples/config.ini /tmp/test.ini
```

**Prevention:**
- Test make targets in CI (GitHub Actions, etc.)
- Use `.PHONY` for targets that don't create files
- Check file existence before operations
- Document target dependencies in comments

---

## Test Quality

### Test Coverage

**Heuristic:** Measure coverage for each package

**Command:**
```bash
# Get coverage for all packages
go test -cover ./...

# Generate detailed HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

**Targets:**
- Critical packages (security, database): >90%
- Library packages: >80%
- Command/CLI packages: >70%
- Overall: >80%

**What to check:**
- Coverage percentage by package
- Which lines are not covered
- Missing error path coverage
- Untested edge cases

---

### Assertion Density

**Heuristic:** Calculate assertions per test function

**Formula:**
```
assertion_density = total_assertions / total_test_functions
```

**Assertions include:**
- `t.Error()`, `t.Errorf()`, `t.Fatal()`, `t.Fatalf()`
- NOT `t.Log()`, `t.Logf()` (these don't fail tests)

**Quality grades:**
- <2.0: Weak (investigate)
- 2.0-4.0: Moderate
- 4.0-8.0: Good
- >8.0: Excellent

**How to count:**
```bash
# Count test functions
grep -c "^func Test" *_test.go

# Count assertions (approximate)
grep -c "t.Error\|t.Fatal" *_test.go
```

**Anti-patterns to find:**
```bash
# Tests using t.Log() instead of assertions
grep -n "t.Log.*if.*{" *_test.go

# Tests with no assertions
grep -A 20 "^func Test" *_test.go | grep -L "t.Error\|t.Fatal"
```

**What to verify:**
- Tests actually fail when code is wrong
- Both positive and negative cases tested
- Unused `want*` fields in table-driven tests
- `t.Log()` not used where `t.Error()` should be

---

### Flaky Tests

**Heuristic:** Run tests multiple times

**Command:**
```bash
# Run tests 10 times
go test -count=10 ./...

# Or in a loop
for i in {1..10}; do
    echo "=== Run $i ==="
    go test ./... || break
done
```

**What causes flakiness:**
- Race conditions (use `-race` to detect)
- Time-dependent logic (sleep, timeouts)
- Random data without seeds
- Network/filesystem dependencies
- Tests not properly isolated

**How to fix:**
- Use mocks instead of real network/filesystem
- Seed random number generators
- Use `time.After()` with reasonable timeouts
- Ensure test cleanup (use `t.TempDir()`, defer cleanup)
- Recognize expected errors (connection close during size limit tests)

---

### Skipped Tests

**Heuristic:** Find tests marked as skipped

**Command:**
```bash
# Find skipped tests
grep -rn "t.Skip(" --include="*_test.go"
```

**What to verify:**
- Why is test skipped? (documented in skip message)
- Is it temporary or permanent?
- Should it be fixed or removed?
- Is there alternative coverage?

**Valid reasons to skip:**
- Integration tests requiring external services (use build tags)
- Tests for unimplemented features (TODO)
- Platform-specific tests

---

### Integration Test Gaps Beyond Unit Tests

**Heuristic:** Unit tests passing doesn't prove end-to-end functionality works

**Problem:** Feature has excellent unit test coverage but may not work in production integration

**What unit tests DON'T prove:**
- Feature is actually wired into application (may be unused code)
- Configuration is properly passed from config file to implementation
- Feature works with real dependencies (database, HTTP, etc.)
- Multiple components work together correctly

**Real-world example from rate limiting:**
```
✓ pkg/ratelimit/ratelimit_test.go - 11 unit tests, 100% coverage
  - Tests token bucket algorithm
  - Tests concurrency safety
  - Tests context cancellation

✗ No integration test proving:
  - Rate limiter is actually created in fetchFeeds
  - Config options (requests_per_minute, rate_limit_burst) are read
  - Multiple fetches to same domain are actually delayed
  - Rate limiting works end-to-end in real usage
```

**Detection strategy:**
```bash
# 1. Find well-tested packages
go test -cover ./... | grep "90\|100%"

# 2. For each, check if feature is used in main application
# Example: ratelimit package exists, but is it used?
grep -r "ratelimit\.New\|ratelimit\.Manager" cmd/ --include="*.go"

# 3. Check if integration tests exist
ls cmd/*_integration_test.go pkg/*_integration_test.go
grep -l "integration\|end.to.end\|e2e" *_test.go
```

**Checklist for "integrated" verification:**
- [ ] Feature is imported and used (not just tested)
- [ ] Config options reach the feature (not just parsed)
- [ ] Integration test exercises full path (config → code → behavior)
- [ ] Can demonstrate feature working in manual test
- [ ] Metrics/logs show feature is active in production

**When integration tests ARE needed:**
- New features that touch multiple components
- Configuration options (prove they're actually used)
- Rate limiting, retries, timeouts (behavioral features)
- Database migrations
- HTTP client behavior changes

**When unit tests alone are OK:**
- Pure functions (no external dependencies)
- Data transformations
- Parsing/serialization
- Utility libraries

**How to write integration tests:**
```go
func TestRateLimitingIntegration(t *testing.T) {
    // 1. Setup: Real config with rate limit settings
    cfg := &config.Config{
        Planet: config.PlanetConfig{
            RequestsPerMinute: 10,  // Very restrictive for testing
            RateLimitBurst:    2,
        },
    }

    // 2. Create real components (not mocks)
    // 3. Exercise the full path
    // 4. Verify behavior (timing, not just success)
    start := time.Now()
    // Fetch 5 feeds from same domain
    // Should take ~30 seconds with 10 req/min limit
    elapsed := time.Since(start)

    if elapsed < 20*time.Second {
        t.Error("Rate limiting not working - requests too fast")
    }
}
```

**Risk assessment:**
- **Low risk**: Feature is simple, obvious if broken (crashes immediately)
- **Medium risk**: Feature is behavioral, may silently not work (this is rate limiting!)
- **High risk**: Feature is security-critical (auth, validation)

**Prevention:**
- For medium/high risk features: require integration test
- Manual testing checklist before marking feature complete
- Observability: logs/metrics that prove feature is working
- Smoke tests in production deployment

---

## Performance

### Benchmarking

**Heuristic:** Benchmark critical paths

**Command:**
```bash
# Run benchmarks
go test -bench=. ./...

# With memory allocation stats
go test -bench=. -benchmem ./...
```

**What to benchmark:**
- Hot paths (frequently called functions)
- Parsing/serialization
- Database queries
- Cryptographic operations

---

### Profiling

**Heuristic:** Profile CPU and memory usage

**Commands:**
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Analyze allocations
go test -benchmem -bench=. | grep "allocs/op"
```

**What to look for:**
- Functions consuming most CPU time
- Excessive memory allocations
- String concatenation in loops (use strings.Builder)
- Regex compilation in loops (compile once, reuse)

---

### HTTP Client Connection Pooling

**Heuristic:** Check if http.Client uses optimized connection pool settings

**Problem:** Go's default HTTP transport only keeps 2 idle connections per host

**Detection:**
```bash
# Find http.Client creation
grep -rn "http.Client{" --include="*.go"

# Check if custom Transport is configured
grep -B 5 -A 10 "http.Client{" --include="*.go" | grep "Transport:"
```

**Default (slow for concurrent requests):**
```go
// Uses http.DefaultTransport with MaxIdleConnsPerHost = 2
client := &http.Client{
    Timeout: 30 * time.Second,
}
```

**Optimized for concurrency:**
```go
// Configure connection pooling
transport := &http.Transport{
    MaxIdleConns:        100,              // Total pool size
    MaxIdleConnsPerHost: 10,               // Per-host (5x default!)
    MaxConnsPerHost:     0,                // No limit on active connections
    IdleConnTimeout:     90 * time.Second, // Keep connections alive
    DisableCompression:  false,            // Enable gzip
    ForceAttemptHTTP2:   true,             // Use HTTP/2 when available
}

client := &http.Client{
    Timeout:   30 * time.Second,
    Transport: transport,
}
```

**Performance impact:**
- 10-20% faster for concurrent HTTP requests
- Reduces connection overhead
- Especially important for services fetching many URLs

**When to optimize:**
- Fetching multiple URLs concurrently
- Making many requests to same hosts
- Background workers, web scrapers, API clients

---

### Database Prepared Statement Reuse

**Heuristic:** Check if queries are prepared once or every call

**Problem:** Preparing statements on every query adds overhead

**Detection:**
```bash
# Find query patterns
grep -rn "db.Query\|db.QueryRow\|db.Exec" --include="*.go" | grep -v Prepare

# Check for statement reuse
grep -rn "stmt.*sql.Stmt" --include="*.go"
```

**Bad pattern (prepares every call):**
```go
func (r *Repo) GetUser(id int) (*User, error) {
    // Query string parsed EVERY call
    row := r.db.QueryRow("SELECT * FROM users WHERE id = ?", id)
    // ...
}
```

**Good pattern (prepare once, reuse):**
```go
type Repo struct {
    db   *sql.DB
    stmt struct {
        getUser    *sql.Stmt
        insertUser *sql.Stmt
        updateUser *sql.Stmt
    }
}

func (r *Repo) init() error {
    var err error

    // Prepare statements once
    r.stmt.getUser, err = r.db.Prepare("SELECT * FROM users WHERE id = ?")
    if err != nil {
        return err
    }

    r.stmt.insertUser, err = r.db.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
    if err != nil {
        return err
    }

    return nil
}

func (r *Repo) GetUser(id int) (*User, error) {
    // Reuses prepared statement
    row := r.stmt.getUser.QueryRow(id)
    // ...
}

func (r *Repo) Close() error {
    r.stmt.getUser.Close()
    r.stmt.insertUser.Close()
    return r.db.Close()
}
```

**Performance impact:**
- 10-20% faster for frequent queries
- Lower database CPU usage
- Better for high-throughput applications

**When to optimize:**
- Frequently-called queries
- Hot paths (called in loops)
- Production services with high QPS

---

### Batch Database Operations

**Heuristic:** Look for INSERT/UPDATE in loops

**Problem:** One transaction per operation is slow

**Detection:**
```bash
# Find loops with database operations
grep -B 5 -A 5 "for.*range" --include="*.go" -r . | grep -E "db.Exec|stmt.Exec|Upsert|Insert|Update"
```

**Bad pattern (N transactions):**
```go
// SLOW - one transaction per entry
func SaveEntries(entries []Entry) error {
    for _, entry := range entries {
        _, err := repo.UpsertEntry(&entry)  // Each is a transaction!
        if err != nil {
            return err
        }
    }
    return nil
}
```

**Good pattern (single transaction):**
```go
// FAST - batch in one transaction
func SaveEntriesBatch(entries []Entry) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // No-op if committed

    stmt, err := tx.Prepare("INSERT OR REPLACE INTO entries (...) VALUES (?, ?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, entry := range entries {
        _, err := stmt.Exec(entry.Title, entry.Content, entry.Published)
        if err != nil {
            return err
        }
    }

    return tx.Commit()  // One commit for all
}
```

**Performance impact:**
- 5-10x faster for batch operations
- Reduces database I/O
- Atomic (all succeed or all fail)

**When to use:**
- Inserting multiple records
- Bulk updates
- Data migrations
- Import/export operations

**What to verify:**
- Loops with database operations use transactions
- Prepared statements reused in loop
- Transaction committed once after loop
- Proper error handling and rollback

---

### Test Coverage Exclusions

**Heuristic:** Verify what's included/excluded in coverage reports

**Problem:** Excluding packages to inflate coverage percentages

**Detection:**
```bash
# Get coverage for all packages
go test -coverprofile=coverage.out ./...

# See what's included
go tool cover -func=coverage.out | head -20

# Check for suspiciously high averages
go test -cover ./... | grep "coverage:"

# Compare package-by-package vs overall
```

**What to verify:**
- All packages included in coverage report (including cmd/)
- Exclusions are documented and justified
- Coverage percentages not cherry-picked
- Low-coverage packages not hidden

**Real-world example:**
```
Documentation claims: "88.4% average coverage (Excellent)"

Reality:
pkg/crawler    96.6%  ✓ Included
pkg/config     93.8%  ✓ Included
pkg/normalizer 80.1%  ✓ Included
cmd/rp         26.6%  ✗ EXCLUDED from "average"

Actual average: (96.6 + 93.8 + 80.1 + 26.6) / 4 = 74.3%
Reported average: (96.6 + 93.8 + 80.1) / 3 = 90.2%
```

**Honest reporting:**
- Include ALL packages (libraries AND commands)
- Note exclusions explicitly: "76% average (excluding cmd/)"
- Separate library coverage from CLI coverage
- Don't hide low-coverage areas

**When exclusions are OK:**
- Generated code (explicitly marked)
- Vendor directories
- Test utilities (in testutil/ packages)
- Platform-specific code (if documented)

---

## Automated Tools

### Essential Tools

**Run on every audit:**
```bash
# Static analysis
go vet ./...

# Race detection
go test -race ./...

# Coverage
go test -cover ./...

# Comprehensive linting
golangci-lint run
```

### Security Tools

```bash
# Security-focused linting
gosec ./...

# Known vulnerability scanning
govulncheck ./...

# Dependency vulnerabilities
nancy go.sum
```

### Code Quality Tools

```bash
# Cyclomatic complexity
gocyclo -over 10 .

# Cognitive complexity
gocognit -over 15 .

# Code duplication
dupl -t 50 .

# Dead code detection
deadcode ./...

# Unused variables/constants
varcheck ./...
structcheck ./...
```

---

## Priority Matrix

| Check Type | Priority | Effort | Impact if Missed |
|------------|----------|--------|------------------|
| SQL injection | **Critical** | Low | Data breach |
| XSS vulnerabilities | **Critical** | Low | User compromise |
| SSRF | **Critical** | Low | Internal network access |
| Race conditions | **Critical** | Medium | Data corruption |
| Resource leaks | **High** | Low | Memory/file exhaustion |
| Ignored errors | **High** | Medium | Silent failures |
| Test coverage | **High** | Medium | Bugs in production |
| TODO accuracy | **High** | Low | False completion status |
| Referenced files exist | **High** | Low | Build failures |
| Build portability | **High** | Low | Broken on other platforms |
| Code duplication | **Medium** | Medium | Maintenance burden |
| Function complexity | **Medium** | Low | Hard to modify |
| Test assertion quality | **Medium** | Medium | False confidence |
| Flaky tests | **Medium** | Medium | CI/CD unreliability |
| Context propagation | **Medium** | Low | No graceful shutdown |
| Unused infrastructure | **Medium** | Medium | Wasted effort |
| Undocumented features | **Medium** | Low | Users miss features |
| Missing interfaces | **Medium** | High | Hard to test |
| Global state | **Medium** | Medium | Thread safety issues |
| HTTP connection pooling | **Low** | Low | 10-20% slower |
| DB prepared statements | **Low** | Medium | 10-20% slower |
| Batch operations | **Low** | Medium | 5-10x slower |
| Coverage exclusions | **Low** | Low | Misleading metrics |
| Performance | **Low** | High | Slow applications |
| Dead code | **Low** | Low | Code bloat |

---

## Audit Workflow Template

### Phase 1: Automated Checks (1 hour)
```bash
#!/bin/bash
# Run all automated checks

echo "=== Running go vet ==="
go vet ./...

echo "=== Running race detector ==="
go test -race ./...

echo "=== Checking coverage ==="
go test -cover ./...

echo "=== Running linter ==="
golangci-lint run

echo "=== Security scan ==="
gosec ./...

echo "=== Detecting flaky tests ==="
go test -count=10 ./...
```

### Phase 2: Pattern Matching (2 hours)
```bash
#!/bin/bash
# Search for common issues

echo "=== Searching for ignored errors ==="
grep -rn "_ = " --include="*.go"

echo "=== Searching for SQL injection risks ==="
grep -rn "fmt.Sprintf.*SELECT\|INSERT\|UPDATE\|DELETE" --include="*.go"

echo "=== Checking template usage ==="
grep -rn '"text/template"' --include="*.go"

echo "=== Finding panics in library code ==="
grep -rn "panic(" --include="*.go" | grep -v "_test.go"

echo "=== Finding skipped tests ==="
grep -rn "t.Skip(" --include="*_test.go"

echo "=== Checking test quality ==="
grep -rn "t\.Log.*if" --include="*_test.go"
```

### Phase 3: Manual Review (4 hours)

**Security (1 hour):**
- Review all database query construction
- Verify HTML template usage and sanitization
- Check URL validation for SSRF prevention
- Verify file path validation

**Resource Management (30 minutes):**
- Check file handles have defer close
- Verify HTTP response bodies closed
- Check database resources cleaned up

**Concurrency (30 minutes):**
- Review goroutine lifecycle management
- Verify mutex usage for shared state
- Check WaitGroup usage

**Error Handling (30 minutes):**
- Review ignored errors (from grep)
- Verify error wrapping with %w
- Check panic usage

**Code Quality (1 hour):**
- Calculate test assertion density
- Find code duplication manually
- Identify functions >100 lines
- Check function complexity

**Tests (30 minutes):**
- Review test coverage gaps
- Check for flaky tests
- Verify assertion effectiveness

### Phase 4: Reporting (30 minutes)

Document findings with:
- Severity (Critical, High, Medium, Low)
- Location (file:line)
- Description of issue
- Recommended fix
- Priority for remediation

---

## Tips for AI Coding Agents

If you're an AI agent using this guide:

1. **Run automated checks first** before declaring task complete
2. **Calculate assertion density** for any test files you generate
3. **Check for these LLM-specific anti-patterns:**
   - Code duplication (you tend to copy-paste instead of refactoring)
   - Long functions (you don't naturally break down complexity)
   - Weak assertions (you focus on happy path)
   - Missing edge cases
4. **Use test-driven refactoring:**
   - Run tests before changing code (baseline)
   - Make incremental changes
   - Run tests after each change
   - Verify no regressions
5. **Verify your changes:**
   - Run `go build` to ensure it compiles
   - Run `go test -race` to check for races
   - Check you didn't introduce duplication

---

## Conclusion

This guide provides systematic heuristics for auditing Go codebases. Use it as:
- **Pre-release checklist** before marking code production-ready
- **PR review guide** for consistent code reviews
- **Self-check for AI agents** after generating code
- **Training material** for developers learning Go best practices

**Remember:** The best audit is the one that finds bugs before users do.

**Recommended frequency:**
- Automated checks: Every PR/commit
- Pattern matching: Weekly
- Full manual audit: Before major releases or quarterly
- Security audit: Before production deployment

---

**Document Version:** 2.1
**Created:** 2025-10-20
**Last Updated:** 2025-10-30
**Validated On:** Multiple Go projects including Rogue Planet feed aggregator (4,000+ lines)

## Changelog

**v2.1 (2025-10-30):**
- Added lessons from v0.4.0 rate limiting implementation
- New check: Version number consistency across code and docs
- New check: Integration test gaps beyond unit tests
- New check: Documentation synchronization after feature implementation
- Expanded TODO comment best practices with future field documentation
- Added spec document lifecycle management

**v2.0 (2025-10-26):**
- Added Documentation Integrity section (3 checks)
- Added Architecture & Design section (2 checks)
- Added Build & Infrastructure section (2 checks)
- Expanded Code Quality section (5 new checks)
- Expanded Concurrency Patterns with context propagation
- Expanded Performance section (4 new checks)
- Updated Quick Start Checklist with 6 new items
- Updated Priority Matrix with 13 new check types
- Total: 17 new checks from real-world Rogue Planet audit

**v1.0 (2025-10-20):**
- Initial release with core security, resource management, and testing checks

---

This is a living document. Please contribute improvements, new heuristics, and tools that you find effective.
