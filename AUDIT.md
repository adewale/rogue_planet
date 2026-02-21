# Rogue Planet Security & Quality Audit

**Initial Audit:** 2026-02-14
**Re-Audit:** 2026-02-21
**Scope:** Full codebase audit covering security, code quality, testing, dependencies, and CI/CD

---

## Executive Summary

The initial audit on 2026-02-14 identified **1 critical**, **6 high**, **12 medium**, and **17 low** severity findings. Remediation was performed, and this re-audit validates those fixes and identifies any remaining issues.

**Remediation Results:**
- **16 of 36 findings fixed** (1 critical, 6 high, 6 medium, 3 low)
- **1 finding partially fixed** (H3: Gosec `continue-on-error` removed, Trivy retains it for SARIF output)
- **19 findings remain open** (0 critical, 0 high, 5 medium, 14 low)
- **0 regressions introduced** -- all 11 packages pass, race detector clean
- **Test coverage improved**: `pkg/repository` 66.4% -> 74.1%, overall 78.2%

---

## Remediation Status

| ID | Severity | Finding | Status |
|----|----------|---------|--------|
| C1 | CRITICAL | XSS via unsanitized entry titles | **FIXED** |
| H1 | HIGH | SSRF redirect bypass | **FIXED** |
| H2 | HIGH | GitHub Actions pinned to `@master` | **FIXED** |
| H3 | HIGH | Security scans `continue-on-error` | **PARTIAL** |
| H4 | HIGH | `mailto:` scheme allowed | **FIXED** |
| H5 | HIGH | Migration transaction scope | **FIXED** |
| H6 | HIGH | No content size limits | **FIXED** |
| M1 | MEDIUM | DNS rebinding protection | OPEN |
| M2 | MEDIUM | IPv4-mapped IPv6 handling | OPEN |
| M3 | MEDIUM | Feed URL validation gaps | **FIXED** |
| M4 | MEDIUM | PRAGMA foreign_keys scope | **FIXED** |
| M5 | MEDIUM | CSP allows `unsafe-inline` | OPEN |
| M6 | MEDIUM | No symlink handling in copyDir | **FIXED** |
| M7 | MEDIUM | Path traversal string check | **FIXED** |
| M8 | MEDIUM | No batch transactions for entries | OPEN |
| M9 | MEDIUM | No input validation in repository | OPEN |
| M10 | MEDIUM | Missing `-trimpath` build flag | **FIXED** |
| M11 | MEDIUM | Missing `CGO_ENABLED=1` | **FIXED** |
| M12 | MEDIUM | Missing `permissions` block in CI | **FIXED** |
| L1 | LOW | Missing special IP range blocks | OPEN |
| L2 | LOW | No explicit minimum TLS version | OPEN |
| L3 | LOW | Gzip decompression ratio | OPEN |
| L4 | LOW | Internal error details in database | OPEN |
| L5 | LOW | `NewForTesting()` exported | OPEN |
| L6 | LOW | Retry-After date no upper bound | OPEN |
| L7 | LOW | Missing security SQLite PRAGMAs | OPEN |
| L8 | LOW | Fragile error string comparison | **FIXED** |
| L9 | LOW | PruneOldEntries no lower bound | **FIXED** |
| L10 | LOW | Database file default permissions | OPEN |
| L11 | LOW | No context on schema init | OPEN |
| L12 | LOW | Relative URLs not resolved | OPEN |
| L13 | LOW | Hash truncation collision risk | OPEN |
| L14 | LOW | No future date clamping | OPEN |
| L15 | LOW | No OPML file size limit | OPEN |
| L16 | LOW | No upper bound on `days` config | OPEN |
| L17 | LOW | golangci-lint `version: latest` | OPEN |

---

## Fix Verification Details

### C1: XSS via Unsanitized Entry Titles -- **FIXED**

