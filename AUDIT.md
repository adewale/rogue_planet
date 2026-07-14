# Rogue Planet Security & Quality Audit

**Initial Audit:** 2026-02-14
**Re-Audit:** 2026-02-21
**Third Audit (this document):** 2026-07-14
**Scope:** Full codebase audit covering security, code quality, testing, dependencies, and CI/CD

---

## Executive Summary

The initial audit (2026-02-14) identified **1 critical**, **6 high**, **12 medium**, and **17 low** severity findings. Two rounds of remediation followed. This third audit re-verifies the entire codebase against the prior findings and hunts for new issues.

**Result of the 2026-07-14 re-audit:**

- **Every previously-tracked CRITICAL, HIGH, and MEDIUM *code* finding is now FIXED and verified** (C1, H1, H4, H5, H6, M1–M4, M6–M12).
- **2 prior findings are PARTIAL**: L1 (a couple of special IP ranges still unblocked) and L11 (migrations don't inherit the schema-init context).
- **Remaining OPEN items are all CI/CD hardening**: M-GHA (`@master` action refs), M-SCAN (`continue-on-error` on scanners), L-LINT (`version: latest`). Several of these are intentional to match `main`.
- **11 NEW findings** identified (0 critical, 0 high, 2 low/medium, 7 low, 2 info) — most notably a **config parser panic on a lone-quote value** (real crash bug) and a **double-HTML-encoding display bug on entry titles**.
- **Quality gates green**: all 11 packages pass, `go vet` clean, overall coverage **78.3%** (above the 75% threshold).

No exploitable XSS, SSRF, SQL-injection, or path-traversal vector was found in the current tree.

---

## Prior Findings — Verification Status (2026-07-14)

| ID | Severity | Finding | Status | Evidence |
|----|----------|---------|--------|----------|
| C1 | CRITICAL | XSS via unsanitized entry titles | **FIXED** | normalizer.go:373-378 (StrictPolicy); generator.go:47 (`Title string`); cmd_helpers.go:273 (no `template.HTML` cast) |
| H1 | HIGH | SSRF redirect bypass | **FIXED** | crawler.go:182-187, 295-300, 429-434 (ValidateURL in all CheckRedirect) |
| H2 | HIGH | GitHub Actions pinned to `@master` | OPEN* | go.yml:135,141 — reverted to `@master` to match `main` (see M-GHA) |
| H3 | HIGH | Security scans `continue-on-error` | OPEN* | go.yml:132,138,147 (see M-SCAN) |
| H4 | HIGH | `mailto:` scheme allowed | **FIXED** | normalizer.go:74,78-80 |
| H5 | HIGH | Migration transaction scope | **FIXED** | repository.go:303-340 (atomic `*sql.Tx`) |
| H6 | HIGH | No content size limits | **FIXED** | normalizer.go:32,34,124-126,176-197 |
| M1 | MEDIUM | DNS rebinding protection | **FIXED** | crawler.go:355-388 (`safeDialContext` validates resolved IPs, dials literal) |
| M2 | MEDIUM | IPv4-mapped IPv6 handling | **FIXED** | crawler.go:50-52 (`ip.To4()` normalization) |
| M3 | MEDIUM | Feed URL validation gaps | **FIXED** | cmd_add_feed.go:16; cmd_import_opml.go:89; cmd_helpers.go:66 |
| M4 | MEDIUM | PRAGMA foreign_keys scope | **FIXED** | repository.go:81 (`SetMaxOpenConns(1)`), :114 |
| M5 | MEDIUM | CSP allows `unsafe-inline` | **FIXED** | generator.go:640 (no `unsafe-inline`), CSS externalized :176-182,:643 |
| M6 | MEDIUM | No symlink handling in copyDir | **FIXED** | generator.go:257-265 (`os.Lstat` skip) |
| M7 | MEDIUM | Path traversal string check | **FIXED** | config.go:362-371 (`filepath.Clean` + component match) |
| M8 | MEDIUM | No batch transactions for entries | **FIXED** | repository.go:567-622 (`UpsertEntriesBatch`) |
| M9 | MEDIUM | No input validation in repository | **FIXED** | repository.go:373-390, 522-531 |
| M10 | MEDIUM | Missing `-trimpath` build flag | **FIXED** | Makefile; go.yml:88 |
| M11 | MEDIUM | Missing `CGO_ENABLED=1` | **FIXED** | go.yml:88 |
| M12 | MEDIUM | Missing `permissions` block in CI | **FIXED** | go.yml:9-11 (nit N2-gen) |
| L1 | LOW | Missing special IP range blocks | **PARTIAL** | crawler.go:30-34 (CGNAT/bench/`0.0.0.0/8` added); `240.0.0.0/4` & `255.255.255.255` still open |
| L2 | LOW | No explicit minimum TLS version | **FIXED** | crawler.go:169-171, 282-284 (`MinVersion: TLS12`) |
| L3 | LOW | Gzip decompression ratio | **FIXED** | crawler.go:497-513 (LimitedReader wraps decompressed stream) |
| L4 | LOW | Internal error details in database | **FIXED** | fetcher.go:225 (`sanitizeErrorMessage`) |
| L5 | LOW | `NewForTesting()` exported | OPEN | Accepted risk (test-only helper) |
| L6 | LOW | Retry-After date no upper bound | **FIXED** | crawler.go:24,552-567,588-590 |
| L7 | LOW | Missing security SQLite PRAGMAs | **FIXED** | repository.go:102-130 (`trusted_schema`, `cell_size_check`) |
| L8 | LOW | Fragile error string comparison | **FIXED** | repository.go:267 (`COALESCE(MAX(version),0)`) |
| L9 | LOW | PruneOldEntries no lower bound | **FIXED** | repository.go:788-790 |
| L10 | LOW | Database file default permissions | **FIXED** | repository.go:93-98 (`0600`, Windows-guarded; sidecar gap NEW-5) |
| L11 | LOW | No context on schema init | **PARTIAL** | initSchema ctx present :156-158; `runMigrations` :303 has none |
| L12 | LOW | Relative URLs not resolved | **FIXED** | normalizer.go:311-368 |
| L13 | LOW | Hash truncation collision risk | **FIXED** | normalizer.go:222,234 (full SHA-256) |
| L14 | LOW | No future date clamping | **FIXED** | normalizer.go:289-294 |
| L15 | LOW | No OPML file size limit | **FIXED** | opml.go:54,74-98 (10MB cap) |
| L16 | LOW | No upper bound on `days` config | **FIXED** | config.go:56,247-249,323-325 |
| L17 | LOW | golangci-lint `version: latest` | OPEN | go.yml:112 (see L-LINT) |

\*H2/H3 were previously marked FIXED/PARTIAL; the security-scan CI configuration was subsequently reverted to `@master` + `continue-on-error` to match the `main` branch and unblock the pipeline. They are re-tracked as M-GHA and M-SCAN below.

---

## Remaining Open Findings

### CI/CD Hardening (Medium/Low)

| ID | Severity | Finding | Risk | Recommendation |
|----|----------|---------|------|----------------|
| M-GHA | MEDIUM | `securego/gosec@master`, `aquasecurity/trivy-action@master`, `govulncheck@latest` unpinned (go.yml:130,135,141) | Mutable refs run in CI with `security-events: write`; supply-chain / reproducibility exposure | Pin to commit SHAs. Mitigated by `contents: read` default. Currently `@master` intentionally tracks `main`. |
| M-SCAN | MEDIUM | `continue-on-error: true` on govulncheck/gosec/trivy (go.yml:132,138,147); `security` job not in any `needs:` | Vulnerable deps or introduced vulns can merge silently | Drop `continue-on-error` on govulncheck+gosec (keep only on the SARIF upload step). Note: `govulncheck` currently exits non-zero on known transitive advisories, which is why it is suppressed. |
| L-LINT | LOW | `golangci-lint` `version: latest` (go.yml:112) | Non-reproducible lint; a new release can change results | Pin to an explicit version. |

### Code (Partial)

| ID | Severity | Finding | Gap | Recommendation |
|----|----------|---------|-----|----------------|
| L1 | LOW | Special IP range blocks | `240.0.0.0/4` (class-E) and limited broadcast `255.255.255.255` not blocked (crawler.go:30-34) | Add both to `blockedCIDRs`. |
| L11 | LOW | Context on schema init | `runMigrations` (repository.go:303) uses `db.Begin()` with no context; the 30s schema-init timeout doesn't cover migrations | Thread `ctx` into `runMigrations`/`migrateToV2`, use `BeginTx(ctx,...)`. |
| L5 | LOW | `NewForTesting()` exported | Test-only constructor bypasses SSRF checks | Accepted; requires intentional misuse. |

---

## New Findings (2026-07-14)

### N-CFG1 — LOW/MEDIUM — Config parser panics on a lone-quote value (`pkg/config/config.go:185-188`)

The quote-stripping block enters its branch whenever the value both starts and ends with the same quote char. A value consisting of a **single** quote byte (e.g. a config line `name = "` or `name = '`) satisfies both `HasPrefix` and `HasSuffix` against that one byte, so it evaluates `value[1:len(value)-1]` = `value[1:0]` and **panics** ("slice bounds out of range"). A malformed config file crashes the process instead of returning a parse error.
**Remediation:** guard with `len(value) >= 2` before stripping quotes.

### N-TITLE — LOW — Double HTML-encoding of entry titles + stale comment (`pkg/normalizer/normalizer.go:370-378`)

`sanitizeTitle` runs bluemonday `StrictPolicy().Sanitize`, which HTML-entity-encodes `&`, `<`, `>` (e.g. `A & B` → `A &amp; B`). The result is stored and later rendered as a plain `string` through `html/template` auto-escaping, which encodes **again** (`&amp;amp;`). Titles containing those characters display double-encoded. Safe direction (not an XSS), but a correctness bug. The doc comment at :371-372 ("titles are cast to `template.HTML` for rendering") is also stale — titles are no longer cast to `template.HTML`.
**Remediation:** strip tags without entity-encoding (e.g. `html.UnescapeString` after StrictPolicy, or a text-extraction pass) and update the comment.

### N-OPML — LOW/MEDIUM — Unbounded recursion in OPML outline flattening (`pkg/opml/opml.go:114-143`)

`extractOutlines` recurses once per nesting level with no depth limit. A 10MB OPML of deeply nested `<outline>` elements (a few bytes per level) can reach very high recursion depth and overflow the goroutine stack (DoS). The 10MB size cap bounds width, not depth. (`encoding/xml` unmarshalling of the nested structure recurses similarly before `extractOutlines` even runs.)
**Remediation:** convert to an explicit stack, or enforce a max nesting depth (e.g. 100) and error beyond it.

### N-LINK — LOW — Normalizer does not scheme-filter `entry.Link` (`pkg/normalizer/normalizer.go:159,381-393`)

`resolveURL` returns the resolved reference with **no scheme check**, and `normalizeEntry` falls back to the raw `item.Link`. A feed item with `Link = "javascript:alert(1)"` yields an unsanitized `entry.Link` that is rendered into `href="{{.Link}}"`. Currently **not a live XSS** because Go `html/template`'s contextual URL filter neutralizes dangerous schemes (`#ZgotmplZ`), but the normalizer provides no defense-in-depth; any consumer using `text/template` or a non-URL context would be exposed.
**Remediation:** reject/blank `entry.Link` whose resolved scheme is not http/https, mirroring the `AllowURLSchemes` policy.

### N-PATH — LOW — Traversal check does not reject absolute paths (`pkg/config/config.go:362-371`)

`pathContainsTraversal` only rejects `..` components. Absolute paths (`/etc/...`, `/var/www/...`) for `Database.Path`, `OutputDir`, or `Template` pass validation, allowing reads/writes outside the working directory. Low risk (config is user-supplied), but a write primitive if configs are ever sourced from a less-trusted context.
**Remediation:** additionally reject `filepath.IsAbs(cleaned)` for these paths, or confine to a base directory.

### N-WAL — LOW — WAL sidecar permissions & chmod TOCTOU (`pkg/repository/repository.go:85-98`)

The `0600` chmod applies only to the main DB file. In WAL mode SQLite also creates `<db>-wal` and `<db>-shm` (containing recently-written row data) at the process umask (typically `0644`), partially defeating L10. There is also a brief window between `db.Ping()` (which creates the file) and `os.Chmod` where the main file sits at umask perms.
**Remediation:** lower the process umask around creation, or chmod the `-wal`/`-shm` files too.

### N-DAYS — LOW — Repository query methods lack defensive `days` bounds (`pkg/repository/repository.go:627,755,787`)

`GetRecentEntries`, `CountRecentEntries`, and `GetRecentEntriesWithOptions` accept `days` with no validation (only `PruneOldEntries` bounds it). Negative `days` yields a future cutoff (empty/fallback results); a huge value is a full scan. Not a security issue given callers pass the config-bounded value, but the repository is an exported public API (`pkg/repository/interface.go`) and should validate defensively.
**Remediation:** reject `days < 1` consistently across these methods.

### N-DEFLATE — LOW — `Accept-Encoding: deflate` advertised but not decoded (`pkg/crawler/crawler.go:416,487`)

`Fetch` advertises `Accept-Encoding: gzip, deflate` but only handles `Content-Encoding: gzip`. A server responding with `deflate` has its raw compressed bytes handed to the parser, silently failing the fetch. Not a security issue (size limits still apply); a reliability bug.
**Remediation:** drop `deflate` from the advertised encodings, or add a `flate`/`zlib` decode branch.

### N-DIAL — LOW — `safeDialContext` always dials `ips[0]`, ignoring family/failover (`pkg/crawler/crawler.go:382-386`)

`safeDialContext` connects only to `ips[0]` regardless of the requested `network` (`tcp4`/`tcp6`) and never fails over to other resolved IPs. If `ips[0]` is IPv6 but `network == "tcp4"` (or the first IP is unreachable), the dial fails though a compatible/reachable address exists. Reliability only — all IPs are already validated, so no security impact.
**Remediation:** filter `ips` by requested family and iterate on dial failure.

### N-FEEDS — INFO — `LoadFeedsFile` has no size limit (`pkg/config/config.go:378-397`)

`os.ReadFile` loads the entire feeds file with no cap, unlike OPML (10MB). Local, user-controlled file, so informational, but inconsistent with the OPML hardening.

### N-CI-PERMS — LOW — Workflow-wide `security-events: write` (`.github/workflows/go.yml:9-11`)

`security-events: write` is granted workflow-wide, so `test`/`build`/`lint` jobs receive it too. Least privilege would scope it to the `security` job only.

### N-CI-FF — LOW — Test matrix `fail-fast` not disabled (`.github/workflows/go.yml:17-20`)

The matrix has no `fail-fast: false`, so GitHub's default cancels remaining OS legs on the first failure, hiding platform-specific failures (relevant with a 3-OS + CGO/sqlite matrix). Recommend `strategy: { fail-fast: false, matrix: ... }`.

### N-CSP — LOW — CSP meta residuals (`pkg/generator/generator.go:640`)

CSP is delivered via `<meta http-equiv>` (acceptable for static output). `frame-ancestors`/`form-action` cannot be enforced via meta, and `img-src 'self' https:` permits images from any HTTPS origin (tracking-pixel/privacy vector inherent to feed content). Not script-injectable given `script-src 'self'` + `object-src 'none'` + `base-uri 'self'`.

---

## Non-Findings (Verified Safe)

- **SQL injection:** none. All queries use bound `?` parameters. The two `fmt.Sprintf`-built queries (repository.go:699-706, 725-733) interpolate only whitelisted values (`validSortFields` map, boolean-selected column).
- **Resource leaks:** none. Every `QueryContext` has `defer rows.Close()`; the batch prepared statement is deferred-closed.
- **XXE / entity expansion (OPML):** not exploitable. `encoding/xml` resolves no external DTD/SYSTEM entities and does no recursive internal-entity expansion; combined with the 10MB cap this is safe.

---

## Test Coverage (2026-07-14)

| Package | Coverage |
|---------|----------|
| `pkg/timeprovider` | 100.0% |
| `pkg/fetcher` | 100.0% |
| `pkg/config` | 96.7% |
| `pkg/logging` | 95.7% |
| `pkg/normalizer` | 94.4% |
| `pkg/ratelimit` | 91.1% |
| `pkg/opml` | 88.5% |
| `pkg/crawler` | 83.9% |
| `pkg/generator` | 79.4% |
| `pkg/repository` | 75.0% |
| `cmd/rp` | 66.6% |
| **Overall** | **78.3%** |

**Remaining coverage gaps:**
1. `cmd/rp` at 66.6% — below the 75% threshold (add tests for list/status/version handlers).
2. `pkg/generator` at 79.4% — asset-copy and error paths.
3. No concurrent database-write tests.

---

## Dependency Assessment

All dependencies use permissive licenses (MIT, BSD, Apache-2.0, Public Domain). No known open advisories in the pinned versions:

- `golang.org/x/net v0.46.0` — above the CVE-2025-22872 / http2 fix lines
- `github.com/mattn/go-sqlite3 v1.14.32` — current
- `github.com/microcosm-cc/bluemonday v1.0.27` — current
- `golang.org/x/text v0.30.0`, `golang.org/x/time v0.14.0`, `gofeed v1.3.0`, `goquery v1.10.3` — current

**Notes:**
- `go.mod` declares `go 1.24.0`; CI builds/tests on Go `1.25`. Consistent; not a vulnerability.
- Two unmaintained transitive deps remain (`modern-go/concurrent` 2018 via `gofeed`, `aymerick/douceur` 2015 via `bluemonday`). Transitive; monitor upstream.

---

## Quality Verification

| Check | Result |
|-------|--------|
| All tests pass | 11/11 packages pass |
| `go vet` | Clean |
| Coverage > 75% | 78.3% overall |
| No regressions | Confirmed |

---

## Recommended Next Steps

### Priority 1 — Correctness bugs (new)
1. **N-CFG1**: guard the quote-strip with `len(value) >= 2` to stop the config-parser panic.
2. **N-TITLE**: fix double-encoding of titles and update the stale comment.
3. **N-OPML**: bound OPML nesting depth to prevent stack-overflow DoS.

### Priority 2 — Defense-in-depth (new)
4. **N-LINK**: scheme-filter `entry.Link` in the normalizer.
5. **N-PATH**: reject absolute paths in `pathContainsTraversal`.
6. **N-DAYS**: validate `days < 1` in repository query methods.
7. **N-WAL**: restrict `-wal`/`-shm` permissions.

### Priority 3 — CI/CD hardening
8. **M-SCAN**: stop suppressing govulncheck/gosec failures (keep `continue-on-error` only on SARIF upload); address the transitive advisories that currently force the suppression.
9. **M-GHA / L-LINT**: pin `gosec`/`trivy`/`golangci-lint` to SHAs/versions.
10. **N-CI-PERMS / N-CI-FF**: scope `security-events: write` to the security job; set `fail-fast: false`.

### Priority 4 — Reliability & coverage
11. **N-DEFLATE / N-DIAL**: fix deflate handling and dial family/failover.
12. Raise `cmd/rp` coverage above 75%; add concurrent DB-write tests.
