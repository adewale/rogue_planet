# Test Coverage Checklist

Quick reference for tracking test implementation progress.

## Package Coverage Status

### pkg/crawler (Target: >85%)

#### Critical Security Tests (P0)
- [ ] SSRF Prevention - All blocked hosts/IPs (17 test cases)
- [ ] HTTP Conditional Requests - ETag/Last-Modified (12 test cases)

#### Core Functionality (P1)
- [ ] Response handling (200, 304, 4xx, 5xx) (6 test cases)
- [ ] Gzip decompression (4 test cases)
- [ ] Size limits (4 test cases)
- [ ] Timeout handling (4 test cases)
- [ ] Redirects (6 test cases)
- [ ] Retry logic with backoff (9 test cases)

**Current Coverage**: __% | **Branch Coverage**: __%

---

### pkg/normalizer (Target: >80%)

#### Critical Security Tests (P0)
- [ ] XSS Prevention - Script removal (30+ test cases)
- [ ] URL scheme validation (7 test cases)

#### Feed Parsing (P1)
- [ ] Multiple feed formats (RSS 1.0/2.0, Atom, JSON) (4 test cases)
- [ ] ID generation fallbacks (6 test cases)
- [ ] Date extraction fallbacks (6 test cases)
- [ ] Author extraction fallbacks (4 test cases)
- [ ] Relative URL resolution (8 test cases)

#### HTML Sanitization Details (P0)
- [ ] Remove: `<script>`, `<iframe>`, `<object>`, `<embed>`, `<base>`
- [ ] Remove: javascript: and data: URIs
- [ ] Remove: All event handlers (onclick, onerror, etc.)
- [ ] Allow: Safe tags (p, a, img, strong, etc.)
- [ ] Allow: http/https URLs only
- [ ] Handle: Malformed/unclosed tags

**Current Coverage**: __% | **Branch Coverage**: __%

---

### pkg/repository (Target: >85%)

#### Critical Data Integrity (P0)
- [ ] UPSERT logic - Insert vs Update branches (7 test cases)
- [ ] Foreign key CASCADE deletes (4 test cases)
- [ ] NULL value handling (10+ test cases)

#### Database Operations (P1)
- [ ] Feed CRUD operations (15 test cases)
- [ ] Entry operations (10 test cases)
- [ ] GetRecentEntries with smart fallback (8 test cases)
- [ ] Cache header updates (8 test cases)
- [ ] Error tracking (5 test cases)
- [ ] Query filtering (active/inactive) (4 test cases)

**Current Coverage**: __% | **Branch Coverage**: __%

---

### pkg/generator (Target: >80%)

#### Template Rendering (P1)
- [ ] Basic template execution (6 test cases)
- [ ] Data binding (8 test cases)
- [ ] Date grouping logic (7 test cases)
- [ ] Feed sidebar rendering (8 test cases)
- [ ] Relative time formatting (13 test cases)
- [ ] File output operations (6 test cases)
- [ ] Static asset copying (6 test cases)

#### Security (P0)
- [ ] XSS prevention in templates (4 test cases)
- [ ] template.HTML usage verification

**Current Coverage**: __% | **Branch Coverage**: __%

---

### pkg/config (Target: >80%)

#### Config Parsing (P1)
- [ ] INI file loading (6 test cases)
- [ ] Default value application (7 test cases)
- [ ] Type parsing (int, bool) (6 test cases)
- [ ] Feed-specific overrides (5 test cases)
- [ ] Validation (5 test cases)

**Current Coverage**: __% | **Branch Coverage**: __%

---

### cmd/rp (Target: >75%)

#### CLI Commands (P1)
- [ ] cmdInit - with/without feeds file (6 test cases)
- [ ] cmdAddFeed - valid/invalid/duplicate (6 test cases)
- [ ] cmdAddAll - batch import (6 test cases)
- [ ] cmdRemoveFeed - exists/cascade (4 test cases)
- [ ] cmdListFeeds - empty/populated (5 test cases)
- [ ] cmdStatus - stats display (4 test cases)
- [ ] cmdUpdate - full workflow (7 test cases)
- [ ] cmdFetch - fetch only (6 test cases)
- [ ] cmdGenerate - generate only (6 test cases)
- [ ] cmdPrune - with/without dry-run (5 test cases)

**Current Coverage**: __% | **Branch Coverage**: __%

---

## Integration Tests Status

### Full Pipeline (P0)
- [ ] Complete end-to-end: init → add → fetch → generate
- [ ] Update cycle with 304 responses
- [ ] Feed management workflow
- [ ] Error recovery workflow
- [ ] Stale feed fallback scenario