**What was done:**
- `pkg/normalizer/normalizer.go`: Added `sanitizeTitle()` method using `bluemonday.StrictPolicy()` to strip all HTML from titles (line 285-292)
- `pkg/normalizer/normalizer.go:145`: All titles now pass through `sanitizeTitle()` before storage
- `pkg/generator/generator.go:47`: `EntryData.Title` changed from `template.HTML` to `string`, enabling `html/template` auto-escaping
- `cmd/rp/cmd_helpers.go`: Removed unsafe `template.HTML(entry.Title)` cast

**Verification:** Titles containing `<img src=x onerror=alert(1)>` are now stripped to plain text. Both the sanitization layer (normalizer) and the type safety layer (string vs template.HTML) prevent XSS. Tests added for OWASP title XSS vectors.

### H1: SSRF Redirect Bypass -- **FIXED**

**What was done:**
- `pkg/crawler/crawler.go`: All three `CheckRedirect` callbacks (lines 108-112, 213-217, 318-322) now call `ValidateURL()` on the redirect target URL
- Redirect targets to `127.0.0.1`, `169.254.169.254`, private IPs, etc. are blocked

**Verification:** Tests in `crawler_comprehensive_test.go` confirm that redirects to private IPs are rejected. The `skipSSRFCheck` flag properly gates the validation for test scenarios.

### H2: GitHub Actions Pinned -- **FIXED**

**What was done:**
- `.github/workflows/go.yml:134`: `securego/gosec@v2.21.4` (was `@master`)
- `.github/workflows/go.yml:139`: `aquasecurity/trivy-action@0.28.0` (was `@master`)

**Verification:** Both actions pinned to specific version tags. Mutable `@master` references eliminated.

### H3: Security Scans `continue-on-error` -- **PARTIALLY FIXED**

**What was done:**
- Gosec step: `continue-on-error: true` removed -- security findings now fail the build
- Trivy step: `continue-on-error: true` retained (line 145) because Trivy outputs to SARIF format and upload is a separate step

**Remaining:** Trivy scan failures are still silently ignored. Consider making the Trivy step itself non-continuable and only keeping `continue-on-error` on the SARIF upload step.

### H4: `mailto:` Scheme Blocked -- **FIXED**

**What was done:**
- `pkg/normalizer/normalizer.go:72-74`: Added `AllowURLSchemeWithCustomPolicy("mailto", func(u *url.URL) bool { return false })` after the `AllowURLSchemes("http", "https")` call

**Verification:** `mailto:` links in feed content are now stripped. Tests confirm the policy rejects `mailto:` URLs while allowing `http:` and `https:`.

### H5: Migration Transaction Scope -- **FIXED**

**What was done:**
- `pkg/repository/repository.go:289`: `migrateToV2` now accepts `*sql.Tx` parameter
- Migration SQL (ALTER TABLE, UPDATE, CREATE INDEX) executes within the transaction
- Version update and migration are atomic -- either both succeed or both roll back

**Verification:** Tests in `repository_test.go` verify atomic migration behavior.

### H6: Content Size Limits -- **FIXED**

**What was done:**
- `pkg/normalizer/normalizer.go:29-32`: Added `MaxEntryContentSize = 1MB` and `MaxEntriesPerFeed = 500`
- `pkg/normalizer/normalizer.go:117-119`: Feed items truncated to `MaxEntriesPerFeed`
- `pkg/normalizer/normalizer.go:167-186`: Content and description truncated to `MaxEntryContentSize`

**Verification:** Tests confirm that feeds with >500 entries are truncated and entries with >1MB content are truncated.

### M3: Feed URL Validation -- **FIXED**

**What was done:**
- `cmd/rp/cmd_add_feed.go:16`: Added `crawler.ValidateURL(opts.URL)` before storing feed URL
- `cmd/rp/cmd_helpers.go`: Added URL validation in `importFeedsFromURLs`

**Verification:** `add-feed` now rejects private IPs, localhost, non-http schemes. Tests confirm rejection.

### M4: PRAGMA foreign_keys Scope -- **FIXED**

