# Industry Code Audit Heuristics Research

## Executive Summary

This research compiles industry-standard code audit methodologies, Go-specific best practices, security checklists, and quality metrics from official sources, academic research, and professional tools. The findings reveal significant gaps in typical code reviews and provide actionable heuristics for comprehensive audits.

---

## 1. Official Go Guidelines

### 1.1 Go Code Review Comments (Official)

**Source**: https://go.dev/wiki/CodeReviewComments

**Key Checks:**

1. **Code Formatting**
   - Run `gofmt` on all code (almost all Go code in the wild uses gofmt)
   - Automatic formatting eliminates most mechanical style issues

2. **Documentation Comments**
   - Comments documenting declarations should be full sentences
   - Comments should begin with the name of the thing being described
   - Package comments must appear adjacent to package clause with no blank line
   - Exported identifiers must be documented

3. **Receiver Types**
   - If method needs to mutate receiver, must use pointer
   - If receiver contains sync.Mutex or similar, must use pointer to avoid copying
   - Be consistent across all methods for a type

4. **Error Handling**
   - Don't ignore errors (check all return values)
   - Don't panic in library code
   - Return errors, don't use panic for normal error handling

5. **Naming Conventions**
   - Use MixedCaps or mixedCaps, not underscores
   - Acronyms should be all capitals (HTTP, URL, not Http, Url)
   - Package names should be lowercase, single word, no underscores

### 1.2 Google's Code Review Standards

**Source**: https://google.github.io/eng-practices/review/

**Key Principles:**

- There is no "perfect" code—only better code
- Technical facts and data overrule opinions and personal preferences
- Style guide is the absolute authority on matters of style
- Reviewers should not require polishing every tiny piece before approval

### 1.3 Effective Go Patterns

**Official patterns that auditors check:**

- Proper use of defer for cleanup
- Avoid goroutine leaks (always have termination path)
- Use context for cancellation and timeouts
- Prefer struct embedding over inheritance
- Use interfaces to define behavior, not data

---

## 2. Professional Linter Checks

### 2.1 golangci-lint Default Enabled Linters

**Source**: https://golangci-lint.run/docs/linters/

**Enabled by Default:**

1. **errcheck** - Checks for unchecked errors (critical bugs)
2. **govet** - Suspicious constructs (see detailed list below)
3. **ineffassign** - Detects unused assignments
4. **staticcheck** - 150+ checks (see section 2.2)
5. **unused** - Unused constants, variables, functions, types

**Recommended Additional Linters:**

6. **gosec** - Security problems (see section 3.2)
7. **gocyclo** - Cyclomatic complexity > 10-15
8. **gocognit** - Cognitive complexity > 10-15
9. **dupl** - Code duplication detection
10. **goconst** - Repeated strings that should be constants
11. **misspell** - Spelling mistakes in comments
12. **unconvert** - Unnecessary type conversions
13. **prealloc** - Slice preallocation opportunities
14. **bodyclose** - HTTP response body not closed
15. **sqlclosecheck** - SQL rows not closed
16. **rowserrcheck** - SQL rows.Err() not checked
17. **noctx** - HTTP requests without context.Context
18. **gocritic** - 100+ opinionated checks (see section 2.4)
19. **bidichk** - Dangerous unicode character sequences
20. **mulint** - Recursive locks (potential deadlocks)

### 2.2 staticcheck Rules (150+ Checks)

**Source**: https://staticcheck.dev/docs/checks

**Categories:**

1. **SA - Static Analysis (Bugs)**
   - SA1000-SA1030: Various misuses
   - SA4000-SA4020: Code that does nothing useful
   - SA5000-SA5012: Correctness issues
   - SA9000-SA9008: Suspicious or dubious code

2. **S - Code Simplifications**
   - S1000-S1040: Opportunities to simplify code

3. **ST - Style Violations**
   - ST1000-ST1023: Stylistic issues

4. **QF - Quick Fixes**
   - QF1000-QF1012: Automated fix suggestions

**Example Critical Checks:**
- SA5009: Invalid Printf calls
- SA5000: Assignment to nil maps
- SA4020: Unreachable case clauses
- SA1019: Use of deprecated APIs
- SA1029: Inappropriate key in call to context.WithValue

**Threshold**: staticcheck should report zero issues for production code

### 2.3 go vet Checks (26 Built-in Analyzers)

**Source**: https://go.dev/src/cmd/vet/doc.go

**Complete List:**

1. **appends** - Missing values after append
2. **asmdecl** - Mismatches between Go and assembly
3. **assign** - Useless assignments
4. **atomic** - Misuse of sync/atomic package
5. **bools** - Common boolean operator mistakes
6. **buildtag** - Malformed +build tags
7. **cgocall** - Cgo pointer passing violations
8. **composites** - Unkeyed composite literals
9. **copylocks** - Locks passed by value
10. **defers** - Common defer mistakes
11. **directive** - Invalid //go: directives
12. **errorsas** - Non-pointer/non-error to errors.As
13. **framepointer** - Assembly issues
14. **httpresponse** - HTTP response handling mistakes
15. **ifaceassert** - Impossible interface assertions
16. **loopclosure** - Loop variable capture bugs
17. **lostcancel** - Context.WithCancel not called
18. **nilfunc** - Useless nil/function comparisons
19. **printf** - Printf format/argument mismatches
20. **shift** - Shifts exceeding integer width
21. **stdmethods** - Incorrect standard interface implementations
22. **structtag** - Malformed struct tags
23. **tests** - Common test/example mistakes
24. **unreachable** - Unreachable code
25. **unsafeptr** - Invalid uintptr to unsafe.Pointer
26. **unusedresult** - Unused function call results

**Threshold**: go vet should report zero issues

### 2.4 go-critic Rules (100+ Opinionated Checks)

**Source**: https://go-critic.com/overview.html

**Categories:**

