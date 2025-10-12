# Testing Plan for Rogue Planet

## Overview

This document outlines the comprehensive testing strategy for Rogue Planet, covering unit tests, integration tests, security tests, and performance benchmarks.

## Testing Principles

1. **Test the contract, not the implementation** - Focus on behavior and interfaces
2. **Isolate components** - Each package should be testable independently
3. **Use table-driven tests** - Go idiom for testing multiple scenarios
4. **Test error paths** - Errors are first-class citizens in Go
5. **Security-first testing** - XSS, SQL injection, and SSRF prevention are critical

## Unit Tests by Component

### 1. Crawler Package (`pkg/crawler`)

#### HTTP Conditional Requests Tests
**Critical: These tests validate CVE prevention**

```go
TestConditionalRequests
  - Test ETag storage and retrieval (exact match, including quotes)
  - Test Last-Modified storage and retrieval (exact format)
  - Test If-None-Match header is sent correctly
  - Test If-Modified-Since header is sent correctly
  - Test 304 Not Modified response handling
  - Test cache update on 200 OK response
  - Test both headers sent when both are available
  - Test no headers sent when cache is empty
```

#### HTTP Client Tests

```go
TestFetchFeed
  - Test successful fetch (200 OK)
  - Test 304 Not Modified (should not return body)
  - Test 404 Not Found
  - Test 429 Too Many Requests (with Retry-After)
  - Test 500 Server Error
  - Test 301 Permanent Redirect (should update URL)
  - Test 302 Temporary Redirect (should keep original URL)
  - Test redirect loop detection (max 5 redirects)
  - Test timeout handling (30 second default)
  - Test connection refused
  - Test DNS lookup failure
  - Test invalid SSL certificate handling

TestUserAgent
  - Test User-Agent header format
  - Test User-Agent includes version number
  - Test User-Agent includes contact URL

TestContentNegotiation
  - Test Accept-Encoding: gzip, deflate
  - Test gzip decompression
  - Test deflate decompression
  - Test uncompressed response

TestResourceLimits
  - Test maximum response size (10MB)
  - Test timeout cancellation
  - Test context cancellation propagation
```

#### Rate Limiting Tests

```go
TestRateLimiting
  - Test per-domain rate limiting
  - Test concurrent fetches respect rate limit
  - Test rate limit doesn't block different domains
  - Test configurable rate limit values

TestWorkerPool
  - Test concurrent fetch with worker pool
  - Test worker pool size configuration (5-20 workers)
  - Test error in one worker doesn't affect others
```

#### URL Validation Tests (SSRF Prevention)

```go
TestURLValidation
  - Test http/https schemes allowed
  - Test ftp/file/gopher schemes rejected
  - Test localhost rejected (localhost, 127.0.0.1, ::1)
  - Test 10.0.0.0/8 rejected (RFC 1918)
  - Test 172.16.0.0/12 rejected (RFC 1918)
  - Test 192.168.0.0/16 rejected (RFC 1918)
  - Test link-local addresses rejected (169.254.0.0/16, fe80::/10)
  - Test public IPs allowed
  - Test IPv6 addresses handled correctly
```

### 2. Normaliser Package (`pkg/normalizer`)

#### Feed Parsing Tests

```go
TestParseFeed
  - Test RSS 1.0 parsing
  - Test RSS 2.0 parsing
  - Test Atom 1.0 parsing
  - Test JSON Feed parsing
  - Test malformed XML handling
  - Test malformed JSON handling
  - Test empty feed
  - Test feed with no entries
```

#### HTML Sanitization Tests
**Critical: These tests prevent CVE-2009-2937 XSS attacks**

```go
TestHTMLSanitization
  - Test <script> tag removal
  - Test <object> tag removal
  - Test <embed> tag removal
  - Test <iframe> tag removal
  - Test <base> tag removal
  - Test javascript: URI removal in href
  - Test javascript: URI removal in src
  - Test data: URI removal in href
  - Test data: URI removal in src
  - Test vbscript: URI removal
  - Test event handler removal (onclick, onerror, onload, etc.)
  - Test style attribute with expression() removal
  - Test safe tags preserved (<p>, <a>, <img>, <strong>, etc.)
  - Test safe attributes preserved (href, src, alt, title)
  - Test nested attack vectors (<scr<script>ipt>)
  - Test encoded attack vectors (%3Cscript%3E)
  - Test Unicode attack vectors (\u003Cscript\u003E)
  - Test HTML entity attack vectors (&lt;script&gt;)
```