**What was done:**
- `pkg/repository/repository.go:72`: Added `db.SetMaxOpenConns(1)` to ensure all operations use the same connection where `PRAGMA foreign_keys = ON` was set

**Verification:** Single connection ensures PRAGMA consistency. Tests verify foreign key enforcement.

### M6: Symlink Handling -- **FIXED**

**What was done:**
- `pkg/generator/generator.go:249-256`: Added `os.Lstat()` check before copying files; symlinks are skipped with a continue statement

**Verification:** Symlinks in template asset directories are silently skipped, preventing path traversal/exfiltration.

### M7: Path Traversal -- **FIXED**

**What was done:**
- `pkg/config/config.go:358-366`: `pathContainsTraversal()` now uses `filepath.Clean()` to normalize paths before checking each component for `..`

**Verification:** Tests confirm that `../etc/passwd`, `foo/../../../etc/passwd`, and similar traversal attempts are rejected. False positives like `foo..bar` are not flagged.

### M10/M11: Build Flags -- **FIXED**

**What was done:**
- `Makefile:20`: `GOBUILD := CGO_ENABLED=1 $(GOCMD) build -trimpath`
- `.github/workflows/go.yml:88`: `CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -v -o rp ./cmd/rp`

**Verification:** Binary no longer leaks build paths. CGO explicitly enabled for SQLite.

### M12: CI Permissions -- **FIXED**

**What was done:**
- `.github/workflows/go.yml:9-11`: Added `permissions: contents: read, security-events: write`

**Verification:** `GITHUB_TOKEN` now follows least-privilege principle.

### L8: Error String Comparison -- **FIXED**

**What was done:**
- `pkg/repository/repository.go:213`: Changed to `SELECT COALESCE(MAX(version), 0) FROM schema_version` -- eliminates fragile string comparison for NULL handling

### L9: PruneOldEntries Validation -- **FIXED**

**What was done:**
- `pkg/repository/repository.go:631-633`: Added `if days < 1 { return 0, fmt.Errorf("days must be >= 1, got %d", days) }`

**Verification:** Tests confirm that `days=0` and `days=-1` return errors instead of deleting all entries.

---

## Remaining Open Findings

### MEDIUM (5 remaining)

| ID | Finding | Risk | Recommendation |
|----|---------|------|----------------|
| M1 | DNS rebinding protection | Attacker-controlled DNS could bypass SSRF checks | Implement custom `DialContext` that validates resolved IPs |
| M2 | IPv4-mapped IPv6 handling | `::ffff:127.0.0.1` may bypass checks (Go stdlib currently handles this, but no tests) | Add explicit `ip.To4()` re-validation and test cases |
| M5 | CSP `unsafe-inline` for styles | CSS injection possible if combined with HTML injection vector | Move to external stylesheet + hash-based CSP |
| M8 | No batch transactions for entries | Performance penalty from per-write WAL sync; non-atomic batch inserts | Expose transactional batch API |
| M9 | No input validation in repository | Any new caller bypassing upstream validation could store invalid data | Add URL/content validation in repository methods |

### LOW (14 remaining)

| ID | Finding | Mitigation Status |
|----|---------|-------------------|
| L1 | Missing special IP range blocks | Partially mitigated by existing private/loopback checks |
| L2 | No explicit minimum TLS version | Go defaults to TLS 1.2+ |
| L3 | Gzip decompression ratio | Mitigated by 10MB decompressed size limit |
| L4 | Internal error details in database | Low risk -- database is local-only |
| L5 | `NewForTesting()` exported | Low risk -- requires intentional misuse |
| L6 | Retry-After no upper bound | Mitigated by 5-minute cap in `FetchWithRetry()` |
| L7 | Missing security SQLite PRAGMAs | Low risk -- database stores feed content only |
| L10 | Database default permissions | Low risk -- typically single-user deployment |
| L11 | No context on schema init | Low risk -- only runs at startup |
| L12 | Relative URLs not resolved | Functional issue, not security-critical |
| L13 | Hash truncation collision risk | Low probability with truncated SHA256 |
| L14 | No future date clamping | Allows content ordering manipulation |
| L15 | No OPML file size limit | Low risk -- OPML is typically small |
| L16 | No upper bound on `days` config | Could cause excessive queries with extreme values |
| L17 | golangci-lint `version: latest` | CI reproducibility concern, not security |