1. **Diagnostic** (enabled by default)
   - Find programming errors
   - Detect suspicious code
   - Examples: appendAssign, badCond, caseOrder, dupBranchBody

2. **Style** (only non-opinionated enabled by default)
   - Replace with more idiomatic forms
   - Examples: commentFormatting, ifElseChain, importShadow, unnamedResult

3. **Performance** (disabled by default)
   - Potential speed/memory issues
   - Examples: appendCombine, hugeParam, rangeExprCopy, sliceClear

**Note**: go-critic is highly opinionated and may have false positives. Use selectively.

---

## 3. Security Audit Best Practices

### 3.1 OWASP API Security Top 10 (2023)

**Source**: https://owasp.org/www-project-api-security/

**Critical Checks:**

1. **API1:2023 - Broken Object Level Authorization**
   - APIs expose endpoints handling object identifiers
   - Check: Every endpoint validates object ownership

2. **API2:2023 - Broken Authentication**
   - Authentication mechanisms poorly implemented
   - Check: Proper token handling, no tokens in URLs

3. **API3:2023 - Broken Object Property Level Authorization**
   - Mass assignment vulnerabilities
   - Check: Validate all input properties

4. **API4:2023 - Unrestricted Resource Consumption**
   - No rate limiting, DoS potential
   - Check: Rate limiting, timeouts, resource quotas

5. **API5:2023 - Broken Function Level Authorization**
   - Missing authorization on admin functions
   - Check: Authorization on all endpoints

6. **API6:2023 - Unrestricted Access to Sensitive Business Flows**
   - Automated attacks on business logic
   - Check: Bot protection, transaction rate limits

7. **API7:2023 - Server Side Request Forgery (SSRF)**
   - API fetches remote resource without validating URL
   - Check: URL validation (block localhost, private IPs)

8. **API8:2023 - Security Misconfiguration**
   - Missing security headers, verbose errors
   - Check: CSP headers, no stack traces in errors

9. **API9:2023 - Improper Inventory Management**
   - Old API versions remain accessible
   - Check: Deprecated endpoint handling

10. **API10:2023 - Unsafe Consumption of APIs**
    - Trusting third-party API data
    - Check: Validate all external data

### 3.2 gosec Security Checks with CWE Mapping

**Source**: https://github.com/securego/gosec

**Rule to CWE Mapping:**

| Rule | CWE | Description |
|------|-----|-------------|
| G101 | CWE-798 | Hard-coded credentials |
| G102 | CWE-200 | Bind to all interfaces (0.0.0.0) |
| G103 | CWE-242 | Unsafe use of unsafe package |
| G104 | CWE-703 | Unhandled errors |
| G106 | CWE-322 | Use of ssh.InsecureIgnoreHostKey |
| G107 | CWE-88 | Potential HTTP request made with variable url |
| G109 | CWE-190 | Integer overflow conversion |
| G110 | CWE-409 | Potential DoS via decompression bomb |
| G201 | CWE-89 | SQL injection (string formatting) |
| G202 | CWE-89 | SQL injection (string concatenation) |
| G203 | CWE-79 | HTML template auto-escaping not used |
| G204 | CWE-78 | Command injection via subprocess |
| G301 | CWE-276 | Poor file permissions (chmod) |
| G302 | CWE-276 | Poor file permissions (open/create) |
| G303 | CWE-377 | Insecure temporary file creation |
| G304 | CWE-22 | File path traversal |
| G305 | CWE-22 | Zip file path traversal |
| G401 | CWE-326 | Weak cryptographic primitive (MD5, SHA1) |
| G402 | CWE-295 | TLS InsecureSkipVerify true |
| G403 | CWE-310 | Weak encryption key (RSA < 2048 bits) |
| G404 | CWE-338 | Weak random number generator (math/rand) |
| G501 | CWE-327 | Blacklisted import: crypto/md5 |
| G502 | CWE-327 | Blacklisted import: crypto/des |
| G503 | CWE-327 | Blacklisted import: crypto/rc4 |
| G504 | CWE-327 | Blacklisted import: net/http/cgi |
| G505 | CWE-327 | Blacklisted import: crypto/sha1 |
| G601 | CWE-118 | Implicit memory aliasing in for loop |

**Threshold**: Zero gosec findings for production code, especially G201, G202, G401, G402, G404

### 3.3 OWASP REST Security Checklist

**Source**: https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html

**Key Requirements:**

1. **HTTPS Everywhere**
   - Only provide HTTPS endpoints
   - Reject HTTP requests or redirect to HTTPS

2. **Authentication & Authorization**
   - Access control at each API endpoint
   - Use standard authentication (JWT, OAuth2)
   - No API keys in URLs

