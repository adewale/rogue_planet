# Changelog

All notable changes to Rogue Planet will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added - Architecture & Testing
- **pkg/fetcher package**: Extracted feed processing business logic
  - Separates orchestration (concurrency, rate limiting) from core logic
  - Dependency injection with interfaces (FeedCrawler, FeedNormalizer, FeedRepository)
  - 88.5% test coverage with 10 comprehensive tests (6 unit + 2 integration + 2 benchmarks)
  - Test-to-code ratio: 4.1:1 (highest in codebase)
- **Concurrency testing suite**: Timing-based tests prevent performance regressions
  - `TestFetchFeed_Concurrency`: Verifies parallel execution (catches 10x slowdowns)
  - `TestFetchFeed_MutexProtectsDatabase`: Validates selective locking
  - Uses atomic counters to measure actual concurrent operations
- **Documentation enhancements**
  - 6 new lessons in LESSONS_LEARNED.md (concurrency, mutex placement, testing)
  - GO_AUDITING_HEURISTICS.md v2.2: mutex placement and concurrency testing sections
  - Comprehensive analysis of performance regression and prevention

### Fixed - Performance
- **Critical 10x performance regression**: Restored concurrent feed processing
  - Issue: Mutex incorrectly placed around entire FetchFeed() operation
  - Impact: HTTP fetching and parsing were serialized instead of concurrent
  - Solution: Selective locking (only database operations protected)
  - Result: 50 feeds in ~10s instead of ~100s

### Changed - Code Quality
- Extracted `setVerboseLogging()` helper to remove duplication
- Improved test coverage: 67.7% → 70.4% overall (+2.7pp)
- Added timing assertions that prevent future performance regressions

### Added - JSON Feed Testing
- **Comprehensive JSON Feed 1.0 and 1.1 test coverage**
  - Test fixtures for JSON Feed 1.0 (author field) and 1.1 (authors array, language field)
  - Edge case testing (missing dates, missing IDs, Unicode, XSS, future dates)
  - Security validation (XSS prevention, malicious URLs)
  - Real-world compatibility testing
- **JSON Feed support documentation**
  - Already supported transparently via gofeed library
  - Users can add JSON Feed URLs (e.g., `https://example.micro.blog/feed.json`)
  - Same security guarantees as RSS/Atom (HTML sanitization, SSRF prevention)

### Planned for 1.0.0
- Feed autodiscovery (parse HTML for RSS/Atom/JSON Feed links)
- Intelligent feed scheduling (adaptive polling)
- Full production deployment documentation
- Binary distribution packages

## [0.4.0] - 2025-10-30

### Added - Production HTTP Performance
- **Per-domain rate limiting** using token bucket algorithm
  - Default: 60 requests/minute per domain with burst of 10
  - Configurable via `requests_per_minute` and `rate_limit_burst`
  - Thread-safe concurrent access with RWMutex
  - Observability via Stats() method
- **Fine-grained HTTP timeouts**
  - `http_timeout_seconds`: Overall request timeout (default: 30)
  - `dial_timeout_seconds`: TCP connection timeout (default: 10)
  - `tls_handshake_timeout_seconds`: TLS handshake timeout (default: 10)
  - `response_header_timeout_seconds`: Response header timeout (default: 10)
- **301 permanent redirect handling**
  - Automatically updates feed URLs in database on 301 responses
  - Logs redirect for transparency
  - Preserves feed metadata (ETag, Last-Modified)
- **Retry-After header support**
  - Respects server-specified retry delays (RFC 7231)
  - Honors HTTP 429 rate limit responses
  - Integrates with exponential backoff retry logic

### Changed
- Updated README.md development status to v0.4.0
- Enhanced "Good Netizen Behavior" documentation
- Added rate limiter implementation notes to CLAUDE.md

### Testing
- 11 new unit tests for rate limiter (concurrency, context cancellation, stats)
- Integration tests verify backwards compatibility with old configs
- All 8 packages passing tests (~37 seconds total)

