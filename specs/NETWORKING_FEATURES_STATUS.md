# Networking Features Status

**Document Status**: This document has been replaced by authoritative sources
**Last Updated**: 2025-10-30

---

## ⚠️ Notice: Document Archived

This networking features status document became outdated after the v0.4.0 release. Rather than maintaining duplicate status information that can drift out of sync, please refer to the authoritative sources below.

---

## Current Networking Feature Status

For the current status of all networking features in Rogue Planet, please see:

### Primary Sources

1. **CHANGELOG.md** - Authoritative record of what's implemented
   - See v0.4.0 section for production HTTP performance features
   - Per-domain rate limiting, fine-grained timeouts, 301 redirects, Retry-After support

2. **README.md** - User-facing feature documentation
   - "Features" section lists HTTP Performance capabilities
   - "Good Netizen Behavior" section describes networking behavior

3. **CLAUDE.md** - Developer implementation details
   - "Good Netizen Behavior" section (lines 218-227)
   - Rate limiter implementation notes (pkg/ratelimit/ section)
   - HTTP conditional requests documentation

### Implementation Details

For detailed implementation information:

- **Rate Limiting**: See `pkg/ratelimit/ratelimit.go` and test suite
- **HTTP Timeouts**: See `pkg/config/config.go` and `pkg/crawler/crawler.go`
- **301 Redirects**: See `pkg/crawler/crawler.go` FinalURL handling
- **Retry Logic**: See `pkg/crawler/crawler.go` FetchWithRetry method

---

## What's Implemented in v0.4.0 ✅

**Core Networking Features**:
- ✅ HTTP conditional requests (ETag/Last-Modified)
- ✅ SSRF prevention (localhost, private IPs blocked)
- ✅ Gzip/deflate decompression
- ✅ Response size limiting (10MB max)
- ✅ Timeout handling (configurable)
- ✅ Redirect following (301/302)
- ✅ User-Agent headers
- ✅ Exponential backoff retry

**v0.4.0 Production HTTP Performance**:
- ✅ Per-domain rate limiting (60 req/min default, configurable)
- ✅ Fine-grained HTTP timeouts (http/dial/TLS/response header)
- ✅ 301 permanent redirect auto-updating
- ✅ Retry-After header support (RFC 7231)
- ✅ HTTP/1.1 connection pooling and keep-alive

---

## What's Planned for v1.0

See `specs/TODO.md` and `specs/v1.0.0-plan.md` for:
- Feed autodiscovery (parse HTML for RSS/Atom links)
- Intelligent feed scheduling (adaptive polling)
- Cache-Control: max-age header respect
- Jitter for thundering herd prevention

---

## Why This Document Was Archived

During v0.4.0 development, we learned that maintaining multiple status documents leads to inconsistency. The authoritative sources (CHANGELOG.md, README.md, CLAUDE.md) are updated as part of feature implementation, ensuring they stay in sync with the code.

**Lesson Learned**: Single source of truth for feature status prevents documentation drift.

---

## Historical Context

This document was originally created to track networking feature implementation across multiple planning documents. After v0.4.0 completion, we consolidated this information into the primary documentation sources listed above.

**Original Purpose**: Track networking features from audit to implementation
**Why Replaced**: Duplicate status information caused synchronization issues
**Replacement**: CHANGELOG.md (authoritative), README.md (user-facing), CLAUDE.md (developer)

---

*For questions about networking features, see CHANGELOG.md or open a GitHub issue.*
*Last comprehensive networking update: v0.4.0 (2025-10-30)*
