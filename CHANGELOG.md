# Changelog

All notable changes to Rogue Planet will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned for 1.0.0
- Public GitHub release
- Full production deployment documentation
- Binary distribution packages
- Community contribution guidelines finalized

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
- Responsive mobile-friendly design
- Custom theme support

---

## Version History Summary

| Version | Date | Description |
|---------|------|-------------|
| 0.1.0 | 2025-10-10 | Initial development release - all core features complete |
| 0.2.0 | 2025-10-13 | OPML support, verify command, improved test coverage |
| 1.0.0 | TBD | Planned public release |

## Links

- [Repository](https://github.com/adewale/rogue_planet)
- [Documentation](README.md)
- [Quick Start](QUICKSTART.md)
- [Workflows](WORKFLOWS.md)
- [Contributing](CONTRIBUTING.md)
- [License](LICENSE)