#### Character Encoding Tests

```go
TestCharacterEncoding
  - Test UTF-8 input preserved
  - Test ISO-8859-1 conversion to UTF-8
  - Test Windows-1252 conversion to UTF-8
  - Test invalid UTF-8 sequences (replace with U+FFFD)
  - Test HTML entities decoded (&nbsp;, &amp;, etc.)
  - Test numeric entities decoded (&#8220;, &#x2014;)
  - Test mixed encoding detection
  - Test BOM handling
```

#### URL Resolution Tests

```go
TestURLResolution
  - Test relative URLs converted to absolute (using feed URL)
  - Test xml:base attribute respected
  - Test nested xml:base attributes
  - Test URLs in href attributes
  - Test URLs in src attributes
  - Test URLs in content HTML
  - Test protocol-relative URLs (//example.com)
  - Test already absolute URLs unchanged
```

#### Date Normalization Tests

```go
TestDateNormalization
  - Test RFC 822 date parsing (Mon, 02 Jan 2006 15:04:05 MST)
  - Test RFC 3339 date parsing (2006-01-02T15:04:05Z)
  - Test ISO 8601 date parsing
  - Test various timezone formats
  - Test missing timezone (assume UTC)
  - Test missing date (use feed date)
  - Test missing feed date (use fetch time)
  - Test future dates (configurable behavior)
  - Test invalid dates (graceful fallback)
  - Test date edge cases (year 2038, leap seconds)
```

#### ID Generation Tests

```go
TestIDGeneration
  - Test existing GUID preserved
  - Test existing ID preserved
  - Test fallback to permalink
  - Test fallback to title+date hash
  - Test fallback to content hash
  - Test stable IDs across multiple parses
  - Test ID uniqueness within feed
```

#### Content Normalization Tests

```go
TestContentNormalization
  - Test full content preferred over summary
  - Test summary used when content absent
  - Test content type detection (html, text, xhtml)
  - Test author extraction from entry
  - Test author fallback to feed level
  - Test multiple authors handling
  - Test categories/tags extraction
  - Test enclosure handling (podcasts, images)
```

### 3. Repository Package (`pkg/repository`)

#### Database Operations Tests

```go
TestDatabaseSetup
  - Test schema creation
  - Test migration from older schema versions
  - Test indexes created correctly
  - Test foreign key constraints
  - Test PRAGMA journal_mode=WAL enabled

TestFeedOperations
  - Test AddFeed with new URL
  - Test AddFeed with duplicate URL (UNIQUE constraint)
  - Test UpdateFeed metadata
  - Test GetFeeds (all feeds)
  - Test GetFeeds (active only)
  - Test GetFeedByURL
  - Test RemoveFeed (cascade deletes entries)
  - Test SetFeedActive/Inactive
```

#### Entry Operations Tests

```go
TestEntryOperations
  - Test UpsertEntry (insert new)
  - Test UpsertEntry (update existing)
  - Test UNIQUE constraint on (feed_id, entry_id)
  - Test GetRecentEntries (last N days)
  - Test GetEntriesSince (timestamp)
  - Test GetEntriesByFeed
  - Test PruneOldEntries (delete by age)
  - Test entry ordering (newest first)
```

#### Caching Operations Tests

```go
TestCacheOperations
  - Test UpdateFeedCache (ETag storage)
  - Test UpdateFeedCache (Last-Modified storage)
  - Test GetFeedCache retrieval
  - Test exact cache value preservation (no modification)
  - Test null cache handling
```

#### Transaction Tests

```go
TestTransactions
  - Test batch insert with transaction
  - Test rollback on error
  - Test commit on success
  - Test concurrent transaction isolation
```

#### SQL Injection Prevention Tests

```go
TestSQLInjectionPrevention
  - Test malicious URL in AddFeed ('; DROP TABLE feeds; --)
  - Test malicious title with SQL keywords
  - Test malicious entry_id
  - Test all inputs properly parameterized
```