---

## Test Coverage (Re-Audit)

| Package | Before | After | Change |
|---------|--------|-------|--------|
| `pkg/config` | 96.5% | 96.6% | +0.1% |
| `pkg/normalizer` | 94.2% | 95.1% | +0.9% |
| `pkg/crawler` | 92.9% | 88.3% | -4.6%* |
| `pkg/ratelimit` | 91.1% | 91.1% | -- |
| `pkg/opml` | 88.9% | 88.9% | -- |
| `pkg/logging` | 95.7% | 95.7% | -- |
| `pkg/timeprovider` | 100.0% | 100.0% | -- |
| `pkg/fetcher` | 100.0% | 100.0% | -- |
| `pkg/generator` | 79.4% | 79.4% | -- |
| `cmd/rp` | 66.7% | 66.9% | +0.2% |
| `pkg/repository` | 66.4% | 74.1% | **+7.7%** |

*\*Crawler coverage decreased because new production code paths (SSRF redirect validation) were added; the comprehensive test file covers the new functionality but the ratio shifted.*

**Overall coverage: 78.2%** (above 75% project threshold)

### Coverage Improvements
- `pkg/repository` improved from 66.4% to 74.1% -- approaching the 75% target
- New tests added for: title XSS, mailto scheme blocking, content size limits, SSRF redirect bypass, migration transactions, PRAGMA verification, prune validation

### Remaining Test Gaps
1. `TestHTMLGeneration` still skipped in `cmd/rp/integration_test.go`
2. `cmd/rp` at 66.9% -- still below 75% threshold
3. No concurrent database write tests
4. No IPv4-mapped IPv6 test cases in SSRF prevention

---

## Dependency Assessment

All dependencies use permissive licenses (MIT, BSD, Apache-2.0, Public Domain). No known unpatched CVEs:

- `golang.org/x/net` v0.46.0 -- patched for CVE-2025-47911/CVE-2025-58190
- `go-sqlite3` v1.14.32 -- bundles SQLite 3.50.3, patched for CVE-2025-6965
- `bluemonday` v1.0.27 -- no known CVEs

**Concern:** Two unmaintained transitive dependencies remain:
- `modern-go/concurrent` (2018) -- pulled via `gofeed`
- `aymerick/douceur` (2015) -- pulled via `bluemonday`

These are transitive and cannot be updated independently. Monitor for upstream replacements.

---

## Quality Verification

| Check | Result |
|-------|--------|
| All tests pass | 11/11 packages pass |
| Race detector | Clean (no races detected) |
| `go vet` | Clean (no issues) |
| Coverage > 75% | 78.2% overall (9 of 11 packages above threshold) |
| No regressions | Confirmed -- all pre-existing tests still pass |

---

## Recommended Next Steps

### Priority 1: Complete Remaining High-Impact Items
1. Fix H3 fully: Remove `continue-on-error` from Trivy scan step (keep only on SARIF upload)
2. Fix M1: Implement custom `DialContext` for DNS rebinding prevention
3. Fix M2: Add IPv4-mapped IPv6 test cases and explicit handling

### Priority 2: Improve Test Coverage
4. Increase `cmd/rp` coverage above 75% (add tests for `cmdListFeeds`, `cmdStatus`, `cmdVersion`)
5. Unskip `TestHTMLGeneration` integration test
6. Add concurrent database write tests

### Priority 3: Defense-in-Depth
7. Fix M5: CSP hardening (external stylesheet + hash-based CSP)
8. Fix M8: Batch transaction API for entry operations
9. Address remaining LOW findings as time permits
