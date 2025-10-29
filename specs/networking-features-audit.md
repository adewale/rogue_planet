# Networking Features Audit: Mentioned but Not Implemented

**Date**: 2025-10-29
**Audit Scope**: All documentation, specs, and source code
**Focus**: Networking, HTTP, performance, and "good netizen" features

---

## Executive Summary

After a thorough sweep through the entire project (documentation, specs, source code, config examples), I found **8 networking-related features** that are mentioned in documentation/specs but not fully implemented in code.

**Categories**:
- 🔴 **High Priority** (3): Critical for production reliability
- 🟡 **Medium Priority** (3): Important for being a good netizen
- 🟢 **Low Priority** (2): Nice-to-have optimizations

---

## 🔴 High Priority: Critical for Production

### 1. Retry-After Header Support (429/503 responses)

**Mentioned in**:
- `specs/rogue-planet-spec.md:152` - "Handle 429 Too Many Requests with backoff"
- `specs/rogue-planet-spec.md:166` - "Honor 429 rate limit responses with Retry-After header"
- `CLAUDE.md:224` - "Honor 429 responses with exponential backoff"

**Current State**: ❌ NOT IMPLEMENTED
- FetchWithRetry implements exponential backoff BUT ignores Retry-After header
- No parsing of Retry-After header from 429 or 503 responses
- No special handling of rate limit responses

**Implementation Gap**:
```go
// Current code in crawler.go:266-310
// FetchWithRetry uses fixed exponential backoff (2^n seconds)
// Should check for Retry-After header first:

func parseRetryAfter(header string) (time.Duration, error) {
    // Parse "Retry-After: 120" (seconds)
    // Parse "Retry-After: Wed, 21 Oct 2015 07:28:00 GMT" (HTTP-date)
}

// In FetchWithRetry:
if resp.StatusCode == 429 || resp.StatusCode == 503 {
    if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
        delay, err := parseRetryAfter(retryAfter)
        if err == nil {
            time.Sleep(delay)
            continue
        }
    }
    // Fall back to exponential backoff if no Retry-After
}
```

