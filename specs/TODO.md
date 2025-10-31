# Rogue Planet - TODO & Roadmap

## Current Status: v0.4.0 ✅ (Complete)

v0.4.0 is complete with production HTTP performance features including rate limiting, 301 redirect handling, and fine-grained timeouts.

---

## Completed Tasks

### Phase 1: Core Packages ✅

#### Crawler Package (96.6% coverage) ✅
- ✅ HTTP conditional requests (ETag/Last-Modified)
- ✅ SSRF prevention (localhost, private IPs blocked)
- ✅ Retry logic with exponential backoff
- ✅ Size limits (10MB max)
- ✅ Timeout handling (30s default)
- ✅ Gzip/deflate decompression
- ✅ Redirect handling (301/302)
- ✅ Comprehensive test suite (13 test cases)

#### Normalizer Package (79.8% coverage) ✅
- ✅ RSS 1.0, RSS 2.0, Atom 1.0 feed parsing
- ✅ HTML sanitization (CVE-2009-2937 prevention)
- ✅ Character encoding handling (UTF-8 normalization)
- ✅ URL resolution (relative to absolute)
- ✅ Date normalization (multiple formats)
- ✅ ID generation (when missing from feeds)
- ✅ Content vs summary extraction
- ✅ Atom Torture Test suite (20 test cases, XHTML/MathML/SVG)
- ✅ Comprehensive test suite (48 total test cases)

#### Repository Package (81.8% coverage) ✅
- ✅ SQLite database with WAL mode
- ✅ Foreign key constraints with CASCADE DELETE
- ✅ Feed CRUD operations
- ✅ Entry upsert with conflict resolution
- ✅ HTTP cache tracking (ETag/Last-Modified)
- ✅ Time-based queries (recent entries)
- ✅ Entry pruning by age
- ✅ Proper NULL handling
- ✅ Comprehensive test suite (12 test cases)

### Phase 2: User Interface ✅

#### Generator Package (86.0% coverage) ✅
- ✅ HTML template renderer
- ✅ Responsive design (mobile-friendly)
- ✅ Classic Planet Planet sidebar with feed list
- ✅ Feed health status display (last updated, error counts)
- ✅ Date grouping functionality ("Today", "Yesterday", etc.)
- ✅ Relative time display ("2 hours ago")
- ✅ Content Security Policy headers
- ✅ Custom template support
- ✅ Safe HTML rendering (template.HTML)
- ✅ Default template embedded in binary
- ✅ Comprehensive test suite (11 test cases)

