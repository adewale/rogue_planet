# Rogue Planet Security & Quality Audit

**Date:** 2026-02-14
**Scope:** Full codebase audit covering security, code quality, testing, dependencies, and CI/CD

---

## Executive Summary

Rogue Planet demonstrates strong engineering fundamentals: parameterized SQL throughout, proper use of `html/template`, bluemonday sanitization, SSRF prevention, and a clean race detector pass. However, the audit identified **1 critical**, **6 high**, **12 medium**, and **17 low** severity findings. The critical finding (unsanitized entry titles cast to `template.HTML`) is the exact class of XSS vulnerability (CVE-2009-2937) that the project explicitly aims to prevent.

---

## Finding Summary

| Severity | Count | Key Areas |
|----------|-------|-----------|
| CRITICAL | 1 | XSS via unsanitized titles |
| HIGH | 6 | SSRF redirect bypass, CI supply chain, `mailto:` scheme leak, migration transactions, security scans ineffective, no content size limits |
| MEDIUM | 12 | DNS rebinding, IPv4-mapped IPv6, feed URL validation gaps, CSP weakness, foreign key pragma scope, path traversal, and more |
| LOW | 17 | Missing TLS min version, IP range gaps, build flags, error leakage, etc. |
| INFORMATIONAL | 20+ | Correct implementations confirmed |

---

## CRITICAL Findings

### C1: Unsanitized Entry Titles Cast to `template.HTML` -- Stored XSS

**Files:** `pkg/normalizer/normalizer.go:126`, `cmd/rp/cmd_helpers.go:266`, `pkg/generator/generator.go:47`

The entry title is extracted with only `strings.TrimSpace()` -- no HTML sanitization is applied. It is then cast to `template.HTML` in `cmd_helpers.go:266`, which tells Go's template engine to render it as raw, unescaped HTML. A malicious feed title like `<img src=x onerror=alert(1)>` would execute JavaScript in the generated page.

The `Content` and `Summary` fields are properly sanitized through bluemonday, but `Title` was missed. This is precisely CVE-2009-2937 (Planet Venus XSS).

**Fix:** Either sanitize titles through `sanitizeHTML()` in the normalizer, or change `EntryData.Title` from `template.HTML` to `string` so `html/template` auto-escapes it. The latter is simpler and safer since titles should not contain HTML.

---

## HIGH Findings

### H1: No SSRF Validation on HTTP Redirect Targets

**Files:** `pkg/crawler/crawler.go:296-311`, `pkg/fetcher/fetcher.go:95-105`

The `CheckRedirect` callback only checks redirect count, not the redirect target URL against the SSRF blocklist. A feed at `https://evil.example.com/feed` could 301-redirect to `http://169.254.169.254/latest/meta-data/`. Worse, the redirected URL is persisted via `UpdateFeedURL()` without validation, causing repeated SSRF on subsequent fetches.

**Fix:** Call `ValidateURL()` on each redirect target in the `CheckRedirect` callback. Also validate `resp.FinalURL` before storing it in the database.

### H2: GitHub Actions Pinned to `@master` -- Supply Chain Risk

**File:** `.github/workflows/go.yml:130,136`

Both `securego/gosec@master` and `aquasecurity/trivy-action@master` use mutable branch references. If either repository is compromised, malicious code executes in CI with full repository access.

**Fix:** Pin to specific commit SHAs (e.g., `securego/gosec@<sha> # vX.Y.Z`).

### H3: Security Scans Use `continue-on-error: true`

**File:** `.github/workflows/go.yml:133,142`

Both Gosec and Trivy are configured with `continue-on-error: true`, meaning security findings never fail the build. Vulnerabilities are silently ignored unless developers check the GitHub Security tab.

**Fix:** Remove `continue-on-error: true` or implement branch protection rules requiring code scanning to pass.

### H4: `mailto:` Scheme Allowed Despite Policy Stating Only http/https

**File:** `pkg/normalizer/normalizer.go:58-61`

`bluemonday.UGCPolicy()` internally calls `AllowStandardURLs()` which adds `mailto` to allowed schemes. The subsequent `AllowURLSchemes("http", "https")` call adds to (not replaces) the existing map. The spec explicitly states "Only allow http/https URL schemes."