3. **Input Validation**
   - Validate all input (whitelist approach)
   - Reject invalid input (don't try to fix it)
   - Use parameterized queries for SQL

4. **Rate Limiting**
   - Implement throttling to prevent DoS
   - Return 429 Too Many Requests

5. **Output Encoding**
   - Proper content-type headers
   - Escape output based on context

6. **Security Headers**
   - Content-Security-Policy
   - X-Content-Type-Options: nosniff
   - X-Frame-Options: DENY

### 3.4 SQL Injection Prevention

**Source**: OWASP SQL Injection Prevention Cheat Sheet

**Go-Specific Requirements:**

1. **ALWAYS Use Prepared Statements**
   ```go
   // CORRECT
   db.Query("SELECT * FROM users WHERE id = ?", userID)

   // WRONG - SQL Injection vulnerability
   db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID))
   ```

2. **Check for String Concatenation in SQL**
   - Search for: `fmt.Sprintf` + database methods
   - Search for: `+` operator + database methods
   - All SQL must use `?` placeholders

3. **Audit All Database Operations**
   - Exec(), Query(), QueryRow() must use placeholders
   - No exceptions (even for "trusted" input)

**Detection Pattern**:
```bash
# Find potential SQL injection
grep -rn "fmt.Sprintf.*Query\|Query.*fmt.Sprintf" .
grep -rn "fmt.Sprintf.*Exec\|Exec.*fmt.Sprintf" .
```

---

## 4. Code Quality Metrics Standards

### 4.1 Cyclomatic Complexity Thresholds

**Sources**: Multiple (NIST, Microsoft, McCabe)

**Industry Standards:**

| Complexity | Rating | Recommendation |
|-----------|--------|----------------|
| 1-10 | Simple | Good, maintainable code |
| 11-15 | Moderate | Consider refactoring |
| 16-20 | Complex | Refactor strongly recommended |
| 21-50 | Very Complex | Difficult to test, must refactor |
| 50+ | Untestable | Cannot be adequately tested |

**Key Guidelines:**

- **NIST Standard**: Limit of 10 (can go to 15 with justification)
- **Steve McConnell (Code Complete)**: 0-5 fine, 6-10 be aware, >10 strongly refactor
- **Microsoft Visual Studio**: Warning at 25
- **MISRA (Safety-Critical)**: Maximum 15

**Recommended Threshold for Go**:
- **Warning**: 10
- **Error**: 15
- **Critical**: 20

### 4.2 Cognitive Complexity vs Cyclomatic Complexity

**Source**: Sonar, various industry sources

**Key Differences:**

**Cyclomatic Complexity:**
- Counts linearly independent paths
- Treats all control flow equally
- Switch with 10 cases = complexity 10
- Focuses on testability
- Easy to calculate, well-established

**Cognitive Complexity:**
- Measures mental effort to understand
- Considers nesting (adds mental burden)
- Switch with 10 cases = complexity 1
- Focuses on readability/maintainability
- Better for human comprehension

**Example:**
```go
// Cyclomatic: 4, Cognitive: 1 (easy to read)
switch day {
case "Mon": return 1
case "Tue": return 2
case "Wed": return 3
case "Thu": return 4
}

// Cyclomatic: 4, Cognitive: 7 (hard to read, nested)
if user != nil {
    if user.IsAdmin() {
        if user.HasPermission("delete") {
            if resource.IsOwnedBy(user) {
                // nested logic
            }
        }
    }
}
```

**Recommendation**: Track both metrics
- Cyclomatic for test coverage planning
- Cognitive for refactoring prioritization

### 4.3 Halstead Complexity Metrics

**Source**: Academic research, Microsoft

**Metrics:**

1. **Halstead Volume** (V)
   - Measures program size
   - V = N × log₂(n)
   - N = total operators + operands
   - n = distinct operators + operands
   - Represents "mental inventory" needed

2. **Halstead Difficulty** (D)
   - How error-prone the code is
   - D = (n₁/2) × (N₂/n₂)
   - Higher = more bugs expected

3. **Halstead Effort** (E)
   - Time to write/understand
   - E = D × V

**Maintainability Index Formula:**

```
MI = MAX(0, (171 - 5.2×ln(V) - 0.23×CC - 16.2×ln(LOC)) × 100/171)
```

Where:
- V = Halstead Volume
- CC = Cyclomatic Complexity
- LOC = Lines of Code

**Maintainability Index Interpretation:**

| MI Range | Maintainability |
|----------|----------------|
| 85-100 | Highly maintainable |
| 65-85 | Moderately maintainable |
| 50-65 | Difficult to maintain |
| 0-50 | Very difficult to maintain |

**Threshold**: MI should be > 65 for all files

### 4.4 Other Quality Metrics

**Code Coverage:**
- Line coverage: > 80% (but not a quality indicator alone)
- Branch coverage: > 75%
- Path coverage: Focus on critical paths
- **Important**: 100% coverage with 0% assertions = useless

**Code Duplication:**
- Threshold: < 3% duplicated blocks
- Tools: dupl, gocyclo

**Function Length:**
- Recommendation: < 50 lines
- Warning: > 100 lines
- Error: > 200 lines

**File Length:**
- Recommendation: < 500 lines
- Warning: > 1000 lines

**Parameter Count:**
- Recommendation: ≤ 3-4 parameters
- Warning: > 5 parameters
- Error: > 7 parameters
- Violation of Single Responsibility if many params

---

## 5. Test Effectiveness Metrics

### 5.1 Coverage vs Mutation Testing

**Source**: Academic research, Codecov, multiple industry sources

**Key Findings:**

**Code Coverage Limitations:**
- Measures what code was executed, not if tests are effective
- 100% coverage doesn't mean quality tests
- Can have 100% line coverage, 0% mutation score
- Coverage tells you what got exercised, not correctness

**Mutation Testing:**
- "Testing your tests"
- Introduces deliberate bugs (mutants)
- Runs test suite to see if mutants caught
- Mutation Score = (Killed Mutants / Total Mutants) × 100%
- Reveals if tests actually assert behavior

**Relationship:**
- Statement frequency coverage > statement/branch coverage
- Correlation with mutation score higher
- Branch coverage necessary but insufficient
- Mutation testing complements coverage

**Practical Application:**

1. **Minimum**: Line coverage > 80%
2. **Better**: Branch coverage > 75%
3. **Best**: Mutation score > 80%

**Go Tools:**
- Coverage: Built-in `go test -cover`
- Mutation: `go-mutesting` (github.com/zimmski/go-mutesting)

### 5.2 Test Smells Detection

**Source**: testsmells.org, academic research

**Categories of Test Smells:**

**1. Code Smells in Tests:**
- **Assertion Roulette**: Multiple assertions without messages
- **Eager Test**: Tests multiple methods (not focused)
- **Lazy Test**: Multiple tests in one method
- **Mystery Guest**: Test uses external files/resources unclear
- **Resource Optimism**: Assumes resources available without checking
- **Test Code Duplication**: Copy-pasted test code
- **Verbose Test**: Too much irrelevant code in test

**2. Behavioral Smells:**
- **Conditional Test Logic**: if/switch in tests (smell)
- **Empty Test**: Test that doesn't test anything
- **Exception Handling**: Test catches exceptions (should fail instead)
- **Ignored Test**: Commented out or skipped tests
- **Sleepy Test**: Uses time.Sleep() for synchronization
- **Unknown Test**: No clear purpose or assertion

**3. Project-Level Smells:**
- **Lack of Cohesion**: Tests spread across many files
- **Test Maverick**: Test doesn't follow project patterns
- **Dead Test**: Test for deleted/refactored code

**Go-Specific Test Smells:**

4. **Not Using Table-Driven Tests** (when appropriate)
5. **Not Using t.Helper()** in test helpers
6. **Using t.Fatal in goroutines** (should use t.Error)
7. **Not Using t.Parallel()** when tests are independent
8. **Magic Numbers** without explanation
9. **Not Cleaning Up Resources** (should use t.Cleanup or defer)

**Detection Tools:**
- TestSmellDetector (Java, but concepts apply)
- Manual code review
- Custom linters

### 5.3 Go Table-Driven Test Best Practices

**Source**: https://go.dev/wiki/TableDrivenTests, Dave Cheney

**Best Practices:**

1. **Use t.Run() for Subtests** (since Go 1.7)
   ```go
   tests := []struct {
       name string
       input int
       want int
   }{
       {"zero", 0, 0},
       {"positive", 5, 25},
   }
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           got := Square(tt.input)
           if got != tt.want {
               t.Errorf("got %d, want %d", got, tt.want)
           }
       })
   }
   ```

2. **Name Every Test Case** (for clarity in output)

3. **Use t.Errorf, Not t.Fatalf** (to see all failures)

4. **Extract Complex Logic to Helpers**

5. **Consider Check Functions** for complex validation

6. **Use Anonymous Structs** for test data

7. **Limit Table Size** (split if too large)

**Anti-Patterns to Avoid:**
- Too many test cases (> 20 suggests overly complex function)
- Complex setup within table (extract to helper)
- Mixing unit and integration tests in same table

---

## 6. Performance Audit Patterns

### 6.1 Memory Allocation and pprof

**Source**: Go official docs, DataDog profiling guide

**Key Profiling Best Practices:**

**1. Profile Types:**
- **heap**: In-use memory (`inuse_space`, `inuse_objects`)
- **allocs**: All allocations over time (`alloc_space`, `alloc_objects`)
- **Use heap for**: Current memory usage
- **Use allocs for**: GC pressure, allocation hot paths

**2. Sampling Rate:**
- Default: 1 sample per 512 KiB allocated
- Modify via `runtime.MemProfileRate`
- Only modify once, at program start (in main())
- Don't change mid-execution (causes incorrect profiles)

**3. Production Safety:**
- Safe to profile in production
- CPU profiling adds ~5% overhead
- Don't expose pprof endpoints publicly
- Secure endpoints with authentication

**4. Profile Isolation:**
- Collect only one profile type at a time
- Memory profiling skews CPU profiles
- Goroutine blocking affects scheduler trace

**5. Analysis Strategy:**
- Avoid profiling during startup/shutdown
- Compare snapshots for differential analysis (leaks)
- Focus on inuse for current memory issues
- Focus on alloc for GC performance

**6. Optimization Mantra:**
- **Reduce**: Turn heap allocations into stack allocations
- **Reuse**: Reuse heap allocations (sync.Pool)
- Minimize pointers on heap (helps GC)

**Common Allocation Hotspots:**
- String concatenation in loops (use strings.Builder)
- Converting []byte to string repeatedly
- Interface boxing
- Unnecessary slice/map growth (preallocate)
- Closure captures (allocates)

### 6.2 Goroutine Leak Detection

**Source**: uber-go/goleak, community best practices

**Detection Tools:**

1. **goleak** (uber-go/goleak)
   ```go
   func TestMain(m *testing.M) {
       goleak.VerifyTestMain(m)
   }

   func TestSomething(t *testing.T) {
       defer goleak.VerifyNone(t)
       // test code
   }
   ```

2. **runtime.NumGoroutine()**
   - Monitor count over time
   - Increasing count = leak

3. **pprof goroutine profile**
   ```bash
   curl http://localhost:6060/debug/pprof/goroutine?debug=2
   ```

4. **Semgrep** (static analysis for leak patterns)

**Common Causes:**

- Blocked channel send/receive
- Infinite loops without exit
- Forgotten goroutines (no cancellation)
- Context not checked in loops
- Missing WaitGroup.Done()
- Deadlocks (mutual waiting)

**Prevention Patterns:**

1. **Always use context for cancellation**
   ```go
   ctx, cancel := context.WithCancel(ctx)
   defer cancel()
   go worker(ctx)
   ```

2. **Check context in loops**
   ```go
   for {
       select {
       case <-ctx.Done():
           return ctx.Err()
       case work := <-workCh:
           process(work)
       }
   }
   ```

3. **Use WaitGroups**
   ```go
   var wg sync.WaitGroup
   wg.Add(1)
   go func() {
       defer wg.Done()
       // work
   }()
   wg.Wait()
   ```

4. **Worker pools** over unlimited goroutines

5. **Timeouts for all I/O**

### 6.3 Race Detector Limitations

**Source**: Go official docs, research papers

**What Race Detector Finds:**
- Data races (conflicting memory access)
- At least one access is a write
- No synchronization between accesses

**What Race Detector MISSES:**

1. **Deadlocks** (synchronization problem, not data race)
2. **Livelocks** (goroutines running but not progressing)
3. **Starvation** (goroutine never gets scheduled)
4. **Logical races** (wrong order, but synchronized)
5. **Races in unexecuted code paths**

**Race Detector Characteristics:**
- Zero false positives (if reported, it's real)
- Can have false negatives (absence doesn't prove correctness)
- Only detects races that occur during execution
- Slows execution ~10x, increases memory 10x

**Usage:**
```bash
go test -race ./...
go run -race ./cmd/app
go build -race  # for production debugging only
```

**Complementary Tools:**

1. **Go deadlock detector** (https://github.com/sasha-s/go-deadlock)
   - Detects potential deadlocks before they happen
   - Only works for mutex deadlocks (not channels)

2. **Static analysis** (research tools)
   - Can find potential races/deadlocks statically
   - Not generally available in standard tools

**Best Practice**: Always run tests with `-race` in CI

### 6.4 N+1 Query Problem Detection

**Source**: Industry blogs, database performance guides

**What is N+1:**
- Fetch list of N items (1 query)
- For each item, fetch related data (N queries)
- Total: N+1 queries instead of 1-2 queries

**Example:**
```go
// BAD: N+1 queries
users := fetchUsers() // 1 query
for _, user := range users {
    posts := fetchPostsByUser(user.ID) // N queries
    // process posts
}

// GOOD: 2 queries
users := fetchUsers() // 1 query
userIDs := extractIDs(users)
posts := fetchPostsByUserIDs(userIDs) // 1 query with WHERE IN
```

**Detection Methods:**

1. **Log Query Counts**
   - Instrument database layer
   - Log number of queries per request
   - Alert if count > threshold

2. **Distributed Tracing**
   - Tools: Zipkin, Jaeger, OpenTelemetry
   - Visualize query patterns
   - Identify sequential queries

3. **Database Monitoring**
   - Look for repetitive queries in logs
   - Same query pattern, different parameters
   - High query count in short time

4. **Anomaly Detection**
   - Set thresholds (e.g., > 10 queries/request)
   - Alert on sudden increases

5. **Code Review Patterns**
   - Loop with database call inside = suspect
   - ORM lazy loading = suspect
   - No eager loading/joins = suspect

**Prevention:**
- Use JOINs to fetch related data
- Eager loading (fetch everything upfront)
- Batch queries (WHERE IN)
- Caching
- DataLoader pattern (batching + caching)

### 6.5 Resource Leak Detection

**Source**: AWS CodeGuru, industry tools

**Critical Resource Types:**

1. **File Handles**
   - Limited per process (usually 1024-65536)
   - Leaking causes "too many open files"

2. **Database Connections**
   - Limited by pool size
   - Leaking causes connection exhaustion

3. **Network Sockets**
   - Limited per system
   - Leaking causes connection failures

4. **Memory**
   - Leaking causes OOM kills

5. **Goroutines**
   - Leaking causes memory growth + CPU waste

**Go Patterns to Check:**

**1. File Handles:**
```go
// CORRECT
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()

// WRONG - leak if error occurs after Open
f, _ := os.Open(path)
// ... code that might return early
f.Close()
```

**2. Database Rows:**
```go
// CORRECT
rows, err := db.Query(...)
if err != nil {
    return err
}
defer rows.Close()
for rows.Next() {
    // scan
}
return rows.Err() // MUST check rows.Err()

// WRONG - leak if not closed
rows, _ := db.Query(...)
for rows.Next() {
    // scan
}
// missing rows.Close()
```

**3. HTTP Response Bodies:**
```go
// CORRECT
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()
io.Copy(ioutil.Discard, resp.Body) // drain before close

// WRONG - leak
resp, _ := http.Get(url)
// missing Close()
```

**4. Context Cancellation:**
```go
// CORRECT
ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel() // MUST call even if context expires

// WRONG - resource leak
ctx, _ := context.WithTimeout(ctx, 5*time.Second)
// missing cancel() call
```

**Detection:**
- Linters: bodyclose, sqlclosecheck, rowserrcheck
- Code review: search for resource acquisition without defer
- Runtime: monitor open file descriptors, connection pool

---

## 7. Concurrency Patterns and Anti-Patterns

### 7.1 Context Usage Best Practices

**Source**: Go official blog, community guides

**Best Practices:**

1. **Always defer cancel()**
   ```go
   ctx, cancel := context.WithCancel(parent)
   defer cancel() // Even if ctx expires, call cancel
   ```

2. **Pass context as first parameter**
   ```go
   func DoWork(ctx context.Context, arg string) error
   ```

3. **Propagate context through call chain**
   - Don't create new root contexts mid-chain
   - Derive from parent context

4. **Don't store contexts in structs** (exception: must document)

5. **Use for cancellation, not error propagation**
   - Context can only be cancelled once
   - Not for multiple errors

6. **Avoid wrapping cancellable context**
   - Multiple cancel points = confusing
   - Keep cancellation logic clear

7. **Use short timeouts near I/O boundaries**
   - Not single large timeout at root
   - Specific timeouts for specific operations

**Common Mistakes:**

- Not checking `<-ctx.Done()` in loops
- Passing `context.Background()` when parent available
- Not propagating cancellation to child goroutines
- Using context.Value for configuration (use parameters)

### 7.2 Error Handling Best Practices

**Source**: Go 1.13 error blog, Dave Cheney

**Modern Error Handling (Go 1.13+):**

**1. Wrapping Errors:**
```go
// Use %w to wrap
return fmt.Errorf("failed to fetch user %d: %w", id, err)

// Check wrapped errors
if errors.Is(err, sql.ErrNoRows) {
    // handle not found
}

// Extract error type
var netErr *net.Error
if errors.As(err, &netErr) {
    // handle network error
}
```

**2. Sentinel Errors:**
```go
var ErrNotFound = errors.New("not found")

// Return wrapped sentinel
return fmt.Errorf("user %q: %w", name, ErrNotFound)

// Check with errors.Is
if errors.Is(err, ErrNotFound) {
    // handle
}
```

**3. Custom Error Types:**
```go
type ValidationError struct {
    Field string
    Err   error
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed on %s: %v", e.Field, e.Err)
}

func (e *ValidationError) Unwrap() error {
    return e.Err
}
```

**Best Practices:**

1. **Wrap errors at each level** (add context)
2. **Use errors.Is for sentinel checks** (not ==)
3. **Use errors.As for type checks** (not type assertion)
4. **Treat errors as opaque when possible** (don't inspect)
5. **Minimize sentinel errors** (prefer opaque)
6. **Don't use errors.Is in hot paths** (5x slower)

**Anti-Patterns:**

- Ignoring errors (`_ = f.Close()`)
- Panic in library code
- Using panic for normal errors
- String matching on error messages
- Not wrapping errors (losing context)
- Over-wrapping (too verbose)

---

## 8. Documentation and Code Organization

### 8.1 Godoc Standards

**Source**: https://go.dev/doc/comment

**Rules:**

1. **All exported identifiers MUST be documented**
   ```go
   // Package foo provides utilities for bar.
   package foo

   // Client connects to the server.
   type Client struct { }

   // Connect establishes a connection.
   func (c *Client) Connect() error { }
   ```

2. **Comments start with name being described**
   ```go
   // Good
   // ParseURL parses a URL string.

   // Bad
   // This function parses a URL string.
   ```

3. **Use full sentences**
   - Start with capital letter
   - End with period

4. **Package comment** adjacent to package clause
   ```go
   // Package http provides HTTP client and server implementations.
   package http
   ```

5. **Deprecation notice**
   ```go
   // Deprecated: Use NewClient instead.
   func OldClient() *Client { }
   ```

6. **BUG comments**
   ```go
   // BUG(username): This doesn't handle IPv6.
   ```

7. **Preformatted text** (indent 4 spaces)

8. **URLs** automatically linked

**Audit Checks:**
- All exported types, functions, constants, variables documented
- Comments grammatically correct
- No TODO/FIXME in godoc (use issue tracker)
- Examples in `example_test.go` files

### 8.2 Package Organization

**Best Practices:**

1. **Flat is better than nested**
   - Avoid deep package hierarchies
   - Prefer `/pkg/user` over `/pkg/domain/user/entity`

2. **Package name is part of API**
   - Avoid stutter: `user.NewUser` not `user.NewUserClient`
   - Short, clear names

3. **Internal packages** for implementation
   - `/internal` not importable from outside module
   - Use for unexported shared code

4. **cmd/** for multiple binaries

5. **Avoid circular dependencies**
   - Use interfaces to invert dependencies
   - Extract shared types to separate package

---

## 9. SOLID Principles Violations

**Source**: Industry code review guides

### 9.1 Single Responsibility Principle (SRP)

**Violations:**

1. **Large types/files** (> 500 lines)
2. **Names with "AND"** (UserAndOrderManager)
3. **Too many methods** (> 10 methods on type)
4. **Too many parameters** (> 5 parameters)
5. **Difficult to name** (ends up as Helper, Utils, Manager)
6. **Hard to test** (needs many mocks)
7. **Feature envy** (uses other type's data more than own)

**Detection:**
- Line count per file/type
- Method count per type
- Parameter count per function
- Test complexity

### 9.2 Open-Closed Principle (OCP)

**Violations:**

1. **Type switches** on concrete types
   ```go
   // Violation
   switch v := x.(type) {
   case *Dog: v.Bark()
   case *Cat: v.Meow()
   }

   // Better: use interface
   x.MakeSound()
   ```

2. **Long if-else chains** checking types

**Detection:**
- `grep -rn "switch.*\.(type)"` (review each)

### 9.3 Liskov Substitution Principle (LSP)

**Violations:**

1. **Subtypes require more** than base type
2. **Subtypes provide less** than base type
3. **Precondition strengthening**
4. **Postcondition weakening**

**Detection:**
- Interface implementation changes behavior radically
- Panic in interface implementation (violates expectations)

### 9.4 Interface Segregation Principle (ISP)

**Violations:**

1. **Fat interfaces** (> 5-7 methods)
2. **Implementations with empty/panic methods**

**Detection:**
```bash
# Find large interfaces
grep -A 20 "type.*interface" *.go | grep "^\s*[A-Z]" | wc -l
```

### 9.5 Dependency Inversion Principle (DIP)

**Violations:**

1. **Depends on concrete types** instead of interfaces
2. **Type assertions** in business logic
3. **Direct instantiation** of dependencies

**Go Pattern**: Accept interfaces, return structs

**Detection:**
- Function parameters are concrete types (often violation)
- grep for type assertions in non-test code

---

## 10. Software Engineering Metrics

**Source**: Academic research, industry tools

### 10.1 Code Churn

**Definition**: % of code modified/replaced within ~3 weeks

**Thresholds:**
- **Healthy**: < 9%
- **Concerning**: 9-14%
- **Problematic**: > 14%

**High churn indicates:**
- Unclear requirements
- Poor initial design
- Technical debt
- Instability

**Measurement:**
```bash
# Files changed multiple times in 3 weeks
git log --since="3 weeks ago" --name-only --pretty=format: | \
    sort | uniq -c | sort -nr
```

### 10.2 Defect Density

**Definition**: Defects per 1000 lines of code

**Calculation:**
```
Defect Density = (Number of Defects / KLOC) × 1000
```

**Thresholds:**
- **Excellent**: < 1 defect/KLOC
- **Good**: 1-5 defects/KLOC
- **Poor**: > 5 defects/KLOC

**Sources:**
- Bug tracker closed bugs
- Post-release defects
- Security vulnerabilities

### 10.3 Technical Debt Ratio (TDR)

**Definition**: Cost to fix issues / cost to develop from scratch

**Calculation (SonarQube):**
```
TDR = Remediation Cost / Development Cost × 100%
```

**Ratings:**
- **A (excellent)**: < 5%
- **B (good)**: 5-10%
- **C (moderate)**: 10-20%
- **D (poor)**: 20-50%
- **E (very poor)**: > 50%

**Components:**
- Code smells × time to fix
- Security vulnerabilities × time to fix
- Bugs × time to fix

---

## 11. Heuristics We Haven't Considered

### 11.1 Dependency Management

**New Checks:**

1. **Outdated dependencies**
   ```bash
   go list -u -m all
   ```

2. **Vulnerable dependencies**
   - Use: `govulncheck` (official Go tool)
   - Check against Go vulnerability database
   - CI integration required

3. **Dependency count** (minimalism)
   - Direct dependencies: aim for < 20
   - Total dependencies: aim for < 100
   - Each dependency is attack surface

4. **License compliance**
   - Check all dependency licenses
   - Tools: google/go-licenses

5. **Deprecated packages**
   - Check for deprecated imports
   - staticcheck detects some (SA1019)

### 11.2 API Design Quality

**New Checks:**

1. **Breaking changes** in public API
   - Tool: `gorelease`
   - Semantic versioning compliance

2. **Function signature consistency**
   - Error always last return value
   - Context always first parameter
   - Options pattern for > 3 params

3. **Interface pollution**
   - Interfaces should have 1-3 methods
   - Interfaces should be small and focused

4. **Premature abstraction**
   - Interface with single implementation (code smell)
   - Exception: for testing

### 11.3 Build and CI/CD

**New Checks:**

1. **Build reproducibility**
   - `go.sum` committed
   - Version pinning
   - No floating dependencies

2. **CI pipeline completeness**
   - All linters run
   - Tests with -race
   - Integration tests
   - Security scans
   - Benchmark tracking

3. **Build time** (developer experience)
   - Full build < 10 minutes (ideal)
   - Full build < 30 minutes (acceptable)
   - Incremental build < 1 minute

### 11.4 Observability

**New Checks:**

1. **Structured logging**
   - No fmt.Println in production code
   - Use structured logger (zap, zerolog, slog)
   - Log levels appropriate

2. **Metrics instrumentation**
   - Prometheus metrics exposed
   - Key operations instrumented
   - Error rates tracked

3. **Tracing**
   - OpenTelemetry integration
   - Trace context propagation
   - Span attributes meaningful

4. **Health checks**
   - /health endpoint
   - /ready endpoint (for k8s)
   - Dependency health checked

### 11.5 Configuration Management

**New Checks:**

1. **No secrets in code**
   - Use environment variables
   - Use secret management (Vault, etc.)
   - .env files in .gitignore

2. **Configuration validation**
   - Validate on startup
   - Fail fast with clear errors
   - Schema validation

3. **Twelve-factor compliance**
   - Config in environment
   - Backing services as attached resources
   - Stateless processes

### 11.6 Error Messages

**New Checks:**

1. **No sensitive data in errors**
   - No passwords, tokens, API keys
   - No file paths (information disclosure)
   - No stack traces to users

2. **Actionable errors**
   - Tell user what went wrong
   - Suggest how to fix
   - Include error codes for lookup

3. **Error consistency**
   - Same error format throughout
   - Same wrapping strategy
   - Same logging pattern

---

## 12. Recommended Tools

### 12.1 Static Analysis

**Essential:**
1. **golangci-lint** - Meta-linter (runs 50+ linters)
   - Config: .golangci.yml
   - CI integration: easy
   - Speed: fast (caching)

2. **govulncheck** - Official vulnerability scanner
   - Checks Go vulnerability database
   - Free, official, required

3. **staticcheck** - Advanced static analysis
   - 150+ checks
   - Very low false positive rate
   - Industry standard

**Recommended:**

4. **gosec** - Security-focused scanner
   - CWE mapping
   - OWASP compliance
   - CI/CD friendly

5. **go-critic** - Opinionated linter
   - 100+ checks
   - Performance suggestions
   - Use selectively (can be noisy)

### 12.2 Testing

**Essential:**

1. **go test** - Built-in test runner
   - Use with `-race`, `-cover`

2. **goleak** - Goroutine leak detector
   - Uber's tool
   - Integrate in test suite

**Recommended:**

3. **testify** - Testing assertions and mocks
   - Popular, but not essential
   - Makes tests more readable

4. **go-mutesting** - Mutation testing
   - Expensive, run periodically
   - Validates test quality

### 12.3 Performance

**Essential:**

1. **pprof** - Built-in profiler
   - CPU, memory, goroutine, blocking
   - Production-ready

2. **benchstat** - Benchmark comparison
   - Statistical analysis
   - Detects regressions

**Recommended:**

3. **go-torch** - Flame graphs
   - Visualize pprof data
   - Find hot paths

4. **vegeta** - Load testing
   - HTTP load generator
   - Performance baselines

### 12.4 Code Quality

**Recommended:**

1. **SonarQube** - Quality platform
   - Technical debt tracking
   - Trends over time
   - CI integration

2. **CodeClimate** - Maintainability analysis
   - Technical debt quantification
   - Pull request analysis

3. **goreportcard.com** - Quick quality check
   - Free for public repos
   - Multiple metrics

### 12.5 Dependency Management

**Essential:**

1. **go mod** - Built-in dependency manager
   - Use `go mod tidy`
   - Commit `go.sum`

2. **govulncheck** - Vulnerability scanner
   - Official tool
   - Required for security

**Recommended:**

3. **dependabot** - Automated updates
   - GitHub integration
   - CVE notifications

4. **nancy** - Dependency scanner
   - Sonatype OSS Index
   - CI integration

### 12.6 Documentation

**Essential:**

1. **godoc** - Built-in documentation
   - `go doc -all`
   - Godoc.org for public packages

**Recommended:**

2. **swag** - Swagger generation from comments
   - API documentation
   - OpenAPI spec generation

---

## 13. Priority Additions to Audit Checklist

### Ranked by Importance:

**CRITICAL (Must Have):**

1. **Security Scanning**
   - [ ] Run `gosec` on entire codebase
   - [ ] Run `govulncheck` for vulnerable dependencies
   - [ ] Check all SQL queries use prepared statements (no string concat)
   - [ ] Verify SSRF prevention (URL validation)
   - [ ] Check for hardcoded secrets (G101)
   - [ ] Verify weak crypto not used (G401, G402, G404)

2. **Resource Leaks**
   - [ ] All file opens have `defer f.Close()`
   - [ ] All `sql.Rows` have `defer rows.Close()` + `rows.Err()` check
   - [ ] All HTTP response bodies closed
   - [ ] All contexts have `defer cancel()`

3. **Error Handling**
   - [ ] All errors checked (errcheck linter)
   - [ ] Errors wrapped with context
   - [ ] No panic in library code

4. **Concurrency**
   - [ ] Run tests with `-race`
   - [ ] Run `goleak` to detect goroutine leaks
   - [ ] All goroutines have exit path (context cancellation)

**HIGH (Should Have):**

5. **Code Complexity**
   - [ ] Cyclomatic complexity < 15 (ideally < 10)
   - [ ] Cognitive complexity < 15
   - [ ] Functions < 50 lines (ideally)
   - [ ] Files < 500 lines (warning at 1000)

6. **Test Quality**
   - [ ] Code coverage > 80% (line), > 75% (branch)
   - [ ] No test smells (assertion roulette, sleepy test, etc.)
   - [ ] Table-driven tests used appropriately
   - [ ] Tests run in parallel where possible

7. **Documentation**
   - [ ] All exported identifiers documented
   - [ ] Package comments present
   - [ ] Comments start with name being described
   - [ ] No TODOs in godoc comments

8. **Static Analysis**
   - [ ] `staticcheck` reports zero issues
   - [ ] `go vet` reports zero issues
   - [ ] `golangci-lint` passes with reasonable config

**MEDIUM (Nice to Have):**

9. **SOLID Principles**
   - [ ] No SRP violations (large types, many parameters)
   - [ ] No OCP violations (type switches on concrete types)
   - [ ] Interfaces small (< 5 methods)

10. **Performance**
    - [ ] Profile with pprof (memory, CPU)
    - [ ] Check for N+1 queries
    - [ ] No unnecessary allocations in hot paths
    - [ ] Preallocate slices when size known

11. **Dependency Quality**
    - [ ] Direct dependencies < 20
    - [ ] All dependencies up-to-date (or explicitly pinned)
    - [ ] Licenses compatible with project
    - [ ] No deprecated packages

**LOW (Optional):**

12. **Metrics Tracking**
    - [ ] Code churn < 14%
    - [ ] Defect density tracked
    - [ ] Technical debt ratio < 10%
    - [ ] Maintainability index > 65

13. **Advanced Testing**
    - [ ] Mutation testing (periodically)
    - [ ] Benchmark regression tests
    - [ ] Property-based testing (for complex algorithms)

---

## 14. Implementation Roadmap

### Phase 1: Security & Correctness (Week 1)
- Set up `gosec` in CI
- Set up `govulncheck` in CI
- Run `errcheck` and fix all violations
- Enable race detector in tests
- Add `goleak` to critical tests

### Phase 2: Static Analysis (Week 2)
- Configure `golangci-lint` with recommended linters
- Fix all `staticcheck` issues
- Fix all `go vet` issues
- Address resource leak patterns

### Phase 3: Testing (Week 3-4)
- Measure current code coverage
- Add tests to reach 80% coverage
- Convert appropriate tests to table-driven
- Run tests with `-race` in CI

### Phase 4: Documentation (Week 5)
- Document all exported identifiers
- Add package comments
- Review and improve error messages
- Add examples for complex APIs

### Phase 5: Metrics & Monitoring (Ongoing)
- Set up complexity tracking
- Track code churn
- Monitor dependency updates
- Review defect density quarterly

---

## 15. Conclusion

**Key Takeaways:**

1. **Security is non-negotiable**: gosec, govulncheck, SQL injection prevention
2. **Concurrency bugs are expensive**: -race, goleak, context patterns
3. **Tests need validation**: Coverage ≠ quality, use mutation testing
4. **Complexity kills maintainability**: Enforce limits on cyclomatic/cognitive complexity
5. **Resource leaks are silent killers**: Defer all cleanup, check all errors
6. **Documentation is code**: All exports must be documented
7. **Automation prevents drift**: Run all checks in CI/CD

**Tools Priority:**
1. golangci-lint (all-in-one)
2. govulncheck (security)
3. goleak (concurrency)
4. pprof (performance)
5. SonarQube (trends)

**Metrics That Matter:**
- Cyclomatic complexity < 15
- Code coverage > 80%
- Zero security findings
- Zero race conditions
- Technical debt ratio < 10%

**Remember**: Perfect is the enemy of good. Prioritize security and correctness, then iterate on quality improvements.

---

## References

1. Go Code Review Comments: https://go.dev/wiki/CodeReviewComments
2. Google Eng Practices: https://google.github.io/eng-practices/
3. golangci-lint: https://golangci-lint.run/
4. staticcheck: https://staticcheck.dev/
5. gosec: https://github.com/securego/gosec
6. OWASP API Security: https://owasp.org/www-project-api-security/
7. Go Race Detector: https://go.dev/doc/articles/race_detector
8. uber-go/goleak: https://github.com/uber-go/goleak
9. Test Smells: https://testsmells.org/
10. Go pprof: https://go.dev/doc/diagnostics

---

**Document Version**: 1.0
**Last Updated**: 2025-10-19
**Research Scope**: Go code audit heuristics, industry best practices, academic research
**Target Audience**: Code reviewers, security auditors, Go developers