#### Config Package (94.7% coverage) ✅
- ✅ INI format parser
- ✅ Simple feeds.txt parser (one URL per line)
- ✅ Comment support (# and ;)
- ✅ Quoted value handling
- ✅ Type validation (int, bool, string)
- ✅ Sensible defaults
- ✅ Per-feed configuration overrides (future-compatible)
- ✅ Comprehensive test suite (24 test cases)

#### CLI Application (cmd/rp) ✅
- ✅ `rp init [-f FILE]` - Initialize new planet (with optional feeds file)
- ✅ `rp add-feed <url>` - Add feed to database
- ✅ `rp add-all -f FILE` - Add multiple feeds from file
- ✅ `rp remove-feed <url>` - Remove feed
- ✅ `rp list-feeds` - Display all feeds with status
- ✅ `rp status` - Show planet status summary
- ✅ `rp update` - Fetch feeds and generate HTML
- ✅ `rp fetch` - Fetch feeds only
- ✅ `rp generate` - Generate HTML only
- ✅ `rp prune --days N` - Remove old entries
- ✅ `rp version` - Show version info
- ✅ Command-line flags (--config, --verbose, --quiet)
- ✅ User-friendly output with ✓ status indicators

### Phase 3: Integration & Polish ✅

#### Documentation ✅
- ✅ README.md with installation and usage
- ✅ WORKFLOWS.md - Comprehensive workflow guide
- ✅ CONTRIBUTING.md - Contributor guidelines
- ✅ CHANGELOG.md - Version history
- ✅ CLAUDE.md for development guidance
- ✅ specs/rogue-planet-spec.md (comprehensive specification)
- ✅ specs/testing-plan.md (testing strategy)
- ✅ specs/research/ATOM_TORTURE_TEST_RESEARCH.md (Distler test analysis)
- ✅ examples/ directory with sample configs
- ✅ .gitignore
- ⏸️ QUICKSTART.md - Deferred to v1.0 (referenced but not yet created)
- ⏸️ THEMES.md - Deferred to v1.0 (referenced but not yet created)

#### Testing ✅
- ✅ 100+ test cases across 5 packages
- ✅ All tests passing
- ✅ 88.7% library average coverage (79.9% overall including cmd/rp)
- ✅ Security tests (XSS, SSRF, SQL injection)
- ✅ End-to-end workflow tested
- ✅ Real-world feed parsing (Daring Fireball Atom, Asymco RSS)
- ✅ Live network tests with build tags
- ✅ Integration tests with saved feed snapshots
- ✅ Smart content fallback tested
- ✅ Race detector: no issues found

#### Example & Verification ✅
- ✅ Tested full workflow with real feed
- ✅ Generated HTML verified (4.2KB output)
- ✅ Binary built successfully (v0.1.0)
- ✅ All commands functional

### Phase 4: OPML Support ✅ (v0.2.0)

#### OPML Package (91.8% coverage) ✅
- ✅ OPML 1.0 and 2.0 parsing and generation
- ✅ Support for xmlUrl/url and text/title attributes
- ✅ RFC 822 date handling
- ✅ Nested outline support (categories/folders)
- ✅ Comprehensive test suite (6 integration tests)

#### OPML Commands ✅
- ✅ `rp import-opml <file> [--dry-run]` - Import feeds from OPML
- ✅ `rp export-opml [--output FILE]` - Export feeds to OPML
- ✅ Dry-run mode for preview before import
- ✅ Duplicate feed detection during import
- ✅ Compatible with Feedly, Inoreader, NewsBlur, The Old Reader

#### Configuration Validation ✅
- ✅ `rp verify` - Validate config and environment
- ✅ Config file syntax validation
- ✅ Database accessibility check
- ✅ Output directory writability check
- ✅ Custom template verification
- ✅ Feed and entry count reporting

### Phase 5: Entry Spam Prevention ✅ (v0.3.0)

#### Entry Spam Prevention Feature ✅
- ✅ `filter_by_first_seen` config option - Filter by discovery date
- ✅ `sort_by` config option - Sort by "published" or "first_seen"
- ✅ Automatic backfill migration for existing databases
- ✅ SQL parameter validation to prevent injection
- ✅ Improved NULL handling with COALESCE
- ✅ 5 unit tests + 2 integration tests
- ✅ 6 new config tests for validation

### Phase 6: Build Automation ✅

#### Makefile ✅
- ✅ `make build` - Build for current platform
- ✅ `make test` - Run all tests
- ✅ `make coverage` - Generate HTML coverage report
- ✅ `make test-race` - Run with race detector
- ✅ `make bench` - Run benchmarks
- ✅ `make fmt` - Format code
- ✅ `make vet` - Run go vet
- ✅ `make lint` - Run linters
- ✅ `make clean` - Remove build artifacts
- ✅ `make install` - Install to GOPATH/bin
- ✅ `make deps` - Download dependencies
- ✅ `make verify` - Verify dependencies
- ✅ `make check` - All quality checks
- ✅ `make quick` - Fast development iteration
- ✅ `make help` - Show all targets

---

## Project Statistics

### Code Metrics
- **Total Lines**: 3,882 (including 1,726 lines of tests)
- **Test-to-Code Ratio**: 44%
- **Packages**: 6 (5 library + 1 CLI)
- **Test Files**: 5
- **Test Cases**: 88

### Coverage by Package
```
Package      Coverage  Status
────────────────────────────────────────────
cmd/rp       26.6%  █████                   Needs Improvement
config       93.8%  ██████████████████      Excellent
crawler      96.6%  ███████████████████     Excellent
generator    84.4%  ████████████████        Very Good
normalizer   85.7%  █████████████████       Very Good
opml         91.8%  ██████████████████      Excellent
repository   80.1%  ████████████████        Good

Overall      79.9%  ███████████████         Good
Library Avg  88.7%  █████████████████       Excellent
```

**Note**: Overall average includes cmd/rp (26.6%). Library average excludes CLI and shows core package quality. Priority is improving cmd/rp coverage above 60%.

### Security
- ✅ XSS Prevention (CVE-2009-2937) - 10 test cases
- ✅ SSRF Prevention - 6 test cases
- ✅ SQL Injection Prevention - All queries parameterized
- ✅ Content Security Policy headers
- ✅ HTML sanitization with bluemonday
- ✅ URL validation (scheme, private IPs)

### Dependencies
- `github.com/mmcdole/gofeed` - Feed parsing (RSS, Atom, JSON Feed)
- `github.com/microcosm-cc/bluemonday` - HTML sanitization
- `github.com/mattn/go-sqlite3` - SQLite driver (CGO)
- `golang.org/x/net` - HTML parsing and charset detection
- `golang.org/x/text` - Character encoding detection

---

## Release Checklist

All items completed:

- ✅ All packages implemented
- ✅ All tests passing
- ✅ Coverage >75% on library packages (cmd/rp: 26.6%, needs improvement)
- ✅ Security audited
- ✅ Documentation complete
- ✅ Examples provided
- ✅ Binary built and tested
- ✅ Makefile for automation
- ✅ Distribution packages ready
- ✅ README with installation guide
- ✅ Version set (0.3.0)

---

## Status: ✅ v0.4.0 - PRODUCTION HTTP PERFORMANCE COMPLETE

The project has evolved significantly beyond the initial release:
- ✅ All v0.4.0 core features implemented and tested
- ✅ OPML import/export support (v0.2.0)
- ✅ Entry spam prevention & stable sort (v0.3.0)
  - Stable Sort Dates (Venus #15) - `sort_by = "first_seen"` option
- ✅ Responsive design (v0.1.0)
  - Responsive Image Constraints (Venus #36) - CSS max-width: 100%
- ✅ Atom Torture Test validation (v0.4.0)
- ✅ Per-domain rate limiting implemented (v0.4.0)
- ✅ 301 permanent redirect handling (v0.4.0)
- ✅ Fine-grained HTTP timeouts (v0.4.0)
- ✅ Retry-After header support (v0.4.0)
- ✅ 88.7% library test coverage
- ✅ Security hardened with SQL injection prevention
- ✅ Build automation in place

**Current Version**: 0.4.0 (Released 2025-10-30)
**Completed Features**: Core aggregation, OPML support, Entry spam prevention, Atom torture tests, Production HTTP performance
**Next Priority**: Feed autodiscovery, intelligent scheduling (P0 features for v1.0)

---

## 🚀 LAUNCH CHECKLIST

### Pre-Launch Tasks (v0.3.0)
- [x] **Create LICENSE file** - MIT License added
- [x] **Create GitHub repository** - Repository created at https://github.com/adewale/rogue_planet
- [x] **Review all documentation** - Updated for v0.3.0, internal consistency verified
- [x] **Final test run** - `make check` passed with 100% success rate
- [ ] **Build verification** - Test binary on clean system
- [x] **Version verification** - Version is 0.3.0 in main.go, crawler.go, Makefile

### GitHub Repository Configuration
- [ ] **Update repository description** - "Modern feed aggregator inspired by Planet Venus, written in Go - v0.3.0 development release"
- [ ] **Add topics/tags** - `rss`, `atom`, `feed-aggregator`, `planet`, `go`, `golang`, `static-site-generator`
- [x] **Add development notice to README** - Added "Development Release v0.3.0" notice
- [ ] **Configure repository settings**
  - [ ] Enable Issues
  - [ ] Enable Discussions (optional)
  - [ ] Configure branch protection for `main`
  - [ ] Verify default branch is `main`

### Code Push & Organization
- [x] **Review and stage changes** - Staged TODO.md, TESTING.md, CHANGELOG.md, README.md, deleted WISHLIST.md
- [x] **Commit documentation updates** - Committed: "Update documentation for v0.3.0 release" (3837446)
- [x] **Push to GitHub** - Successfully pushed to main (45c05c1..3837446)
- [ ] **Verify README renders correctly** - Check GitHub renders installation instructions properly

### Release Creation (v0.3.0)
- [ ] **Create v0.3.0 release tag** - `git tag -a v0.3.0 -m "Release v0.3.0 - Entry Spam Prevention"`
- [ ] **Push tag to GitHub** - `git push origin v0.3.0`
- [ ] **Create GitHub Release** - Draft release on GitHub
  - [ ] Title: "Rogue Planet v0.3.0 - Entry Spam Prevention"
  - [ ] Copy release notes from CHANGELOG.md
  - [ ] **Mark as pre-release** (not production-ready yet)
  - [ ] Upload pre-built binaries (optional for 0.3.0)
  - [ ] Include development disclaimer

### Optional CI/CD Setup
- [ ] **GitHub Actions workflow** - Automated testing on push/PR
- [ ] **Automated releases** - Build binaries for multiple platforms
- [ ] **Code coverage reporting** - Integrate with codecov.io
- [ ] **Automated dependency updates** - Dependabot configuration

### Community Setup
- [ ] **Announce on social media** - Twitter, Mastodon, Reddit (r/golang, r/selfhosted)
- [ ] **Submit to directories**
  - [ ] awesome-go list
  - [ ] awesome-selfhosted list
  - [ ] Go package discovery (pkg.go.dev auto-indexes)
- [ ] **Write blog post** - Technical writeup about the project
- [ ] **Create demo site** - Host example planet somewhere public

---

## Roadmap

### v1.0.0 - Production Ready (P0 Features)

**Status**: 1 of 3 original P0 features delivered early in v0.4.0
**Bonus**: 2 additional v1.x P1 features already implemented

**Completed Early (from v1.0.0 P0)**:
- ✅ 301 Permanent Redirect Handling - Delivered in v0.4.0 (originally planned for v1.0.0)

**Completed (from v1.x P1 wishlist)**:
- ✅ Responsive Image Constraints (Venus #36) - Implemented in v0.1.0
- ✅ Stable Sort Dates (Venus #15) - Implemented in v0.3.0 as `sort_by = "first_seen"`

**Remaining for v1.0.0**:

#### Feed Autodiscovery
**Problem**: Users give website URLs (https://blog.example.com/) instead of feed URLs (https://blog.example.com/feed.xml). Browser support for RSS discovery removed in Firefox/Chrome.

**Solution**: Parse HTML `<link rel="alternate">` tags to find RSS/Atom feeds
- Support RSS, Atom, and JSON Feed autodiscovery
- Handle multiple feeds per site (let user choose)
- Add `rp discover <url>` command

**Effort**: 2 days
**Priority**: P0 - Critical UX improvement

---

#### Intelligent Feed Scheduling
**Problem**: All feeds fetched at same interval regardless of update frequency. Wastes resources on slow-updating feeds, misses updates on fast-updating feeds.

**Solution**: Adaptive polling based on feed characteristics
- Adaptive polling based on historical update frequency
- Respect `Cache-Control: max-age` headers from feeds
- Exponential backoff for failing feeds (1h → 2h → 4h → 8h → 24h)
- Add jitter to prevent thundering herd (don't fetch all feeds at :00)
- Store `last_updated`, `update_frequency`, `next_fetch` in database

**Effort**: 1 week
**Priority**: P0 - Scalability and efficiency

---

### v1.x - User Experience (P1 Features)

#### HTML Escaping in Titles (Venus #24)
**Problem**: When feed titles contain HTML tags (like `<dialog>`, `<foo>`), they're stripped instead of displayed as literal text.

**Current Behavior**: HTML sanitization strips tags from all fields including titles.

**Solution**:
- Titles, author names, and feed names should have HTML entities escaped (show `&lt;dialog&gt;` as `<dialog>`)
- Content and summary fields continue using full HTML sanitization (allow safe subset, block dangerous tags)

**Implementation**: Use different sanitization policies for metadata (escape all HTML) vs content (allow safe HTML subset)

**Effort**: Low
**Priority**: P1 - Data accuracy issue

**Source**: https://github.com/rubys/venus/issues/24

---

### v1.x - Feed Management (P1 Features)

#### Auto-Reactivate Inactive Feeds (Venus #34)
**Problem**: When a feed goes inactive (author stops posting) and later resumes, it stays marked as inactive in database. Requires manual intervention to reactivate.

**Current Behavior**: Feeds marked inactive are skipped during fetches. Must be manually reactivated via database update.

**Solution**: When an inactive feed has new entries (detected by newer published dates than last fetch), automatically mark it active again
```
1. Fetch inactive feeds periodically (e.g., daily instead of hourly)
2. If new entries found: set active=1, fetch_error_count=0
3. Log reactivation for visibility
```

**Rationale**: Blogs often go dormant and resume. Automatic reactivation improves user experience.

**Effort**: Medium
**Priority**: P1

**Source**: https://github.com/rubys/venus/issues/34

---

#### Character Encoding Detection
**Problem**: Feeds often declare wrong encoding or omit encoding declaration entirely. Results in garbled text.

**Current Behavior**: Relies on feed-declared encoding.

**Solution**: Use charset detection library with fallback chain
- Try declared encoding first
- Use charset detection (chardet, encoding/japanese, etc.)
- Fallback to UTF-8
- Log encoding mismatches

**Effort**: Medium
**Priority**: P1 - Data quality issue

---

#### HTTP Authentication Support
**Problem**: Some feeds require Basic or Digest authentication. Cannot be fetched without credentials.

**Current Behavior**: No authentication support.

**Solution**: Add per-feed auth configuration
```ini
[https://private.example.com/feed.xml]
auth_type = basic
auth_user = username
auth_pass = password
```

Store credentials securely, pass in HTTP requests.

**Effort**: Medium
**Priority**: P1 - Access to private feeds

---

### v2.0 - Advanced Features (P2)

#### Allow data-* Attributes (Venus #19)
**Problem**: HTML5 `data-*` attributes are harmless metadata but get stripped by sanitization, breaking third-party scripts and widgets that users add to templates.

**Current Behavior**: bluemonday strips data-* attributes from all elements.

**Solution**: Allow data-* attributes on safe elements (div, span, article, section, etc.)
```go
policy.AllowDataAttributes()
```

**Rationale**:
- Enables template customization with JavaScript libraries
- data-* attributes are client-side only, no XSS risk
- Standard HTML5 practice for storing custom data

**Effort**: Trivial
**Priority**: P2 - Template customization

**Source**: https://github.com/rubys/venus/pull/19

---

#### Future Dates Configuration
**Problem**: Some feeds contain entries with future dates (scheduled posts, timezone errors). Need policy for handling these.

**Current Behavior**: Accepted as-is.

**Solution**: Configuration option to ignore, accept, or clamp future-dated entries
```ini
[planet]
future_dates = accept | ignore_entry | clamp_to_now
```

- `accept`: Keep as-is (default)
- `ignore_entry`: Skip entries with pub date > now
- `clamp_to_now`: Set pub date to fetch time if in future

**Rationale**: Already in spec (line 244), needs implementation. Prevents timeline pollution from incorrectly dated entries.

**Effort**: Low
**Priority**: P2

---

#### Enhanced Media Tag Support (Venus #18)
**Problem**: HTML5 video/audio tags need attributes like `preload`, `poster`, `controls` to work well, but sanitization strips them.

**Current Behavior**: Only basic video/audio tags allowed, missing important attributes.

**Solution**: Configure bluemonday to allow safe media attributes:
- `<video>`: preload, poster, controls, width, height
- `<audio>`: preload, controls
- `<source>`: src, type

**Rationale**: Better embedding of multimedia content from feeds.

**Effort**: Low
**Priority**: P2

**Source**: https://github.com/rubys/venus/pull/18

---

#### Embedded Content Handling
**Problem**: YouTube embeds, tweets, audio players don't display properly or get stripped entirely.

**Current Behavior**: Most iframes blocked for security.

**Solution**: Whitelist trusted iframe domains
```ini
[planet]
trusted_iframe_domains = youtube.com, youtube-nocookie.com, vimeo.com, codepen.io
```

- Only allow iframes with https URLs from approved domains
- Set strict sandbox attribute: `sandbox="allow-scripts allow-same-origin"`

**Security Note**: Requires careful implementation to avoid SSRF and XSS risks.

**Effort**: Medium
**Priority**: P2 - Common use case

---

---

## Completed v1.x/v2.x Features (Delivered Early)

The following features from the v1.x and v2.x wishlist were implemented before v1.0.0:

### ✅ Responsive Image Constraints (Venus #36) - v0.1.0
**Feature**: Images in feed content automatically constrained to page width
**Implementation**: CSS in default template
```css
.entry-content img {
    max-width: 100%;
    height: auto;
}
```
**Location**: pkg/generator/generator.go:510-512
**Benefit**: Prevents layout overflow on mobile devices
**Priority**: Was v1.x P1, delivered in initial release

---

### ✅ Stable Sort Dates (Venus #15) - v0.3.0
**Feature**: Entries can be sorted by discovery time instead of published date
**Implementation**: `sort_by = "first_seen"` config option
**Benefit**: Entries don't "jump" in timeline when authors make corrections
**Location**: pkg/config/config.go, pkg/repository/repository.go
**Priority**: Was v1.x P1, delivered in v0.3.0

---

### v2.x - Nice to Have (P3)

#### MathJax Support (Venus #33)
**Problem**: Mathematical equations in feeds (especially from Blogger, academic blogs) don't render properly. MathML and TeX markup display as raw code.

**Current Behavior**: Math markup passes through as text.

**Solution**:
1. Document how to add MathJax to custom templates
2. Ensure sanitization allows MathML tags or TeX delimiters
3. Optional: Add `enable_mathjax` config option to include in default template

**Rationale**: Important for academic/technical planet sites aggregating math/science blogs.

**Effort**: Low
**Priority**: P3 - Niche use case

**Source**: https://github.com/rubys/venus/issues/33

---

#### Atom Torture Test Research and Validation ✅
**Status**: ✅ **COMPLETED in v0.4.0**

**Implementation**:
- ✅ Comprehensive research document created: `specs/research/ATOM_TORTURE_TEST_RESEARCH.md`
- ✅ Accessed and analyzed original Jacques Distler blog post (April 18, 2006)
- ✅ Added 2 missing critical test cases:
  1. XHTML case-sensitivity test (The Distler Test: "a **b**c **D**e f")
  2. xml:base relative URL resolution test
- ✅ Updated all test fixture headers with original source references
- ✅ Updated test suite documentation with background and policy decisions
- ✅ All 20 torture tests passing

**Test Fixtures**:
- testdata/atom-torture-xhtml.xml (7 test entries - added 2)
- testdata/atom-torture-mathml.xml (6 test entries)
- testdata/atom-torture-svg.xml (7 test entries)
- pkg/normalizer/normalizer_torture_test.go (20 test functions)

**Key Findings**:
- ✅ gofeed correctly uses XML parsing (case-sensitive)
- ✅ xml:base support working correctly
- ✅ MathML/SVG stripped for security (documented policy, correct decision)
- ✅ Comprehensive XSS prevention validated

**Reference**: https://golem.ph.utexas.edu/~distler/blog/archives/000793.html

**Completed**: 2025-10-30

---

#### GitHub Gist Embeds (Venus #28)
**Problem**: GitHub Gist embeds (iframe-based) don't display because iframes are blocked by sanitization.

**Current Behavior**: All iframes stripped for security.

**Solution**: Covered under "Embedded Content Handling" above - allow iframes from trusted domains including gist.github.com.

**Rationale**: Common use case for developer planet sites.

**Effort**: Medium (part of embedded content feature)
**Priority**: P3 - Niche

**Source**: https://github.com/rubys/venus/issues/28

---

### Future (Unscheduled)

---

## Not Implementing

These features are explicitly out of scope:

### ❌ Web-Based Admin Interface
**Reason**: Against static-output philosophy. Adds attack surface, complexity, maintenance burden. CLI is sufficient for planet operators.

### ❌ PostgreSQL/MySQL Support
**Reason**: SQLite is perfect for this use case. Other databases add deployment complexity with no benefit for feed aggregation workload.

### ❌ Multiple Template Engines
**Reason**: Go templates are sufficient. Adding Jinja2 would require embedding Python or complex reimplementation. Not worth complexity.

### ❌ Feed Content Modification
**Reason**: Aggregator should aggregate, not modify. Content transformation belongs elsewhere (user scripts, external tools).

---

## Already Implemented ✅

These Planet Venus issues are already addressed in Rogue Planet:

### ✅ User-Agent Headers (Venus #29)
- **Status**: Implemented in v0.1.0
- **Implementation**: Rogue Planet sends proper User-Agent header on all requests
- **Configuration**: `user_agent` in config.ini
- **Default**: `RoguePlanet/0.3 (+https://github.com/adewale/rogue_planet)`

### ✅ XSS Prevention (CVE-2009-2937)
- **Status**: Core security feature (spec lines 548-762)
- **Implementation**: Full HTML sanitization using bluemonday
- Strips script tags, event handlers, dangerous URIs
- Uses html/template for auto-escaping
- Content Security Policy headers in generated HTML

### ✅ Video Autoplay Filtering (Venus #1)
- **Status**: Handled by bluemonday sanitization
- **Implementation**: Dangerous attributes like autoplay stripped from all tags
- **Source**: https://github.com/rubys/venus/issues/1

### ✅ HTTP Conditional Requests
- **Status**: Core feature (spec lines 64-156)
- **Implementation**: Proper ETag/Last-Modified support
- Reduces bandwidth and server load
- Well-behaved feed fetching

### ✅ OPML Import/Export
- **Status**: v0.2.0 feature
- **Implementation**: Full OPML 1.0/2.0 support
- **Commands**: `rp import-opml`, `rp export-opml`
- Compatible with Feedly, Inoreader, NewsBlur, The Old Reader

### ✅ Entry Spam Prevention
- **Status**: v0.3.0 feature
- **Implementation**: `filter_by_first_seen` and `sort_by` config options
- Prevents flooding timeline when adding new feeds
- See [ENTRY_SPAM.md](ENTRY_SPAM.md) for complete details

---

## Contributing

For future contributors:

1. Read `CLAUDE.md` for development guidance
2. Follow `specs/testing-plan.md` for testing
3. Maintain coverage >75%
4. Run `make check` before committing
5. All security features are mandatory

Before implementing wishlist items:
1. Check if already implemented in current version
2. Discuss approach in GitHub issues
3. Ensure test coverage >75%
4. Update documentation
5. Follow patterns in CLAUDE.md

## License

MIT License - See [LICENSE](../LICENSE) file for details.

---

---

## v0.4.0 - COMPLETED ✅ (2025-10-30)

### Release Summary

Version 0.4.0 delivers production-ready HTTP performance features:

**Completed Features**:
- ✅ Per-domain rate limiting using token bucket algorithm
  - Default: 60 requests/minute per domain with burst of 10
  - Configurable via `requests_per_minute` and `rate_limit_burst`
  - Thread-safe concurrent access with RWMutex
- ✅ Fine-grained HTTP timeouts
  - `http_timeout_seconds`, `dial_timeout_seconds`, `tls_handshake_timeout_seconds`, `response_header_timeout_seconds`
  - All configurable with sensible defaults
- ✅ 301 permanent redirect handling
  - Automatically updates feed URLs in database on 301 responses
  - Logs redirect for transparency
- ✅ Retry-After header support
  - Respects server-specified retry delays (RFC 7231)
  - Honors HTTP 429 rate limit responses
- ✅ Atom Torture Test research and implementation
  - Comprehensive research document
  - 2 new test cases added (XHTML case-sensitivity, xml:base)
  - 20 torture tests passing

**Documentation Updates**:
- ✅ Updated CHANGELOG.md with v0.4.0 release notes
- ✅ Updated README.md with HTTP performance features
- ✅ Updated CLAUDE.md with rate limiter implementation notes
- ✅ Updated examples/config.ini with all new HTTP options
- ✅ Updated GO_AUDITING_HEURISTICS.md with lessons learned
- ✅ Updated all outdated spec documents

**Testing**:
- ✅ 11 new unit tests for rate limiter (concurrency, context cancellation, stats)
- ✅ All 8 packages passing tests
- ✅ Maintained >85% library test coverage

**Dependencies**:
- ✅ Added `golang.org/x/time v0.14.0` for rate limiting

---

*Released: 2025-10-30*
*Status: v0.4.0 complete - Production HTTP performance implemented*
*Next: v1.0.0 planning - Feed autodiscovery and intelligent scheduling*