### Real-World Feeds (P1)
- [ ] Daring Fireball (Atom) parsing
- [ ] Asymco (RSS) parsing
- [ ] Malformed feed recovery
- [ ] Mixed feed formats (Atom + RSS + JSON)

### Security End-to-End (P0)
- [ ] XSS prevention full pipeline
- [ ] SSRF prevention full pipeline
- [ ] SQL injection prevention verification
- [ ] Path traversal prevention

### Performance (P2)
- [ ] Concurrent fetch (50 feeds)
- [ ] Large database (10,000 entries)
- [ ] Memory usage profiling

---

## Edge Cases Status

### Network Errors (P1)
- [ ] Timeout handling
- [ ] DNS failures
- [ ] TLS errors
- [ ] HTTP error codes
- [ ] Redirect chains

### Feed Content (P1)
- [ ] Empty feeds
- [ ] Encoding issues (UTF-8, Windows-1252)
- [ ] Date edge cases (missing, invalid, future)
- [ ] Huge content (>10MB)
- [ ] Special characters (emoji, RTL, symbols)

### Database (P1)
- [ ] Concurrent access (WAL mode)
- [ ] Disk full scenario
- [ ] Very long strings

### File System (P2)
- [ ] Permission errors
- [ ] Path issues (spaces, Unicode)
- [ ] Symlinks

---

## Test Execution Commands

```bash
# Quick development cycle
make quick              # fmt + test + build

# Full quality check
make check              # fmt + vet + test + race

# Coverage report
make coverage           # HTML report in coverage/

# Specific package
go test ./pkg/crawler -v

# With race detector
go test -race ./...

# Integration tests
make test-integration

# Network tests (requires internet)
go test -tags=network ./pkg/crawler -v

# Benchmarks
make bench
```

---

## Coverage Targets Summary

| Package           | Line Coverage | Branch Coverage | Status |
|-------------------|---------------|-----------------|--------|
| pkg/crawler       | >85%         | >90% (security) | [ ]    |
| pkg/normalizer    | >80%         | >90% (security) | [ ]    |
| pkg/repository    | >85%         | >95% (UPSERT)   | [ ]    |
| pkg/generator     | >80%         | >85%            | [ ]    |
| pkg/config        | >80%         | >85%            | [ ]    |
| cmd/rp            | >75%         | >80%            | [ ]    |
| **Overall**       | **>75%**     | **>85%**        | [ ]    |

---

## Priority Legend

- **P0 (Critical)**: Must pass before release - security and data integrity
- **P1 (High)**: Core functionality - should pass before release
- **P2 (Medium)**: Edge cases and performance - nice to have
- **P3 (Low)**: Benchmarks and stress tests - future work

---

## Next Actions

1. **Start with P0 tests**: Security and data integrity first
2. **Run coverage report**: `make coverage` to see current state
3. **Fill gaps systematically**: Work through each package
4. **Add integration tests**: Full workflows and real-world scenarios
5. **Set up CI/CD**: Automate test execution on PRs
6. **Monitor and improve**: Regular coverage reviews

---

## Notes

- Use `t.TempDir()` for automatic cleanup
- Use `httptest.NewServer()` for mock HTTP servers
- Use table-driven tests for similar scenarios
- Keep unit tests fast (<100ms each)
- Use `-tags=network` for tests requiring internet
- Run race detector regularly: `go test -race ./...`

---

## Test File Organization

```
pkg/crawler/
  ✓ crawler.go
  ✓ crawler_test.go              # Unit tests
  ✓ crawler_live_test.go         # Network tests
  ✓ crawler_user_agent_test.go   # Feature-specific

pkg/normalizer/
  ✓ normalizer.go
  ✓ normalizer_test.go           # Unit tests
  ✓ normalizer_realworld_test.go # Real feed tests

pkg/repository/
  ✓ repository.go
  ✓ repository_test.go           # Database tests

pkg/generator/
  ✓ generator.go
  ✓ generator_test.go
  ✓ generator_integration_test.go

cmd/rp/
  ✓ main.go
  ✓ commands.go
  ✓ commands_test.go             # CLI tests
  ✓ integration_test.go          # Full pipeline
  ✓ realworld_integration_test.go

testdata/
  [ ] daring-fireball.xml        # Add feed snapshots
  [ ] asymco.xml
  [ ] malicious-feed.xml         # XSS test cases
  [ ] malformed-feed.xml
```

---

## Tracking Progress

Update this checklist as tests are implemented. Use `make coverage` to verify actual coverage numbers and replace placeholders (__%) with real data.

**Last Updated**: 2025-10-09
**Next Review**: [Set quarterly review date]