### Dependencies
- Added `golang.org/x/time v0.14.0` for rate limiting

## [0.3.0] - 2025-10-16

### Added - Entry Spam Prevention
- **`filter_by_first_seen` config option**: Filter entries by discovery date instead of published date
  - Prevents flooding timeline when adding new feeds with historical content
  - Configurable: `filter_by_first_seen = true` in `[planet]` section
- **`sort_by` config option**: Sort by "published" or "first_seen" (Venus #15: Stable Sort Dates)
  - Stable chronological ordering (entries don't "jump" when authors update them)
  - Configurable: `sort_by = "first_seen"` in `[planet]` section
  - Solves long-standing Planet Venus issue with timeline instability
- **Automatic database migration**: Backfills `first_seen` for existing entries using COALESCE

### Security - SQL Injection Hardening
- **SQL parameter validation**: Whitelist validation for dynamic SQL field names
- **Improved NULL handling**: COALESCE-based fallback in migration queries
- **Config validation**: `sort_by` field validated during config parsing

### Testing
- 5 new unit tests for entry spam prevention
- 2 new integration tests for filter/sort behavior
- 6 new config validation tests
- All tests passing with improved edge case coverage

### Documentation
- Updated README.md with development status notice
- Created LESSONS_FROM_MARS.md analyzing two different "Mars" aggregators
- Merged WISHLIST.md into TODO.md for unified roadmap
- TODO.md now includes P0/P1/P2/P3 feature priorities with Venus issue references
- Updated CHANGELOG.md with v0.3.0 release notes

## [0.2.0] - 2025-10-13

### Added - OPML Support
- **OPML Package**: Full OPML 1.0 and 2.0 support with 91.8% test coverage
  - Parse and generate OPML files
  - Support for both `xmlUrl`/`url` and `text`/`title` attributes
  - RFC 822 date handling
  - Nested outline support (categories/folders)
- `rp import-opml <file> [--dry-run]` - Import feeds from OPML file
- `rp export-opml [--output FILE]` - Export feeds to OPML format
- Dry-run mode for previewing imports before committing
- Duplicate feed detection during import
- Compatibility with Feedly, Inoreader, NewsBlur, The Old Reader exports

### Added - Configuration Validation
- `rp verify` - Validate configuration and environment
  - Config file syntax validation
  - Database accessibility check
  - Output directory writability check
  - Custom template existence verification
  - Feed and entry count reporting

### Added - Package Documentation
- Godoc-compliant package comments for all 6 packages
- Package-level documentation for crawler, normalizer, generator, config, repository, opml

### Improved - Test Coverage
- Generator package: 54.3% → 86.0% coverage
  - Tests for feed sidebar rendering
  - Owner info display tests
  - Template function tests (formatDate, relativeTime, truncate, stripHTML)
  - Static asset copying tests
  - Subtitle support tests
- Repository package: 75.3% → 85.6% coverage
  - UpdateFeedError function tests
  - CountEntries function tests
  - CountRecentEntries function tests
  - Database initialization error tests
- Overall project: Maintained >85% average coverage across core packages

### Improved - Integration Tests
- 6 new OPML integration tests with proper test isolation
- Added `setupTestDir()` helper to prevent database conflicts
- Fixed nil pointer panics in existing command tests
- All tests use `t.TempDir()` for automatic cleanup

### Fixed
- Test isolation: Tests now change working directory to temp dirs
- Command tests: Added proper nil checks before accessing `err.Error()`
- Error handling: All error messages follow Go conventions (lowercase, no trailing punctuation)

### Documentation
- Updated README.md with OPML commands in organized command sections
- Updated WORKFLOWS.md with comprehensive OPML import/export workflows
  - Backing up feed lists
  - Migrating from other feed readers
  - Sharing feed lists
  - Merging multiple OPML files
- Updated QUICKSTART.md with OPML and verify commands
- Updated command help text with new commands

## [0.1.0] - 2025-10-10

### Added - Core Functionality
- **Crawler Package**: HTTP conditional requests (ETag/Last-Modified), SSRF prevention, retry logic, gzip decompression
- **Normalizer Package**: RSS 1.0/2.0, Atom 1.0 parsing, HTML sanitization (CVE-2009-2937 prevention), character encoding handling
- **Repository Package**: SQLite database with WAL mode, foreign key constraints, HTTP cache tracking
- **Generator Package**: Static HTML generation, responsive design, classic Planet sidebar, custom template support
- **Config Package**: INI format parser, simple feeds.txt support, configuration validation

### Added - CLI Commands
- `rp init [-f FILE]` - Initialize new planet with optional feeds import
- `rp add-feed <url>` - Add individual feed
- `rp add-all -f FILE` - Bulk feed import
- `rp remove-feed <url>` - Remove feed
- `rp list-feeds` - Display all feeds with status
- `rp status` - Show planet statistics
- `rp update` - Fetch feeds and regenerate site
- `rp fetch` - Fetch feeds without generating
- `rp generate` - Generate site without fetching
- `rp prune --days N` - Remove old entries
- `rp version` - Show version information

### Added - Documentation
- README.md with installation and usage instructions
- QUICKSTART.md - 5-minute setup guide
- WORKFLOWS.md - Comprehensive workflow documentation
- TESTING.md - Testing strategy and guidelines
- CLAUDE.md - Development guidance for AI assistance
- CONTRIBUTING.md - Contribution guidelines
- specs/rogue-planet-spec.md - Complete technical specification
- specs/testing-plan.md - Testing strategy
- examples/ - Example configurations and themes

### Added - Testing
- 100+ test cases across all packages
- 80%+ average test coverage
- Unit tests for all core packages
- Integration tests for full pipeline
- Real-world feed snapshots (Daring Fireball, Asymco)
- Live network tests (with build tags)
- Security tests (XSS, SSRF, SQL injection prevention)
- Race detector validation

### Added - Build & Automation
- Makefile with 18 targets
- `make quick` - Fast development iteration
- `make check` - All quality checks (fmt, vet, test, race)
- `make coverage` - HTML coverage reports
- `make install` - Install to GOPATH/bin with smart path detection

### Security
- **XSS Prevention**: HTML sanitization using bluemonday
- **SSRF Protection**: URL validation blocking localhost and private IPs
- **SQL Injection Prevention**: Prepared statements throughout
- **Content Security Policy**: Headers in generated HTML
- **Safe HTML Rendering**: Go's html/template with proper escaping

### Features
- Classic Planet Planet sidebar showing all subscribed feeds
- Feed health status display (last updated, error counts)
- Smart content fallback (recent entries if time window empty)
- Date grouping ("Today", "Yesterday", etc.)
- Relative time display ("2 hours ago")
- **Responsive mobile-friendly design** (Venus #36: Responsive Image Constraints)
  - Images automatically constrained to page width with CSS max-width: 100%
  - Prevents layout overflow on mobile devices
- Custom theme support

---

## Version History Summary

| Version | Date | Description |
|---------|------|-------------|
| 0.1.0 | 2025-10-10 | Initial development release - all core features complete |
| 0.2.0 | 2025-10-13 | OPML support, verify command, improved test coverage |
| 0.3.0 | 2025-10-16 | Entry spam prevention, SQL injection hardening, docs consolidation |
| 1.0.0 | TBD | Planned production release (autodiscovery, 301 handling, intelligent scheduling) |

## Links

- [Repository](https://github.com/adewale/rogue_planet)
- [Documentation](README.md)
- [Quick Start](QUICKSTART.md)
- [Workflows](WORKFLOWS.md)
- [Contributing](CONTRIBUTING.md)
- [License](LICENSE)
