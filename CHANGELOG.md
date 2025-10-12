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
| 1.0.0 | TBD | Planned public release |

## Links

- [Repository](https://github.com/adewale/rogue_planet)
- [Documentation](README.md)
- [Quick Start](QUICKSTART.md)
- [Workflows](WORKFLOWS.md)
- [Contributing](CONTRIBUTING.md)
- [License](LICENSE)