### 4. Generator Package (`pkg/generator`)

#### Template Tests

```go
TestTemplateRendering
  - Test valid template renders
  - Test template variables populated correctly
  - Test auto-escaping of HTML in titles
  - Test safe HTML content (from sanitizer) renders unescaped
  - Test missing variables handled gracefully
  - Test template syntax errors detected
  - Test custom template loading
```

#### Output Generation Tests

```go
TestOutputGeneration
  - Test HTML5 validation
  - Test semantic HTML structure
  - Test entries sorted by date (newest first)
  - Test date grouping (optional)
  - Test feed attribution displayed
  - Test links to original posts
  - Test empty entries list handled
  - Test output file creation
  - Test output file permissions (0644)
```

#### Date Formatting Tests

```go
TestDateFormatting
  - Test relative dates ("2 hours ago")
  - Test absolute dates ("January 2, 2025")
  - Test timezone handling
  - Test locale-specific formatting (if supported)
```

### 5. Config Package (`pkg/config`)

#### Configuration Parsing Tests

```go
TestConfigParsing
  - Test INI format parsing
  - Test feeds.txt parsing (one URL per line)
  - Test comments ignored (#)
  - Test empty lines ignored
  - Test default values applied
  - Test per-feed overrides
  - Test invalid URLs rejected
  - Test missing config file handled
```

#### Configuration Validation Tests

```go
TestConfigValidation
  - Test required fields present
  - Test URL format validation
  - Test numeric range validation (fetch intervals, days)
  - Test path validation (output directory exists)
  - Test conflicting settings detected
```

## Integration Tests

### End-to-End Pipeline Tests

```go
TestFullPipeline
  - Test complete flow: fetch → normalize → store → generate
  - Test with real feed URLs (snapshot testing)
  - Test with multiple feeds
  - Test with feeds of different formats
  - Test site regeneration with cached data
  - Test update cycle (fetch new entries only)
```

### Feed Format Integration Tests

Create test fixtures in `testdata/` directory:

```
testdata/
  feeds/
    rss1.xml          - Valid RSS 1.0 feed
    rss2.xml          - Valid RSS 2.0 feed
    atom.xml          - Valid Atom 1.0 feed
    jsonfeed.json     - Valid JSON Feed
    malformed.xml     - Malformed XML
    empty.xml         - Empty feed
    no-dates.xml      - Feed with missing dates
    no-ids.xml        - Feed with missing entry IDs
    relative-urls.xml - Feed with relative URLs
    xss-attack.xml    - Feed with XSS attempts
    huge-feed.xml     - Large feed (for performance testing)
```

```go
TestFeedFormatIntegration
  - Test each format in testdata/feeds/
  - Test parsing → normalization → storage
  - Verify normalized output matches expected results
  - Use golden file testing for outputs
```

### Conditional Request Integration Tests

Requires mock HTTP server:

```go
TestConditionalRequestIntegration
  - Test first fetch (no cache)
  - Test second fetch (cache hit, 304)
  - Test third fetch (cache miss, 200 with new content)
  - Test ETag-only server
  - Test Last-Modified-only server
  - Test server supporting both
  - Test server supporting neither
```

### Error Recovery Integration Tests

```go
TestErrorRecovery
  - Test one failing feed doesn't block others
  - Test retry logic with exponential backoff
  - Test recovery after temporary failure
  - Test permanent failure handling (disable feed)
  - Test partial feed parse (some entries valid)
```

## Security Tests

### XSS Attack Tests
**Critical: Prevent CVE-2009-2937**

```go
TestXSSPrevention
  - Test <script>alert(1)</script>
  - Test <img src=x onerror=alert(1)>
  - Test <a href="javascript:alert(1)">
  - Test <iframe src="evil.com">
  - Test <object data="evil.com">
  - Test <embed src="evil.com">
  - Test <base href="evil.com">
  - Test <link rel="import" href="evil.com">
  - Test <svg><script>alert(1)</script></svg>
  - Test <math><script>alert(1)</script></math>
  - Test encoded attacks (%3Cscript%3E)
  - Test unicode attacks (\u003Cscript\u003E)
  - Test nested attacks (<scr<script>ipt>)
  - Test mutation XSS attacks
  - Test CSS injection (expression(), behavior:)
  - Test HTML entity attacks (&lt;script&gt;)
```