**Fix:** Build a custom policy from `bluemonday.NewPolicy()` instead of `UGCPolicy()`, or explicitly remove `mailto` after creating the UGC policy.

### H5: Migration Transaction Does Not Enclose Migration SQL

**File:** `pkg/repository/repository.go:260-281`

The transaction `tx` is created, but `migrateFn()` (e.g., `migrateToV2`) executes SQL against `r.db` directly rather than `tx`. If the migration succeeds but the version update fails, the schema is modified but the version table doesn't reflect it, causing a stuck state on next startup.

**Fix:** Pass `*sql.Tx` into migration functions so all DDL/DML executes within the transaction.

### H6: No Content Length Limit in Normalizer

**File:** `pkg/normalizer/normalizer.go:81,147-157`

The normalizer has no per-entry content size limit or entry count limit. A feed with a single multi-megabyte `<content>` block or thousands of entries could cause memory exhaustion. While the crawler limits total feed size to 10MB, individual entry content is unbounded.

**Fix:** Add per-entry content size limits (e.g., 1MB) and total entry count limits (e.g., 500).

---

## MEDIUM Findings

### M1: No DNS Rebinding Protection

**File:** `pkg/crawler/crawler.go:79-82` (DialContext)

`ValidateURL()` checks hostname strings, but DNS resolution happens separately in the default `net.Dialer`. An attacker-controlled DNS record could initially resolve to a public IP (passing validation), then change to `127.0.0.1` before TCP connection.

**Fix:** Implement a custom `DialContext` that resolves DNS and checks the resolved IP against the SSRF blocklist before connecting.

### M2: IPv4-Mapped IPv6 Addresses Not Explicitly Handled

**File:** `pkg/crawler/crawler.go:244-259`

No explicit handling of IPv4-mapped IPv6 addresses like `::ffff:127.0.0.1` or `::ffff:10.0.0.1`. While Go's `IsLoopback()` handles `::ffff:127.0.0.1` correctly, there are no tests for these cases. Any regression in Go's stdlib would silently break the protection.

**Fix:** Call `ip.To4()` and re-validate. Add test cases for `http://[::ffff:127.0.0.1]/feed` and similar.

### M3: Feed URLs Not Validated on `add-feed` Command

**Files:** `cmd/rp/cmd_add_feed.go:8-29`, `cmd/rp/cmd_helpers.go:60-73`

The `add-feed` and `add-all` commands store URLs without SSRF or scheme validation. Contrast: `import-opml` does call `crawler.ValidateURL()` before adding. Malicious URLs get stored and exported to OPML files that other readers may consume.

**Fix:** Call `crawler.ValidateURL()` in `cmdAddFeed` and `importFeedsFromURLs` before storing.

### M4: `PRAGMA foreign_keys` Per-Connection Scope

**File:** `pkg/repository/repository.go:62-78`

`PRAGMA foreign_keys = ON` is set via `db.Exec()`, which applies to one connection from the `database/sql` pool. Additional pool connections won't have foreign keys enabled, silently breaking CASCADE DELETE.

**Fix:** Set `db.SetMaxOpenConns(1)` or use DSN parameter `?_foreign_keys=1` or register a connection init hook.

### M5: CSP Allows `'unsafe-inline'` for Styles

**File:** `pkg/generator/generator.go:425`

The CSP includes `style-src 'self' 'unsafe-inline'`, which allows CSS injection for UI redressing/phishing if combined with an HTML injection vector.

**Fix:** Move CSS to an external stylesheet and use hash-based CSP for the known `<style>` block.

### M6: No Symlink Handling in `copyDir`/`copyFile`

**File:** `pkg/generator/generator.go:222-310`

The static asset copy doesn't check for symlinks. A malicious custom template could symlink to sensitive files that would be copied to the web-accessible output directory.

**Fix:** Use `os.Lstat()` to detect and skip symlinks.

### M7: Path Traversal Check Uses Simple String Contains

**File:** `pkg/config/config.go:335-344`

Uses `strings.Contains(path, "..")` which doesn't handle encoded paths, symlinks, or normalized traversal. The generator doesn't validate paths independently.