**Impact**:
- Risk of being permanently banned by aggressive servers
- Wasted bandwidth on repeated failures
- Poor netizen behavior (ignoring server's explicit instructions)

**Effort**: 2-3 hours

**References**: RFC 7231 Section 7.1.3

---

### 2. Cache-Control max-age Support

**Mentioned in**:
- `specs/rogue-planet-spec.md:159` - "Respect Cache-Control: max-age header if present"
- `CLAUDE.md:222` - "Respect Cache-Control headers"
- `v1.0.0-plan.md:109-111` - Phase 5: Respect Cache-Control

**Current State**: ⚠️ PARTIALLY PLANNED
- Listed in networking-implementation-plan.md Phase 4.6
- Parser function sketched but not implemented
- Not integrated with scheduling logic

**Implementation Gap**:
```go
// Need to parse: "Cache-Control: max-age=3600, public"
// Extract max-age value (seconds)
// Use as minimum fetch interval (don't fetch more often)
```

**Impact**:
- May fetch feeds more often than server wants
- Risk of being rate-limited
- Wasted bandwidth

**Effort**: 2 hours (already in networking plan Phase 4.6)

**Note**: Already covered in networking-implementation-plan.md but highlighting here for completeness

---

### 3. Connection Timeout Configuration

**Mentioned in**:
- `specs/rogue-planet-spec.md:146` - "Timeout handling (default: 30s, configurable)"
- `specs/rogue-planet-spec.md:155` - "Connection pooling with appropriate timeouts"
- `networking-implementation-plan.md:113` - connection_timeout config

**Current State**: ⚠️ HARDCODED
- Default timeout: 30 seconds (crawler.go:25)
- Not configurable via config.ini
- No per-feed timeout override

**Implementation Gap**:
```ini
# Add to config.ini
[planet]
http_timeout = 30         # Default: 30 seconds
connect_timeout = 10      # TCP connection timeout
read_timeout = 30         # Read response timeout
```

```go
// In config.go
type PlanetConfig struct {
    HTTPTimeout    int  // Seconds
    ConnectTimeout int  // Seconds
    ReadTimeout    int  // Seconds
}

// In crawler.go
transport := &http.Transport{
    DialContext: (&net.Dialer{
        Timeout: time.Duration(config.ConnectTimeout) * time.Second,
    }).DialContext,
    ResponseHeaderTimeout: time.Duration(config.ReadTimeout) * time.Second,
}
```

**Impact**:
- Cannot tune for slow/fast feeds
- May timeout too quickly on slow servers
- May wait too long on dead servers

**Effort**: 2-3 hours

---

## 🟡 Medium Priority: Good Netizen Features

### 4. robots.txt Support (Optional)

**Mentioned in**:
- `specs/rogue-planet-spec.md:147` - "Respect robots.txt (optional but recommended)"
- `specs/rogue-planet-spec.md:174` - "golang.org/x/time/rate for rate limiting"

**Current State**: ❌ NOT IMPLEMENTED
- No robots.txt parsing
- No User-Agent matching against robots.txt rules
- Feeds are fetched regardless of robots.txt Disallow rules

**Implementation Gap**:
```go
// Need to fetch /robots.txt for each domain
// Parse User-agent: * and User-agent: RoguePlanet
// Check Disallow: rules before fetching feeds
// Cache robots.txt per domain (TTL: 24 hours)
```

**Libraries**:
- `github.com/temoto/robotstxt` - Popular robots.txt parser
- Or implement RFC 9309 manually

**Impact**:
- May violate site policies
- Risk of being blocked
- Not a good netizen

**Effort**: 4-6 hours (new dependency, caching layer)

**Priority Rationale**: Marked "optional" in spec, RSS/Atom feeds typically want to be crawled

---

### 5. Minimum Fetch Frequency Enforcement

**Mentioned in**:
- `specs/rogue-planet-spec.md:168` - "Implement maximum fetch frequency (e.g., never faster than every 15 minutes)"
- `CLAUDE.md:226` - "Default fetch interval: 1 hour (not more frequent than 15 minutes)"
- `v1.0.0-plan.md:94` - "Min interval: 15 minutes (respect server load)"

**Current State**: ⚠️ PARTIALLY ENFORCED
- Config validation: concurrent_fetches 1-50 (config.go:225)
- No minimum time between fetches enforced
- No per-domain cooldown

**Implementation Gap**:
```go
// Global minimum: No feed fetched < 15 minutes since last fetch
// Check last_fetched timestamp before allowing fetch
// Enforce in GetFeedsReadyForFetch()

if time.Since(feed.LastFetched) < 15*time.Minute {
    // Skip - too soon
    continue
}
```

**Impact**:
- Risk of hammering servers
- May be blocked by aggressive rate limiting
- Poor netizen behavior

**Effort**: 1-2 hours (already in intelligent scheduling plan)

**Note**: Will be addressed by Phase 4 of networking-implementation-plan.md

---

### 6. Jitter for Thundering Herd Prevention

**Mentioned in**:
- `specs/rogue-planet-spec.md:160` - "Implement jitter to avoid thundering herd"
- `v1.0.0-plan.md:104-106` - Phase 4: Jitter
- `networking-implementation-plan.md` - Phase 4.4

**Current State**: ❌ NOT IMPLEMENTED
- All feeds fetched at same time (when `rp update` runs)
- No randomization of fetch times
- Thundering herd problem if many planets use same cron schedule

**Implementation Gap**:
```go
// Add ±10% random jitter to fetch interval
jitterRange := float64(interval) * 0.10
jitter := time.Duration((rand.Float64()*2-1) * jitterRange)
nextFetch := now.Add(interval + jitter)
```

**Impact**:
- All feeds hit at :00 if cron runs hourly
- Bursts of load on feed servers
- Contributes to thundering herd if many users

**Effort**: 1 hour (simple random offset)

**Note**: Already in networking-implementation-plan.md Phase 4.4

---

## 🟢 Low Priority: Nice-to-Have

### 7. Per-Feed Configuration Overrides

**Mentioned in**:
- `specs/rogue-planet-spec.md:416` - Shows `concurrent_fetches = 5` in config
- `specs/TODO.md:66` - "Per-feed configuration overrides (future-compatible)"
- Config parsing handles unknown sections (config.go:136-138)

**Current State**: ⚠️ PARSER READY, NOT USED
- Config parser ignores unknown sections (forward compatibility)
- No per-feed overrides implemented
- All feeds use global config

**Implementation Gap**:
```ini
# Example per-feed config
[https://slow-blog.com/feed.xml]
fetch_interval = 14400  # 4 hours (slow-updating blog)
timeout = 60            # 60 second timeout (slow server)

[https://fast-blog.com/feed.xml]
fetch_interval = 900    # 15 minutes (breaking news)
```

**Tables Needed**:
```sql
-- Add feed_config table
CREATE TABLE feed_config (
    feed_id INTEGER PRIMARY KEY,
    timeout INTEGER,
    fetch_interval INTEGER,
    max_retries INTEGER,
    FOREIGN KEY(feed_id) REFERENCES feeds(id) ON DELETE CASCADE
);
```

**Impact**:
- Cannot tune problematic feeds individually
- One-size-fits-all approach

**Effort**: 1-2 days (schema change, config loading, application logic)

**Priority Rationale**: Can work around with global config, rare need for per-feed tuning

---

### 8. 503 Service Unavailable Handling

**Mentioned in**:
- Implicit in Retry-After (#1 above)
- Standard HTTP error handling expected

**Current State**: ⚠️ BASIC HANDLING
- FetchWithRetry treats 503 like any 5xx error
- No special backoff for temporary unavailability
- No Retry-After header parsing (see #1)

**Implementation Gap**:
```go
// Treat 503 specially:
// - Check Retry-After header (if present)
// - Use longer backoff than normal errors
// - Don't count against fetch_error_count (temporary issue)

if resp.StatusCode == 503 {
    // Service temporarily unavailable
    if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
        // Honor server's request
    } else {
        // Use conservative backoff (e.g., 5 minutes)
        time.Sleep(5 * time.Minute)
    }
}
```

**Impact**:
- May retry too aggressively during outages
- May mark feeds as failed when server is just temporarily down

**Effort**: 30 minutes (combine with Retry-After implementation)

**Note**: Overlaps with #1 (Retry-After support)

---

## Feature Comparison Matrix

| Feature | Mentioned In | Implemented | Priority | Effort | In Network Plan |
|---------|-------------|-------------|----------|--------|----------------|
| Retry-After header | Spec, CLAUDE.md | ❌ No | 🔴 High | 2-3h | ❌ No |
| Cache-Control max-age | Spec, CLAUDE.md, v1.0 | ⚠️ Planned | 🔴 High | 2h | ✅ Phase 4.6 |
| Timeout config | Spec | ⚠️ Hardcoded | 🔴 High | 2-3h | ⚠️ Partial |
| robots.txt | Spec (optional) | ❌ No | 🟡 Medium | 4-6h | ❌ No |
| Min fetch frequency | Spec, CLAUDE.md, v1.0 | ⚠️ Partial | 🟡 Medium | 1-2h | ✅ Phase 4 |
| Jitter | Spec, v1.0 | ❌ No | 🟡 Medium | 1h | ✅ Phase 4.4 |
| Per-feed config | TODO.md | ⚠️ Parser ready | 🟢 Low | 1-2d | ❌ No |
| 503 handling | Implicit | ⚠️ Basic | 🟢 Low | 30m | ❌ No |

---

## Summary Statistics

**Total Features Found**: 8

**By Implementation Status**:
- ❌ Not Implemented: 3 (38%)
- ⚠️ Partially Implemented: 5 (62%)
- ✅ Fully Implemented: 0 (0%)

**By Priority**:
- 🔴 High: 3 (38%)
- 🟡 Medium: 3 (38%)
- 🟢 Low: 2 (25%)

**Already in Networking Plan**: 4 features (50%)
**New Findings**: 4 features (50%)

---

## Recommendations

### Immediate Actions (v0.5.0)
1. **Implement Retry-After header support** (2-3h, high impact)
2. **Add timeout configuration** (2-3h, improves reliability)
3. **Implement 503 special handling** (30m, piggyback on Retry-After)

**Total Effort**: ~1 day
**Benefit**: Critical for production reliability, prevents bans

---

### Short Term (v0.6.0 - with rate limiting)
4. **Enforce minimum fetch frequency** (1-2h)
5. **Add jitter** (1h)

**Total Effort**: 2-3 hours (already in network plan Phase 4)
**Benefit**: Better netizen behavior, reduces server load spikes

---

### Medium Term (v1.1.0+)
6. **robots.txt support** (4-6h, new dependency)
   - Assess actual need based on user feedback
   - Many feed servers don't use robots.txt for RSS/Atom

7. **Per-feed configuration** (1-2 days, schema change)
   - Defer until users request it
   - Can work around with global config

---

### Cache-Control (Already Planned)
Feature #2 is already in networking-implementation-plan.md Phase 4.6, no additional action needed beyond executing that plan.

---

## Integration with Existing Plans

### networking-implementation-plan.md Coverage

**Already Covered** ✅:
- Connection pooling (Phase 1.1)
- 301 redirects (Phase 2)
- Rate limiting (Phase 3)
- Intelligent scheduling (Phase 4)
  - Min fetch frequency (Phase 4)
  - Jitter (Phase 4.4)
  - Cache-Control (Phase 4.6)

**Missing from Plan** ❌:
- Retry-After header support (HIGH PRIORITY!)
- Timeout configuration
- robots.txt support
- Per-feed configuration
- 503 special handling

---

## Suggested Plan Updates

### Add to Phase 1 (Foundation)
**Phase 1.4: Retry-After Header Support**
- Parse Retry-After from 429/503 responses
- Use server-specified delay before exponential backoff
- Add tests for RFC 7231 date and seconds formats

**Phase 1.5: Timeout Configuration**
- Add http_timeout, connect_timeout, read_timeout to config
- Apply to http.Transport
- Document in examples/config.ini

### Add to Phase 2 (Quick Wins)
**Phase 2.4: 503 Service Unavailable Handling**
- Special handling for temporary failures
- Don't increment fetch_error_count
- Longer backoff than normal errors

### Defer to Future
- robots.txt support (v1.1 or later)
- Per-feed configuration overrides (v1.1 or later)

---

## Documentation Gaps

Files mentioning networking features but lacking implementation notes:

1. **CLAUDE.md**:
   - Line 182: Still lists rate limiting as dependency (needs update after P2.1)
   - Line 222: "Respect Cache-Control" - no implementation note
   - Line 224: "Honor 429 responses" - no Retry-After mention

2. **README.md**:
   - Doesn't mention timeout configuration
   - Doesn't mention Retry-After support

3. **specs/rogue-planet-spec.md**:
   - Lines 145-170: Comprehensive requirements but no implementation status

4. **examples/config.ini**:
   - Missing timeout configuration options
   - Missing rate limiting options (to be added in Phase 3)

---

## Testing Coverage

**Existing Tests** ✅:
- `pkg/crawler/crawler_test.go`: Retry logic (13 tests)
- `pkg/crawler/crawler_comprehensive_test.go`: Timeout handling

**Missing Tests** ❌:
- Retry-After header parsing (seconds format)
- Retry-After header parsing (HTTP-date format)
- Cache-Control max-age parsing
- Timeout configuration application
- 503 vs 500 error handling
- robots.txt parsing and application

---

## Code References

**Mentioned But Not Implemented**:
```bash
# Grep results showing mentions without implementation:
specs/rogue-planet-spec.md:166: Honor 429 rate limit responses with Retry-After header
specs/rogue-planet-spec.md:159: Respect Cache-Control: max-age header
specs/rogue-planet-spec.md:147: Respect robots.txt (optional)
specs/rogue-planet-spec.md:168: Implement maximum fetch frequency
CLAUDE.md:222: Respect Cache-Control headers
CLAUDE.md:224: Honor 429 responses with exponential backoff
```

**Implemented**:
```bash
# Current implementations:
pkg/crawler/crawler.go:24: const DefaultTimeout = 30 * time.Second
pkg/crawler/crawler.go:266: func FetchWithRetry (exponential backoff)
pkg/config/config.go:166: case "concurrent_fetches"
cmd/rp/commands.go:803: Semaphore pattern for concurrency
```

---

## Conclusion

This audit found **8 networking-related features** mentioned in documentation/specs but not fully implemented.

**Key Findings**:
1. **4 features already in networking-implementation-plan.md** - will be addressed when that plan is executed
2. **4 new findings** - should be added to the implementation plan
3. **3 high-priority items** need immediate attention for production readiness
4. Most features are "good netizen" behaviors that prevent being blocked

**Action Items**:
1. ✅ Update networking-implementation-plan.md with 4 new features
2. ✅ Prioritize Retry-After and timeout config for v0.5.0
3. ✅ Document current status in specs to prevent confusion
4. ✅ Add missing test coverage for new features

**Next Steps**:
1. Review this audit with stakeholders
2. Decide on feature priorities
3. Update networking-implementation-plan.md
4. Begin implementation starting with high-priority items

---

**Audit Completed**: 2025-10-29
**Auditor**: Claude Code
**Files Reviewed**: 50+ (all docs, specs, source, tests)
**Search Terms**: timeout, retry, backoff, throttle, rate, connection, pool, redirect, cache-control, robots.txt, jitter, 429, 503, Retry-After