### SSRF Attack Tests

```go
TestSSRFPrevention
  - Test http://localhost rejected
  - Test http://127.0.0.1 rejected
  - Test http://[::1] rejected
  - Test http://10.0.0.1 rejected (private network)
  - Test http://192.168.1.1 rejected (private network)
  - Test http://169.254.169.254 rejected (link-local)
  - Test http://metadata.google.internal rejected
  - Test redirect to localhost rejected
  - Test DNS rebinding scenario
  - Test IPv6 private addresses rejected
```

### SQL Injection Tests

```go
TestSQLInjectionPrevention
  - Test '; DROP TABLE feeds; -- in URL
  - Test ' OR '1'='1 in search
  - Test UNION SELECT in entry_id
  - Test comment injection (--,  /*, */)
  - Test stacked queries (;)
  - Test blind SQL injection attempts
  - Verify all queries use parameterization
```

### Path Traversal Tests

```go
TestPathTraversalPrevention
  - Test ../../../etc/passwd in paths
  - Test absolute paths outside output directory
  - Test symlink following (if applicable)
  - Test output path validation
```

## Performance Tests

### Benchmark Tests

```go
BenchmarkFetchFeeds
  - Benchmark concurrent fetching (5, 10, 20 workers)
  - Benchmark with different feed sizes
  - Benchmark with cached vs uncached fetches

BenchmarkParseFeed
  - Benchmark RSS 1.0 parsing
  - Benchmark RSS 2.0 parsing
  - Benchmark Atom parsing
  - Benchmark JSON Feed parsing
  - Benchmark large feeds (1000+ entries)

BenchmarkSanitizeHTML
  - Benchmark simple HTML sanitization
  - Benchmark complex HTML with many tags
  - Benchmark HTML with nested structures

BenchmarkDatabaseOperations
  - Benchmark insert 1000 entries
  - Benchmark query recent entries (1 day, 7 days, 30 days)
  - Benchmark update existing entries
  - Benchmark with different index strategies

BenchmarkGenerateHTML
  - Benchmark template rendering (100, 500, 1000 entries)
  - Benchmark with different template complexities
```

### Load Tests

```go
TestConcurrentFetching
  - Test 100 feeds fetched concurrently
  - Test 1000 feeds fetched concurrently
  - Measure memory usage
  - Measure connection pool exhaustion
  - Test worker pool saturation

TestDatabaseConcurrency
  - Test concurrent reads and writes
  - Test WAL mode performance
  - Test connection pool limits
```

### Memory Tests

```go
TestMemoryUsage
  - Test memory with 100 feeds
  - Test memory with 10,000 entries
  - Test memory with large feed content
  - Test for memory leaks (repeated cycles)
  - Test goroutine leak detection
```

## Regression Tests

### Historical Bug Prevention

Based on lessons learned from Planet/Venus:

```go
TestCharacterEncodingBugs
  - Test Windows-1252 mislabeled as UTF-8
  - Test mixed encodings in single feed
  - Test emoji and Unicode edge cases

TestDateParsingBugs
  - Test non-standard date formats seen in wild
  - Test timezone offset edge cases
  - Test daylight saving time transitions

TestHTMLParsingBugs
  - Test unclosed tags
  - Test mismatched tags
  - Test invalid nesting
  - Test HTML5 entities
```

## Test Data Management

### Test Fixtures

Store test data in `testdata/` directory:

```
testdata/
  feeds/              - Sample feed files
  expected/           - Expected output (golden files)
  malicious/          - XSS and attack vectors
  encodings/          - Various character encodings
  templates/          - Test templates
```

### Mock Servers

Use `httptest` for HTTP testing:

```go
// Example mock server for unit tests (httptest chooses random port)
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  // Check conditional request headers
  if r.Header.Get("If-None-Match") == `"abc123"` {
    w.WriteHeader(http.StatusNotModified)
    return
  }

  w.Header().Set("ETag", `"abc123"`)
  w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
  w.Write(feedData)
}))
defer server.Close()
```

**Port Allocation for Integration Tests**