**Fix:** Use `filepath.Clean` + prefix check to ensure paths stay within expected directories.

### M8: No Batch Transactions for Entry Operations

**File:** `pkg/repository/repository.go:442`

`UpsertEntry` performs individual INSERT/UPDATE without transaction wrapping. Batch inserts from a feed fetch are not atomic and suffer performance penalties from per-write WAL sync.

**Fix:** Expose a transactional batch API or accept `*sql.Tx` in `UpsertEntry`.

### M9: No Input Validation in Repository Layer

**File:** `pkg/repository/repository.go` (multiple methods)

`AddFeed`, `UpdateFeedURL`, `UpsertEntry` accept arbitrary values without validation. While upstream layers validate, any new caller bypassing them could store invalid data.

### M10: Missing `-trimpath` Build Flag

**File:** `Makefile:28`

The binary embeds full local filesystem paths without `-trimpath`, leaking build environment information and preventing reproducible builds.

### M11: Missing Explicit `CGO_ENABLED=1` in Makefile

**File:** `Makefile:57`

The project depends on `mattn/go-sqlite3` (CGO), but doesn't explicitly set `CGO_ENABLED=1`, causing confusing failures in CGO-disabled environments.

### M12: Missing `permissions` Block in GitHub Actions

**File:** `.github/workflows/go.yml`

No `permissions` block declared, granting `GITHUB_TOKEN` broad read/write access. Should restrict to `contents: read` and `security-events: write`.

---

## LOW Findings

### L1: Missing Blocks for Special IP Ranges

`pkg/crawler/crawler.go:237-259` -- `0.0.0.0/8` (beyond `0.0.0.0`), `100.64.0.0/10` (CGNAT), `198.18.0.0/15` (benchmarking), broader multicast not blocked.

### L2: No Explicit Minimum TLS Version

`pkg/crawler/crawler.go:71-92` -- Go defaults to TLS 1.2+, but not explicitly configured.

### L3: Gzip Decompression Ratio Not Checked

`pkg/crawler/crawler.go:354-362` -- Mitigated by 10MB limit on decompressed output, but no compression ratio check.

### L4: Internal Error Details Stored in Database

`pkg/fetcher/fetcher.go:192` -- Raw `err.Error()` (which may contain hostnames/IPs) stored via `UpdateFeedError`.

### L5: `NewForTesting()` Exported Without Build Constraints

`pkg/crawler/crawler.go:120-125` -- Disables SSRF checks and is accessible from production code.

### L6: Retry-After Date Has No Upper Bound

`pkg/crawler/crawler.go:416-421` -- Mitigated by 5-minute cap in `FetchWithRetry()`.

### L7: Missing Security-Hardening SQLite PRAGMAs

`pkg/repository/repository.go:62-88` -- `trusted_schema=OFF`, `cell_size_check=ON`, `busy_timeout` not set.

### L8: Fragile Error String Comparison

`pkg/repository/repository.go:210` -- Compares against specific error message string for NULL handling. Use `COALESCE(MAX(version), 0)` instead.

### L9: `PruneOldEntries` Has No Lower Bound on `days`

`pkg/repository/repository.go:628` -- `days=0` or negative would delete all entries.

### L10: Database File Created with Default Permissions

`pkg/repository/repository.go:63` -- Default 0644 permissions; consider 0600 for defense-in-depth.

### L11: No Context/Timeout on Schema Initialization

`pkg/repository/repository.go:99-153` -- Migration queries use `r.db.Exec()` without context, blocking indefinitely if database is locked.

### L12: Relative URLs in Content Not Resolved

`pkg/normalizer/normalizer.go:243-250` -- `baseURL` parameter accepted but never used; relative URLs remain broken in output.

### L13: Hash Truncation to 64 Bits + Empty-Content Collision Risk

`pkg/normalizer/normalizer.go:174-191` -- Truncated SHA256 increases collision probability. Entries with identical empty content generate the same ID.

### L14: No Future Date Clamping

`pkg/normalizer/normalizer.go:214-232` -- Future dates accepted without capping, allowing content ordering manipulation.

### L15: No OPML File Size Limit

`pkg/opml/opml.go:76-85` -- `ParseFile` reads entire file with `os.ReadFile` without size check.

