# Go Code Auditing Heuristics

**Purpose:** Comprehensive checklist for auditing Go codebases
**Audience:** Developers, code reviewers, AI agents, security auditors
**Version:** 1.0
**Last Updated:** 2025-10-20

---

## Table of Contents

1. [Quick Start Checklist](#quick-start-checklist)
2. [Security Auditing](#security-auditing)
3. [Resource Management](#resource-management)
4. [Concurrency Patterns](#concurrency-patterns)
5. [Error Handling](#error-handling)
6. [Code Quality](#code-quality)
7. [Test Quality](#test-quality)
8. [Performance](#performance)
9. [Automated Tools](#automated-tools)
10. [Priority Matrix](#priority-matrix)

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
| Code duplication | **Medium** | Medium | Maintenance burden |
| Function complexity | **Medium** | Low | Hard to modify |
| Test assertion quality | **Medium** | Medium | False confidence |
| Flaky tests | **Medium** | Medium | CI/CD unreliability |
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

**Document Version:** 1.0
**Created:** 2025-10-20
**Validated On:** Multiple Go projects including feed aggregator (4,000 lines)

This is a living document. Please contribute improvements, new heuristics, and tools that you find effective.
