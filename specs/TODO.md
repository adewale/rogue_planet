# Rogue Planet - TODO

## ✅ v0.1.0 FEATURES COMPLETE

All planned features for the initial release have been implemented, tested, and documented.

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
- ✅ Comprehensive test suite (28 test cases)

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

#### Generator Package (85.2% coverage) ✅
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
- ✅ QUICKSTART.md - 5-minute setup guide
- ✅ WORKFLOWS.md - Comprehensive workflow guide
- ✅ CONTRIBUTING.md - Contributor guidelines
- ✅ CHANGELOG.md - Version history
- ✅ CLAUDE.md for development guidance
- ✅ specs/rogue-planet-spec.md (comprehensive specification)
- ✅ specs/testing-plan.md (testing strategy)
- ✅ examples/ directory with sample configs
- ✅ .gitignore

#### Testing ✅
- ✅ 100+ test cases across 5 packages
- ✅ All tests passing
- ✅ 80%+ average test coverage
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
config       94.7%  ██████████████████  Excellent
crawler      96.6%  ███████████████████ Excellent
generator    86.0%  █████████████████   Excellent
normalizer   79.8%  ███████████████     Good
opml         91.8%  ██████████████████  Excellent
repository   81.8%  ████████████████    Very Good
Average      88.4%  █████████████████   Excellent
```

### Security
- ✅ XSS Prevention (CVE-2009-2937) - 10 test cases
- ✅ SSRF Prevention - 6 test cases
- ✅ SQL Injection Prevention - All queries parameterized
- ✅ Content Security Policy headers
- ✅ HTML sanitization with bluemonday
- ✅ URL validation (scheme, private IPs)

### Dependencies
- `github.com/mmcdole/gofeed` - Feed parsing
- `github.com/microcosm-cc/bluemonday` - HTML sanitization
- `github.com/mattn/go-sqlite3` - SQLite driver
- `golang.org/x/time/rate` - Rate limiting

---

## Future Enhancements (Optional)

These are **not required** but documented for future consideration:

### Web UI
- [ ] Configuration interface
- [ ] Feed management dashboard
- [ ] Real-time update status

### Additional Formats
- [ ] Multiple output formats (RSS/Atom feeds)
- [ ] JSON Feed output
- [ ] Archive pages (by month/year)

### Advanced Features
- [ ] WebSub/PubSubHubbub support
- [ ] OPML import/export
- [ ] Feed discovery from website URLs
- [ ] Tag/category support
- [ ] Full-text search
- [ ] Plugin system for custom filters

### DevOps
- [ ] Docker image (Dockerfile exists as placeholder)
- [ ] Docker Compose setup
- [ ] Kubernetes manifests
- [ ] GitHub Actions CI/CD workflow
- [ ] Automated releases

### Performance
- [ ] Benchmarking suite
- [ ] Performance optimization
- [ ] Caching strategies
- [ ] CDN integration guide

---

## Release Checklist

All items completed:

- ✅ All packages implemented
- ✅ All tests passing
- ✅ Coverage >75% on all packages
- ✅ Security audited
- ✅ Documentation complete
- ✅ Examples provided
- ✅ Binary built and tested
- ✅ Makefile for automation
- ✅ Distribution packages ready
- ✅ README with installation guide
- ✅ Version set (0.3.0)

---

## Status: 🚀 v0.3.0 - ENTRY SPAM PREVENTION COMPLETE

The project has evolved significantly beyond the initial release:
- ✅ All core features implemented and tested
- ✅ OPML import/export support (v0.2.0)
- ✅ Entry spam prevention feature (v0.3.0)
- ✅ Documentation complete and up-to-date
- ✅ 88.4% average test coverage (up from 80.6%)
- ✅ Security hardened with SQL injection prevention
- ✅ Build automation in place

**Current Version**: 0.3.0 (Development Release)
**Completed Features**: Core aggregation, OPML support, Entry spam prevention
**Next Priority**: Feed autodiscovery, 301 redirect handling, intelligent scheduling (P0 features for v1.0)

---

## 🚀 LAUNCH CHECKLIST

### Pre-Launch Tasks (v0.1.0)
- [ ] **Create LICENSE file** - Add open source license (MIT, Apache 2.0, or GPL)
- [ ] **Review all documentation** - Ensure all docs are up-to-date and accurate
- [ ] **Final test run** - Execute `make check` to verify all tests pass
- [ ] **Build verification** - Test binary on clean system
- [ ] **Version verification** - Confirm version is 0.3.0 in code (main.go, pkg/crawler/crawler.go, CHANGELOG)

### GitHub Repository Setup
- [ ] **Create GitHub repository** - Initialize on GitHub
- [ ] **Add repository description** - "Modern feed aggregator inspired by Planet Venus, written in Go - v0.3.0 development release"
- [ ] **Add topics/tags** - `rss`, `atom`, `feed-aggregator`, `planet`, `go`, `golang`, `static-site-generator`
- [ ] **Add development notice** - Clearly mark as 0.3.x development release in README
- [ ] **Configure repository settings**
  - [ ] Enable Issues
  - [ ] Enable Discussions (optional)
  - [ ] Configure branch protection for `main`
  - [ ] Set default branch to `main`

### Initial Git Push
- [ ] **Stage all files** - `git add` all project files (see GITHUB_PUSH_LIST.md)
- [ ] **Verify staging** - Review `git status` to ensure correct files staged
- [ ] **Create initial commit** - Commit with message: "Initial development release: Rogue Planet v0.3.0"
- [ ] **Add remote** - `git remote add origin <repo-url>`
- [ ] **Push to GitHub** - `git push -u origin main`

### Release Creation (v0.3.0)
- [ ] **Create v0.3.0 release tag** - `git tag -a v0.3.0 -m "Release v0.3.0 - Entry Spam Prevention"`
- [ ] **Push tag to GitHub** - `git push origin v0.3.0`
- [ ] **Create GitHub Release** - Draft release on GitHub
  - [ ] Title: "Rogue Planet v0.3.0 - Entry Spam Prevention"
  - [ ] Copy release notes from CHANGELOG.md
  - [ ] **Mark as pre-release** (not production-ready yet)
  - [ ] Upload pre-built binaries (optional for 0.3.0)
  - [ ] Include development disclaimer

### Post-Launch Documentation
- [x] **Add GitHub URL to code** - Update User-Agent in crawler.go with repo URL (DONE: already has github.com/adewale/rogue_planet)
- [ ] **Update README badges** (optional)
  - [ ] Build status badge
  - [ ] Coverage badge (codecov.io or coveralls.io)
  - [ ] Go Report Card badge
  - [ ] License badge
- [ ] **Create SECURITY.md** - Document security policy and vulnerability reporting
- [ ] **Create issue templates** - Bug report and feature request templates
- [ ] **Create pull request template**

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

### Monitoring & Maintenance
- [ ] **Set up GitHub notifications** - Monitor issues and PRs
- [ ] **Plan maintenance schedule** - Dependency updates, security patches
- [ ] **Create project roadmap** - Document future enhancement plans
- [ ] **Set up analytics** (optional) - Track adoption and usage

---

## Contributing

For future contributors:

1. Read `CLAUDE.md` for development guidance
2. Follow `specs/testing-plan.md` for testing
3. Maintain coverage >75%
4. Run `make check` before committing
5. All security features are mandatory

## License

MIT License - See [LICENSE](LICENSE) file for details.

---

*Last Updated: 2025-10-16*
*Status: v0.3.0 - Entry spam prevention complete, P0 features planned for v1.0*

---

## Roadmap to v1.0.0

### Critical Features (P0) - Next Priority

- [ ] **Feed Autodiscovery** (2 days)
  - Parse HTML `<link rel="alternate">` tags to find RSS/Atom feeds
  - Support RSS, Atom, and JSON Feed autodiscovery
  handle it

- [ ] **301 Redirect Handling** (1 day)
  - Detect permanent redirects (301 Moved Permanently)
  - Update stored feed URL in database automatically
  - Prevents wasting bandwidth on redirects forever
  - Critical for long-term reliability as feeds migrate HTTP→HTTPS

- [ ] **Intelligent Feed Scheduling** (1 week)
  - Adaptive polling based on historical update frequency
  - Respect Cache-Control: max-age headers from feeds
  - Exponential backoff for failing feeds (1h → 2h → 4h → 8h → 24h)
  - Add jitter to prevent thundering herd (don't fetch all feeds at :00)
  
### Important Features (P1) - Follow-up Priority
- [ ] Character encoding detection and fallback
- [ ] HTTP authentication support (Basic/Digest Auth)
- [ ] Embedded content handling (YouTube, tweets, audio players)

### Planned Enhancements
- [ ] Additional real-world testing
- [ ] Additional feed format edge cases
- [ ] Additional themes (dark mode, minimalist)