When running integration tests that require specific ports (not using httptest):
- Use ports **9000+** to avoid conflicts with other local services
- Recommended allocation:
  - `localhost:9000` - Main test feed server
  - `localhost:9001` - Secondary test feed server (for multi-source tests)
  - `localhost:9002` - Mock server with rate limiting
  - `localhost:9003` - Mock server for redirect tests
  - `localhost:9004+` - Additional test servers as needed

## Test Coverage Goals

- **Overall coverage**: 80%+ (measured with `go test -cover`)
- **Critical paths**: 100% (sanitization, conditional requests, URL validation)
- **Error paths**: 80%+ (all error returns tested)
- **Security code**: 100% (XSS prevention, SSRF prevention, SQL injection prevention)

## Continuous Testing

### Pre-commit Checks

```bash
# Run before every commit
go fmt ./...
go vet ./...
golangci-lint run
go test ./...
go test -race ./...
```

### CI/CD Pipeline

```yaml
# GitHub Actions example
- Run all unit tests
- Run integration tests
- Run security tests
- Run benchmarks (track performance over time)
- Check test coverage
- Run static analysis (gosec, staticcheck)
- Test on multiple Go versions (1.21, 1.22, 1.23)
- Test on multiple platforms (Linux, macOS, Windows)
```

## Manual Testing

### Real-World Feed Testing

Maintain a list of diverse real-world feeds for manual testing:

- Personal blogs (WordPress, Medium, Ghost)
- News sites (RSS 2.0)
- Podcasts (with enclosures)
- Reddit/Lobsters (Atom)
- GitHub releases (Atom)
- YouTube channels (Atom)
- Mastodon/fediverse feeds (Atom)

Test with these feeds regularly to catch edge cases.

### Security Audit

Before each release:

1. Run OWASP ZAP or similar security scanner on generated HTML
2. Manual code review of sanitization logic
3. Manual review of all HTTP request handling
4. Dependency vulnerability scan (`go list -json -m all | nancy sleuth`)
5. Test with intentionally malicious feeds from testdata/malicious/

## Test Execution Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detector
go test -race ./...

# Run specific package tests
go test ./pkg/crawler
go test ./pkg/normalizer -v

# Run specific test
go test ./pkg/crawler -run TestConditionalRequests

# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...

# Run tests with timeout
go test -timeout 30s ./...

# Run tests in parallel
go test -parallel 4 ./...

# Run only short tests (skip integration tests)
go test -short ./...
```

## Test Organization

Use Go's standard testing idioms:

```go
// Table-driven tests
func TestURLValidation(t *testing.T) {
  tests := []struct {
    name    string
    url     string
    wantErr bool
  }{
    {"valid http", "http://example.com/feed", false},
    {"valid https", "https://example.com/feed", false},
    {"localhost rejected", "http://localhost/feed", true},
    {"private IP rejected", "http://192.168.1.1/feed", true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      err := ValidateURL(tt.url)
      if (err != nil) != tt.wantErr {
        t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
      }
    })
  }
}

// Subtests for organization
func TestCrawler(t *testing.T) {
  t.Run("ConditionalRequests", func(t *testing.T) { /* ... */ })
  t.Run("RateLimiting", func(t *testing.T) { /* ... */ })
  t.Run("ErrorHandling", func(t *testing.T) { /* ... */ })
}

// Helper functions
func mustParseFeed(t *testing.T, filename string) *Feed {
  t.Helper()
  data, err := os.ReadFile(filename)
  if err != nil {
    t.Fatal(err)
  }
  feed, err := ParseFeed(data)
  if err != nil {
    t.Fatal(err)
  }
  return feed
}
```

## Success Criteria

The testing plan is successful when:

1. ✅ All unit tests pass with >80% coverage
2. ✅ All security tests pass (100% coverage on security-critical code)
3. ✅ Integration tests pass with real-world feeds
4. ✅ Performance benchmarks meet targets (10+ feeds/second)
5. ✅ No XSS vulnerabilities found in security audit
6. ✅ No SSRF vulnerabilities found in security audit
7. ✅ No SQL injection vulnerabilities found in security audit
8. ✅ Memory usage stays within acceptable bounds (<100MB for 1000 feeds)
9. ✅ Race detector finds no data races
10. ✅ Manual testing with diverse real-world feeds succeeds
