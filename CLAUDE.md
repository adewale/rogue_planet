# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Rogue Planet (rp) is a modern feed aggregator written in Go, inspired by Venus/Planet. It downloads RSS and Atom feeds from multiple sources and aggregates them into a single chronological stream published as a static HTML page.

## Core Architecture

The system follows a clear pipeline architecture with four distinct phases:

```
Config → Crawler → Normaliser → Repository → Site Generator → HTML Output
```

**1. Crawler**: Fetches feed content via HTTP/HTTPS with proper conditional request support (ETag/Last-Modified) and gzip decompression

**2. Normaliser**: Transforms all feed formats (RSS 1.0/2.0, Atom, JSON Feed) into a canonical internal format and sanitizes HTML content

**3. Repository**: Stores normalized entries in SQLite database with proper indexing for date-based queries. Implements smart fallback to show recent entries if nothing in configured time window.

**4. Site Generator**: Queries recent entries from database and generates static HTML using Go's `html/template` with responsive layout and classic Planet Planet sidebar showing all subscribed feeds

## Critical Implementation Requirements

### HTTP Conditional Requests (MANDATORY)

The crawler MUST implement proper HTTP conditional requests to avoid being rate-limited or banned:

- Store `ETag` and `Last-Modified` headers EXACTLY as received from servers (including quotes in ETag)
- Send `If-None-Match` (with stored ETag) and `If-Modified-Since` (with stored Last-Modified) on subsequent requests
- Handle 304 Not Modified responses correctly (don't re-parse or re-store)
- NEVER fabricate or modify these header values
- Update stored values on EVERY successful response

This is non-negotiable for a well-behaved feed aggregator. See specs/rogue-planet-spec.md lines 64-156 for detailed implementation.

### Security Requirements (CVE-2009-2937 Prevention)

**HTML Sanitization is CRITICAL** - Planet Venus suffered from XSS vulnerabilities:

- Use `github.com/microcosm-cc/bluemonday` for HTML sanitization
- Sanitize ALL feed content before storage
- Strip JavaScript URIs in src/href attributes (`javascript:`, `data:`)
- Remove ALL event handlers (onclick, onerror, etc.)
- Remove dangerous tags: `<script>`, `<object>`, `<embed>`, `<iframe>`, `<base>`
- Only allow http/https URL schemes in links and images
- Use Go's `html/template` (not `text/template`) for auto-escaping
- Implement Content Security Policy headers in generated HTML

See specs/rogue-planet-spec.md lines 548-762 for complete security requirements.

### SSRF Prevention

Validate all feed URLs to prevent Server-Side Request Forgery:

- Only allow http/https schemes
- Block localhost, 127.0.0.1, ::1
- Block private IP ranges (RFC 1918)
- Block link-local addresses

### Database Schema

SQLite database with two main tables:

**feeds table**: Stores feed metadata including `etag`, `last_modified`, `fetch_error_count`, `next_fetch` for intelligent scheduling

**entries table**: Stores normalized entries with UNIQUE constraint on (feed_id, entry_id) to prevent duplicates

See specs/rogue-planet-spec.md lines 256-302 for complete schema.

## Development Commands

```bash
# Quick Development (recommended during active development)
make quick                     # Format + test + build (fast iteration)
make check                     # All quality checks: fmt + vet + test + race

# Build the project
make build                     # Build to bin/rp
go build -o rp ./cmd/rp       # Or build directly with go

# Run tests
make test                      # Run all tests (excludes network tests)
make test-short                # Run tests without verbose output
go test ./...                  # Or run with go

# Run specific package tests
go test ./pkg/crawler -v
go test ./pkg/normalizer -v
go test ./pkg/repository -v
go test ./pkg/generator -v

# Run integration tests
make test-integration          # Full pipeline integration tests

# Run with race detector
make test-race                 # Detect race conditions
go test -race ./...

# Coverage reporting
make coverage                  # Generate HTML coverage report in coverage/coverage.html

# Run live network tests (requires internet)
go test -tags=network ./pkg/crawler -v

# Code quality
make fmt                       # Format all Go code
make vet                       # Run go vet
make lint                      # Run golangci-lint (if installed)

# Build for production
make build                     # With optimizations
CGO_ENABLED=1 go build -ldflags="-s -w" -o rp ./cmd/rp

# Install to GOPATH/bin
make install                   # Install globally

# Generate example planet
make run-example               # Create test planet in /tmp/rogue-planet-example
```

## Project Structure

```
cmd/rp/              - CLI entry point and command handlers
  main.go            - Main entry point with command routing
  commands.go        - Command implementation (cmdInit, cmdAddFeed, cmdUpdate, etc.)
  logger_test.go     - Logging utilities for tests
  *_test.go          - Integration tests
  opml_integration_test.go - OPML integration tests
pkg/crawler/         - HTTP fetching with conditional request support
  crawler.go         - Core HTTP fetching with ETag/Last-Modified
  crawler_test.go    - Unit tests
  crawler_live_test.go - Network tests (requires -tags=network)
pkg/normalizer/      - Feed parsing and HTML sanitization
  normalizer.go      - Feed parsing and content sanitization
  normalizer_test.go - Unit tests with mock feeds
  normalizer_realworld_test.go - Tests with real feed snapshots
pkg/repository/      - SQLite database operations
  repository.go      - Database schema and operations
  repository_test.go - Database tests with t.TempDir()
pkg/generator/       - Static HTML generation
  generator.go       - Template rendering and HTML generation
  generator_test.go  - Template tests
  generator_integration_test.go - Full generation pipeline tests
pkg/config/          - Configuration parsing (INI format)
  config.go          - Config loading and validation
  config_test.go     - Config parsing tests
pkg/opml/            - OPML parsing and generation (feed list import/export)
  opml.go            - OPML 1.0/2.0 parsing and generation
  opml_test.go       - OPML parsing tests (91.8% coverage)
pkg/ratelimit/       - Per-domain rate limiting
  ratelimit.go       - Rate limiter manager using token bucket algorithm
  ratelimit_test.go  - Rate limiter tests including concurrency tests
testdata/            - Test fixtures (saved feed snapshots)
specs/               - Specifications and design documents
examples/            - Example configurations and themes
```

## Code Organization Patterns

**Command Pattern**: All CLI commands in `cmd/rp/commands.go` follow a consistent pattern:
- Each command has an `Options` struct (e.g., `InitOptions`, `UpdateOptions`)
- Each command has a `cmd*` function (e.g., `cmdInit`, `cmdUpdate`) that takes options
- Output is written to `opts.Output` for testability (can be os.Stdout or test buffer)
- Database and config paths are configurable via options for testing

**Testing Pattern**: Tests use `t.TempDir()` for automatic cleanup:
```go
func TestSomething(t *testing.T) {
    dir := t.TempDir()  // Automatically cleaned up after test
    dbPath := filepath.Join(dir, "test.db")
    repo, err := repository.New(dbPath)
    // ... test code
}
```

## Key Dependencies

- Feed parsing: `github.com/mmcdole/gofeed`
- HTML sanitization: `github.com/microcosm-cc/bluemonday`
- SQLite driver: `github.com/mattn/go-sqlite3` (CGO) or `modernc.org/sqlite` (pure Go)
- Rate limiting: `golang.org/x/time/rate`
- HTML parsing: `golang.org/x/net/html`
- Character encoding: `golang.org/x/text/encoding`

## CLI Interface

Implemented commands:

```bash
# Core Commands
rp init [-f FILE]             # Initialize new planet configuration
rp add-feed <url>             # Add a new feed
rp add-all -f FILE            # Add multiple feeds from a file
rp remove-feed <url>          # Remove a feed
rp list-feeds                 # List all configured feeds
rp status                     # Show planet status (feed and entry counts)

# Operation Commands
rp update                     # Fetch all feeds and regenerate site
rp fetch                      # Fetch feeds without generating
rp generate                   # Regenerate site without fetching
rp prune --days N             # Prune entries older than N days

# Import/Export Commands
rp import-opml <file> [--dry-run]  # Import feeds from OPML file
rp export-opml [--output FILE]     # Export feeds to OPML format

# Utility Commands
rp verify                     # Validate configuration and environment
rp version                    # Show version information
```

## Important Design Principles

1. **Static Output**: Generate complete HTML files, not dynamic pages - this survives traffic spikes and scales infinitely

2. **Separate Concerns**: Keep fetch, normalize, store, and generate as distinct phases that can run independently

3. **Good Netizen Behavior**:
   - Include version and contact URL in User-Agent: `RoguePlanet/1.0 (+https://yoursite.com/about)`
   - Respect Cache-Control headers
   - Per-domain rate limiting (default: 60 req/min with burst of 10)
   - Honor 429 responses with exponential backoff and jitter (±10% randomization)
   - Handle 301/308 permanent redirects by auto-updating feed URLs
   - Default fetch interval: 1 hour (not more frequent than 15 minutes)

4. **Graceful Degradation**: One failing feed should not break the entire aggregation - log errors and continue

5. **Data Normalization**: Convert all feed formats to a single canonical internal format (Atom-style) before storage

6. **Performance**: Concurrent fetching with worker pool (5-20 workers), connection pooling, prepared statements, proper database indexes

## Error Handling

- Wrap errors with context using `fmt.Errorf("operation failed: %w", err)`
- Log all fetch errors but continue processing other feeds
- Store fetch errors in database with `fetch_error` and `fetch_error_count` fields
- Implement exponential backoff for repeatedly failing feeds
- Don't expose internal paths or stack traces in user-facing errors

## Testing Strategy

**Test Organization**:
- Unit tests in each package (*_test.go in same directory)
- Integration tests in cmd/rp/ (test full command workflows)
- Network tests use build tag: `// +build network` at top of file
- Real-world tests use saved feed snapshots from testdata/

**Test Coverage Requirements**:
- Maintain >75% coverage across all core packages
- Generate coverage reports with `make coverage` (outputs to coverage/coverage.html)
- Use `go test -cover ./...` for quick coverage check

**Test Types**:
- **Unit tests**: Each component (crawler, normalizer, repository, generator)
- **Format tests**: Various feed formats (RSS 1.0, RSS 2.0, Atom 1.0, JSON Feed)
  - JSON Feed 1.0 and 1.1 test fixtures in `testdata/`
  - Tests for `author` (v1.0) vs `authors` array (v1.1)
  - Edge cases: missing dates/IDs, Unicode, future dates, malicious content
  - Security: XSS prevention, URL scheme filtering
- **Edge case tests**: Malformed feeds, missing dates, missing IDs, encoding issues
- **Security tests**: HTML sanitization with XSS attempts
- **Integration tests**: Full pipeline from init → add feeds → fetch → generate
- **Real-world tests**: Saved snapshots from Daring Fireball, Asymco, etc.
- **Network tests**: Live feed fetching (run with `-tags=network`)

**Testing Best Practices**:
- Use `t.TempDir()` for automatic temp directory cleanup
- Mock HTTP servers for reliable, fast integration tests
- Test both success and error paths
- Verify database state after operations
- Check generated HTML contains expected content

## SQL Guidelines

- ALWAYS use prepared statements with placeholders (`?` or `$1`)
- NEVER concatenate user input into SQL strings
- Use transactions for batch operations
- Enable WAL mode for better concurrency: `PRAGMA journal_mode=WAL`
- Use appropriate indexes on date columns (published, updated, first_seen)

## Template Variables

The site generator makes these variables available in templates:

```
{{.Title}}           - Site title
{{.Subtitle}}        - Site subtitle (optional)
{{.Link}}            - Site link
{{.Updated}}         - Last updated timestamp
{{.Generator}}       - Generator name and version
{{.OwnerName}}       - Planet owner name
{{.OwnerEmail}}      - Planet owner email
{{.GroupByDate}}     - Whether to group entries by date
{{.Entries}}         - Array of entries (sorted newest first)
  {{.Title}}         - Entry title
  {{.Link}}          - Entry permalink
  {{.Author}}        - Entry author
  {{.FeedTitle}}     - Source feed title
  {{.FeedLink}}      - Source feed link
  {{.Published}}     - Published date
  {{.Updated}}       - Updated date
  {{.Content}}       - Sanitized HTML content (safe to output)
  {{.Summary}}       - Sanitized HTML summary
  {{.PublishedRelative}} - Relative time string ("2 hours ago")
{{.DateGroups}}      - Entries grouped by date (if GroupByDate enabled)
  {{.Date}}          - Date for this group
  {{.DateStr}}       - Formatted date string
  {{.Entries}}       - Entries for this date
{{.Feeds}}           - Array of feeds (for sidebar)
  {{.Title}}         - Feed title
  {{.Link}}          - Feed website link
  {{.URL}}           - Feed URL
  {{.LastUpdated}}   - Last fetch time
  {{.ErrorCount}}    - Number of consecutive fetch errors
```

## Configuration Format

Support both simple feeds.txt (one URL per line) and extended config.ini format with per-feed overrides. See specs/rogue-planet-spec.md lines 395-429.

## Feed Handling Edge Cases

Real-world feeds are messy. Handle these scenarios:

- Missing or invalid dates (use feed date, then fetch time as fallback)
- Missing entry IDs (generate from permalink or content hash)
- Relative URLs in content (resolve to absolute using feed URL or xml:base)
- Incorrect character encoding declarations
- Malformed HTML in feed content
- Future dates (configurable: ignore or accept)
- 301/302/308 redirects (update feed URL in database for 301 and 308 permanent redirects)
- Huge feed files (limit to 10MB)
- Network timeouts (30 second default)

## Common Development Workflows

**Making a change to the crawler**:
```bash
# 1. Make your changes to pkg/crawler/crawler.go
# 2. Run tests with race detector
go test -race ./pkg/crawler -v
# 3. Run integration tests
make test-integration
# 4. Quick build and test
make quick
```

**Adding a new CLI command**:
```bash
# 1. Add command case in cmd/rp/main.go
# 2. Create run* function in main.go
# 3. Implement cmd* function in cmd/rp/commands.go with Options struct
# 4. Write tests in cmd/rp/commands_test.go
# 5. Update printUsage() help text
# 6. Run make test and make quick
```

**Debugging feed parsing issues**:
```bash
# 1. Save feed snapshot to testdata/
curl -o testdata/problematic-feed.xml https://example.com/feed
# 2. Create test in pkg/normalizer/normalizer_realworld_test.go
# 3. Run normalizer tests
go test ./pkg/normalizer -v
# 4. Use -v flag to see detailed parsing output
```

**Running the full integration test suite**:
```bash
make quick                     # Fast: fmt + test + build
make check                     # Thorough: fmt + vet + test + race
make coverage                  # With coverage report
```

## Historical Context

Rogue Planet learns from 20+ years of feed aggregator history (Planet, Venus, Mars, Pluto, Moonmoon):

- **Static output** proven superior to dynamic pages (traffic resilience)
- **HTTP conditional requests** are mandatory - poorly-behaved readers get blocked
- **HTML sanitization** is critical - Venus suffered from CVE-2009-2937 XSS vulnerability
- **Simple is sustainable** - fewer features, fewer bugs, longer project life
- **Character encoding is hard** - feeds often lie about their encoding
- **Entry IDs are unreliable** - need fallback ID generation strategies

See specs/rogue-planet-spec.md lines 763-1000 for complete lessons learned.

## Implementation Notes

**Crawler (pkg/crawler/crawler.go)**:
- Implements SSRF prevention via `ValidateURL()` (blocks localhost, private IPs)
- Has `NewForTesting()` constructor that disables SSRF checks for tests
- Stores ETag and Last-Modified headers EXACTLY as received (line 233-235)
- Handles gzip decompression automatically (line 202-211)
- Limits response size to 10MB (MaxFeedSize constant)

**Normalizer (pkg/normalizer/normalizer.go)**:
- Uses bluemonday.UGCPolicy() for HTML sanitization
- Only allows http/https URL schemes (line 54)
- Generates stable IDs using SHA256 hash if GUID missing (line 150-179)
- Handles multiple author formats (entry-level and feed-level)
- Falls back through: PublishedParsed → UpdatedParsed → feed.UpdatedParsed → fetchTime

**Repository (pkg/repository/repository.go)**:
- Enables WAL mode automatically for better concurrency (line 63)
- Enables foreign keys for CASCADE DELETE (line 69)
- `GetRecentEntries()` has smart fallback: tries time window first, falls back to 50 most recent (line 267-314)
- All queries use prepared statements (never string concatenation)
- Times stored as RFC3339 strings in SQLite

**Generator (pkg/generator/generator.go)**:
- Default template embedded in const `defaultTemplate` (line 356-631)
- Supports custom templates via `NewWithTemplate(path)`
- Template uses Go's `html/template` for automatic escaping
- Content is marked as `template.HTML` only AFTER sanitization
- Includes CSP header in HTML output (line 361)

**OPML (pkg/opml/opml.go)**:
- Full OPML 1.0 and 2.0 support with 91.8% test coverage
- Supports both `xmlUrl`/`url` and `text`/`title` attribute naming conventions
- RFC 822 date parsing and generation (`time.RFC1123Z` format)
- Handles nested `<outline>` elements (categories/folders)
- `Parse()` reads OPML from io.Reader, `Generate()` creates OPML from feed list
- `Marshal()` converts to XML bytes, `Write()` writes to file
- Compatible with exports from Feedly, Inoreader, NewsBlur, The Old Reader
- Used by `cmdImportOPML()` and `cmdExportOPML()` commands

**Rate Limiter (pkg/ratelimit/ratelimit.go)**:
- Per-domain rate limiting using token bucket algorithm from `golang.org/x/time/rate`
- Thread-safe with RWMutex for concurrent access
- Lazy limiter creation with double-checked locking pattern
- Default: 60 requests/minute per domain with burst of 10
- Configurable via `requests_per_minute` and `rate_limit_burst` in config.ini
- Fail-open strategy: allows requests on URL parsing errors
- `Wait()` blocks until rate limit allows request (respects context cancellation)
- `Allow()` checks if request would be allowed without blocking
- `Stats()` provides observability into rate limiter state per domain

**Verify Command (cmd/rp/commands.go: cmdVerify)**:
- Validates config.ini syntax and accessibility
- Checks database file exists and can be opened
- Verifies output directory is writable
- Validates custom template file exists (if specified)
- Reports feed and entry counts
- Returns brief, error-focused output