### L16: No Upper Bound on `days` Config Value

`pkg/config/config.go:241-248` -- Validated `>= 1` but no maximum; extreme values cause excessive queries.

### L17: golangci-lint Using `version: latest`

`.github/workflows/go.yml:109` -- Non-deterministic CI results.

---

## Test Coverage Assessment

| Package | Coverage | Status |
|---------|----------|--------|
| `pkg/config` | 96.5% | Excellent |
| `pkg/normalizer` | 94.2% | Excellent |
| `pkg/crawler` | 92.9% | Excellent |
| `pkg/ratelimit` | 91.1% | Excellent |
| `pkg/opml` | 88.9% | Good |
| `pkg/logging` | 95.7% | Excellent |
| `pkg/timeprovider` | 100.0% | Perfect |
| `pkg/fetcher` | 100.0% | Perfect |
| `pkg/generator` | 79.4% | Adequate |
| `cmd/rp` | **66.7%** | **Below 75% threshold** |
| `pkg/repository` | **66.4%** | **Below 75% threshold** |

### Key Test Gaps

1. **`TestHTMLGeneration` is permanently skipped** (`cmd/rp/integration_test.go:137`) -- the most complete end-to-end CLI test is disabled with `t.Skip()`
2. **`cmd/rp` (66.7%)**: Missing dedicated tests for `cmdListFeeds`, `cmdStatus`, `cmdVersion`, and error paths in `cmdUpdate`/`cmdFetch`
3. **`pkg/repository` (66.4%)**: Missing tests for `ClearFeedError`, `GetFeedByID`, `UpdateFeedNextFetch`, 50-entry fallback limit, and WAL mode verification
4. **No concurrent database write tests** despite production use of worker pools (5-20 concurrent fetchers)
5. **No CSS injection test** in the otherwise exemplary XSS test suite
6. **No `vbscript:` scheme test** despite being called out in the testing plan

### Test Strengths

- XSS prevention tests are exemplary (18 OWASP vectors)
- Table-driven tests used consistently
- `t.Parallel()` used throughout
- Clean race detector pass across all packages
- Real-world feed snapshot tests (Daring Fireball, Asymco)
- Deterministic time testing via `FakeClock`

---

## Dependency Assessment

All dependencies use permissive licenses (MIT, BSD, Apache-2.0, Public Domain). No known unpatched CVEs in current dependency versions:

- `golang.org/x/net` v0.46.0 -- patched for CVE-2025-47911/CVE-2025-58190
- `go-sqlite3` v1.14.32 -- bundles SQLite 3.50.3, patched for CVE-2025-6965
- `bluemonday` v1.0.27 -- no known CVEs

**Concern:** Two unmaintained transitive dependencies (`modern-go/concurrent` from 2018, `aymerick/douceur` from 2015) pulled in via `gofeed` and `bluemonday` respectively.

---

## Prioritized Remediation Plan

### Immediate (Security-Critical)

1. **Fix C1**: Sanitize entry titles or change `EntryData.Title` to `string` type
2. **Fix H1**: Add SSRF validation in `CheckRedirect` callback and before `UpdateFeedURL`
3. **Fix H2**: Pin GitHub Actions to commit SHAs
4. **Fix H3**: Remove `continue-on-error: true` from security scan steps

### Short-Term (High Impact)

5. **Fix H4**: Remove `mailto:` from allowed URL schemes in sanitizer
6. **Fix H5**: Pass `*sql.Tx` to migration functions
7. **Fix H6**: Add per-entry content size and count limits in normalizer
8. **Fix M3**: Add URL validation to `add-feed` and `add-all` commands
9. **Fix M4**: Ensure `PRAGMA foreign_keys` applies to all connections

### Medium-Term (Defense-in-Depth)

10. **Fix M1**: Custom DialContext for DNS rebinding prevention
11. **Fix M2**: Explicit IPv4-mapped IPv6 handling with tests
12. **Fix M5-M7**: CSP hardening, symlink handling, path traversal improvements
13. **Fix M10-M12**: Build and CI improvements
14. Increase `cmd/rp` and `pkg/repository` test coverage above 75%
15. Unskip `TestHTMLGeneration` integration test
